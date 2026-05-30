package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/ymhhh/go-common/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestFromConfig_JSON_stdout(t *testing.T) {
	opts := config.Options{
		"logger": map[string]any{
			"level":  "debug",
			"format": "json",
			"output": "discard",
			"json": map[string]any{
				"prettyPrint": true,
			},
		},
	}
	c := opts.ToConfig()

	l, err := FromConfig(c, "logger")
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	defer func() { _ = l.Close() }()

	if l.Level != logrus.DebugLevel {
		t.Fatalf("level: %v", l.Level)
	}
	if _, ok := l.Formatter.(*logrus.JSONFormatter); !ok {
		t.Fatalf("formatter: %T", l.Formatter)
	}
}

func TestFromConfig_TextFormatterOptions(t *testing.T) {
	opts := config.Options{
		"logger": map[string]any{
			"level":  "info",
			"format": "text",
			"output": "discard",
			"text": map[string]any{
				"disableColors": true,
				"fullTimestamp": true,
			},
		},
	}
	c := opts.ToConfig()

	l, err := FromConfig(c, "logger")
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	defer func() { _ = l.Close() }()

	tf, ok := l.Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Fatalf("formatter: %T", l.Formatter)
	}
	if tf.DisableColors != true || tf.FullTimestamp != true {
		t.Fatalf("text formatter opts: %+v", tf)
	}
}

func TestFromConfig_FileRotate(t *testing.T) {
	opts := config.Options{
		"logger": map[string]any{
			"level":  "info",
			"format": "text",
			"output": "file:./tmp.log",
			"file": map[string]any{
				"rotate": map[string]any{
					"enabled":    true,
					"maxSizeMB":  1,
					"maxBackups": 2,
					"maxAgeDays": 3,
					"compress":   true,
				},
			},
		},
	}
	c := opts.ToConfig()

	l, err := FromConfig(c)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	defer func() { _ = l.Close() }()

	if _, ok := l.Out.(*lumberjack.Logger); !ok {
		t.Fatalf("out: %T", l.Out)
	}
}

func TestFromConfig_RelativeFilePathUsesConfigDir(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "conf")
	logDir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}

	cfgPath := filepath.Join(configDir, "app.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
logger:
  level: info
  format: text
  output: file:./logs/app.log
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	runDir := filepath.Join(root, "run")
	if err := os.Mkdir(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(runDir); err != nil {
		t.Fatalf("chdir run: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	}()

	c, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	l, err := FromConfig(c)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	l.Info("config-relative log")
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(logDir, "app.log"))
	if err != nil {
		t.Fatalf("read config-relative log: %v", err)
	}
	if !strings.Contains(string(b), "config-relative log") {
		t.Fatalf("log file missing message: %q", string(b))
	}
}

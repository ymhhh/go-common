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

func TestFromConfig_FilePathWithoutOutputUsesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	opts := config.Options{
		"logger": map[string]any{
			"level":  "info",
			"format": "text",
			"file": map[string]any{
				"path": path,
			},
		},
	}
	c := opts.ToConfig()

	l, err := FromConfig(c)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	l.Info("file path fallback")
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(b), "file path fallback") {
		t.Fatalf("log file missing entry: %q", string(b))
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

func TestInitGlobal_ReusesLoggerAcrossReload(t *testing.T) {
	resetGlobal := func() {
		globalMu.Lock()
		l := global
		global = nil
		globalMu.Unlock()
		if l != nil {
			_ = l.Close()
		}
	}
	resetGlobal()
	t.Cleanup(resetGlobal)

	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.log")
	secondPath := filepath.Join(dir, "second.log")
	configForPath := func(path string) config.Config {
		opts := config.Options{
			"logger": map[string]any{
				"level":  "info",
				"format": "text",
				"output": "file:" + path,
			},
		}
		return opts.ToConfig()
	}

	if err := InitGlobal(configForPath(firstPath)); err != nil {
		t.Fatalf("InitGlobal first: %v", err)
	}
	entry := L().WithField("component", "worker")
	entry.Info("before reload")

	if err := InitGlobal(configForPath(secondPath)); err != nil {
		t.Fatalf("InitGlobal second: %v", err)
	}
	entry.Info("after reload from cached entry")
	L().Info("after reload from fresh entry")

	globalMu.RLock()
	l := global
	globalMu.RUnlock()
	if err := l.Close(); err != nil {
		t.Fatalf("Close global: %v", err)
	}

	secondLog, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("read second log: %v", err)
	}
	if !strings.Contains(string(secondLog), "after reload from cached entry") {
		t.Fatalf("cached entry did not write to reloaded global logger: %q", string(secondLog))
	}
	if !strings.Contains(string(secondLog), "after reload from fresh entry") {
		t.Fatalf("fresh entry missing from second log: %q", string(secondLog))
	}
}

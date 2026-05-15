package logger

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestInitGlobal_ReusesLoggerAcrossReload(t *testing.T) {
	resetGlobalForTest(t)
	t.Cleanup(func() {
		resetGlobalForTest(t)
	})

	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.log")
	secondPath := filepath.Join(dir, "second.log")

	if err := InitGlobal(fileLoggerConfig(firstPath)); err != nil {
		t.Fatalf("InitGlobal first: %v", err)
	}

	stale := L().WithField("case", "stale")
	firstLogger := stale.Logger
	stale.Info("before reload")

	if err := InitGlobal(fileLoggerConfig(secondPath)); err != nil {
		t.Fatalf("InitGlobal second: %v", err)
	}
	if got := L().Logger; got != firstLogger {
		t.Fatalf("global logger was replaced; stale entries will keep the old output")
	}

	stale.Info("after reload")
	L().Info("current after reload")

	first, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatalf("read first log: %v", err)
	}
	second, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("read second log: %v", err)
	}

	if !bytes.Contains(first, []byte("before reload")) {
		t.Fatalf("first log missing pre-reload entry: %s", first)
	}
	if bytes.Contains(first, []byte("after reload")) {
		t.Fatalf("stale entry wrote to old log after reload: %s", first)
	}
	if !bytes.Contains(second, []byte("after reload")) {
		t.Fatalf("second log missing stale post-reload entry: %s", second)
	}
	if !bytes.Contains(second, []byte("current after reload")) {
		t.Fatalf("second log missing current post-reload entry: %s", second)
	}
}

func fileLoggerConfig(path string) config.Config {
	opts := config.Options{
		"logger": map[string]any{
			"level":  "info",
			"format": "text",
			"output": "file:" + path,
			"text": map[string]any{
				"disableColors": true,
			},
		},
	}
	return opts.ToConfig()
}

func resetGlobalForTest(t *testing.T) {
	t.Helper()

	globalMu.Lock()
	closer := globalCloser
	global = nil
	globalCloser = nil
	globalMu.Unlock()

	if closer != nil {
		_ = closer.Close()
	}
}

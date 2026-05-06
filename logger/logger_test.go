package logger

import (
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

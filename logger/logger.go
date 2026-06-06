package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/sirupsen/logrus"
	"github.com/ymhhh/go-common/config"
)

type Fields = logrus.Fields
type FieldMap = logrus.FieldMap

// Config is the config schema for this package.
//
// Example YAML:
//
//	logger:
//	  level: info
//	  format: json         # text|json
//	  output: stdout       # stdout|stderr|discard|/path/to/file.log|file:/path/to/file.log
//	  reportCaller: false
//	  text:
//	    disableColors: true
//	    fullTimestamp: true
type Config struct {
	Level        string `json:"level" yaml:"level"`
	Format       string `json:"format" yaml:"format"` // text|json
	Output       string `json:"output" yaml:"output"` // stdout|stderr|discard|path|file:path
	ReportCaller bool   `json:"reportCaller" yaml:"reportCaller"`

	File struct {
		// Path is used when Output is "file" or empty but File.Path is set.
		// It is also used as a default for Output like "file:" (missing path).
		Path string `json:"path" yaml:"path"`

		Rotate struct {
			Enabled    bool `json:"enabled" yaml:"enabled"`
			MaxSizeMB  int  `json:"maxSizeMB" yaml:"maxSizeMB"`   // default 100
			MaxBackups int  `json:"maxBackups" yaml:"maxBackups"` // default 7
			MaxAgeDays int  `json:"maxAgeDays" yaml:"maxAgeDays"` // default 7
			Compress   bool `json:"compress" yaml:"compress"`
			LocalTime  bool `json:"localTime" yaml:"localTime"`
		} `json:"rotate" yaml:"rotate"`
	} `json:"file" yaml:"file"`

	Text struct {
		DisableColors bool `json:"disableColors" yaml:"disableColors"`
		FullTimestamp bool `json:"fullTimestamp" yaml:"fullTimestamp"`
	} `json:"text" yaml:"text"`

	JSON struct {
		PrettyPrint bool `json:"prettyPrint" yaml:"prettyPrint"`
	} `json:"json" yaml:"json"`
}

// Logger wraps logrus.Logger and holds resources if needed.
type Logger struct {
	*logrus.Logger
	closeOnce sync.Once
	closer    io.Closer
}

func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	var err error
	l.closeOnce.Do(func() {
		if l.closer != nil {
			err = l.closer.Close()
		}
	})
	return err
}

// FromConfig creates a logger from a config tree.
//
// By default it reads the "logger" subtree. If you pass a single path, it reads
// that subtree instead. If the path is empty string, the whole config is used.
func FromConfig(c config.Config, path ...string) (*Logger, error) {
	p := "logger"
	if len(path) > 1 {
		return nil, fmt.Errorf("logger: FromConfig expects 0 or 1 path, got %d", len(path))
	}
	if len(path) == 1 {
		p = path[0]
	}

	var cfg Config
	if p == "" {
		if err := c.Object(&cfg); err != nil {
			return nil, err
		}
	} else {
		if err := c.Object(&cfg, config.WithObjectPath(p)); err != nil {
			return nil, err
		}
	}
	baseDir := ""
	if bc, ok := c.(interface{ BaseDir() string }); ok {
		baseDir = bc.BaseDir()
	}
	return newWithBaseDir(cfg, baseDir)
}

// New constructs a new configured logger.
func New(cfg Config) (*Logger, error) {
	return newWithBaseDir(cfg, "")
}

func newWithBaseDir(cfg Config, baseDir string) (*Logger, error) {
	l := logrus.New()

	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}

	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		return nil, fmt.Errorf("logger: parse level %q: %w", cfg.Level, err)
	}
	l.SetLevel(level)
	l.SetReportCaller(cfg.ReportCaller)

	switch strings.ToLower(cfg.Format) {
	case "text":
		l.SetFormatter(&logrus.TextFormatter{
			DisableColors: cfg.Text.DisableColors,
			FullTimestamp: cfg.Text.FullTimestamp,
		})
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: cfg.JSON.PrettyPrint,
		})
	default:
		return nil, fmt.Errorf("logger: unsupported format: %q", cfg.Format)
	}

	path, err := resolveOutputPath(cfg, baseDir)
	if err != nil {
		return nil, err
	}

	out, closer, err := createWriter(path, cfg)
	if err != nil {
		return nil, err
	}
	l.SetOutput(out)

	return &Logger{Logger: l, closer: closer}, nil
}

func resolveOutputPath(cfg Config, baseDir string) (string, error) {
	s := strings.TrimSpace(cfg.Output)

	// Handle special output targets.
	switch {
	case s == "":
		if p := strings.TrimSpace(cfg.File.Path); p != "" {
			s = p
		} else {
			return "stderr", nil
		}
	case strings.EqualFold(s, "stderr"):
		return "stderr", nil
	case strings.EqualFold(s, "stdout"):
		return "stdout", nil
	case strings.EqualFold(s, "discard"):
		return "discard", nil
	case strings.EqualFold(s, "file"):
		s = strings.TrimSpace(cfg.File.Path)
		if s == "" {
			return "", fmt.Errorf("logger: empty output path")
		}
	default:
		// file:/abs/or/rel/path
		if strings.HasPrefix(strings.ToLower(s), "file:") {
			s = strings.TrimSpace(s[len("file:"):])
		}
	}

	if s == "" {
		return "", fmt.Errorf("logger: empty output path")
	}

	if !filepath.IsAbs(s) {
		if baseDir != "" {
			s = filepath.Join(baseDir, s)
		} else if wd, err := os.Getwd(); err == nil {
			s = filepath.Join(wd, s)
		} else {
			return "", fmt.Errorf("logger: cannot resolve relative output path %q: %w", s, err)
		}
	}

	return s, nil
}

func createWriter(path string, cfg Config) (io.Writer, io.Closer, error) {
	switch path {
	case "stderr":
		return os.Stderr, nil, nil
	case "stdout":
		return os.Stdout, nil, nil
	case "discard":
		return io.Discard, nil, nil
	}

	// rotate output (lumberjack)
	if cfg.File.Rotate.Enabled ||
		cfg.File.Rotate.MaxSizeMB > 0 ||
		cfg.File.Rotate.MaxBackups > 0 ||
		cfg.File.Rotate.MaxAgeDays > 0 ||
		cfg.File.Rotate.Compress ||
		cfg.File.Rotate.LocalTime {
		maxSize := cfg.File.Rotate.MaxSizeMB
		if maxSize <= 0 {
			maxSize = 100
		}
		maxBackups := cfg.File.Rotate.MaxBackups
		if maxBackups <= 0 {
			maxBackups = 7
		}
		maxAge := cfg.File.Rotate.MaxAgeDays
		if maxAge <= 0 {
			maxAge = 7
		}
		lj := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   cfg.File.Rotate.Compress,
			LocalTime:  cfg.File.Rotate.LocalTime,
		}
		// Ensure the directory exists.
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("logger: open output file %q: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return nil, nil, fmt.Errorf("logger: close output file %q: %w", path, err)
		}
		return lj, lj, nil
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("logger: open output file %q: %w", path, err)
	}
	return f, f, nil
}

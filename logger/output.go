package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func rotateEnabled(cfg Config) bool {
	r := cfg.File.Rotate
	return r.Enabled || r.MaxSizeMB > 0 || r.MaxBackups > 0 ||
		r.MaxAgeDays > 0 || r.Compress
}

func resolveOutputPath(cfg Config, baseDir string) (string, error) {
	s := strings.TrimSpace(cfg.Output)

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

	if rotateEnabled(cfg) {
		return newRotatingWriter(path, normalizeRotateConfig(&cfg))
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("logger: open output file %q: %w", path, err)
	}
	return f, f, nil
}

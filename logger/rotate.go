package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golift.io/rotatorr"
	"golift.io/rotatorr/compressor"
	"golift.io/rotatorr/introtator"
)

type rotateConfig struct {
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func normalizeRotateConfig(cfg *Config) rotateConfig {
	rc := rotateConfig{
		MaxSizeMB:  cfg.File.Rotate.MaxSizeMB,
		MaxBackups: cfg.File.Rotate.MaxBackups,
		MaxAgeDays: cfg.File.Rotate.MaxAgeDays,
		Compress:   cfg.File.Rotate.Compress,
	}
	if rc.MaxSizeMB <= 0 {
		rc.MaxSizeMB = 100
	}
	if rc.MaxBackups <= 0 {
		rc.MaxBackups = 7
	}
	if rc.MaxAgeDays <= 0 {
		rc.MaxAgeDays = 7
	}
	return rc
}

func newRotatingWriter(path string, rc rotateConfig) (io.Writer, io.Closer, error) {
	layout := &introtator.Layout{
		FileCount:  rc.MaxBackups,
		PostRotate: buildPostRotate(path, rc),
	}

	w, err := rotatorr.New(&rotatorr.Config{
		Filepath: path,
		FileSize: int64(rc.MaxSizeMB) * 1024 * 1024,
		FileMode: 0o644,
		DirMode:  0o755,
		Rotatorr: layout,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("logger: create rotating writer %q: %w", path, err)
	}
	return w, w, nil
}

func buildPostRotate(mainPath string, rc rotateConfig) func(fileName, newFile string) {
	return func(_, newFile string) {
		if rc.Compress {
			compressor.CompressBackground(newFile, nil)
		}
		purgeExpiredBackups(mainPath, rc.MaxAgeDays)
	}
}

func purgeExpiredBackups(mainPath string, maxAgeDays int) {
	if maxAgeDays <= 0 {
		return
	}

	dir := filepath.Dir(mainPath)
	prefix := backupPrefix(filepath.Base(mainPath))
	cutoff := time.Now().Add(-time.Duration(maxAgeDays) * 24 * time.Hour)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		suffix := strings.TrimPrefix(name, prefix)
		suffix = strings.TrimSuffix(suffix, ".gz")
		suffix = strings.TrimSuffix(suffix, ".log")
		if _, err := strconv.Atoi(suffix); err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

func backupPrefix(baseName string) string {
	return strings.TrimSuffix(baseName, ".log") + "."
}

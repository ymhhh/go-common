package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

type backupFile struct {
	path    string
	index   int
	modTime time.Time
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
			compressor.CompressBackground(newFile, func(_ *compressor.Report) {
				purgeBackups(mainPath, rc)
			})
			return
		}
		purgeBackups(mainPath, rc)
	}
}

func listBackupFiles(mainPath string) ([]backupFile, error) {
	dir := filepath.Dir(mainPath)
	prefix := backupPrefix(filepath.Base(mainPath))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]backupFile, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		suffix := strings.TrimPrefix(name, prefix)
		suffix = strings.TrimSuffix(suffix, ".gz")
		suffix = strings.TrimSuffix(suffix, ".log")
		index, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, backupFile{
			path:    filepath.Join(dir, name),
			index:   index,
			modTime: info.ModTime(),
		})
	}
	return files, nil
}

func purgeBackups(mainPath string, rc rotateConfig) {
	files, err := listBackupFiles(mainPath)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-time.Duration(rc.MaxAgeDays) * 24 * time.Hour)
	remaining := make([]backupFile, 0, len(files))

	for _, file := range files {
		if rc.MaxAgeDays > 0 && file.modTime.Before(cutoff) {
			_ = os.Remove(file.path)
			continue
		}
		remaining = append(remaining, file)
	}

	if !rc.Compress || rc.MaxBackups <= 0 {
		return
	}

	byIndex := make(map[int][]backupFile, len(remaining))
	for _, file := range remaining {
		byIndex[file.index] = append(byIndex[file.index], file)
	}

	indices := make([]int, 0, len(byIndex))
	for index := range byIndex {
		indices = append(indices, index)
	}
	sort.Ints(indices)

	for _, index := range indices {
		if index <= rc.MaxBackups {
			continue
		}
		for _, file := range byIndex[index] {
			_ = os.Remove(file.path)
		}
	}
}

func backupPrefix(baseName string) string {
	return strings.TrimSuffix(baseName, ".log") + "."
}

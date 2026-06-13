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

// suffixRotator rotates backups as app.log.1, app.log.2 (number suffix).
type suffixRotator struct {
	FileCount  int
	PostRotate func(fileName, newFile string)
}

func (s *suffixRotator) Dirs(fileName string) ([]string, error) {
	return []string{filepath.Dir(fileName)}, nil
}

func (s *suffixRotator) Post(fileName, newFile string) {
	if s.PostRotate != nil {
		s.PostRotate(fileName, newFile)
	}
}

func (s *suffixRotator) Rotate(fileName string) (string, error) {
	indices, err := backupIndices(fileName)
	if err != nil {
		return "", err
	}

	maxIndex := 0
	for _, index := range indices {
		if index > maxIndex {
			maxIndex = index
		}
	}

	for index := maxIndex; index >= 1; index-- {
		if err := bumpBackup(fileName, index); err != nil {
			return "", err
		}
	}

	newFile := backupPath(fileName, 1)
	if err := os.Rename(fileName, newFile); err != nil {
		return "", fmt.Errorf("rotate log file: %w", err)
	}

	if err := removeBackupsAbove(fileName, s.FileCount); err != nil {
		return "", err
	}

	return newFile, nil
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
	layout := &suffixRotator{
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

func backupNamePrefix(mainPath string) string {
	return filepath.Base(mainPath) + "."
}

func backupPath(mainPath string, index int) string {
	return mainPath + "." + strconv.Itoa(index)
}

func backupIndices(mainPath string) ([]int, error) {
	dir := filepath.Dir(mainPath)
	prefix := backupNamePrefix(mainPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]struct{})
	for _, entry := range entries {
		index, ok := parseBackupIndex(entry.Name(), prefix)
		if !ok {
			continue
		}
		seen[index] = struct{}{}
	}

	indices := make([]int, 0, len(seen))
	for index := range seen {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	return indices, nil
}

func parseBackupIndex(name, prefix string) (int, bool) {
	if !strings.HasPrefix(name, prefix) {
		return 0, false
	}

	suffix := strings.TrimPrefix(name, prefix)
	suffix = strings.TrimSuffix(suffix, ".gz")
	index, err := strconv.Atoi(suffix)
	if err != nil || index <= 0 {
		return 0, false
	}
	return index, true
}

func bumpBackup(mainPath string, index int) error {
	for _, path := range []string{backupPath(mainPath, index), backupPath(mainPath, index) + ".gz"} {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		dst := backupPath(mainPath, index+1)
		if strings.HasSuffix(path, ".gz") {
			dst += ".gz"
		}

		_ = os.Remove(dst)
		if err := os.Rename(path, dst); err != nil {
			return fmt.Errorf("rotate backup %q -> %q: %w", path, dst, err)
		}
	}
	return nil
}

func removeBackupsAbove(mainPath string, maxBackups int) error {
	if maxBackups < 1 {
		return nil
	}

	indices, err := backupIndices(mainPath)
	if err != nil {
		return err
	}

	for _, index := range indices {
		if index <= maxBackups {
			continue
		}
		for _, path := range []string{backupPath(mainPath, index), backupPath(mainPath, index) + ".gz"} {
			_ = os.Remove(path)
		}
	}
	return nil
}

func listBackupFiles(mainPath string) ([]backupFile, error) {
	dir := filepath.Dir(mainPath)
	prefix := backupNamePrefix(mainPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]backupFile, 0, len(entries))
	for _, entry := range entries {
		index, ok := parseBackupIndex(entry.Name(), prefix)
		if !ok {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, backupFile{
			path:    filepath.Join(dir, entry.Name()),
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

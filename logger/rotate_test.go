package logger

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golift.io/rotatorr/compressor"
	"golift.io/rotatorr/filer"
)

func TestSuffixRotator_RotateUsesTrailingIndex(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")

	if err := os.WriteFile(mainPath, []byte("current"), 0o644); err != nil {
		t.Fatalf("write main log: %v", err)
	}
	if err := os.WriteFile(mainPath+".1", []byte("first"), 0o644); err != nil {
		t.Fatalf("write first backup: %v", err)
	}

	rotator := &suffixRotator{FileCount: 2}
	newFile, err := rotator.Rotate(mainPath)
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if newFile != mainPath+".1" {
		t.Fatalf("newFile: got %q want %q", newFile, mainPath+".1")
	}

	for _, name := range []string{"app.log.1", "app.log.2"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected backup %s: %v", name, err)
		}
	}
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		t.Fatalf("expected active log to be rotated, err=%v", err)
	}
}

func TestPurgeBackups_EnforcesMaxBackupsForCompressedArchives(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")

	backups := []string{"app.log.1.gz", "app.log.2.gz", "app.log.3.gz"}
	for _, name := range backups {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
			t.Fatalf("write backup %s: %v", name, err)
		}
	}

	purgeBackups(mainPath, rotateConfig{
		MaxBackups: 2,
		MaxAgeDays: 7,
		Compress:   true,
	}, "")

	for _, name := range []string{"app.log.1.gz", "app.log.2.gz"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected backup %s to remain: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "app.log.3.gz")); !os.IsNotExist(err) {
		t.Fatalf("expected app.log.3.gz to be removed, err=%v", err)
	}
}

func TestBuildPostRotate_PurgesAfterCompressionCompletes(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")
	rotating := mainPath + ".1"

	if err := os.WriteFile(rotating, []byte("rotating backup"), 0o644); err != nil {
		t.Fatalf("write rotating backup: %v", err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(rotating, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes rotating backup: %v", err)
	}

	post := buildPostRotate(mainPath, rotateConfig{
		MaxBackups: 2,
		MaxAgeDays: 1,
		Compress:   true,
	})
	post("", rotating)

	if _, err := os.Stat(rotating); !os.IsNotExist(err) {
		t.Fatalf("expected rotating backup to be replaced by compressed archive, err=%v", err)
	}
	if _, err := os.Stat(rotating + ".gz"); err != nil {
		t.Fatalf("expected compressed backup: %v", err)
	}
}

func TestBuildPostRotate_WaitsForCompressionBeforeReturning(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")
	rotating := mainPath + ".1"

	if err := os.WriteFile(rotating, []byte("rotating backup"), 0o644); err != nil {
		t.Fatalf("write rotating backup: %v", err)
	}

	originalFiler := compressor.Filer
	blocker := &blockingCompressFiler{
		Filer:     originalFiler,
		blockPath: rotating,
		started:   make(chan struct{}),
		release:   make(chan struct{}),
	}
	compressor.Filer = blocker
	t.Cleanup(func() {
		blocker.releaseCompression()
		compressor.Filer = originalFiler
	})

	post := buildPostRotate(mainPath, rotateConfig{
		MaxBackups: 2,
		MaxAgeDays: 7,
		Compress:   true,
	})

	done := make(chan struct{})
	go func() {
		post("", rotating)
		close(done)
	}()

	select {
	case <-blocker.started:
	case <-done:
		t.Fatal("post-rotate returned before compression started")
	case <-time.After(time.Second):
		t.Fatal("compression did not start")
	}

	select {
	case <-done:
		t.Fatal("post-rotate returned before compression finished")
	default:
	}

	blocker.releaseCompression()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("post-rotate did not return after compression finished")
	}
}

type blockingCompressFiler struct {
	filer.Filer
	blockPath   string
	started     chan struct{}
	release     chan struct{}
	once        sync.Once
	releaseOnce sync.Once
}

func (f *blockingCompressFiler) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if name == f.blockPath && flag == os.O_RDONLY {
		f.once.Do(func() {
			close(f.started)
		})
		<-f.release
	}
	return f.Filer.OpenFile(name, flag, perm)
}

func (f *blockingCompressFiler) releaseCompression() {
	f.releaseOnce.Do(func() {
		close(f.release)
	})
}

func TestPurgeBackups_RemovesExpiredBackupsWithoutCompress(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")
	expired := mainPath + ".9"

	if err := os.WriteFile(expired, []byte("old"), 0o644); err != nil {
		t.Fatalf("write expired backup: %v", err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(expired, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes expired backup: %v", err)
	}

	purgeBackups(mainPath, rotateConfig{
		MaxBackups: 7,
		MaxAgeDays: 1,
		Compress:   false,
	}, "")

	if _, err := os.Stat(expired); !os.IsNotExist(err) {
		t.Fatalf("expected expired backup to be removed, err=%v", err)
	}
}

func TestPurgeBackups_EnforcesMaxBackupsWithoutCompress(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")

	backups := []string{"app.log.1", "app.log.2", "app.log.3", "app.log.4"}
	for _, name := range backups {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
			t.Fatalf("write backup %s: %v", name, err)
		}
	}

	purgeBackups(mainPath, rotateConfig{
		MaxBackups: 2,
		MaxAgeDays: 7,
		Compress:   false,
	}, "")

	for _, name := range []string{"app.log.1", "app.log.2"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected backup %s to remain: %v", name, err)
		}
	}
	for _, name := range []string{"app.log.3", "app.log.4"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected backup %s to be removed, err=%v", name, err)
		}
	}
}

func TestPurgeBackups_SkipsFileBeingCompressed(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")
	compressing := mainPath + ".1"

	backups := []string{"app.log.1", "app.log.2", "app.log.3"}
	for _, name := range backups {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
			t.Fatalf("write backup %s: %v", name, err)
		}
	}

	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(compressing, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes compressing backup: %v", err)
	}

	purgeBackups(mainPath, rotateConfig{
		MaxBackups: 2,
		MaxAgeDays: 1,
		Compress:   true,
	}, compressing)

	if _, err := os.Stat(compressing); err != nil {
		t.Fatalf("file being compressed should not be deleted: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "app.log.3")); !os.IsNotExist(err) {
		t.Fatalf("expected app.log.3 to be removed, err=%v", err)
	}
}

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

func TestBuildPostRotate_DefersPurgeUntilCompressionCompletes(t *testing.T) {
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

	if _, err := os.Stat(rotating); err != nil {
		t.Fatalf("rotating backup removed before background compression finished: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(rotating + ".gz"); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("compression did not finish in time")
}

func TestBuildPostRotate_WaitsForCompressionBeforeReturning(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "app.log")
	rotating := mainPath + ".1"

	if err := os.WriteFile(rotating, []byte("rotating backup"), 0o644); err != nil {
		t.Fatalf("write rotating backup: %v", err)
	}

	originalFiler := compressor.Filer
	blocking := &blockingCompressFiler{
		Filer:     filer.Default(),
		blockName: rotating,
		entered:   make(chan struct{}),
		release:   make(chan struct{}),
	}
	compressor.Filer = blocking
	t.Cleanup(func() {
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
	case <-blocking.entered:
	case <-time.After(time.Second):
		t.Fatal("compression did not start")
	}

	returnedEarly := false
	select {
	case <-done:
		returnedEarly = true
	default:
	}

	close(blocking.release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("post did not return after compression was released")
	}

	if returnedEarly {
		t.Fatal("post returned before compression completed")
	}
	if _, err := os.Stat(rotating); !os.IsNotExist(err) {
		t.Fatalf("expected uncompressed backup to be removed after compression, err=%v", err)
	}
	if _, err := os.Stat(rotating + ".gz"); err != nil {
		t.Fatalf("expected compressed backup: %v", err)
	}
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

type blockingCompressFiler struct {
	filer.Filer
	blockName string
	entered   chan struct{}
	release   chan struct{}
	once      sync.Once
}

func (f *blockingCompressFiler) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if name == f.blockName {
		f.once.Do(func() {
			close(f.entered)
			<-f.release
		})
	}
	return f.Filer.OpenFile(name, flag, perm)
}

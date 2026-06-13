package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	})

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
	})

	if _, err := os.Stat(expired); !os.IsNotExist(err) {
		t.Fatalf("expected expired backup to be removed, err=%v", err)
	}
}

package screenshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewServiceCleansStorageDirOnStartup(t *testing.T) {
	storageDir := t.TempDir()

	oldFile := filepath.Join(storageDir, "old.png")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	oldSubdir := filepath.Join(storageDir, "nested")
	if err := os.Mkdir(oldSubdir, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}

	oldNestedFile := filepath.Join(oldSubdir, "old.txt")
	if err := os.WriteFile(oldNestedFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	svc, err := NewService(storageDir)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected storage dir to be empty, found %d entries", len(entries))
	}

	if svc.storageDir != storageDir {
		t.Fatalf("storageDir = %q, want %q", svc.storageDir, storageDir)
	}
}

func TestSaveWritesScreenshotFile(t *testing.T) {
	storageDir := t.TempDir()

	svc, err := NewService(storageDir)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	data := []byte("png-data")
	id, err := svc.Save(data)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if len(id) != idLength {
		t.Fatalf("id length = %d, want %d", len(id), idLength)
	}

	got, err := os.ReadFile(svc.Path(id))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != string(data) {
		t.Fatalf("file contents = %q, want %q", got, data)
	}
}

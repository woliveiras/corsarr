package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectorAcceptsWritableDirectoryWithHardlinks(t *testing.T) {
	baseDirectory := t.TempDir()

	status := NewInspector().Inspect(baseDirectory)

	if status.State != StateReady {
		t.Fatalf("expected ready storage, got %q: %s", status.State, status.TechnicalDetail)
	}
	if !status.Writable {
		t.Fatal("expected storage to be writable")
	}
	if !status.Hardlinks {
		t.Fatal("expected temporary filesystem to support hardlinks")
	}
	if !filepath.IsAbs(status.Path) {
		t.Fatalf("expected normalized absolute path, got %q", status.Path)
	}

	entries, err := os.ReadDir(baseDirectory)
	if err != nil {
		t.Fatalf("read inspected directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected probe artifacts to be removed, found %d", len(entries))
	}
}

func TestInspectorRejectsMissingPathWithoutCreatingIt(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "does-not-exist")

	status := NewInspector().Inspect(missingPath)

	if status.State != StateInvalid {
		t.Fatalf("expected invalid storage, got %q", status.State)
	}
	if _, err := os.Stat(missingPath); !os.IsNotExist(err) {
		t.Fatalf("expected missing path to remain absent, got %v", err)
	}
}

func TestInspectorRejectsFilePath(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "media.txt")
	if err := os.WriteFile(filePath, []byte("media"), 0o600); err != nil {
		t.Fatalf("create fixture file: %v", err)
	}

	status := NewInspector().Inspect(filePath)

	if status.State != StateInvalid {
		t.Fatalf("expected file path to be invalid, got %q", status.State)
	}
}

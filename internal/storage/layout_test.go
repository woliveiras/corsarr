package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLayoutPreparerCreatesIdempotentCorsarrTree(t *testing.T) {
	baseDirectory := t.TempDir()
	markerPath := filepath.Join(baseDirectory, "keep-me.txt")
	if err := os.WriteFile(markerPath, []byte("user data"), 0o600); err != nil {
		t.Fatalf("create marker: %v", err)
	}

	preparer := NewLayoutPreparer()
	first, err := preparer.Prepare(baseDirectory, []string{"prowlarr", "radarr"})
	if err != nil {
		t.Fatalf("prepare storage layout: %v", err)
	}
	second, err := preparer.Prepare(baseDirectory, []string{"radarr", "prowlarr"})
	if err != nil {
		t.Fatalf("prepare storage layout again: %v", err)
	}
	if first.RootPath != second.RootPath {
		t.Fatalf("expected stable root path, got %q and %q", first.RootPath, second.RootPath)
	}

	requiredDirectories := []string{
		"config/prowlarr",
		"config/radarr",
		"media/downloads/incomplete",
		"media/downloads/complete",
		"media/library/movies",
		"media/library/tv",
		"media/library/music",
		"media/library/books",
	}
	for _, relativePath := range requiredDirectories {
		info, err := os.Stat(filepath.Join(first.RootPath, relativePath))
		if err != nil {
			t.Fatalf("expected directory %q: %v", relativePath, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", relativePath)
		}
	}

	marker, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("read preserved marker: %v", err)
	}
	if string(marker) != "user data" {
		t.Fatalf("expected preexisting data to remain unchanged, got %q", marker)
	}

	configInfo, err := os.Stat(filepath.Join(first.RootPath, "config"))
	if err != nil {
		t.Fatalf("stat config directory: %v", err)
	}
	if permissions := configInfo.Mode().Perm(); permissions != 0o700 {
		t.Fatalf("expected private config permissions 0700, got %04o", permissions)
	}
}

func TestLayoutPreparerRejectsUnsafeApplicationID(t *testing.T) {
	baseDirectory := t.TempDir()
	outsidePath := filepath.Join(baseDirectory, "escaped")

	_, err := NewLayoutPreparer().Prepare(baseDirectory, []string{"../escaped"})
	if err == nil {
		t.Fatal("expected unsafe application ID to be rejected")
	}
	if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no escaped directory, got %v", statErr)
	}
}

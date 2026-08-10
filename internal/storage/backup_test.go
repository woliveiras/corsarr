package storage

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupManagerArchivesOnlyApplicationConfig(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Corsarr")
	config := filepath.Join(root, "config", "radarr")
	if err := os.MkdirAll(filepath.Join(config, "db"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config, "db", "radarr.db"), []byte("database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "media", "library", "movies"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "library", "movies", "movie.mkv"), []byte("media"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager := NewBackupManager()
	manager.now = func() time.Time { return time.Date(2026, 8, 10, 22, 0, 0, 0, time.UTC) }
	result, err := manager.Backup(root, "radarr")
	if err != nil {
		t.Fatalf("backup config: %v", err)
	}
	if result.FileCount != 1 || len(result.SHA256) != 64 {
		t.Fatalf("unexpected result %#v", result)
	}

	archive, err := os.Open(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = archive.Close() }()
	gz, err := gzip.NewReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(gz)
	var files []string
	for {
		header, nextErr := reader.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		if header.Typeflag == tar.TypeReg {
			files = append(files, header.Name)
		}
	}
	if len(files) != 1 || files[0] != "db/radarr.db" {
		t.Fatalf("unexpected backup entries %v", files)
	}
}

func TestBackupManagerRejectsSymlinkWithoutPublishingArchive(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Corsarr")
	config := filepath.Join(root, "config", "sonarr")
	if err := os.MkdirAll(config, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "media"), filepath.Join(config, "outside")); err != nil {
		t.Fatal(err)
	}

	if _, err := NewBackupManager().Backup(root, "sonarr"); err == nil {
		t.Fatal("expected symlink rejection")
	}
	matches, err := filepath.Glob(filepath.Join(root, "backups", "config", "sonarr", "*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no published backup, got %v", matches)
	}
}

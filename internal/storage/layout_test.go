package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestApplicationDataManagerArchivesOnlyTargetConfiguration(t *testing.T) {
	baseDirectory := t.TempDir()
	preparer := NewLayoutPreparer()
	layout, err := preparer.Prepare(baseDirectory, []string{"radarr", "sonarr"})
	if err != nil {
		t.Fatalf("prepare storage layout: %v", err)
	}
	markerPath := filepath.Join(layout.RootPath, "config", "radarr", "config.xml")
	if err := os.WriteFile(markerPath, []byte("radarr settings"), 0o600); err != nil {
		t.Fatalf("write application config: %v", err)
	}
	mediaMarker := filepath.Join(layout.RootPath, "media", "library", "movies", "keep.mkv")
	if err := os.WriteFile(mediaMarker, []byte("media"), 0o600); err != nil {
		t.Fatalf("write media marker: %v", err)
	}

	manager := NewApplicationDataManager()
	manager.now = func() time.Time {
		return time.Date(2026, time.August, 10, 12, 30, 0, 123, time.UTC)
	}
	archived, err := manager.Archive(baseDirectory, "radarr")
	if err != nil {
		t.Fatalf("archive application config: %v", err)
	}
	if !archived.Archived {
		t.Fatal("expected application config to be archived")
	}
	if _, err := os.Stat(filepath.Join(layout.RootPath, "config", "radarr")); !os.IsNotExist(err) {
		t.Fatalf("expected live application config to be absent, got %v", err)
	}
	archivedMarker, err := os.ReadFile(filepath.Join(archived.ArchivePath, "config.xml"))
	if err != nil || string(archivedMarker) != "radarr settings" {
		t.Fatalf("expected archived config marker, data=%q err=%v", archivedMarker, err)
	}
	if _, err := os.Stat(filepath.Join(layout.RootPath, "config", "sonarr")); err != nil {
		t.Fatalf("expected unrelated app config preserved: %v", err)
	}
	if _, err := os.Stat(mediaMarker); err != nil {
		t.Fatalf("expected shared media preserved: %v", err)
	}
}

func TestApplicationDataManagerRejectsUnsafeTarget(t *testing.T) {
	baseDirectory := t.TempDir()
	if _, err := NewApplicationDataManager().Archive(baseDirectory, "../movies"); err == nil {
		t.Fatal("expected unsafe application ID to be rejected")
	}
}

func TestApplicationDataManagerReportsMissingConfigurationWithoutCreatingTrash(t *testing.T) {
	baseDirectory := t.TempDir()
	manager := NewApplicationDataManager()

	status, err := manager.Archive(baseDirectory, "radarr")
	if err != nil {
		t.Fatalf("archive missing application config: %v", err)
	}
	if status.Archived {
		t.Fatal("expected missing config to remain a no-op")
	}
	if _, err := os.Stat(filepath.Join(baseDirectory, "Corsarr", "trash")); !os.IsNotExist(err) {
		t.Fatalf("expected no trash tree for missing data, got %v", err)
	}
}

func TestApplicationDataManagerInspectsOnlyKnownApplicationDirectory(t *testing.T) {
	baseDirectory := t.TempDir()
	layout, err := NewLayoutPreparer().Prepare(baseDirectory, []string{"radarr"})
	if err != nil {
		t.Fatalf("prepare storage layout: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(layout.RootPath, "config", "radarr", "config.xml"),
		[]byte("settings"),
		0o600,
	); err != nil {
		t.Fatalf("write application config: %v", err)
	}

	status, err := NewApplicationDataManager().Inspect(baseDirectory, "radarr")
	if err != nil {
		t.Fatalf("inspect application data: %v", err)
	}
	if !status.Present || status.ApplicationID != "radarr" {
		t.Fatalf("expected Radarr config present, got %#v", status)
	}
	if _, err := NewApplicationDataManager().Inspect(baseDirectory, "../movies"); err == nil {
		t.Fatal("expected unsafe inspection target to be rejected")
	}
}

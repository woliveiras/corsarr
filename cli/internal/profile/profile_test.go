package profile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProfile(t *testing.T) {
	name := "test-profile"
	p := NewProfile(name)

	if p.Name != name {
		t.Errorf("Expected name %s, got %s", name, p.Name)
	}

	if p.Version != ProfileVersion {
		t.Errorf("Expected version %s, got %s", ProfileVersion, p.Version)
	}

	if p.VPN.Enabled {
		t.Error("Expected VPN to be disabled by default")
	}

	if len(p.Services) != 0 {
		t.Error("Expected empty services list")
	}

	if len(p.Environment) != 0 {
		t.Error("Expected empty environment map")
	}
}

func TestSaveAndLoadProfile(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test profile
	p := NewProfile("test-save-load")
	p.Description = "Test profile"
	p.Services = []string{"radarr", "sonarr", "prowlarr"}
	p.Environment = map[string]string{
		"PUID": "1000",
		"PGID": "1000",
	}
	p.VPN.Enabled = true
	p.VPN.Provider = "nordvpn"

	// Save profile
	if err := SaveProfile(p); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Load profile
	loaded, err := LoadProfile("test-save-load")
	if err != nil {
		t.Fatalf("Failed to load profile: %v", err)
	}

	// Verify loaded data
	if loaded.Name != p.Name {
		t.Errorf("Expected name %s, got %s", p.Name, loaded.Name)
	}

	if loaded.Description != p.Description {
		t.Errorf("Expected description %s, got %s", p.Description, loaded.Description)
	}

	if len(loaded.Services) != len(p.Services) {
		t.Errorf("Expected %d services, got %d", len(p.Services), len(loaded.Services))
	}

	if !loaded.VPN.Enabled {
		t.Error("Expected VPN to be enabled")
	}

	if loaded.VPN.Provider != "nordvpn" {
		t.Errorf("Expected provider nordvpn, got %s", loaded.VPN.Provider)
	}
}

func TestLoadNonExistentProfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	_, err := LoadProfile("nonexistent")
	if err == nil {
		t.Error("Expected error when loading nonexistent profile")
	}
}

func TestListProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initially no profiles
	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}

	// Create some profiles
	p1 := NewProfile("profile1")
	p1.Description = "First profile"
	p1.Services = []string{"radarr"}
	SaveProfile(p1)

	p2 := NewProfile("profile2")
	p2.Description = "Second profile"
	p2.Services = []string{"sonarr", "prowlarr"}
	SaveProfile(p2)

	// List profiles
	profiles, err = ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

	// Verify metadata
	found1 := false
	found2 := false
	for _, p := range profiles {
		if p.Name == "profile1" {
			found1 = true
			if p.Description != "First profile" {
				t.Errorf("Expected description 'First profile', got %s", p.Description)
			}
			if len(p.Services) != 1 {
				t.Errorf("Expected 1 service, got %d", len(p.Services))
			}
		}
		if p.Name == "profile2" {
			found2 = true
			if len(p.Services) != 2 {
				t.Errorf("Expected 2 services, got %d", len(p.Services))
			}
		}
	}

	if !found1 || !found2 {
		t.Error("Not all profiles found in list")
	}
}

func TestDeleteProfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create and save profile
	p := NewProfile("test-delete")
	if err := SaveProfile(p); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Verify it exists
	if !ProfileExists("test-delete") {
		t.Error("Profile should exist before deletion")
	}

	// Delete profile
	if err := DeleteProfile("test-delete"); err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Verify it doesn't exist
	if ProfileExists("test-delete") {
		t.Error("Profile should not exist after deletion")
	}

	// Try to delete again (should error)
	err := DeleteProfile("test-delete")
	if err == nil {
		t.Error("Expected error when deleting nonexistent profile")
	}
}

func TestExportProfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create profile
	p := NewProfile("test-export")
	p.Description = "Export test"
	p.Services = []string{"radarr", "sonarr"}
	p.VPN.Enabled = true
	if err := SaveProfile(p); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Export to file
	exportPath := filepath.Join(tmpDir, "exported.json")
	if err := ExportProfile("test-export", exportPath); err != nil {
		t.Fatalf("Failed to export profile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Exported file should exist")
	}

	// Verify file content (basic check)
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Exported file should not be empty")
	}
}

func TestImportProfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create and export a profile
	p := NewProfile("test-import")
	p.Description = "Import test"
	p.Services = []string{"jellyfin", "jellyseerr"}
	p.VPN.Enabled = false
	SaveProfile(p)

	exportPath := filepath.Join(tmpDir, "import-test.json")
	ExportProfile("test-import", exportPath)

	// Delete the original profile
	DeleteProfile("test-import")

	// Import the profile
	imported, err := ImportProfile(exportPath)
	if err != nil {
		t.Fatalf("Failed to import profile: %v", err)
	}

	// Verify imported data
	if imported.Name != "test-import" {
		t.Errorf("Expected name test-import, got %s", imported.Name)
	}

	if imported.Description != "Import test" {
		t.Errorf("Expected description 'Import test', got %s", imported.Description)
	}

	if len(imported.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(imported.Services))
	}

	if imported.VPN.Enabled {
		t.Error("Expected VPN to be disabled")
	}

	// Save imported profile
	if err := SaveProfile(imported); err != nil {
		t.Fatalf("Failed to save imported profile: %v", err)
	}

	// Verify it can be loaded
	loaded, err := LoadProfile("test-import")
	if err != nil {
		t.Fatalf("Failed to load imported profile: %v", err)
	}

	if loaded.Name != "test-import" {
		t.Error("Imported profile was not saved correctly")
	}
}

func TestProfileExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Profile doesn't exist initially
	if ProfileExists("nonexistent") {
		t.Error("Profile should not exist")
	}

	// Create profile
	p := NewProfile("exists-test")
	SaveProfile(p)

	// Profile should exist now
	if !ProfileExists("exists-test") {
		t.Error("Profile should exist")
	}
}

func TestGetMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create profile with full data
	p := NewProfile("metadata-test")
	p.Description = "Test metadata"
	p.Services = []string{"radarr", "sonarr", "prowlarr"}
	p.Environment = map[string]string{
		"PUID": "1000",
		"PGID": "1000",
		"TZ":   "UTC",
	}
	SaveProfile(p)

	// Get metadata
	meta, err := GetMetadata("metadata-test")
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	// Verify metadata
	if meta.Name != "metadata-test" {
		t.Errorf("Expected name metadata-test, got %s", meta.Name)
	}

	if meta.Description != "Test metadata" {
		t.Errorf("Expected description 'Test metadata', got %s", meta.Description)
	}

	if len(meta.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(meta.Services))
	}

	if meta.Version != ProfileVersion {
		t.Errorf("Expected version %s, got %s", ProfileVersion, meta.Version)
	}
}

func TestProfileTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create profile
	p := NewProfile("timestamp-test")
	initialCreated := p.CreatedAt
	initialUpdated := p.UpdatedAt

	// Save profile
	SaveProfile(p)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Load and modify
	loaded, _ := LoadProfile("timestamp-test")
	loaded.Description = "Updated description"

	// Save again
	SaveProfile(loaded)

	// Load again
	updated, _ := LoadProfile("timestamp-test")

	// CreatedAt should be the same
	if !updated.CreatedAt.Equal(initialCreated) {
		t.Error("CreatedAt should not change")
	}

	// UpdatedAt should be different (newer)
	if !updated.UpdatedAt.After(initialUpdated) {
		t.Error("UpdatedAt should be updated")
	}
}

package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileStoreReturnsEmptyCurrentStateWhenFileIsMissing(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "desktop-state.json"))

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load missing state: %v", err)
	}
	if loaded.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("expected current schema %d, got %d", CurrentSchemaVersion, loaded.SchemaVersion)
	}
	if loaded.StoragePath != "" || len(loaded.Applications) != 0 {
		t.Fatalf("expected empty state, got %#v", loaded)
	}
}

func TestFileStorePersistsDesktopStateWithPrivatePermissions(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "nested", "desktop-state.json")
	store := NewFileStore(statePath)
	want := DesktopState{
		SchemaVersion: CurrentSchemaVersion,
		StoragePath:   "/Users/test/Media",
		Applications:  []string{"prowlarr", "radarr"},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("save desktop state: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load desktop state: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded state mismatch\nwant: %#v\n got: %#v", want, got)
	}

	info, err := os.Stat(statePath)
	if err != nil {
		t.Fatalf("stat desktop state: %v", err)
	}
	if permissions := info.Mode().Perm(); permissions != 0o600 {
		t.Fatalf("expected private state permissions 0600, got %04o", permissions)
	}

	updated := DesktopState{
		SchemaVersion: CurrentSchemaVersion,
		StoragePath:   "/Users/test/Other Media",
		Applications:  []string{"jellyfin"},
	}
	if err := store.Save(updated); err != nil {
		t.Fatalf("replace desktop state: %v", err)
	}
	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("reload replaced desktop state: %v", err)
	}
	if !reflect.DeepEqual(reloaded, updated) {
		t.Fatalf("replaced state mismatch\nwant: %#v\n got: %#v", updated, reloaded)
	}
}

func TestFileStoreRejectsCorruptedState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "desktop-state.json")
	if err := os.WriteFile(statePath, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write corrupted state: %v", err)
	}

	_, err := NewFileStore(statePath).Load()
	if err == nil {
		t.Fatal("expected corrupted state to be rejected")
	}
}

func TestFileStoreMigratesSchemaOneSetupWithoutInventingConsent(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "desktop-state.json")
	legacy := []byte(`{"schemaVersion":1,"storagePath":"/Users/test/Media","applications":["radarr"]}`)
	if err := os.WriteFile(statePath, legacy, 0o600); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}

	loaded, err := NewFileStore(statePath).Load()
	if err != nil {
		t.Fatalf("load legacy state: %v", err)
	}
	if loaded.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("expected migrated schema %d, got %d", CurrentSchemaVersion, loaded.SchemaVersion)
	}
	if loaded.RuntimeConsentVersion != "" || loaded.RuntimeConsentAcceptedAt != "" {
		t.Fatalf("expected consent to remain absent, got %#v", loaded)
	}
}

func TestFileStoreMigratesSchemaTwoWithoutInventingStartAtLogin(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "desktop-state.json")
	legacy := []byte(`{"schemaVersion":2,"storagePath":"/Users/test/Media","applications":["radarr"],"runtimeConsentVersion":"2026-08-10.2","runtimeConsentAcceptedAt":"2026-08-10T18:30:00Z"}`)
	if err := os.WriteFile(statePath, legacy, 0o600); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}

	loaded, err := NewFileStore(statePath).Load()
	if err != nil {
		t.Fatalf("load schema two state: %v", err)
	}
	if loaded.SchemaVersion != CurrentSchemaVersion || loaded.StartAtLogin {
		t.Fatalf("unexpected migrated state %#v", loaded)
	}
}

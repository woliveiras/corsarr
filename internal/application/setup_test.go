package application

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/services"
	statefile "github.com/woliveiras/corsarr/internal/state"
)

func TestSetupServicePersistsValidatedApplicationSelection(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		StoragePath:   "/Users/test/Media",
		Applications:  []string{},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.SaveApplications([]string{"radarr", "prowlarr", "radarr"})
	if err != nil {
		t.Fatalf("save application selection: %v", err)
	}
	want := []string{"prowlarr", "qbittorrent", "radarr"}
	if !reflect.DeepEqual(status.Applications, want) {
		t.Fatalf("expected deterministic unique selection %v, got %v", want, status.Applications)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected one state save, got %d", store.saveCalls)
	}
}

func TestSetupServiceRejectsUnknownApplicationWithoutChangingState(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"prowlarr"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	_, err = service.SaveApplications([]string{"prowlarr", "attacker-controlled-id"})
	if err == nil {
		t.Fatal("expected unknown application to be rejected")
	}
	if store.saveCalls != 0 {
		t.Fatalf("expected rejected selection not to be saved, got %d saves", store.saveCalls)
	}
	if !reflect.DeepEqual(store.desktopState.Applications, []string{"prowlarr"}) {
		t.Fatalf("expected original selection to remain unchanged, got %v", store.desktopState.Applications)
	}
}

func TestSetupServiceSerializesConcurrentStorageAndApplicationUpdates(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &slowMemoryStateStore{
		memoryStateStore: memoryStateStore{desktopState: statefile.DesktopState{
			SchemaVersion: statefile.CurrentSchemaVersion,
			Applications:  []string{},
		}},
		firstLoadStarted: make(chan struct{}),
	}
	service := NewSetupService(NewCatalog(registry), store)

	errors := make(chan error, 2)
	go func() {
		_, saveErr := service.SaveStorage("/Users/test/Media")
		errors <- saveErr
	}()
	<-store.firstLoadStarted
	go func() {
		_, saveErr := service.SaveApplications([]string{"jellyfin"})
		errors <- saveErr
	}()

	for range 2 {
		if saveErr := <-errors; saveErr != nil {
			t.Fatalf("save setup concurrently: %v", saveErr)
		}
	}
	loaded, err := service.Load()
	if err != nil {
		t.Fatalf("load merged setup: %v", err)
	}
	if loaded.StoragePath != "/Users/test/Media" {
		t.Fatalf("expected storage update to survive, got %q", loaded.StoragePath)
	}
	if !reflect.DeepEqual(loaded.Applications, []string{"jellyfin"}) {
		t.Fatalf("expected application update to survive, got %v", loaded.Applications)
	}
}

func TestSetupServiceRequiresExplicitCurrentTermsForInstallation(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		StoragePath:   "/Users/test/Media",
		Applications:  []string{"jellyfin"},
	}}
	service := NewSetupService(NewCatalog(registry), store)
	service.now = func() time.Time {
		return time.Date(2026, 8, 10, 18, 30, 0, 0, time.UTC)
	}

	before, err := service.Load()
	if err != nil {
		t.Fatalf("load setup before consent: %v", err)
	}
	if !before.CanPrepare || before.CanInstall || before.TermsAccepted {
		t.Fatalf("expected preparation without installation authority, got %#v", before)
	}

	after, err := service.AcceptCurrentTerms()
	if err != nil {
		t.Fatalf("accept current terms: %v", err)
	}
	if !after.CanInstall || !after.TermsAccepted || after.TermsVersion != CurrentTermsVersion {
		t.Fatalf("expected current consent to enable installation, got %#v", after)
	}
	if store.desktopState.RuntimeConsentAcceptedAt != "2026-08-10T18:30:00Z" {
		t.Fatalf("unexpected consent timestamp %q", store.desktopState.RuntimeConsentAcceptedAt)
	}
}

type memoryStateStore struct {
	desktopState statefile.DesktopState
	loadErr      error
	saveErr      error
	saveCalls    int
}

type slowMemoryStateStore struct {
	memoryStateStore
	firstLoadStarted chan struct{}
	firstLoadOnce    sync.Once
}

func (s *slowMemoryStateStore) Load() (statefile.DesktopState, error) {
	desktopState, err := s.memoryStateStore.Load()
	s.firstLoadOnce.Do(func() { close(s.firstLoadStarted) })
	time.Sleep(50 * time.Millisecond)
	return desktopState, err
}

func (s *memoryStateStore) Load() (statefile.DesktopState, error) {
	return s.desktopState, s.loadErr
}

func (s *memoryStateStore) Save(desktopState statefile.DesktopState) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saveCalls++
	s.desktopState = desktopState
	return nil
}

package application

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/autostart"
	"github.com/woliveiras/corsarr/internal/quality"
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

	status, err := service.SaveApplications([]string{"radarr", "radarr"})
	if err != nil {
		t.Fatalf("save application selection: %v", err)
	}
	want := []string{"radarr"}
	if !reflect.DeepEqual(status.Applications, want) {
		t.Fatalf("expected deterministic unique selection %v, got %v", want, status.Applications)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected one state save, got %d", store.saveCalls)
	}
}

func TestSetupServiceRequiresQualityStepForARRSelectionAndDefaultsToBalanced1080p(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion:  statefile.CurrentSchemaVersion,
		StoragePath:    "/Users/test/Media",
		Applications:   []string{},
		OnboardingStep: OnboardingStepApplications,
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.SaveApplications([]string{"radarr"})
	if err != nil {
		t.Fatalf("save Arr selection: %v", err)
	}
	if !status.QualityProfileRequired || status.QualityProfilePreset != "balanced-1080p" ||
		status.QualityProfileVersion == "" || store.desktopState.QualityProfileVersion == "" {
		t.Fatalf("expected balanced quality default, got %#v", status)
	}
	status, err = service.AdvanceOnboarding()
	if err != nil || status.OnboardingStep != OnboardingStepQuality {
		t.Fatalf("advance to conditional quality step: status=%#v err=%v", status, err)
	}
}

func TestSetupServicePersistsOnlyKnownQualityPresetForARRSelection(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"sonarr"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.SaveQualityProfilePreset("economy")
	if err != nil {
		t.Fatalf("save quality preset: %v", err)
	}
	if status.QualityProfilePreset != "economy" || store.desktopState.QualityProfilePreset != "economy" ||
		status.QualityProfileVersion == "" || store.desktopState.QualityProfileVersion == "" {
		t.Fatalf("expected persisted preset, status=%#v state=%#v", status, store.desktopState)
	}
	if _, err := service.SaveQualityProfilePreset("invented"); err == nil {
		t.Fatal("expected unknown quality preset to be rejected")
	}
	if store.desktopState.QualityProfilePreset != "economy" {
		t.Fatalf("rejected preset changed state: %#v", store.desktopState)
	}
}

func TestSetupServiceDoesNotAddQualityStepWithoutARRSelection(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion:  statefile.CurrentSchemaVersion,
		Applications:   []string{"jellyfin"},
		OnboardingStep: OnboardingStepApplications,
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.AdvanceOnboarding()
	if err != nil {
		t.Fatalf("keep non-Arr onboarding installable: %v", err)
	}
	if status.OnboardingStep != OnboardingStepApplications || status.QualityProfileRequired {
		t.Fatalf("unexpected quality step without Arr: %#v", status)
	}
}

func TestSetupServiceKeepsCompletedLegacyARRSetupUnmanagedAndInstallable(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion, StoragePath: "/Users/test/Media",
		Applications: []string{"radarr"}, OnboardingCompleted: true,
		RuntimeConsentVersion: CurrentTermsVersion, RuntimeConsentAcceptedAt: "2026-08-12T12:00:00Z",
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.Load()
	if err != nil {
		t.Fatalf("load completed legacy setup: %v", err)
	}
	if !status.CanInstall || status.QualityProfilePreset != "unmanaged" ||
		status.QualityProfileVersion != "" {
		t.Fatalf("legacy quality ownership was invented or blocked installation: %#v", status)
	}
	status, err = service.SaveApplications([]string{"radarr", "jellyfin"})
	if err != nil {
		t.Fatalf("save completed legacy selection: %v", err)
	}
	if status.QualityProfilePreset != "unmanaged" ||
		status.QualityProfileVersion != quality.PresetCatalogVersion {
		t.Fatalf("completed legacy setup did not persist explicit unmanaged state: %#v", status)
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

func TestSetupServiceRejectsNewApplicationWithoutAutomatedSetup(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"prowlarr"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	_, err = service.SaveApplications([]string{"prowlarr", "fileflows"})
	if err == nil {
		t.Fatal("expected application without automated setup to be rejected")
	}
	if store.saveCalls != 0 {
		t.Fatalf("expected rejected selection not to be saved, got %d saves", store.saveCalls)
	}
	if !reflect.DeepEqual(store.desktopState.Applications, []string{"prowlarr"}) {
		t.Fatalf("expected original selection to remain unchanged, got %v", store.desktopState.Applications)
	}
}

func TestSetupServicePreservesLegacyApplicationWithoutAutomatedSetup(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"fileflows"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.SaveApplications([]string{"fileflows", "prowlarr"})
	if err != nil {
		t.Fatalf("preserve legacy manual application: %v", err)
	}
	if !reflect.DeepEqual(status.Applications, []string{"fileflows", "prowlarr"}) {
		t.Fatalf("expected legacy selection to remain manageable, got %v", status.Applications)
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

func TestSetupServiceCompletesOnboardingOnlyForInstallableSetup(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		StoragePath:   "/Users/test/Media", Applications: []string{"jellyfin"},
		RuntimeConsentVersion:    CurrentTermsVersion,
		RuntimeConsentAcceptedAt: "2026-08-11T05:00:00Z",
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.CompleteOnboarding()
	if err != nil {
		t.Fatalf("complete onboarding: %v", err)
	}
	if !status.OnboardingCompleted || !store.desktopState.OnboardingCompleted {
		t.Fatalf("expected persisted onboarding completion, status=%#v state=%#v", status, store.desktopState)
	}
}

func TestSetupServiceDoesNotRepeatCompletedOnboardingWhenTermsChange(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion:            statefile.CurrentSchemaVersion,
		OnboardingCompleted:      true,
		OnboardingStep:           OnboardingStepComplete,
		RuntimeConsentVersion:    "previous-terms-version",
		RuntimeConsentAcceptedAt: "2026-08-10T05:00:00Z",
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.Load()
	if err != nil {
		t.Fatalf("load completed onboarding: %v", err)
	}
	if !status.OnboardingCompleted || status.TermsAccepted {
		t.Fatalf("terms renewal repeated one-time onboarding: %#v", status)
	}
}

func TestSetupServiceRejectsOnboardingCompletionWithoutFullSetup(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		StoragePath:   "/Users/test/Media", Applications: []string{"jellyfin"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	if _, err := service.CompleteOnboarding(); err == nil {
		t.Fatal("expected incomplete onboarding to be rejected")
	}
	if store.saveCalls != 0 || store.desktopState.OnboardingCompleted {
		t.Fatalf("rejected onboarding completion changed state: %#v", store.desktopState)
	}
}

func TestSetupServicePersistsSequentialOnboardingProgress(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.AdvanceOnboarding()
	if err != nil || status.OnboardingStep != OnboardingStepPermissions {
		t.Fatalf("advance welcome: status=%#v err=%v", status, err)
	}
	if _, err := service.AdvanceOnboarding(); err == nil {
		t.Fatal("expected permissions step without consent to be rejected")
	}
	if _, err := service.AcceptCurrentTerms(); err != nil {
		t.Fatalf("accept terms: %v", err)
	}
	status, err = service.AdvanceOnboarding()
	if err != nil || status.OnboardingStep != OnboardingStepEnvironment {
		t.Fatalf("advance permissions: status=%#v err=%v", status, err)
	}
	status, err = service.AdvanceOnboarding()
	if err != nil || status.OnboardingStep != OnboardingStepStorage {
		t.Fatalf("advance environment: status=%#v err=%v", status, err)
	}
	if _, err := service.AdvanceOnboarding(); err == nil {
		t.Fatal("expected storage step without selected folder to be rejected")
	}
	if _, err := service.SaveStorage("/Users/test/Media"); err != nil {
		t.Fatalf("save storage: %v", err)
	}
	status, err = service.AdvanceOnboarding()
	if err != nil || status.OnboardingStep != OnboardingStepApplications {
		t.Fatalf("advance storage: status=%#v err=%v", status, err)
	}
}

func TestSetupServicePersistsExplicitStartAtLoginPreference(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{},
	}}
	login := &fakeAutostartManager{status: autostart.Status{Supported: true}}
	service := NewSetupService(NewCatalog(registry), store, login)

	status, err := service.SetStartAtLogin(true)
	if err != nil {
		t.Fatalf("enable start at login: %v", err)
	}
	if !status.StartAtLogin || !status.StartAtLoginSupported || !store.desktopState.StartAtLogin {
		t.Fatalf("unexpected enabled setup %#v state=%#v", status, store.desktopState)
	}
	if login.setCalls != 1 || !login.lastEnabled {
		t.Fatalf("expected one native enable, got %#v", login)
	}
}

func TestSetupServiceReportsNativeApprovalWithoutClaimingEnabled(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{},
	}}
	login := &fakeAutostartManager{
		status:         autostart.Status{Supported: true},
		statusAfterSet: autostart.Status{Supported: true, RequiresApproval: true},
	}
	service := NewSetupService(NewCatalog(registry), store, login)

	status, err := service.SetStartAtLogin(true)
	if err != nil {
		t.Fatalf("request start at login: %v", err)
	}
	if !status.StartAtLogin || !status.StartAtLoginRequiresApproval {
		t.Fatalf("expected approval-required preference, got %#v", status)
	}
	if !store.desktopState.StartAtLogin {
		t.Fatal("expected requested preference to persist while approval is pending")
	}
}

func TestSetupServiceAllowsExplicitJellyfinLANOnlyWhenSelected(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"jellyfin"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	status, err := service.SetJellyfinLAN(true)
	if err != nil {
		t.Fatalf("enable Jellyfin LAN: %v", err)
	}
	if !status.JellyfinLANEnabled || !store.desktopState.AllowJellyfinLAN {
		t.Fatalf("expected persisted LAN choice, status=%#v state=%#v", status, store.desktopState)
	}

	status, err = service.SaveApplications([]string{"radarr"})
	if err != nil {
		t.Fatalf("replace application selection: %v", err)
	}
	if status.JellyfinLANEnabled || store.desktopState.AllowJellyfinLAN {
		t.Fatalf("LAN choice survived Jellyfin removal, status=%#v state=%#v", status, store.desktopState)
	}
}

func TestSetupServiceRejectsJellyfinLANWithoutJellyfin(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	store := &memoryStateStore{desktopState: statefile.DesktopState{
		SchemaVersion: statefile.CurrentSchemaVersion,
		Applications:  []string{"radarr"},
	}}
	service := NewSetupService(NewCatalog(registry), store)

	if _, err := service.SetJellyfinLAN(true); err == nil {
		t.Fatal("expected Jellyfin LAN without Jellyfin to be rejected")
	}
	if store.saveCalls != 0 {
		t.Fatalf("rejected LAN choice changed state %d times", store.saveCalls)
	}
}

type memoryStateStore struct {
	desktopState statefile.DesktopState
	loadErr      error
	saveErr      error
	saveCalls    int
}

type fakeAutostartManager struct {
	status         autostart.Status
	statusAfterSet autostart.Status
	setErr         error
	setCalls       int
	lastEnabled    bool
}

func (m *fakeAutostartManager) Status() (autostart.Status, error) {
	return m.status, nil
}

func (m *fakeAutostartManager) SetEnabled(enabled bool) (autostart.Status, error) {
	m.setCalls++
	m.lastEnabled = enabled
	if m.setErr != nil {
		return m.status, m.setErr
	}
	if m.statusAfterSet.Supported {
		m.status = m.statusAfterSet
	} else {
		m.status.Enabled = enabled
	}
	return m.status, nil
}

func (m *fakeAutostartManager) OpenSystemSettings() error { return nil }

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

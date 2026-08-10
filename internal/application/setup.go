package application

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/woliveiras/corsarr/internal/autostart"
	statefile "github.com/woliveiras/corsarr/internal/state"
)

type SetupStatus struct {
	StoragePath                  string   `json:"storagePath,omitempty"`
	Applications                 []string `json:"applications"`
	CanPrepare                   bool     `json:"canPrepare"`
	CanInstall                   bool     `json:"canInstall"`
	TermsVersion                 string   `json:"termsVersion"`
	TermsAccepted                bool     `json:"termsAccepted"`
	StartAtLogin                 bool     `json:"startAtLogin"`
	StartAtLoginSupported        bool     `json:"startAtLoginSupported"`
	StartAtLoginRequiresApproval bool     `json:"startAtLoginRequiresApproval"`
	JellyfinLANEnabled           bool     `json:"jellyfinLanEnabled"`
}

const CurrentTermsVersion = "2026-08-10.2"

type SetupService struct {
	catalog   *Catalog
	store     statefile.Store
	mu        sync.Mutex
	now       func() time.Time
	autostart autostart.Manager
}

func NewSetupService(
	catalog *Catalog,
	store statefile.Store,
	autostartManagers ...autostart.Manager,
) *SetupService {
	manager := autostart.NewPlatformManager("unsupported")
	if len(autostartManagers) > 0 && autostartManagers[0] != nil {
		manager = autostartManagers[0]
	}
	return &SetupService{catalog: catalog, store: store, now: time.Now, autostart: manager}
}

func (s *SetupService) AcceptCurrentTerms() (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	desktopState.RuntimeConsentVersion = CurrentTermsVersion
	desktopState.RuntimeConsentAcceptedAt = s.now().UTC().Format(time.RFC3339)
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save runtime consent: %w", err)
	}
	return s.status(desktopState)
}

func (s *SetupService) Load() (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	return s.status(desktopState)
}

func (s *SetupService) SaveStorage(path string) (SetupStatus, error) {
	if strings.TrimSpace(path) == "" {
		return SetupStatus{}, fmt.Errorf("storage path is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.StoragePath = path
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save desktop storage: %w", err)
	}
	return s.status(desktopState)
}

func (s *SetupService) SaveApplications(applicationIDs []string) (SetupStatus, error) {
	applications, err := s.validatedApplications(applicationIDs)
	if err != nil {
		return SetupStatus{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = applications
	if !containsApplication(applications, "jellyfin") {
		desktopState.AllowJellyfinLAN = false
	}
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save desktop applications: %w", err)
	}
	return s.status(desktopState)
}

func (s *SetupService) SetJellyfinLAN(enabled bool) (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	if enabled && !containsApplication(desktopState.Applications, "jellyfin") {
		return SetupStatus{}, fmt.Errorf("Jellyfin must be selected before enabling LAN access")
	}
	desktopState.AllowJellyfinLAN = enabled
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save Jellyfin LAN access: %w", err)
	}
	return s.status(desktopState)
}

func (s *SetupService) SetStartAtLogin(enabled bool) (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	before, err := s.autostart.Status()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("inspect start-at-login setting: %w", err)
	}
	after, err := s.autostart.SetEnabled(enabled)
	if err != nil {
		return SetupStatus{}, err
	}
	desktopState.StartAtLogin = after.Enabled || after.RequiresApproval
	if err := s.store.Save(desktopState); err != nil {
		_, rollbackErr := s.autostart.SetEnabled(before.Enabled || before.RequiresApproval)
		return SetupStatus{}, errors.Join(
			fmt.Errorf("save start-at-login setting: %w", err),
			rollbackErr,
		)
	}
	return setupStatus(desktopState, after), nil
}

func (s *SetupService) OpenStartAtLoginSettings() error {
	return s.autostart.OpenSystemSettings()
}

func (s *SetupService) validatedApplications(applicationIDs []string) ([]string, error) {
	unique := make(map[string]struct{}, len(applicationIDs))
	for _, id := range applicationIDs {
		if _, known := s.catalog.byID[id]; !known {
			return nil, fmt.Errorf("application is not available in the desktop catalog: %s", id)
		}
		if err := s.includeDependencies(id, unique); err != nil {
			return nil, err
		}
	}

	applications := make([]string, 0, len(unique))
	for id := range unique {
		applications = append(applications, id)
	}
	sort.Strings(applications)
	return applications, nil
}

func (s *SetupService) includeDependencies(id string, selected map[string]struct{}) error {
	if _, alreadySelected := selected[id]; alreadySelected {
		return nil
	}
	application, exists := s.catalog.byID[id]
	if !exists {
		return fmt.Errorf("required application is not available in the desktop catalog: %s", id)
	}
	selected[id] = struct{}{}
	for _, dependencyID := range application.Dependencies {
		if err := s.includeDependencies(dependencyID, selected); err != nil {
			return err
		}
	}
	return nil
}

func (s *SetupService) knownApplications(applicationIDs []string) []string {
	known := make([]string, 0, len(applicationIDs))
	seen := make(map[string]struct{}, len(applicationIDs))
	for _, id := range applicationIDs {
		if _, exists := s.catalog.byID[id]; !exists {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		known = append(known, id)
	}
	sort.Strings(known)
	return known
}

func (s *SetupService) status(desktopState statefile.DesktopState) (SetupStatus, error) {
	loginStatus, err := s.autostart.Status()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("inspect start-at-login setting: %w", err)
	}
	return setupStatus(desktopState, loginStatus), nil
}

func setupStatus(desktopState statefile.DesktopState, loginStatus autostart.Status) SetupStatus {
	canPrepare := desktopState.StoragePath != "" && len(desktopState.Applications) > 0
	termsAccepted := desktopState.RuntimeConsentVersion == CurrentTermsVersion &&
		desktopState.RuntimeConsentAcceptedAt != ""
	return SetupStatus{
		StoragePath:                  desktopState.StoragePath,
		Applications:                 desktopState.Applications,
		CanPrepare:                   canPrepare,
		CanInstall:                   canPrepare && termsAccepted,
		TermsVersion:                 CurrentTermsVersion,
		TermsAccepted:                termsAccepted,
		StartAtLogin:                 loginStatus.Enabled || loginStatus.RequiresApproval,
		StartAtLoginSupported:        loginStatus.Supported,
		StartAtLoginRequiresApproval: loginStatus.RequiresApproval,
		JellyfinLANEnabled:           desktopState.AllowJellyfinLAN,
	}
}

func containsApplication(applications []string, expected string) bool {
	for _, applicationID := range applications {
		if applicationID == expected {
			return true
		}
	}
	return false
}

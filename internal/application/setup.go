package application

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/woliveiras/corsarr/internal/autostart"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/quality"
	statefile "github.com/woliveiras/corsarr/internal/state"
)

type SetupStatus struct {
	Language                     string   `json:"language,omitempty"`
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
	OnboardingCompleted          bool     `json:"onboardingCompleted"`
	OnboardingStep               string   `json:"onboardingStep"`
	QualityProfileRequired       bool     `json:"qualityProfileRequired"`
	QualityProfilePreset         string   `json:"qualityProfilePreset,omitempty"`
	QualityProfileVersion        string   `json:"qualityProfileVersion,omitempty"`
}

func (s *SetupService) SaveLanguagePreference(languageCode string) (SetupStatus, error) {
	languageCode, err := i18n.NormalizeLanguage(languageCode)
	if err != nil {
		return SetupStatus{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Language = languageCode
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save desktop language: %w", err)
	}
	return s.status(desktopState)
}

const CurrentTermsVersion = "2026-08-10.2"

const (
	OnboardingStepWelcome      = "welcome"
	OnboardingStepPermissions  = "permissions"
	OnboardingStepEnvironment  = "environment"
	OnboardingStepStorage      = "storage"
	OnboardingStepApplications = "applications"
	OnboardingStepQuality      = "quality"
	OnboardingStepComplete     = "complete"
)

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

// CompleteOnboarding persists the one-time first-run boundary only after the
// reviewed setup has current consent, storage, and at least one application.
// The desktop calls this only after installation reports complete.
func (s *SetupService) CompleteOnboarding() (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	termsAccepted := desktopState.RuntimeConsentVersion == CurrentTermsVersion &&
		desktopState.RuntimeConsentAcceptedAt != ""
	qualityRequired := containsARRApplication(desktopState.Applications)
	preset := normalizedQualityPreset(desktopState)
	if desktopState.StoragePath == "" || len(desktopState.Applications) == 0 || !termsAccepted ||
		(qualityRequired && (!quality.ValidPreset(preset) ||
			desktopState.QualityProfileVersion != quality.PresetCatalogVersion)) {
		return SetupStatus{}, fmt.Errorf("onboarding setup is incomplete")
	}
	desktopState.OnboardingCompleted = true
	desktopState.OnboardingStep = OnboardingStepComplete
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save onboarding completion: %w", err)
	}
	return s.status(desktopState)
}

// AdvanceOnboarding persists the furthest reviewed step so an interrupted
// first run resumes without silently skipping consent or storage selection.
func (s *SetupService) AdvanceOnboarding() (SetupStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	current := normalizedOnboardingStep(desktopState)
	switch current {
	case OnboardingStepWelcome:
		desktopState.OnboardingStep = OnboardingStepPermissions
	case OnboardingStepPermissions:
		termsAccepted := desktopState.RuntimeConsentVersion == CurrentTermsVersion &&
			desktopState.RuntimeConsentAcceptedAt != ""
		if !termsAccepted {
			return SetupStatus{}, fmt.Errorf("current terms must be accepted before continuing onboarding")
		}
		desktopState.OnboardingStep = OnboardingStepEnvironment
	case OnboardingStepEnvironment:
		desktopState.OnboardingStep = OnboardingStepStorage
	case OnboardingStepStorage:
		if strings.TrimSpace(desktopState.StoragePath) == "" {
			return SetupStatus{}, fmt.Errorf("storage must be selected before continuing onboarding")
		}
		desktopState.OnboardingStep = OnboardingStepApplications
	case OnboardingStepApplications:
		if containsARRApplication(desktopState.Applications) {
			desktopState.OnboardingStep = OnboardingStepQuality
		} else {
			return s.status(desktopState)
		}
	case OnboardingStepQuality, OnboardingStepComplete:
		return s.status(desktopState)
	default:
		return SetupStatus{}, fmt.Errorf("unsupported onboarding step: %s", current)
	}
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save onboarding progress: %w", err)
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
	if err := s.rejectNewManualApplications(applications, desktopState.Applications); err != nil {
		return SetupStatus{}, err
	}
	desktopState.Applications = applications
	if containsARRApplication(applications) {
		if desktopState.OnboardingCompleted {
			if !quality.ValidPreset(desktopState.QualityProfilePreset) {
				desktopState.QualityProfilePreset = string(quality.PresetUnmanaged)
				desktopState.QualityProfileVersion = quality.PresetCatalogVersion
			}
		} else {
			if !quality.ValidPreset(desktopState.QualityProfilePreset) {
				desktopState.QualityProfilePreset = string(quality.PresetBalanced1080p)
			}
			desktopState.QualityProfileVersion = quality.PresetCatalogVersion
		}
	} else {
		desktopState.QualityProfilePreset = ""
		desktopState.QualityProfileVersion = ""
		if desktopState.OnboardingStep == OnboardingStepQuality {
			desktopState.OnboardingStep = OnboardingStepApplications
		}
	}
	if !containsApplication(applications, "jellyfin") {
		desktopState.AllowJellyfinLAN = false
	}
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save desktop applications: %w", err)
	}
	return s.status(desktopState)
}

func (s *SetupService) rejectNewManualApplications(
	requested []string,
	existing []string,
) error {
	alreadySelected := make(map[string]struct{}, len(existing))
	for _, id := range existing {
		alreadySelected[id] = struct{}{}
	}
	for _, id := range requested {
		application := s.catalog.byID[id]
		if application.AutomatedSetup {
			continue
		}
		if _, legacySelection := alreadySelected[id]; legacySelection {
			continue
		}
		return fmt.Errorf(
			"application does not have automated desktop setup: %s",
			id,
		)
	}
	return nil
}

func (s *SetupService) SaveQualityProfilePreset(preset string) (SetupStatus, error) {
	if !quality.ValidPreset(preset) {
		return SetupStatus{}, fmt.Errorf("quality profile preset is not supported: %s", preset)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	desktopState, err := s.store.Load()
	if err != nil {
		return SetupStatus{}, fmt.Errorf("load desktop setup: %w", err)
	}
	desktopState.Applications = s.knownApplications(desktopState.Applications)
	if !containsARRApplication(desktopState.Applications) {
		return SetupStatus{}, fmt.Errorf("Radarr or Sonarr must be selected before choosing a quality profile")
	}
	desktopState.QualityProfilePreset = preset
	desktopState.QualityProfileVersion = quality.PresetCatalogVersion
	if err := s.store.Save(desktopState); err != nil {
		return SetupStatus{}, fmt.Errorf("save quality profile preset: %w", err)
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
		return SetupStatus{}, fmt.Errorf("jellyfin must be selected before enabling LAN access")
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
		unique[id] = struct{}{}
	}

	applications := make([]string, 0, len(unique))
	for id := range unique {
		applications = append(applications, id)
	}
	sort.Strings(applications)
	return applications, nil
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
	qualityRequired := containsARRApplication(desktopState.Applications)
	qualityPreset := normalizedQualityPreset(desktopState)
	qualityReady := desktopState.OnboardingCompleted || !qualityRequired || (quality.ValidPreset(qualityPreset) &&
		desktopState.QualityProfileVersion == quality.PresetCatalogVersion)
	return SetupStatus{
		Language:                     desktopState.Language,
		StoragePath:                  desktopState.StoragePath,
		Applications:                 desktopState.Applications,
		CanPrepare:                   canPrepare,
		CanInstall:                   canPrepare && termsAccepted && qualityReady,
		TermsVersion:                 CurrentTermsVersion,
		TermsAccepted:                termsAccepted,
		StartAtLogin:                 loginStatus.Enabled || loginStatus.RequiresApproval,
		StartAtLoginSupported:        loginStatus.Supported,
		StartAtLoginRequiresApproval: loginStatus.RequiresApproval,
		JellyfinLANEnabled:           desktopState.AllowJellyfinLAN,
		OnboardingCompleted:          desktopState.OnboardingCompleted,
		OnboardingStep:               normalizedOnboardingStep(desktopState),
		QualityProfileRequired:       qualityRequired,
		QualityProfilePreset:         qualityPreset,
		QualityProfileVersion:        desktopState.QualityProfileVersion,
	}
}

func normalizedQualityPreset(desktopState statefile.DesktopState) string {
	if !containsARRApplication(desktopState.Applications) {
		return ""
	}
	if quality.ValidPreset(desktopState.QualityProfilePreset) {
		return desktopState.QualityProfilePreset
	}
	if desktopState.OnboardingCompleted {
		return string(quality.PresetUnmanaged)
	}
	return string(quality.PresetBalanced1080p)
}

func containsARRApplication(applications []string) bool {
	return containsApplication(applications, "radarr") || containsApplication(applications, "sonarr")
}

func normalizedOnboardingStep(desktopState statefile.DesktopState) string {
	if desktopState.OnboardingCompleted {
		return OnboardingStepComplete
	}
	if desktopState.OnboardingStep == "" {
		return OnboardingStepWelcome
	}
	return desktopState.OnboardingStep
}

func containsApplication(applications []string, expected string) bool {
	for _, applicationID := range applications {
		if applicationID == expected {
			return true
		}
	}
	return false
}

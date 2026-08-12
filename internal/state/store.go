package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const CurrentSchemaVersion = 6

type DesktopState struct {
	SchemaVersion            int      `json:"schemaVersion"`
	StoragePath              string   `json:"storagePath,omitempty"`
	Applications             []string `json:"applications"`
	RuntimeConsentVersion    string   `json:"runtimeConsentVersion,omitempty"`
	RuntimeConsentAcceptedAt string   `json:"runtimeConsentAcceptedAt,omitempty"`
	StartAtLogin             bool     `json:"startAtLogin,omitempty"`
	AllowJellyfinLAN         bool     `json:"allowJellyfinLan,omitempty"`
	OnboardingCompleted      bool     `json:"onboardingCompleted,omitempty"`
	OnboardingStep           string   `json:"onboardingStep,omitempty"`
	QualityProfilePreset     string   `json:"qualityProfilePreset,omitempty"`
	QualityProfileVersion    string   `json:"qualityProfileVersion,omitempty"`
}

type Store interface {
	Load() (DesktopState, error)
	Save(state DesktopState) error
}

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func DefaultPath() (string, error) {
	configurationDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user configuration directory: %w", err)
	}
	return filepath.Join(configurationDirectory, "Corsarr", "desktop-state.json"), nil
}

func (s *FileStore) Load() (DesktopState, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return emptyDesktopState(), nil
	}
	if err != nil {
		return DesktopState{}, fmt.Errorf("read desktop state: %w", err)
	}

	var desktopState DesktopState
	if err := json.Unmarshal(data, &desktopState); err != nil {
		return DesktopState{}, fmt.Errorf("decode desktop state: %w", err)
	}
	if desktopState.SchemaVersion >= 1 && desktopState.SchemaVersion < CurrentSchemaVersion {
		desktopState.SchemaVersion = CurrentSchemaVersion
	}
	if desktopState.SchemaVersion != CurrentSchemaVersion {
		return DesktopState{}, fmt.Errorf(
			"unsupported desktop state schema: %d",
			desktopState.SchemaVersion,
		)
	}
	if desktopState.Applications == nil {
		desktopState.Applications = []string{}
	}
	return desktopState, nil
}

func (s *FileStore) Save(desktopState DesktopState) error {
	desktopState.SchemaVersion = CurrentSchemaVersion
	if desktopState.Applications == nil {
		desktopState.Applications = []string{}
	}

	data, err := json.MarshalIndent(desktopState, "", "  ")
	if err != nil {
		return fmt.Errorf("encode desktop state: %w", err)
	}
	data = append(data, '\n')

	stateDirectory := filepath.Dir(s.path)
	if err := os.MkdirAll(stateDirectory, 0o700); err != nil {
		return fmt.Errorf("create desktop state directory: %w", err)
	}

	temporaryFile, err := os.CreateTemp(stateDirectory, ".desktop-state-*")
	if err != nil {
		return fmt.Errorf("create temporary desktop state: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer func() { _ = os.Remove(temporaryPath) }()

	if err := temporaryFile.Chmod(0o600); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("protect temporary desktop state: %w", err)
	}
	if _, err := temporaryFile.Write(data); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("write temporary desktop state: %w", err)
	}
	if err := temporaryFile.Sync(); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("sync temporary desktop state: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("close temporary desktop state: %w", err)
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("replace desktop state: %w", err)
	}
	return nil
}

func emptyDesktopState() DesktopState {
	return DesktopState{
		SchemaVersion: CurrentSchemaVersion,
		Applications:  []string{},
	}
}

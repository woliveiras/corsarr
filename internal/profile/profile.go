package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Profile represents a saved configuration
type Profile struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	Version     string            `json:"version" yaml:"version"`
	VPN         VPNConfig         `json:"vpn" yaml:"vpn"`
	Services    []string          `json:"services" yaml:"services"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	OutputDir   string            `json:"output_dir" yaml:"output_dir"`
}

// VPNConfig holds VPN-related configuration
type VPNConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Country  string `json:"country,omitempty" yaml:"country,omitempty"`
	City     string `json:"city,omitempty" yaml:"city,omitempty"`
}

// Metadata contains profile summary information
type Metadata struct {
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
	Version     string    `json:"version" yaml:"version"`
	Services    []string  `json:"services" yaml:"services"`
}

const (
	// ProfileVersion is the current profile format version
	ProfileVersion = "1.0.0"
	// DefaultProfileDir is the default directory for storing profiles
	DefaultProfileDir = ".corsarr/profiles"
)

// GetProfileDir returns the profile directory path
func GetProfileDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, DefaultProfileDir), nil
}

// ensureProfileDir creates the profile directory if it doesn't exist
func ensureProfileDir() error {
	profileDir, err := GetProfileDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(profileDir, 0755)
}

// getProfilePath returns the full path for a profile file
func getProfilePath(name string) (string, error) {
	profileDir, err := GetProfileDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(profileDir, name+".yaml"), nil
}

// NewProfile creates a new profile with default values
func NewProfile(name string) *Profile {
	now := time.Now()
	return &Profile{
		Name:        name,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     ProfileVersion,
		Services:    []string{},
		Environment: make(map[string]string),
		VPN: VPNConfig{
			Enabled: false,
		},
	}
}

// SaveProfile saves a profile to disk
func SaveProfile(profile *Profile) error {
	if err := ensureProfileDir(); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	profilePath, err := getProfilePath(profile.Name)
	if err != nil {
		return err
	}

	// Update timestamp
	profile.UpdatedAt = time.Now()
	profile.Version = ProfileVersion

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// LoadProfile loads a profile from disk
func LoadProfile(name string) (*Profile, error) {
	profilePath, err := getProfilePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles returns a list of all saved profiles
func ListProfiles() ([]*Metadata, error) {
	profileDir, err := GetProfileDir()
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return []*Metadata{}, nil
	}

	entries, err := os.ReadDir(profileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile directory: %w", err)
	}

	var profiles []*Metadata
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .yaml extension
		profile, err := LoadProfile(name)
		if err != nil {
			continue // Skip invalid profiles
		}

		profiles = append(profiles, &Metadata{
			Name:        profile.Name,
			Description: profile.Description,
			CreatedAt:   profile.CreatedAt,
			UpdatedAt:   profile.UpdatedAt,
			Version:     profile.Version,
			Services:    profile.Services,
		})
	}

	return profiles, nil
}

// DeleteProfile removes a profile from disk
func DeleteProfile(name string) error {
	profilePath, err := getProfilePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(profilePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile '%s' not found", name)
		}
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ExportProfile exports a profile to a specific path in JSON format
func ExportProfile(name, outputPath string) error {
	profile, err := LoadProfile(name)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportProfile imports a profile from a JSON or YAML file
func ImportProfile(inputPath string) (*Profile, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	var profile Profile

	// Try YAML first
	if err := yaml.Unmarshal(data, &profile); err != nil {
		// Try JSON
		if jsonErr := json.Unmarshal(data, &profile); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse profile (tried YAML and JSON): %w", err)
		}
	}

	// Update metadata
	profile.UpdatedAt = time.Now()
	if profile.Version == "" {
		profile.Version = ProfileVersion
	}

	return &profile, nil
}

// ProfileExists checks if a profile exists
func ProfileExists(name string) bool {
	profilePath, err := getProfilePath(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(profilePath)
	return err == nil
}

// GetMetadata returns metadata for a profile without loading the full profile
func GetMetadata(name string) (*Metadata, error) {
	profile, err := LoadProfile(name)
	if err != nil {
		return nil, err
	}

	return &Metadata{
		Name:        profile.Name,
		Description: profile.Description,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
		Version:     profile.Version,
		Services:    profile.Services,
	}, nil
}

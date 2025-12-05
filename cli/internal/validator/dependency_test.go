package validator

import (
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestDependencyValidator(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tests := []struct {
		name         string
		serviceIDs   []string
		vpnEnabled   bool
		expectErrors bool
		errorCount   int
	}{
		{
			name:         "No dependencies",
			serviceIDs:   []string{"jellyfin"},
			vpnEnabled:   false,
			expectErrors: false,
			errorCount:   0,
		},
		{
			name:         "Missing dependency - Radarr without qBittorrent",
			serviceIDs:   []string{"radarr"}, // Depends on qbittorrent and prowlarr
			vpnEnabled:   false,
			expectErrors: true,
			errorCount:   2, // Missing qbittorrent and prowlarr
		},
		{
			name:         "Missing dependency - Jellyseerr without Jellyfin",
			serviceIDs:   []string{"jellyseerr"}, // Depends on jellyfin
			vpnEnabled:   false,
			expectErrors: true,
			errorCount:   1, // Missing jellyfin
		},
		{
			name:         "VPN enabled without Gluetun",
			serviceIDs:   []string{"sonarr"},
			vpnEnabled:   true,
			expectErrors: true,
			errorCount:   3, // Missing Gluetun + missing dependencies (qbittorrent, prowlarr)
		},
		{
			name:         "Valid with VPN and all dependencies",
			serviceIDs:   []string{"gluetun", "qbittorrent", "prowlarr", "radarr"},
			vpnEnabled:   true,
			expectErrors: false,
			errorCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(registry, tt.serviceIDs, "/data", "/output", tt.vpnEnabled)
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}

			validator := NewDependencyValidator(config)
			result := validator.Validate()

			if tt.expectErrors && !result.HasErrors() {
				t.Error("Expected errors but got none")
			}

			if !tt.expectErrors && result.HasErrors() {
				t.Errorf("Expected no errors but got %d: %v", len(result.Errors), result.Errors)
			}

			if len(result.Errors) != tt.errorCount {
				t.Errorf("Expected %d errors, got %d", tt.errorCount, len(result.Errors))
			}
		})
	}
}

func TestGetMissingDependencies(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Radarr depends on qbittorrent and prowlarr
	config, err := NewConfig(registry, []string{"radarr"}, "/data", "/output", false)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	missing := GetMissingDependencies(config)
	
	if len(missing) == 0 {
		t.Error("Expected missing dependencies for Radarr")
	}

	if deps, ok := missing["Radarr"]; ok {
		if len(deps) != 2 {
			t.Errorf("Radarr should have 2 missing dependencies (qbittorrent, prowlarr), got %d: %v", len(deps), deps)
		}
	} else {
		t.Error("Radarr not found in missing dependencies map")
	}
}

func TestSuggestDependencies(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Jellyseerr depends on jellyfin
	config, err := NewConfig(registry, []string{"jellyseerr"}, "/data", "/output", false)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	suggestions := SuggestDependencies(config)
	
	if len(suggestions) == 0 {
		t.Error("Expected dependency suggestions for Jellyseerr")
	}

	// Should suggest Jellyfin
	found := false
	for _, suggestion := range suggestions {
		if suggestion == "Jellyfin" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected Jellyfin in suggestions, got: %v", suggestions)
	}
}

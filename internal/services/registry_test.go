package services

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	if registry == nil {
		t.Fatal("Registry is nil")
	}

	count := registry.GetServiceCount()
	if count == 0 {
		t.Error("Expected services to be loaded, got 0")
	}

	t.Logf("✅ Loaded %d services", count)
}

func TestGetService(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tests := []struct {
		name      string
		serviceID string
		wantErr   bool
	}{
		{"Valid service - qbittorrent", "qbittorrent", false},
		{"Valid service - radarr", "radarr", false},
		{"Valid service - sonarr", "sonarr", false},
		{"Invalid service", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := registry.GetService(tt.serviceID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for service %s, got nil", tt.serviceID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for service %s: %v", tt.serviceID, err)
				}
				if service == nil {
					t.Errorf("Expected service %s to be found", tt.serviceID)
				}
				if service != nil && service.ID != tt.serviceID {
					t.Errorf("Expected service ID %s, got %s", tt.serviceID, service.ID)
				}
			}
		})
	}
}

func TestGetServicesByCategory(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tests := []struct {
		category    ServiceCategory
		minExpected int
	}{
		{CategoryDownload, 1},   // at least qbittorrent
		{CategoryIndexer, 1},    // at least prowlarr
		{CategoryMedia, 2},      // at least radarr and sonarr
		{CategoryStreaming, 1},  // at least jellyfin
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			services := registry.GetServicesByCategory(tt.category)
			if len(services) < tt.minExpected {
				t.Errorf("Expected at least %d services in category %s, got %d",
					tt.minExpected, tt.category, len(services))
			}
			t.Logf("Category %s has %d services", tt.category, len(services))
		})
	}
}

func TestValidateDependencies(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tests := []struct {
		name       string
		services   []string
		shouldPass bool
	}{
		{
			name:       "Valid: Basic stack",
			services:   []string{"qbittorrent", "prowlarr", "radarr", "sonarr"},
			shouldPass: true,
		},
		{
			name:       "Invalid: Sonarr without dependencies",
			services:   []string{"sonarr"},
			shouldPass: false,
		},
		{
			name:       "Invalid: Jellyseerr without Jellyfin",
			services:   []string{"qbittorrent", "prowlarr", "jellyseerr"},
			shouldPass: false,
		},
		{
			name:       "Valid: Complete stack with Jellyfin",
			services:   []string{"qbittorrent", "prowlarr", "radarr", "sonarr", "jellyfin", "jellyseerr"},
			shouldPass: true,
		},
		{
			name:       "Valid: Just Jellyfin (no dependencies)",
			services:   []string{"jellyfin"},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.ValidateDependencies(tt.services)
			if tt.shouldPass && err != nil {
				t.Errorf("Expected to pass but failed: %v", err)
			} else if !tt.shouldPass && err == nil {
				t.Errorf("Expected to fail but passed")
			}
		})
	}
}

func TestFilterByVPNCompatibility(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test with VPN enabled
	t.Run("VPN enabled", func(t *testing.T) {
		services := registry.FilterByVPNCompatibility(true)
		if len(services) == 0 {
			t.Error("Expected services when VPN is enabled, got 0")
		}

		// Check that gluetun is not in the list
		for _, service := range services {
			if service.Category == CategoryVPN {
				t.Error("VPN service should not be in filtered list")
			}
		}

		t.Logf("VPN enabled: %d compatible services", len(services))
	})

	// Test with VPN disabled
	t.Run("VPN disabled", func(t *testing.T) {
		services := registry.FilterByVPNCompatibility(false)
		if len(services) == 0 {
			t.Error("Expected services when VPN is disabled, got 0")
		}

		// Check that no service requires VPN
		for _, service := range services {
			if service.RequiresVPN {
				t.Errorf("Service %s requires VPN but VPN is disabled", service.Name)
			}
		}

		t.Logf("VPN disabled: %d compatible services", len(services))
	})
}

func TestGetAllServices(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	services := registry.GetAllServices()
	if len(services) == 0 {
		t.Error("Expected services, got 0")
	}

	// Verify all services have required fields
	for _, service := range services {
		if service.ID == "" {
			t.Error("Service has empty ID")
		}
		if service.Name == "" {
			t.Errorf("Service %s has empty Name", service.ID)
		}
		if service.Image == "" {
			t.Errorf("Service %s has empty Image", service.ID)
		}
		if service.ContainerName == "" {
			t.Errorf("Service %s has empty ContainerName", service.ID)
		}
	}

	t.Logf("✅ All %d services are valid", len(services))
}

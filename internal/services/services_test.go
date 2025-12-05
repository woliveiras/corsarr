package services

import (
	"testing"
)

func TestService_GetTranslationKey(t *testing.T) {
	service := &Service{
		ID:   "radarr",
		Name: "Radarr",
	}

	expected := "services.radarr"
	result := service.GetTranslationKey()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestService_GetNameKey(t *testing.T) {
	service := &Service{
		ID:   "sonarr",
		Name: "Sonarr",
	}

	expected := "services.sonarr.name"
	result := service.GetNameKey()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestService_GetDescriptionKey(t *testing.T) {
	service := &Service{
		ID:   "prowlarr",
		Name: "Prowlarr",
	}

	expected := "services_prowlarr_description"
	result := service.GetDescriptionKey()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestService_IsCompatibleWithVPN(t *testing.T) {
	tests := []struct {
		name       string
		service    *Service
		vpnEnabled bool
		expected   bool
	}{
		{
			name: "Service requires VPN and VPN is enabled",
			service: &Service{
				ID:          "test1",
				RequiresVPN: true,
			},
			vpnEnabled: true,
			expected:   true,
		},
		{
			name: "Service requires VPN but VPN is disabled",
			service: &Service{
				ID:          "test2",
				RequiresVPN: true,
			},
			vpnEnabled: false,
			expected:   false,
		},
		{
			name: "Service supports VPN and VPN is enabled",
			service: &Service{
				ID:          "test3",
				RequiresVPN: false,
				SupportsVPN: true,
			},
			vpnEnabled: true,
			expected:   true,
		},
		{
			name: "Service supports VPN but VPN is disabled",
			service: &Service{
				ID:          "test4",
				RequiresVPN: false,
				SupportsVPN: true,
			},
			vpnEnabled: false,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.IsCompatibleWithVPN(tt.vpnEnabled)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestService_HasDependencies(t *testing.T) {
	tests := []struct {
		name     string
		service  *Service
		expected bool
	}{
		{
			name: "Service with dependencies",
			service: &Service{
				ID:           "sonarr",
				Dependencies: []string{"qbittorrent", "prowlarr"},
			},
			expected: true,
		},
		{
			name: "Service without dependencies",
			service: &Service{
				ID:           "jellyfin",
				Dependencies: []string{},
			},
			expected: false,
		},
		{
			name: "Service with nil dependencies",
			service: &Service{
				ID:           "bazarr",
				Dependencies: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.HasDependencies()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

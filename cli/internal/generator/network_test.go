package generator

import (
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestGetNetworkConfig(t *testing.T) {
	tests := []struct {
		name    string
		vpnMode bool
		wantNil bool
	}{
		{
			name:    "VPN mode - no custom network",
			vpnMode: true,
			wantNil: true,
		},
		{
			name:    "Bridge mode - has custom network",
			vpnMode: false,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetNetworkConfig(tt.vpnMode)
			
			if tt.wantNil {
				if config != nil {
					t.Error("Expected nil config for VPN mode")
				}
			} else {
				if config == nil {
					t.Fatal("Expected network config, got nil")
				}
				if config.Name != "media" {
					t.Errorf("Expected network name 'media', got '%s'", config.Name)
				}
				if config.Driver != "bridge" {
					t.Errorf("Expected driver 'bridge', got '%s'", config.Driver)
				}
			}
		})
	}
}

func TestConfigureServiceNetworking(t *testing.T) {
	tests := []struct {
		name        string
		service     *services.Service
		vpnMode     bool
		expectError bool
	}{
		{
			name: "VPN mode - compatible service",
			service: &services.Service{
				ID:          "radarr",
				SupportsVPN: true,
				Category:    services.CategoryMedia,
			},
			vpnMode:     true,
			expectError: false,
		},
		{
			name: "VPN mode - Gluetun service",
			service: &services.Service{
				ID:       "gluetun",
				Category: services.CategoryVPN,
			},
			vpnMode:     true,
			expectError: false,
		},
		{
			name: "VPN mode - service without explicit VPN support",
			service: &services.Service{
				ID:          "test",
				RequiresVPN: false,
				SupportsVPN: false,
				Category:    services.CategoryMedia,
			},
			vpnMode:     true,
			expectError: false, // Should pass because IsCompatibleWithVPN returns true for non-required services
		},
		{
			name: "Bridge mode - service with network config",
			service: &services.Service{
				ID: "radarr",
				Network: services.NetworkConfig{
					BridgeMode: services.BridgeModeConfig{
						Hostname: "radarr",
						Networks: []string{"media"},
					},
				},
			},
			vpnMode:     false,
			expectError: false,
		},
		{
			name: "Bridge mode - service without network config",
			service: &services.Service{
				ID: "test",
				Network: services.NetworkConfig{
					BridgeMode: services.BridgeModeConfig{
						Networks: []string{},
					},
				},
			},
			vpnMode:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigureServiceNetworking(tt.service, tt.vpnMode)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateNetworkConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		services    []*services.Service
		vpnMode     bool
		expectError bool
	}{
		{
			name: "Valid VPN configuration",
			services: []*services.Service{
				{
					ID:          "radarr",
					SupportsVPN: true,
					Category:    services.CategoryMedia,
				},
				{
					ID:          "sonarr",
					SupportsVPN: true,
					Category:    services.CategoryMedia,
				},
			},
			vpnMode:     true,
			expectError: false,
		},
		{
			name: "Valid bridge configuration",
			services: []*services.Service{
				{
					ID: "radarr",
					Network: services.NetworkConfig{
						BridgeMode: services.BridgeModeConfig{
							Networks: []string{"media"},
						},
					},
				},
			},
			vpnMode:     false,
			expectError: false,
		},
		{
			name: "Invalid VPN configuration - service requires VPN but VPN disabled",
			services: []*services.Service{
				{
					ID:          "flaresolverr",
					RequiresVPN: true,
					SupportsVPN: true,
					Category:    services.CategoryIndexer,
				},
			},
			vpnMode:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetworkConfiguration(tt.services, tt.vpnMode)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetExposedPorts(t *testing.T) {
	tests := []struct {
		name         string
		services     []*services.Service
		vpnMode      bool
		expectedLen  int
	}{
		{
			name: "VPN mode - collects all ports",
			services: []*services.Service{
				{
					ID:       "gluetun",
					Category: services.CategoryVPN,
					Ports: []services.PortMapping{
						{Host: "8000", Container: "8000"},
					},
				},
				{
					ID: "radarr",
					Ports: []services.PortMapping{
						{Host: "7878", Container: "7878"},
					},
				},
				{
					ID: "sonarr",
					Ports: []services.PortMapping{
						{Host: "8989", Container: "8989"},
					},
				},
			},
			vpnMode:     true,
			expectedLen: 2, // Excludes Gluetun ports
		},
		{
			name: "Bridge mode - returns nil",
			services: []*services.Service{
				{
					ID: "radarr",
					Ports: []services.PortMapping{
						{Host: "7878", Container: "7878"},
					},
				},
			},
			vpnMode:     false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := GetExposedPorts(tt.services, tt.vpnMode)
			
			if len(ports) != tt.expectedLen {
				t.Errorf("Expected %d ports, got %d", tt.expectedLen, len(ports))
			}
		})
	}
}

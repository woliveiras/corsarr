package generator

import (
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestNewComposeStrategy(t *testing.T) {
	tests := []struct {
		name        string
		vpnMode     bool
		expectType  string
	}{
		{
			name:       "VPN mode creates VPNModeStrategy",
			vpnMode:    true,
			expectType: "*generator.VPNModeStrategy",
		},
		{
			name:       "Bridge mode creates BridgeModeStrategy",
			vpnMode:    false,
			expectType: "*generator.BridgeModeStrategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewComposeStrategy(tt.vpnMode)
			if strategy == nil {
				t.Fatal("Strategy is nil")
			}
			
			// Check template path to verify strategy type
			path := strategy.GetTemplatePath()
			if tt.vpnMode && path != "templates/docker-compose/vpn-mode.tmpl" {
				t.Errorf("Expected VPN template path, got %s", path)
			}
			if !tt.vpnMode && path != "templates/docker-compose/bridge-mode.tmpl" {
				t.Errorf("Expected bridge template path, got %s", path)
			}
		})
	}
}

func TestVPNModeStrategy_GenerateCompose(t *testing.T) {
	strategy := &VPNModeStrategy{}

	t.Run("Success with Gluetun", func(t *testing.T) {
		services := []*services.Service{
			{
				ID:            "gluetun",
				Name:          "Gluetun",
				Category:      services.CategoryVPN,
				Image:         "qmcgaw/gluetun:latest",
				ContainerName: "gluetun",
				Restart:       "unless-stopped",
			},
			{
				ID:            "radarr",
				Name:          "Radarr",
				Category:      services.CategoryMedia,
				Image:         "lscr.io/linuxserver/radarr:latest",
				ContainerName: "radarr",
				Restart:       "unless-stopped",
				Network: services.NetworkConfig{
					VPNMode: services.VPNModeConfig{
						NetworkMode: "service:gluetun",
					},
				},
			},
		}

		content, err := strategy.GenerateCompose(services)
		if err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		if content == "" {
			t.Error("Generated content is empty")
		}
	})

	t.Run("Error without Gluetun", func(t *testing.T) {
		services := []*services.Service{
			{
				ID:       "radarr",
				Category: services.CategoryMedia,
			},
		}

		_, err := strategy.GenerateCompose(services)
		if err == nil {
			t.Error("Expected error when Gluetun is missing")
		}
	})
}

func TestBridgeModeStrategy_GenerateCompose(t *testing.T) {
	strategy := &BridgeModeStrategy{}

	services := []*services.Service{
		{
			ID:            "radarr",
			Name:          "Radarr",
			Category:      services.CategoryMedia,
			Image:         "lscr.io/linuxserver/radarr:latest",
			ContainerName: "radarr",
			Restart:       "unless-stopped",
			Network: services.NetworkConfig{
				BridgeMode: services.BridgeModeConfig{
					Hostname: "radarr",
					Networks: []string{"media"},
				},
			},
			Ports: []services.PortMapping{
				{Host: "7878", Container: "7878", Protocol: "tcp"},
			},
		},
	}

	content, err := strategy.GenerateCompose(services)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	if content == "" {
		t.Error("Generated content is empty")
	}
}

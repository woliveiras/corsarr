package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestNewComposeGenerator(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tmpDir := t.TempDir()
	generator := NewComposeGenerator(registry, tmpDir)

	if generator == nil {
		t.Fatal("Generator is nil")
	}
	if generator.registry == nil {
		t.Fatal("Generator registry is nil")
	}
	if generator.outputDir != tmpDir {
		t.Errorf("Expected outputDir %s, got %s", tmpDir, generator.outputDir)
	}
}

func TestComposeGenerator_Preview(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tmpDir := t.TempDir()
	generator := NewComposeGenerator(registry, tmpDir)

	tests := []struct {
		name        string
		serviceIDs  []string
		vpnMode     bool
		expectError bool
		checkFor    []string
	}{
		{
			name:        "Bridge mode - basic services",
			serviceIDs:  []string{"qbittorrent", "prowlarr", "radarr"},
			vpnMode:     false,
			expectError: false,
			checkFor:    []string{"services:", "qbittorrent:", "prowlarr:", "radarr:", "networks:", "media:"},
		},
		{
			name:        "VPN mode - includes Gluetun",
			serviceIDs:  []string{"qbittorrent", "prowlarr"},
			vpnMode:     true,
			expectError: false,
			checkFor:    []string{"services:", "gluetun:", "qbittorrent:", "network_mode:"},
		},
		{
			name:        "Invalid service",
			serviceIDs:  []string{"nonexistent"},
			vpnMode:     false,
			expectError: true,
			checkFor:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := generator.Preview(tt.serviceIDs, tt.vpnMode)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if content == "" {
				t.Error("Generated content is empty")
			}

			for _, check := range tt.checkFor {
				if !strings.Contains(content, check) {
					t.Errorf("Expected content to contain '%s'", check)
				}
			}

			t.Logf("Generated compose length: %d bytes", len(content))
		})
	}
}

func TestComposeGenerator_Generate(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tmpDir := t.TempDir()
	generator := NewComposeGenerator(registry, tmpDir)

	t.Run("Generate without backup", func(t *testing.T) {
		err := generator.Generate([]string{"jellyfin"}, false, false)
		if err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		// Check file exists
		outputPath := filepath.Join(tmpDir, "docker-compose.yml")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("docker-compose.yml was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		if !strings.Contains(string(content), "jellyfin:") {
			t.Error("Generated file doesn't contain jellyfin service")
		}
	})

	t.Run("Generate with backup", func(t *testing.T) {
		// First create an existing file
		existingPath := filepath.Join(tmpDir, "docker-compose.yml")
		err := os.WriteFile(existingPath, []byte("existing content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Generate with backup - include all dependencies
		err = generator.Generate([]string{"qbittorrent", "prowlarr", "radarr", "sonarr"}, false, true)
		if err != nil {
			t.Fatalf("Failed to generate with backup: %v", err)
		}

		// Check backup was created
		files, err := filepath.Glob(filepath.Join(tmpDir, "docker-compose.yml.backup.*"))
		if err != nil {
			t.Fatalf("Failed to list backup files: %v", err)
		}
		if len(files) == 0 {
			t.Error("No backup file was created")
		}
	})

	t.Run("Generate VPN mode adds Gluetun", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		generator2 := NewComposeGenerator(registry, tmpDir2)

		err := generator2.Generate([]string{"qbittorrent", "prowlarr", "radarr"}, true, false)
		if err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(tmpDir2, "docker-compose.yml"))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if !strings.Contains(string(content), "gluetun:") {
			t.Error("VPN mode should include Gluetun service")
		}
	})
}

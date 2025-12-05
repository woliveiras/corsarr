package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewEnvGenerator(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewEnvGenerator(tmpDir)

	if generator == nil {
		t.Fatal("Generator is nil")
	}
	if generator.outputDir != tmpDir {
		t.Errorf("Expected outputDir %s, got %s", tmpDir, generator.outputDir)
	}
}

func TestNewDefaultEnvConfig(t *testing.T) {
	config := NewDefaultEnvConfig()

	if config == nil {
		t.Fatal("Config is nil")
	}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"ComposeProjectName", config.ComposeProjectName, "corsarr"},
		{"ARRPath", config.ARRPath, "/opt/corsarr/"},
		{"Timezone", config.Timezone, "UTC"},
		{"PUID", config.PUID, "1000"},
		{"PGID", config.PGID, "1000"},
		{"UMASK", config.UMASK, "002"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.value)
			}
		})
	}

	if config.CustomEnv == nil {
		t.Error("CustomEnv should be initialized")
	}
}

func TestEnvGenerator_Preview(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewEnvGenerator(tmpDir)

	tests := []struct {
		name     string
		config   *EnvConfig
		checkFor []string
	}{
		{
			name:     "Basic config without VPN",
			config:   NewDefaultEnvConfig(),
			checkFor: []string{"COMPOSE_PROJECT_NAME=", "ARRPATH=", "TZ=", "PUID=", "PGID=", "UMASK="},
		},
		{
			name: "Config with VPN",
			config: &EnvConfig{
				ComposeProjectName: "test",
				ARRPath:            "/test/",
				Timezone:           "America/New_York",
				PUID:               "1001",
				PGID:               "1001",
				UMASK:              "022",
				VPNConfig: &VPNConfig{
					ServiceProvider: "mullvad",
					Type:            "wireguard",
					DNSAddress:      "1.1.1.1",
				},
			},
			checkFor: []string{
				"COMPOSE_PROJECT_NAME=test",
				"TZ=America/New_York",
				"VPN_SERVICE_PROVIDER=mullvad",
				"VPN_TYPE=wireguard",
				"DNS_ADDRESS=1.1.1.1",
			},
		},
		{
			name: "Config with custom env vars",
			config: &EnvConfig{
				ComposeProjectName: "custom",
				ARRPath:            "/custom/",
				Timezone:           "UTC",
				PUID:               "1000",
				PGID:               "1000",
				UMASK:              "002",
				CustomEnv: map[string]string{
					"CUSTOM_VAR1": "value1",
					"CUSTOM_VAR2": "value2",
				},
			},
			checkFor: []string{
				"COMPOSE_PROJECT_NAME=custom",
				"CUSTOM_VAR1=value1",
				"CUSTOM_VAR2=value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := generator.Preview(tt.config)
			if err != nil {
				t.Fatalf("Preview failed: %v", err)
			}

			if content == "" {
				t.Error("Generated content is empty")
			}

			for _, check := range tt.checkFor {
				if !strings.Contains(content, check) {
					t.Errorf("Expected content to contain '%s'", check)
				}
			}

			t.Logf("Generated env length: %d bytes", len(content))
		})
	}
}

func TestEnvGenerator_Generate(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewEnvGenerator(tmpDir)

	t.Run("Generate without backup", func(t *testing.T) {
		config := NewDefaultEnvConfig()
		err := generator.Generate(config, false)
		if err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		// Check file exists
		outputPath := filepath.Join(tmpDir, ".env")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error(".env was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		if !strings.Contains(string(content), "COMPOSE_PROJECT_NAME=corsarr") {
			t.Error("Generated file doesn't contain expected content")
		}
	})

	t.Run("Generate with backup", func(t *testing.T) {
		// First create an existing file
		existingPath := filepath.Join(tmpDir, ".env")
		err := os.WriteFile(existingPath, []byte("EXISTING=value"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Generate with backup
		config := NewDefaultEnvConfig()
		err = generator.Generate(config, true)
		if err != nil {
			t.Fatalf("Failed to generate with backup: %v", err)
		}

		// Check backup was created
		files, err := filepath.Glob(filepath.Join(tmpDir, ".env.backup.*"))
		if err != nil {
			t.Fatalf("Failed to list backup files: %v", err)
		}
		if len(files) == 0 {
			t.Error("No backup file was created")
		}

		// Verify backup contains old content
		backupContent, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatalf("Failed to read backup: %v", err)
		}
		if !strings.Contains(string(backupContent), "EXISTING=value") {
			t.Error("Backup doesn't contain original content")
		}
	})
}

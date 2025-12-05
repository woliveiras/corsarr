package generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

//go:embed templates/env.tmpl
var envTemplateFS embed.FS

// EnvGenerator handles .env file generation
type EnvGenerator struct {
	outputDir string
}

// EnvConfig holds all environment variables
type EnvConfig struct {
	ComposeProjectName string
	ARRPath            string
	Timezone           string
	PUID               string
	PGID               string
	UMASK              string
	VPNConfig          *VPNConfig
	CustomEnv          map[string]string
}

// VPNConfig holds VPN-specific configuration
type VPNConfig struct {
	ServiceProvider     string
	Type                string
	WireguardPrivateKey string
	WireguardPublicKey  string
	WireguardAddresses  string
	ServerCountries     string
	PortForwarding      string
	DNSAddress          string
}

// NewEnvGenerator creates a new env generator
func NewEnvGenerator(outputDir string) *EnvGenerator {
	return &EnvGenerator{
		outputDir: outputDir,
	}
}

// Generate creates a .env file based on the configuration
func (g *EnvGenerator) Generate(config *EnvConfig, backup bool) error {
	// Backup existing file if requested
	if backup {
		if err := g.backupExistingFile(); err != nil {
			return fmt.Errorf("failed to backup existing file: %w", err)
		}
	}

	// Generate env file
	content, err := g.renderTemplate(config)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Write file with secure permissions (0600 - owner read/write only)
	outputPath := filepath.Join(g.outputDir, ".env")
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// renderTemplate processes the template with the given data
func (g *EnvGenerator) renderTemplate(config *EnvConfig) (string, error) {
	// Read template file
	tmplContent, err := envTemplateFS.ReadFile("templates/env.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("env").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// backupExistingFile creates a backup of the existing .env
func (g *EnvGenerator) backupExistingFile() error {
	sourcePath := filepath.Join(g.outputDir, ".env")
	
	// Check if file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(g.outputDir, fmt.Sprintf(".env.backup.%s", timestamp))

	// Read original file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	// Write backup with secure permissions (0600 - owner read/write only)
	if err := os.WriteFile(backupPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// Preview generates the env content without writing to file
func (g *EnvGenerator) Preview(config *EnvConfig) (string, error) {
	return g.renderTemplate(config)
}

// NewDefaultEnvConfig creates a default environment configuration
func NewDefaultEnvConfig() *EnvConfig {
	return &EnvConfig{
		ComposeProjectName: "corsarr",
		ARRPath:            "/opt/corsarr/",
		Timezone:           "UTC",
		PUID:               "1000",
		PGID:               "1000",
		UMASK:              "002",
		CustomEnv:          make(map[string]string),
	}
}

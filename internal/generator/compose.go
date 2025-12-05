package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/woliveiras/corsarr/internal/services"
)

// ComposeGenerator handles docker-compose.yml generation
type ComposeGenerator struct {
	registry *services.Registry
	strategy ComposeStrategy
	outputDir string
}

// NewComposeGenerator creates a new compose generator
func NewComposeGenerator(registry *services.Registry, outputDir string) *ComposeGenerator {
	return &ComposeGenerator{
		registry:  registry,
		outputDir: outputDir,
	}
}

// SetStrategy sets the compose generation strategy
func (g *ComposeGenerator) SetStrategy(vpnMode bool) {
	g.strategy = NewComposeStrategy(vpnMode)
}

// Generate creates a docker-compose.yml file based on selected services
func (g *ComposeGenerator) Generate(serviceIDs []string, vpnMode bool, backup bool) error {
	// Set strategy based on VPN mode
	g.SetStrategy(vpnMode)

	// Load and prepare services
	selectedServices, err := g.prepareServices(serviceIDs, vpnMode)
	if err != nil {
		return err
	}

	// Validate dependencies
	if err := g.validateServices(selectedServices); err != nil {
		return err
	}

	// Backup existing file if requested
	if backup {
		if err := g.backupExistingFile(); err != nil {
			return fmt.Errorf("failed to backup existing file: %w", err)
		}
	}

	// Generate compose using strategy
	content, err := g.strategy.GenerateCompose(selectedServices)
	if err != nil {
		return fmt.Errorf("failed to generate compose: %w", err)
	}

	// Write file
	outputPath := filepath.Join(g.outputDir, "docker-compose.yml")
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// prepareServices loads services and adds Gluetun if needed
func (g *ComposeGenerator) prepareServices(serviceIDs []string, vpnMode bool) ([]*services.Service, error) {
	selectedServices := make([]*services.Service, 0, len(serviceIDs))
	
	for _, id := range serviceIDs {
		service, err := g.registry.GetService(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get service %s: %w", id, err)
		}
		selectedServices = append(selectedServices, service)
	}

	// Add Gluetun if VPN mode is enabled and not already in the list
	if vpnMode {
		hasGluetun := false
		for _, s := range selectedServices {
			if s.ID == "gluetun" {
				hasGluetun = true
				break
			}
		}
		if !hasGluetun {
			gluetun, err := g.registry.GetService("gluetun")
			if err != nil {
				return nil, fmt.Errorf("failed to get gluetun service: %w", err)
			}
			// Prepend Gluetun to the list
			selectedServices = append([]*services.Service{gluetun}, selectedServices...)
		}
	}

	return selectedServices, nil
}

// validateServices validates service dependencies
func (g *ComposeGenerator) validateServices(selectedServices []*services.Service) error {
	serviceIDs := make([]string, len(selectedServices))
	for i, s := range selectedServices {
		serviceIDs[i] = s.ID
	}
	
	if err := g.registry.ValidateDependencies(serviceIDs); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}
	
	return nil
}



// backupExistingFile creates a backup of the existing docker-compose.yml
func (g *ComposeGenerator) backupExistingFile() error {
	sourcePath := filepath.Join(g.outputDir, "docker-compose.yml")
	
	// Check if file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(g.outputDir, fmt.Sprintf("docker-compose.yml.backup.%s", timestamp))

	// Read original file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// Preview generates the compose content without writing to file
func (g *ComposeGenerator) Preview(serviceIDs []string, vpnMode bool) (string, error) {
	// Set strategy based on VPN mode
	g.SetStrategy(vpnMode)

	// Load and prepare services
	selectedServices, err := g.prepareServices(serviceIDs, vpnMode)
	if err != nil {
		return "", err
	}

	// Generate using strategy
	return g.strategy.GenerateCompose(selectedServices)
}

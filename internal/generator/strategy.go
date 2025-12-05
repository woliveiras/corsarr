package generator

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/woliveiras/corsarr/internal/services"
)

//go:embed templates/docker-compose/*.tmpl
var templatesFS embed.FS

// ComposeStrategy defines the interface for compose generation strategies
type ComposeStrategy interface {
	GenerateCompose(selectedServices []*services.Service) (string, error)
	GetTemplatePath() string
}

// VPNModeStrategy generates compose for VPN mode
type VPNModeStrategy struct{}

// BridgeModeStrategy generates compose for bridge mode
type BridgeModeStrategy struct{}

// VPNComposeData holds data for VPN mode template
type VPNComposeData struct {
	Services     []*services.Service
	Gluetun      *services.Service
	ExposedPorts []services.PortMapping
}

// BridgeComposeData holds data for bridge mode template
type BridgeComposeData struct {
	Services []*services.Service
}

// NewComposeStrategy creates the appropriate strategy based on VPN mode
func NewComposeStrategy(vpnMode bool) ComposeStrategy {
	if vpnMode {
		return &VPNModeStrategy{}
	}
	return &BridgeModeStrategy{}
}

// GenerateCompose implements ComposeStrategy for VPN mode
func (s *VPNModeStrategy) GenerateCompose(selectedServices []*services.Service) (string, error) {
	// Separate Gluetun from other services
	var gluetun *services.Service
	var otherServices []*services.Service

	for _, svc := range selectedServices {
		if svc.Category == services.CategoryVPN {
			gluetun = svc
		} else {
			otherServices = append(otherServices, svc)
		}
	}

	if gluetun == nil {
		return "", fmt.Errorf("gluetun service not found in VPN mode")
	}

	// Get exposed ports for Gluetun
	exposedPorts := GetExposedPorts(selectedServices, true)

	data := VPNComposeData{
		Services:     otherServices,
		Gluetun:      gluetun,
		ExposedPorts: exposedPorts,
	}

	return renderTemplate(s.GetTemplatePath(), "compose-vpn", data)
}

// GetTemplatePath returns the template path for VPN mode
func (s *VPNModeStrategy) GetTemplatePath() string {
	return "templates/docker-compose/vpn-mode.tmpl"
}

// GenerateCompose implements ComposeStrategy for bridge mode
func (s *BridgeModeStrategy) GenerateCompose(selectedServices []*services.Service) (string, error) {
	data := BridgeComposeData{
		Services: selectedServices,
	}

	return renderTemplate(s.GetTemplatePath(), "compose-bridge", data)
}

// GetTemplatePath returns the template path for bridge mode
func (s *BridgeModeStrategy) GetTemplatePath() string {
	return "templates/docker-compose/bridge-mode.tmpl"
}

// renderTemplate is a helper function to render templates
func renderTemplate(templatePath, templateName string, data interface{}) (string, error) {
	tmplContent, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

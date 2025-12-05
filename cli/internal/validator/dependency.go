package validator

import (
	"fmt"
	"strings"
)

// DependencyValidator checks service dependencies
type DependencyValidator struct {
	config *Config
}

// NewDependencyValidator creates a new dependency validator
func NewDependencyValidator(config *Config) *DependencyValidator {
	return &DependencyValidator{config: config}
}

// Validate checks if all service dependencies are satisfied
func (dv *DependencyValidator) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Build a set of selected service IDs for quick lookup
	selectedIDs := make(map[string]bool)
	for _, service := range dv.config.Services {
		selectedIDs[service.ID] = true
	}

	// Check each service's dependencies
	for _, service := range dv.config.Services {
		for _, depID := range service.Dependencies {
			if !selectedIDs[depID] {
				// Dependency not selected - get the dependency service name
				depService, err := dv.config.Registry.GetService(depID)
				depName := depID
				if err == nil {
					depName = depService.Name
				}

				result.AddError(
					"dependencies",
					fmt.Sprintf("Service '%s' requires '%s' but it is not selected", service.Name, depName),
					SeverityError,
				)
			}
		}

		// Check VPN requirements
		if service.RequiresVPN && !dv.config.VPNEnabled {
			result.AddError(
				"vpn",
				fmt.Sprintf("Service '%s' requires VPN but VPN is not enabled", service.Name),
				SeverityError,
			)
		}
	}

	// If VPN is enabled, ensure Gluetun is included
	if dv.config.VPNEnabled && !selectedIDs["gluetun"] {
		result.AddError(
			"vpn",
			"VPN mode is enabled but Gluetun service is not included",
			SeverityCritical,
		)
	}

	return result
}

// GetMissingDependencies returns a map of services to their missing dependencies
func GetMissingDependencies(config *Config) map[string][]string {
	missing := make(map[string][]string)
	
	selectedIDs := make(map[string]bool)
	for _, service := range config.Services {
		selectedIDs[service.ID] = true
	}

	for _, service := range config.Services {
		missingDeps := []string{}
		
		for _, depID := range service.Dependencies {
			if !selectedIDs[depID] {
				depService, err := config.Registry.GetService(depID)
				depName := depID
				if err == nil {
					depName = depService.Name
				}
				missingDeps = append(missingDeps, depName)
			}
		}

		if len(missingDeps) > 0 {
			missing[service.Name] = missingDeps
		}
	}

	return missing
}

// SuggestDependencies returns services that should be added to satisfy dependencies
func SuggestDependencies(config *Config) []string {
	suggestions := make(map[string]bool)
	
	selectedIDs := make(map[string]bool)
	for _, service := range config.Services {
		selectedIDs[service.ID] = true
	}

	for _, service := range config.Services {
		for _, depID := range service.Dependencies {
			if !selectedIDs[depID] {
				suggestions[depID] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(suggestions))
	for id := range suggestions {
		service, err := config.Registry.GetService(id)
		if err == nil {
			result = append(result, service.Name)
		}
	}

	return result
}

// FormatDependencyError creates a user-friendly error message for dependency issues
func FormatDependencyError(serviceName string, missingDeps []string) string {
	if len(missingDeps) == 0 {
		return ""
	}

	if len(missingDeps) == 1 {
		return fmt.Sprintf("Service '%s' requires '%s'", serviceName, missingDeps[0])
	}

	return fmt.Sprintf("Service '%s' requires: %s", serviceName, strings.Join(missingDeps, ", "))
}

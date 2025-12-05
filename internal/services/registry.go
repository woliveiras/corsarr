package services

import (
	"embed"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed templates/services/*.yaml
var servicesFS embed.FS

// Registry manages all available services
type Registry struct {
	services map[string]*Service
	byCategory map[ServiceCategory][]*Service
}

// NewRegistry creates a new service registry
func NewRegistry() (*Registry, error) {
	registry := &Registry{
		services: make(map[string]*Service),
		byCategory: make(map[ServiceCategory][]*Service),
	}

	if err := registry.loadServices(); err != nil {
		return nil, err
	}

	return registry, nil
}

// loadServices loads all service definitions from embedded YAML files
func (r *Registry) loadServices() error {
	// List of service definition files
	serviceFiles := []string{
		"qbittorrent.yaml",
		"prowlarr.yaml",
		"flaresolverr.yaml",
		"sonarr.yaml",
		"radarr.yaml",
		"lidarr.yaml",
		"lazylibrarian.yaml",
		"bazarr.yaml",
		"jellyfin.yaml",
		"jellyseerr.yaml",
		"fileflows.yaml",
		"gluetun.yaml",
	}

	for _, filename := range serviceFiles {
		data, err := servicesFS.ReadFile(fmt.Sprintf("templates/services/%s", filename))
		if err != nil {
			// Service file might not exist yet, skip
			continue
		}

		var service Service
		if err := yaml.Unmarshal(data, &service); err != nil {
			return fmt.Errorf("failed to parse service file %s: %w", filename, err)
		}

		r.services[service.ID] = &service
		r.byCategory[service.Category] = append(r.byCategory[service.Category], &service)
	}

	// Sort services by name within each category
	for category := range r.byCategory {
		sort.Slice(r.byCategory[category], func(i, j int) bool {
			return r.byCategory[category][i].Name < r.byCategory[category][j].Name
		})
	}

	return nil
}

// GetService returns a service by ID
func (r *Registry) GetService(id string) (*Service, error) {
	service, exists := r.services[id]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", id)
	}
	return service, nil
}

// GetAllServices returns all available services
func (r *Registry) GetAllServices() []*Service {
	services := make([]*Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	
	// Sort by category and name
	sort.Slice(services, func(i, j int) bool {
		if services[i].Category == services[j].Category {
			return services[i].Name < services[j].Name
		}
		return services[i].Category < services[j].Category
	})
	
	return services
}

// GetServicesByCategory returns all services in a specific category
func (r *Registry) GetServicesByCategory(category ServiceCategory) []*Service {
	return r.byCategory[category]
}

// GetServicesByIDs returns services matching the provided IDs
func (r *Registry) GetServicesByIDs(ids []string) ([]*Service, error) {
	services := make([]*Service, 0, len(ids))
	
	for _, id := range ids {
		service, err := r.GetService(id)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	
	return services, nil
}

// ValidateDependencies checks if all dependencies are satisfied
func (r *Registry) ValidateDependencies(selectedIDs []string) error {
	selectedMap := make(map[string]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}

	for _, id := range selectedIDs {
		service, err := r.GetService(id)
		if err != nil {
			return err
		}

		for _, depID := range service.Dependencies {
			if !selectedMap[depID] {
				depService, _ := r.GetService(depID)
				depName := depID
				if depService != nil {
					depName = depService.Name
				}
				return fmt.Errorf("service '%s' requires '%s' but it is not selected", service.Name, depName)
			}
		}
	}

	return nil
}

// FilterByVPNCompatibility filters services based on VPN mode
func (r *Registry) FilterByVPNCompatibility(vpnEnabled bool) []*Service {
	filtered := make([]*Service, 0)
	
	for _, service := range r.services {
		// Skip VPN service itself from the list
		if service.Category == CategoryVPN {
			continue
		}
		
		if service.IsCompatibleWithVPN(vpnEnabled) {
			filtered = append(filtered, service)
		}
	}
	
	return filtered
}

// GetServiceCount returns the total number of services
func (r *Registry) GetServiceCount() int {
	return len(r.services)
}

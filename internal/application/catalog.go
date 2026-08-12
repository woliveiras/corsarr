package application

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"

	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/services"
)

const localApplicationHost = "127.0.0.1"

var recommendedApplicationIDs = []string{
	"bazarr",
	"jellyfin",
	"jellyseerr",
	"prowlarr",
	"qbittorrent",
	"radarr",
	"sonarr",
}

// Desktop exposes every application with a safe local Web UI for lifecycle
// management. Only this allowlist can also be newly selected for one-click
// installation because each entry has a bounded, idempotent provisioning
// adapter.
var desktopApplicationIDs = map[string]struct{}{
	"bazarr": {}, "jellyfin": {}, "jellyseerr": {}, "lazylibrarian": {},
	"lidarr": {}, "prowlarr": {}, "qbittorrent": {}, "radarr": {}, "sonarr": {},
}

// ApplicationSummary is the runtime-independent application data exposed to
// presentation layers such as Corsarr Desktop.
type ApplicationSummary struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	URL            string   `json:"url"`
	Optional       bool     `json:"optional"`
	AutomatedSetup bool     `json:"automatedSetup"`
	Dependencies   []string `json:"dependencies"`
}

func (c *Catalog) InstallationOrder(applicationIDs []string) ([]string, error) {
	selected := make(map[string]struct{}, len(applicationIDs))
	for _, id := range applicationIDs {
		if _, exists := c.byID[id]; !exists {
			return nil, fmt.Errorf("application is not available in the desktop catalog: %s", id)
		}
		selected[id] = struct{}{}
	}
	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	ordered := make([]string, 0, len(ids))
	visiting := make(map[string]bool, len(ids))
	visited := make(map[string]bool, len(ids))
	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("application dependency cycle includes %s", id)
		}
		visiting[id] = true
		application := c.byID[id]
		for _, dependencyID := range application.Dependencies {
			if _, included := selected[dependencyID]; !included {
				continue
			}
			if err := visit(dependencyID); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		ordered = append(ordered, id)
		return nil
	}
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// Catalog exposes only user-facing applications from the service registry.
type Catalog struct {
	applications []ApplicationSummary
	byID         map[string]ApplicationSummary
}

// NewCatalog creates a desktop-safe view of the existing service registry.
func NewCatalog(registry *services.Registry) *Catalog {
	return newCatalog(registry, nil)
}

// NewLocalizedCatalog creates the same runtime-safe catalog while translating
// user-facing service names and descriptions through Corsarr's embedded locale.
func NewLocalizedCatalog(registry *services.Registry, translator *i18n.I18n) *Catalog {
	return newCatalog(registry, translator)
}

func newCatalog(registry *services.Registry, translator *i18n.I18n) *Catalog {
	applications := make([]ApplicationSummary, 0)
	byID := make(map[string]ApplicationSummary)

	for _, service := range registry.GetAllServices() {
		if service.WebUI == nil {
			continue
		}

		applicationURL, ok := localApplicationURL(service.WebUI.Port)
		if !ok {
			continue
		}

		name := service.Name
		description := service.Description
		if translator != nil {
			if translated := translator.T(service.GetNameKey()); translated != service.GetNameKey() {
				name = translated
			}
			if translated := translator.T(service.GetDescriptionKey()); translated != service.GetDescriptionKey() {
				description = translated
			}
		}

		_, automatedSetup := desktopApplicationIDs[service.ID]
		summary := ApplicationSummary{
			ID:             service.ID,
			Name:           name,
			Description:    description,
			Category:       string(service.Category),
			URL:            applicationURL,
			Optional:       service.Optional,
			AutomatedSetup: automatedSetup,
			Dependencies:   append([]string{}, service.Dependencies...),
		}

		applications = append(applications, summary)
		byID[summary.ID] = summary
	}

	return &Catalog{applications: applications, byID: byID}
}

// ListApplications returns a copy so callers cannot mutate catalog state.
func (c *Catalog) ListApplications() []ApplicationSummary {
	applications := make([]ApplicationSummary, len(c.applications))
	copy(applications, c.applications)
	return applications
}

// RecommendedApplicationIDs returns the reviewed starter stack for people who
// want movies and TV shows without choosing every supporting application.
func (c *Catalog) RecommendedApplicationIDs() ([]string, error) {
	applicationIDs := append([]string(nil), recommendedApplicationIDs...)
	if _, err := c.InstallationOrder(applicationIDs); err != nil {
		return nil, fmt.Errorf("recommended desktop applications are unavailable: %w", err)
	}

	return applicationIDs, nil
}

// ResolveApplicationURL returns an allowlisted local URL for a catalog ID.
func (c *Catalog) ResolveApplicationURL(id string) (string, error) {
	application, ok := c.byID[id]
	if !ok {
		return "", fmt.Errorf("application is not available in the desktop catalog: %s", id)
	}

	return application.URL, nil
}

func localApplicationURL(portValue string) (string, bool) {
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return "", false
	}

	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(localApplicationHost, portValue),
	}).String(), true
}

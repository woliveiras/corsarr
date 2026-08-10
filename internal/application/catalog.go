package application

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/woliveiras/corsarr/internal/services"
)

const localApplicationHost = "127.0.0.1"

// ApplicationSummary is the runtime-independent application data exposed to
// presentation layers such as Corsarr Desktop.
type ApplicationSummary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	URL          string   `json:"url"`
	Optional     bool     `json:"optional"`
	Dependencies []string `json:"dependencies"`
}

// Catalog exposes only user-facing applications from the service registry.
type Catalog struct {
	applications []ApplicationSummary
	byID         map[string]ApplicationSummary
}

// NewCatalog creates a desktop-safe view of the existing service registry.
func NewCatalog(registry *services.Registry) *Catalog {
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

		summary := ApplicationSummary{
			ID:           service.ID,
			Name:         service.Name,
			Description:  service.Description,
			Category:     string(service.Category),
			URL:          applicationURL,
			Optional:     service.Optional,
			Dependencies: append([]string(nil), service.Dependencies...),
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

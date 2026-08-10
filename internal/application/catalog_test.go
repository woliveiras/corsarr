package application

import (
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestCatalogListsOnlyApplicationsWithWebInterfaces(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	catalog := NewCatalog(registry)
	applications := catalog.ListApplications()

	radarr, ok := findApplication(applications, "radarr")
	if !ok {
		t.Fatal("expected Radarr in desktop catalog")
	}
	if radarr.URL != "http://127.0.0.1:7878" {
		t.Fatalf("expected safe local Radarr URL, got %q", radarr.URL)
	}
	if radarr.Category != "media" {
		t.Fatalf("expected Radarr category media, got %q", radarr.Category)
	}

	for _, infrastructureID := range []string{"flaresolverr", "gluetun"} {
		if _, ok := findApplication(applications, infrastructureID); ok {
			t.Fatalf("did not expect infrastructure service %q in desktop catalog", infrastructureID)
		}
	}
}

func TestCatalogResolvesOnlyKnownApplicationURLs(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	catalog := NewCatalog(registry)

	url, err := catalog.ResolveApplicationURL("jellyfin")
	if err != nil {
		t.Fatalf("resolve Jellyfin URL: %v", err)
	}
	if url != "http://127.0.0.1:8096" {
		t.Fatalf("expected safe local Jellyfin URL, got %q", url)
	}

	if _, err := catalog.ResolveApplicationURL("https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary URL to be rejected")
	}
	if _, err := catalog.ResolveApplicationURL("gluetun"); err == nil {
		t.Fatal("expected non-UI infrastructure service to be rejected")
	}
}

func TestCatalogOrdersDependenciesBeforeSelectedApplications(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	catalog := NewCatalog(registry)

	ordered, err := catalog.InstallationOrder([]string{"radarr", "prowlarr", "qbittorrent"})
	if err != nil {
		t.Fatalf("order applications: %v", err)
	}
	positions := make(map[string]int, len(ordered))
	for index, id := range ordered {
		positions[id] = index
	}
	if positions["qbittorrent"] > positions["radarr"] || positions["prowlarr"] > positions["radarr"] {
		t.Fatalf("expected dependencies before radarr, got %v", ordered)
	}
}

func TestCatalogRequiresArrApplicationsBeforeBazarr(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	catalog := NewCatalog(registry)

	selected := []string{"bazarr", "radarr", "sonarr", "prowlarr", "qbittorrent"}
	ordered, err := catalog.InstallationOrder(selected)
	if err != nil {
		t.Fatalf("order applications: %v", err)
	}
	positions := make(map[string]int, len(ordered))
	for index, id := range ordered {
		positions[id] = index
	}
	if positions["radarr"] > positions["bazarr"] || positions["sonarr"] > positions["bazarr"] {
		t.Fatalf("expected Radarr and Sonarr before Bazarr, got %v", ordered)
	}
	if _, err := catalog.InstallationOrder([]string{"bazarr"}); err == nil {
		t.Fatal("expected Bazarr without Arr dependencies to be rejected")
	}
}

func findApplication(applications []ApplicationSummary, id string) (ApplicationSummary, bool) {
	for _, application := range applications {
		if application.ID == id {
			return application, true
		}
	}

	return ApplicationSummary{}, false
}

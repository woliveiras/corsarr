package application

import (
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/i18n"
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
	fileflows, ok := findApplication(applications, "fileflows")
	if !ok || fileflows.AutomatedSetup {
		t.Fatal("expected FileFlows lifecycle access without one-click automated setup")
	}
	lazyLibrarian, ok := findApplication(applications, "lazylibrarian")
	if !ok || !lazyLibrarian.AutomatedSetup {
		t.Fatal("expected LazyLibrarian automated setup in the desktop catalog")
	}
}

func TestCatalogAlwaysExposesDependenciesAsAnArray(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	for _, application := range NewCatalog(registry).ListApplications() {
		if application.Dependencies == nil {
			t.Fatalf("application %q exposes null dependencies", application.ID)
		}
	}
}

func TestLocalizedCatalogUsesEmbeddedDesktopTranslations(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	translator, err := i18n.New("pt-br")
	if err != nil {
		t.Fatalf("create translator: %v", err)
	}

	catalog := NewLocalizedCatalog(registry, translator)
	radarr, ok := findApplication(catalog.ListApplications(), "radarr")
	if !ok {
		t.Fatal("expected Radarr in localized catalog")
	}
	if radarr.Description != "buscador e gerenciador de filmes" {
		t.Fatalf("unexpected localized description %q", radarr.Description)
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

func TestCatalogOrdersSelectedIntegrationsBeforeConsumers(t *testing.T) {
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
		t.Fatalf("expected selected integrations before radarr, got %v", ordered)
	}
}

func TestCatalogAllowsAConsumerWithoutManagedIntegrations(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	catalog := NewCatalog(registry)

	ordered, err := catalog.InstallationOrder([]string{"sonarr"})
	if err != nil {
		t.Fatalf("order standalone Sonarr: %v", err)
	}
	if !reflect.DeepEqual(ordered, []string{"sonarr"}) {
		t.Fatalf("expected only selected Sonarr, got %v", ordered)
	}
}

func TestCatalogOrdersSelectedARRApplicationsBeforeBazarr(t *testing.T) {
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
	standalone, err := catalog.InstallationOrder([]string{"bazarr"})
	if err != nil || !reflect.DeepEqual(standalone, []string{"bazarr"}) {
		t.Fatalf("expected standalone Bazarr to remain selected, got %v, err=%v", standalone, err)
	}
}

func TestCatalogRecommendedApplicationsFormCompleteStarterStack(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	catalog := NewCatalog(registry)

	recommended, err := catalog.RecommendedApplicationIDs()
	if err != nil {
		t.Fatalf("resolve recommended applications: %v", err)
	}
	want := []string{
		"bazarr",
		"jellyfin",
		"jellyseerr",
		"prowlarr",
		"qbittorrent",
		"radarr",
		"sonarr",
	}
	if !reflect.DeepEqual(recommended, want) {
		t.Fatalf("expected reviewed starter stack %v, got %v", want, recommended)
	}
	if _, err := catalog.InstallationOrder(recommended); err != nil {
		t.Fatalf("recommended stack has incomplete dependencies: %v", err)
	}

	recommended[0] = "mutated"
	again, err := catalog.RecommendedApplicationIDs()
	if err != nil {
		t.Fatalf("resolve recommended applications again: %v", err)
	}
	if again[0] != "bazarr" {
		t.Fatalf("caller mutated catalog preset: %v", again)
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

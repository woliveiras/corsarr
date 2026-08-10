package legal

import (
	"testing"

	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestCatalogHasCompleteNoticeForEveryDesktopApplication(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	runtimeCatalog, err := runtimecatalog.NewRuntimeCatalog(registry)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog(registry, runtimeCatalog)
	if err != nil {
		t.Fatalf("create legal catalog: %v", err)
	}

	applicationCount := 0
	for _, notice := range catalog.ListNotices() {
		if notice.ComponentType != ComponentApplication {
			continue
		}
		applicationCount++
		if notice.Name == "" || notice.Purpose == "" || notice.License == "" ||
			notice.ImageMaintainer == "" || notice.ApprovedImage == "" || len(notice.Links) < 3 {
			t.Fatalf("incomplete legal notice %#v", notice)
		}
	}
	if applicationCount != 10 {
		t.Fatalf("expected all 10 desktop applications, got %d", applicationCount)
	}
}

func TestCatalogResolvesOnlyKnownHTTPSLinks(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	runtimeCatalog, err := runtimecatalog.NewRuntimeCatalog(registry)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog(registry, runtimeCatalog)
	if err != nil {
		t.Fatal(err)
	}

	link, err := catalog.ResolveLink("radarr", LinkLicense)
	if err != nil || link != "https://github.com/Radarr/Radarr/blob/develop/LICENSE" {
		t.Fatalf("unexpected license link %q, %v", link, err)
	}
	if _, err := catalog.ResolveLink("radarr", "https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary link kind to be rejected")
	}
	if _, err := catalog.ResolveLink("unknown", LinkOfficial); err == nil {
		t.Fatal("expected unknown component to be rejected")
	}
}

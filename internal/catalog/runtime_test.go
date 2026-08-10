package catalog

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestRuntimeCatalogCoversEveryDesktopApplicationWithImmutableSpec(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create service registry: %v", err)
	}
	catalog, err := NewRuntimeCatalog(registry)
	if err != nil {
		t.Fatalf("create runtime catalog: %v", err)
	}
	root := filepath.Join(t.TempDir(), "Corsarr")

	for _, service := range registry.GetAllServices() {
		if service.WebUI == nil {
			continue
		}
		t.Run(service.ID, func(t *testing.T) {
			spec, resolveErr := catalog.Resolve(service.ID, root, RuntimeOptions{
				Timezone: "Europe/Madrid",
				PUID:     1000,
				PGID:     1000,
			})
			if resolveErr != nil {
				t.Fatalf("resolve runtime spec: %v", resolveErr)
			}
			if err := spec.Validate(); err != nil {
				t.Fatalf("validate resolved spec: %v", err)
			}
			if strings.Contains(spec.Image, ":latest") || !strings.Contains(spec.Image, "@sha256:") {
				t.Fatalf("expected immutable image, got %q", spec.Image)
			}
			for _, port := range spec.Ports {
				if port.Exposure != runtime.ExposureLoopback {
					t.Fatalf("expected safe loopback default, got %#v", port)
				}
			}
		})
	}
}

func TestRuntimeCatalogRejectsUnknownApplication(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create service registry: %v", err)
	}
	catalog, err := NewRuntimeCatalog(registry)
	if err != nil {
		t.Fatalf("create runtime catalog: %v", err)
	}

	_, err = catalog.Resolve("unknown", t.TempDir(), RuntimeOptions{})
	if err == nil {
		t.Fatal("expected unknown application to be rejected")
	}
}

package catalog

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/services"
)

const verifyRemoteManifestsEnvironment = "CORSARR_VERIFY_REMOTE_MANIFESTS"

// TestRuntimeCatalogRemotePlatformContract queries only the immutable OCI
// indexes already approved by the catalog. It does not pull image layers.
func TestRuntimeCatalogRemotePlatformContract(t *testing.T) {
	if os.Getenv(verifyRemoteManifestsEnvironment) != "1" {
		t.Skip("set CORSARR_VERIFY_REMOTE_MANIFESTS=1 to query approved OCI indexes")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatalf("find Docker client: %v", err)
	}

	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create service registry: %v", err)
	}
	runtimeCatalog, err := NewRuntimeCatalog(registry)
	if err != nil {
		t.Fatalf("create runtime catalog: %v", err)
	}

	for _, service := range registry.GetAllServices() {
		if service.WebUI == nil {
			continue
		}
		t.Run(service.ID, func(t *testing.T) {
			image, imageErr := runtimeCatalog.ApprovedImage(service.ID)
			if imageErr != nil {
				t.Fatalf("resolve approved image: %v", imageErr)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			output, commandErr := exec.CommandContext(
				ctx,
				"docker",
				"buildx",
				"imagetools",
				"inspect",
				"--raw",
				image,
			).Output()
			if commandErr != nil {
				t.Fatalf("inspect immutable OCI index: %v", commandErr)
			}
			var index struct {
				Manifests []struct {
					Platform struct {
						OS           string `json:"os"`
						Architecture string `json:"architecture"`
					} `json:"platform"`
				} `json:"manifests"`
			}
			if err := json.Unmarshal(output, &index); err != nil {
				t.Fatalf("decode OCI index: %v", err)
			}
			platforms := make(map[string]bool)
			for _, manifest := range index.Manifests {
				platforms[manifest.Platform.OS+"/"+manifest.Platform.Architecture] = true
			}
			for _, required := range []string{"linux/amd64", "linux/arm64"} {
				if !platforms[required] {
					t.Fatalf("approved OCI index is missing %s", required)
				}
			}
		})
	}
}

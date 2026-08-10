package runtime

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

const dockerContractImageEnvironment = "CORSARR_DOCKER_CONTRACT_IMAGE"

// TestDockerManagerRealContract is opt-in because it creates temporary Docker
// resources on the developer's current context. The immutable fixture image
// must already exist locally; the test never invokes Pull.
func TestDockerManagerRealContract(t *testing.T) {
	image := strings.TrimSpace(os.Getenv(dockerContractImageEnvironment))
	if image == "" {
		t.Skip("set CORSARR_DOCKER_CONTRACT_IMAGE to an immutable image already present locally")
	}
	if err := validateImageReference(image); err != nil {
		t.Fatalf("validate local contract image: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runner := OSCommandRunner{}
	dockerPath, err := runner.LookPath("docker")
	if err != nil {
		t.Fatalf("find Docker client: %v", err)
	}
	if _, err := runner.Run(ctx, dockerPath, "image", "inspect", image); err != nil {
		t.Fatalf("contract image must already exist locally: %v", err)
	}

	networkExisted := dockerContractNetworkExisted(t, ctx, runner, dockerPath)
	manager := NewDockerManager(runner, 30*time.Second)
	const applicationID = "contract-test"
	if _, err := manager.Inspect(ctx, applicationID); !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("contract container name is not clean: %v", err)
	}
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, inspectErr := manager.Inspect(cleanupContext, applicationID); inspectErr == nil {
			if removeErr := manager.Remove(cleanupContext, applicationID); removeErr != nil {
				t.Errorf("remove contract container during cleanup: %v", removeErr)
			}
		}
		if !networkExisted {
			dockerContractRemoveOwnedNetwork(t, cleanupContext, runner, dockerPath)
		}
	})

	if err := manager.EnsureNetwork(ctx); err != nil {
		t.Fatalf("ensure owned network: %v", err)
	}
	spec := ContainerSpec{
		ApplicationID: applicationID,
		Image:         image,
		Environment:   map[string]string{"TZ": "Etc/UTC"},
	}
	if err := manager.Create(ctx, spec); err != nil {
		t.Fatalf("create contract container: %v", err)
	}
	created, err := manager.Inspect(ctx, applicationID)
	if err != nil || created.State != ContainerStateCreated || created.Image != image {
		t.Fatalf("inspect created contract container: status=%#v err=%v", created, err)
	}
	if err := manager.Start(ctx, applicationID); err != nil {
		t.Fatalf("start contract container: %v", err)
	}
	running, err := manager.Inspect(ctx, applicationID)
	if err != nil || running.State != ContainerStateRunning {
		t.Fatalf("inspect running contract container: status=%#v err=%v", running, err)
	}
	if _, err := manager.Logs(ctx, applicationID, 5); err != nil {
		t.Fatalf("read bounded owned logs: %v", err)
	}
	if err := manager.Restart(ctx, applicationID); err != nil {
		t.Fatalf("restart contract container: %v", err)
	}
	if err := manager.Stop(ctx, applicationID); err != nil {
		t.Fatalf("stop contract container: %v", err)
	}
	stopped, err := manager.Inspect(ctx, applicationID)
	if err != nil || stopped.State != ContainerStateStopped {
		t.Fatalf("inspect stopped contract container: status=%#v err=%v", stopped, err)
	}
	if err := manager.Remove(ctx, applicationID); err != nil {
		t.Fatalf("remove contract container: %v", err)
	}
	if _, err := manager.Inspect(ctx, applicationID); !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("verify contract container removal: %v", err)
	}
}

func dockerContractNetworkExisted(
	t *testing.T,
	ctx context.Context,
	runner OSCommandRunner,
	dockerPath string,
) bool {
	t.Helper()
	output, err := runner.Run(
		ctx,
		dockerPath,
		"network", "inspect", CorsarrNetworkName, "--format", networkOwnershipFormat,
	)
	if err == nil {
		if strings.TrimSpace(output) != managedLabelValue {
			t.Fatalf("existing Corsarr network is not owned by Corsarr")
		}
		return true
	}
	if !indicatesMissingResource(err.Error(), "network") {
		t.Fatalf("inspect contract network: %v", err)
	}
	return false
}

func dockerContractRemoveOwnedNetwork(
	t *testing.T,
	ctx context.Context,
	runner OSCommandRunner,
	dockerPath string,
) {
	t.Helper()
	output, err := runner.Run(
		ctx,
		dockerPath,
		"network", "inspect", CorsarrNetworkName, "--format", networkOwnershipFormat,
	)
	if err != nil {
		if !indicatesMissingResource(err.Error(), "network") {
			t.Errorf("inspect contract network during cleanup: %v", err)
		}
		return
	}
	if strings.TrimSpace(output) != managedLabelValue {
		t.Errorf("refuse to remove contract network after ownership changed")
		return
	}
	if _, err := runner.Run(ctx, dockerPath, "network", "rm", CorsarrNetworkName); err != nil {
		t.Errorf("remove contract network during cleanup: %v", err)
	}
}

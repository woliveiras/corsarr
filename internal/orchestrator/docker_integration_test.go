package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/provisioning"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

const dockerContractImageEnvironment = "CORSARR_DOCKER_CONTRACT_IMAGE"
const dockerRollbackImageEnvironment = "CORSARR_DOCKER_ROLLBACK_IMAGE"

// TestInstallerRealDockerContract verifies the transactional install path with
// an immutable image already present on the current Docker context. It replaces
// Pull with a local-image assertion, publishes one ephemeral loopback port, and
// removes all resources it creates.
func TestInstallerRealDockerContract(t *testing.T) {
	image := strings.TrimSpace(os.Getenv(dockerContractImageEnvironment))
	if image == "" {
		t.Skip("set CORSARR_DOCKER_CONTRACT_IMAGE to an immutable image already present locally")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runner := containerruntime.OSCommandRunner{}
	dockerPath, err := runner.LookPath("docker")
	if err != nil {
		t.Fatalf("find Docker client: %v", err)
	}
	if _, err := runner.Run(ctx, dockerPath, "image", "inspect", image); err != nil {
		t.Fatalf("contract image must already exist locally: %v", err)
	}

	hostPort := availableLoopbackPort(t)
	mountPath := t.TempDir()
	manager := containerruntime.NewDockerManager(runner, 30*time.Second)
	runtime := &locallySeededRuntime{Manager: manager, image: image}
	const applicationID = "installer-contract"
	if _, err := manager.Inspect(ctx, applicationID); !errors.Is(err, containerruntime.ErrResourceNotFound) {
		t.Fatalf("contract container name is not clean: %v", err)
	}
	networkExisted := installerContractNetworkExisted(t, ctx, runner, dockerPath)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, inspectErr := manager.Inspect(cleanupContext, applicationID); inspectErr == nil {
			if removeErr := manager.Remove(cleanupContext, applicationID); removeErr != nil {
				t.Errorf("remove installer contract container during cleanup: %v", removeErr)
			}
		}
		if !networkExisted {
			installerContractRemoveOwnedNetwork(t, cleanupContext, runner, dockerPath)
		}
	})

	resolver := contractSpecResolver{spec: containerruntime.ContainerSpec{
		ApplicationID: applicationID,
		Image:         image,
		Ports: []containerruntime.PortBinding{{
			HostPort: hostPort, ContainerPort: 8025,
			Protocol: containerruntime.ProtocolTCP, Exposure: containerruntime.ExposureLoopback,
		}},
		Mounts: []containerruntime.BindMount{{
			HostPath: mountPath, ContainerPath: "/corsarr-contract", ReadOnly: true,
		}},
		Environment: map[string]string{"TZ": "Etc/UTC"},
	}}
	readiness := provisioning.NewHTTPReadiness(
		contractEndpointResolver{url: fmt.Sprintf("http://127.0.0.1:%d", hostPort)},
		30*time.Second,
		100*time.Millisecond,
	)
	installer := NewInstaller(runtime, resolver, readiness)

	status, err := installer.Install(
		ctx,
		applicationID,
		t.TempDir(),
		catalog.RuntimeOptions{},
	)
	if err != nil || status.State != containerruntime.ContainerStateRunning {
		t.Fatalf("install through real Docker contract: status=%#v err=%v", status, err)
	}
	if runtime.pullCalls != 1 {
		t.Fatalf("expected one locally verified pull boundary, got %d", runtime.pullCalls)
	}
	verifyInstallerContractMount(t, ctx, runner, dockerPath)

	status, err = installer.Install(
		ctx,
		applicationID,
		t.TempDir(),
		catalog.RuntimeOptions{},
	)
	if err != nil || status.State != containerruntime.ContainerStateRunning {
		t.Fatalf("reconcile existing real container: status=%#v err=%v", status, err)
	}
	if runtime.pullCalls != 1 {
		t.Fatalf("idempotent reconcile crossed pull boundary again: %d", runtime.pullCalls)
	}

	if err := manager.Remove(ctx, applicationID); err != nil {
		t.Fatalf("remove installer contract container: %v", err)
	}
}

func TestUpdaterRealDockerRollbackContract(t *testing.T) {
	previousImage := strings.TrimSpace(os.Getenv(dockerContractImageEnvironment))
	approvedImage := strings.TrimSpace(os.Getenv(dockerRollbackImageEnvironment))
	if previousImage == "" || approvedImage == "" {
		t.Skip("set both Docker contract image variables to immutable images already present locally")
	}
	if previousImage == approvedImage {
		t.Fatal("rollback contract requires two different immutable image references")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runner := containerruntime.OSCommandRunner{}
	dockerPath, err := runner.LookPath("docker")
	if err != nil {
		t.Fatalf("find Docker client: %v", err)
	}
	for _, image := range []string{previousImage, approvedImage} {
		if _, err := runner.Run(ctx, dockerPath, "image", "inspect", image); err != nil {
			t.Fatalf("contract image must already exist locally: %v", err)
		}
	}

	manager := containerruntime.NewDockerManager(runner, 30*time.Second)
	runtime := &locallySeededRuntime{Manager: manager, image: approvedImage}
	const applicationID = "updater-contract"
	if _, err := manager.Inspect(ctx, applicationID); !errors.Is(err, containerruntime.ErrResourceNotFound) {
		t.Fatalf("updater contract container name is not clean: %v", err)
	}
	networkExisted := installerContractNetworkExisted(t, ctx, runner, dockerPath)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, inspectErr := manager.Inspect(cleanupContext, applicationID); inspectErr == nil {
			if removeErr := manager.Remove(cleanupContext, applicationID); removeErr != nil {
				t.Errorf("remove updater contract container during cleanup: %v", removeErr)
			}
		}
		if !networkExisted {
			installerContractRemoveOwnedNetwork(t, cleanupContext, runner, dockerPath)
		}
	})

	if err := manager.EnsureNetwork(ctx); err != nil {
		t.Fatalf("ensure updater contract network: %v", err)
	}
	previousSpec := containerruntime.ContainerSpec{
		ApplicationID: applicationID,
		Image:         previousImage,
		Environment:   map[string]string{"TZ": "Etc/UTC"},
	}
	if err := manager.Create(ctx, previousSpec); err != nil {
		t.Fatalf("create previous contract container: %v", err)
	}
	if err := manager.Start(ctx, applicationID); err != nil {
		t.Fatalf("start previous contract container: %v", err)
	}

	rootPath := t.TempDir()
	configPath := rootPath + "/config/" + applicationID
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		t.Fatalf("create updater contract config: %v", err)
	}
	if err := os.WriteFile(configPath+"/settings.json", []byte("{}"), 0o600); err != nil {
		t.Fatalf("write updater contract config: %v", err)
	}
	approvedSpec := previousSpec
	approvedSpec.Image = approvedImage
	resolver := contractSpecResolver{spec: approvedSpec}
	failOnce := &failFirstContractReadiness{}
	updater := NewUpdater(runtime, resolver, failOnce, storage.NewBackupManager())

	rolledBack, updateErr := updater.Update(
		ctx,
		applicationID,
		rootPath,
		catalog.RuntimeOptions{},
	)
	if updateErr == nil || !rolledBack.RolledBack || rolledBack.Updated {
		t.Fatalf("expected verified rollback, result=%#v err=%v", rolledBack, updateErr)
	}
	restored, err := manager.Inspect(ctx, applicationID)
	if err != nil || restored.State != containerruntime.ContainerStateRunning ||
		restored.Image != previousImage {
		t.Fatalf("previous container was not restored: status=%#v err=%v", restored, err)
	}

	updater = NewUpdater(runtime, resolver, alwaysReadyContract{}, storage.NewBackupManager())
	updated, err := updater.Update(ctx, applicationID, rootPath, catalog.RuntimeOptions{})
	if err != nil || !updated.Updated || updated.RolledBack {
		t.Fatalf("expected successful real update, result=%#v err=%v", updated, err)
	}
	current, err := manager.Inspect(ctx, applicationID)
	if err != nil || current.State != containerruntime.ContainerStateRunning ||
		current.Image != approvedImage {
		t.Fatalf("approved container is not running: status=%#v err=%v", current, err)
	}
	if err := manager.Remove(ctx, applicationID); err != nil {
		t.Fatalf("remove updater contract container: %v", err)
	}
}

func verifyInstallerContractMount(
	t *testing.T,
	ctx context.Context,
	runner containerruntime.OSCommandRunner,
	dockerPath string,
) {
	t.Helper()
	output, err := runner.Run(
		ctx,
		dockerPath,
		"inspect", "corsarr-installer-contract", "--format", `{{json .Mounts}}`,
	)
	if err != nil {
		t.Fatalf("inspect installer contract mounts: %v", err)
	}
	var mounts []struct {
		Type        string `json:"Type"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	}
	if err := json.Unmarshal([]byte(output), &mounts); err != nil {
		t.Fatalf("decode installer contract mounts: %v", err)
	}
	for _, mount := range mounts {
		if mount.Type == "bind" && mount.Source != "" &&
			mount.Destination == "/corsarr-contract" && !mount.RW {
			return
		}
	}
	t.Fatalf("expected read-only host bind mount, got %#v", mounts)
}

type locallySeededRuntime struct {
	containerruntime.Manager
	image     string
	pullCalls int
}

func (r *locallySeededRuntime) Pull(_ context.Context, image string) error {
	r.pullCalls++
	if image != r.image {
		return fmt.Errorf("unexpected contract image %q", image)
	}
	return nil
}

type contractSpecResolver struct {
	spec containerruntime.ContainerSpec
}

func (r contractSpecResolver) Resolve(
	string,
	string,
	catalog.RuntimeOptions,
) (containerruntime.ContainerSpec, error) {
	return r.spec, nil
}

type contractEndpointResolver struct {
	url string
}

type failFirstContractReadiness struct {
	calls int
}

func (r *failFirstContractReadiness) Wait(context.Context, string) error {
	r.calls++
	if r.calls == 1 {
		return errors.New("intentional contract health failure")
	}
	return nil
}

type alwaysReadyContract struct{}

func (alwaysReadyContract) Wait(context.Context, string) error { return nil }

func (r contractEndpointResolver) ResolveApplicationURL(string) (string, error) {
	return r.url, nil
}

func availableLoopbackPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve loopback port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("release loopback port: %v", err)
	}
	return port
}

func installerContractNetworkExisted(
	t *testing.T,
	ctx context.Context,
	runner containerruntime.OSCommandRunner,
	dockerPath string,
) bool {
	t.Helper()
	output, err := runner.Run(
		ctx,
		dockerPath,
		"network", "inspect", containerruntime.CorsarrNetworkName,
		"--format", `{{index .Labels "io.corsarr.managed"}}`,
	)
	if err == nil {
		if strings.TrimSpace(output) != "true" {
			t.Fatalf("existing Corsarr network is not owned by Corsarr")
		}
		return true
	}
	if !installerContractNetworkMissing(err) {
		t.Fatalf("inspect installer contract network: %v", err)
	}
	return false
}

func installerContractRemoveOwnedNetwork(
	t *testing.T,
	ctx context.Context,
	runner containerruntime.OSCommandRunner,
	dockerPath string,
) {
	t.Helper()
	output, err := runner.Run(
		ctx,
		dockerPath,
		"network", "inspect", containerruntime.CorsarrNetworkName,
		"--format", `{{index .Labels "io.corsarr.managed"}}`,
	)
	if err != nil {
		if !installerContractNetworkMissing(err) {
			t.Errorf("inspect installer contract network during cleanup: %v", err)
		}
		return
	}
	if strings.TrimSpace(output) != "true" {
		t.Errorf("refuse to remove installer contract network after ownership changed")
		return
	}
	if _, err := runner.Run(
		ctx,
		dockerPath,
		"network", "rm", containerruntime.CorsarrNetworkName,
	); err != nil {
		t.Errorf("remove installer contract network during cleanup: %v", err)
	}
}

func installerContractNetworkMissing(err error) bool {
	detail := strings.ToLower(err.Error())
	return strings.Contains(detail, "no such network") ||
		(strings.Contains(detail, "network") && strings.Contains(detail, "not found"))
}

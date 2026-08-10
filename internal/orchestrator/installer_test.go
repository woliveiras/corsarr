package orchestrator

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

func TestInstallerCompletesOwnedContainerLifecycle(t *testing.T) {
	manager := &fakeRuntimeManager{inspectStatus: containerruntime.ContainerStatus{
		ApplicationID: "radarr", State: containerruntime.ContainerStateRunning,
	}}
	resolver := &fakeSpecResolver{spec: validInstallerSpec("radarr")}
	installer := NewInstaller(manager, resolver)

	status, err := installer.Install(
		context.Background(),
		"radarr",
		"/Users/test/Media/Corsarr",
		catalog.RuntimeOptions{Timezone: "Europe/Madrid", PUID: 1000, PGID: 1000},
	)
	if err != nil {
		t.Fatalf("install application: %v", err)
	}
	if status.State != containerruntime.ContainerStateRunning {
		t.Fatalf("expected running container, got %#v", status)
	}
	want := []string{"network", "pull", "create", "start", "inspect"}
	if !reflect.DeepEqual(manager.operations, want) {
		t.Fatalf("unexpected install operations\nwant: %v\n got: %v", want, manager.operations)
	}
}

func TestInstallerRemovesIncompleteContainerAfterStartFailure(t *testing.T) {
	manager := &fakeRuntimeManager{startErr: errors.New("runtime stopped")}
	installer := NewInstaller(manager, &fakeSpecResolver{spec: validInstallerSpec("radarr")})

	_, err := installer.Install(
		context.Background(),
		"radarr",
		"/Users/test/Media/Corsarr",
		catalog.RuntimeOptions{},
	)
	if err == nil {
		t.Fatal("expected start failure")
	}
	want := []string{"network", "pull", "create", "start", "remove"}
	if !reflect.DeepEqual(manager.operations, want) {
		t.Fatalf("expected incomplete container cleanup\nwant: %v\n got: %v", want, manager.operations)
	}
}

func validInstallerSpec(applicationID string) containerruntime.ContainerSpec {
	return containerruntime.ContainerSpec{
		ApplicationID: applicationID,
		Image:         "example.invalid/app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
}

type fakeSpecResolver struct {
	spec containerruntime.ContainerSpec
	err  error
}

func (r *fakeSpecResolver) Resolve(
	string,
	string,
	catalog.RuntimeOptions,
) (containerruntime.ContainerSpec, error) {
	return r.spec, r.err
}

type fakeRuntimeManager struct {
	operations    []string
	startErr      error
	inspectStatus containerruntime.ContainerStatus
}

func (m *fakeRuntimeManager) EnsureNetwork(context.Context) error {
	m.operations = append(m.operations, "network")
	return nil
}

func (m *fakeRuntimeManager) Pull(context.Context, string) error {
	m.operations = append(m.operations, "pull")
	return nil
}

func (m *fakeRuntimeManager) Create(context.Context, containerruntime.ContainerSpec) error {
	m.operations = append(m.operations, "create")
	return nil
}

func (m *fakeRuntimeManager) Start(context.Context, string) error {
	m.operations = append(m.operations, "start")
	return m.startErr
}

func (m *fakeRuntimeManager) Stop(context.Context, string) error { return nil }

func (m *fakeRuntimeManager) Restart(context.Context, string) error { return nil }

func (m *fakeRuntimeManager) Remove(context.Context, string) error {
	m.operations = append(m.operations, "remove")
	return nil
}

func (m *fakeRuntimeManager) Inspect(context.Context, string) (containerruntime.ContainerStatus, error) {
	m.operations = append(m.operations, "inspect")
	return m.inspectStatus, nil
}

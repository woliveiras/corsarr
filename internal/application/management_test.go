package application

import (
	"context"
	"testing"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestManagementServiceReportsInstalledAndMissingApplications(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{statuses: map[string]containerruntime.ContainerStatus{
		"radarr": {ApplicationID: "radarr", State: containerruntime.ContainerStateRunning},
	}}
	service := NewManagementService(NewCatalog(registry), runtime)

	statuses := service.ListStatuses(context.Background())
	if findManagedStatus(statuses, "radarr").State != ManagedStateRunning {
		t.Fatalf("expected Radarr running, got %#v", findManagedStatus(statuses, "radarr"))
	}
	if findManagedStatus(statuses, "jellyfin").State != ManagedStateNotInstalled {
		t.Fatalf("expected Jellyfin not installed, got %#v", findManagedStatus(statuses, "jellyfin"))
	}
}

func TestManagementServiceRejectsUnknownLifecycleTarget(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{}
	service := NewManagementService(NewCatalog(registry), runtime)

	if err := service.Restart(context.Background(), "attacker-controlled"); err == nil {
		t.Fatal("expected unknown application to be rejected")
	}
	if runtime.lastOperation != "" {
		t.Fatalf("expected no runtime operation, got %q", runtime.lastOperation)
	}
}

type managementRuntime struct {
	statuses      map[string]containerruntime.ContainerStatus
	lastOperation string
}

func (m *managementRuntime) EnsureNetwork(context.Context) error { return nil }
func (m *managementRuntime) Pull(context.Context, string) error  { return nil }
func (m *managementRuntime) Create(context.Context, containerruntime.ContainerSpec) error {
	return nil
}
func (m *managementRuntime) Inspect(_ context.Context, id string) (containerruntime.ContainerStatus, error) {
	status, exists := m.statuses[id]
	if !exists {
		return containerruntime.ContainerStatus{}, containerruntime.ErrResourceNotFound
	}
	return status, nil
}
func (m *managementRuntime) Start(context.Context, string) error {
	m.lastOperation = "start"
	return nil
}
func (m *managementRuntime) Stop(context.Context, string) error {
	m.lastOperation = "stop"
	return nil
}
func (m *managementRuntime) Restart(context.Context, string) error {
	m.lastOperation = "restart"
	return nil
}
func (m *managementRuntime) Remove(context.Context, string) error {
	m.lastOperation = "remove"
	return nil
}
func (m *managementRuntime) Logs(context.Context, string, int) (string, error) {
	return "", nil
}

func findManagedStatus(statuses []ManagedApplicationStatus, id string) ManagedApplicationStatus {
	for _, status := range statuses {
		if status.ApplicationID == id {
			return status
		}
	}
	return ManagedApplicationStatus{}
}

package application

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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

func TestManagementServiceReportsApprovedImageUpdate(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{statuses: map[string]containerruntime.ContainerStatus{
		"radarr": {ApplicationID: "radarr", State: containerruntime.ContainerStateRunning, Image: "old-image"},
	}}
	service := NewManagementService(NewCatalog(registry), runtime, approvedImageStub{image: "approved-image"})

	status := findManagedStatus(service.ListStatuses(context.Background()), "radarr")
	if !status.UpdateAvailable || status.ApprovedImage != "approved-image" {
		t.Fatalf("expected available approved update, got %#v", status)
	}
}

func TestManagementServiceKeepsRuntimeFailureOutOfDesktopPayload(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{errors: map[string]error{
		"radarr": errors.New("private socket /Users/test/runtime.sock failed"),
	}}
	service := NewManagementService(NewCatalog(registry), runtime)

	status := findManagedStatus(service.ListStatuses(context.Background()), "radarr")
	if status.State != ManagedStateAttention || status.TechnicalDetail == "" {
		t.Fatalf("expected internal runtime detail, got %#v", status)
	}
	if status.Issue == nil || status.Issue.Code != "application_status_unavailable" {
		t.Fatalf("expected actionable status issue, got %#v", status.Issue)
	}
	payload, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("encode status: %v", err)
	}
	if strings.Contains(string(payload), "runtime.sock") || strings.Contains(string(payload), "technicalDetail") {
		t.Fatalf("desktop status exposed raw runtime detail: %s", payload)
	}
	if !strings.Contains(string(payload), "application_status_unavailable") {
		t.Fatalf("desktop status omitted bounded issue: %s", payload)
	}
}

func TestManagementServiceBlocksRemovalRequiredByInstalledApplications(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{statuses: map[string]containerruntime.ContainerStatus{
		"qbittorrent": {ApplicationID: "qbittorrent", State: containerruntime.ContainerStateRunning},
		"radarr":      {ApplicationID: "radarr", State: containerruntime.ContainerStateRunning},
		"sonarr":      {ApplicationID: "sonarr", State: containerruntime.ContainerStateStopped},
	}}
	service := NewManagementService(NewCatalog(registry), runtime)

	status := findManagedStatus(service.ListStatuses(context.Background()), "qbittorrent")
	want := []string{"radarr", "sonarr"}
	if !reflect.DeepEqual(status.RemovalBlockedBy, want) {
		t.Fatalf("expected removal blockers %v, got %#v", want, status)
	}
	if err := service.Remove(context.Background(), "qbittorrent"); !errors.Is(err, ErrApplicationRequired) {
		t.Fatalf("expected dependency-aware removal rejection, got %v", err)
	}
	if runtime.lastOperation != "" {
		t.Fatalf("blocked removal reached runtime: %q", runtime.lastOperation)
	}
}

func TestManagementServiceAllowsLeafApplicationRemoval(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{statuses: map[string]containerruntime.ContainerStatus{
		"qbittorrent": {ApplicationID: "qbittorrent", State: containerruntime.ContainerStateRunning},
		"radarr":      {ApplicationID: "radarr", State: containerruntime.ContainerStateRunning},
	}}
	service := NewManagementService(NewCatalog(registry), runtime)

	if err := service.Remove(context.Background(), "radarr"); err != nil {
		t.Fatalf("remove leaf application: %v", err)
	}
	if runtime.lastOperation != "remove" {
		t.Fatalf("expected runtime removal, got %q", runtime.lastOperation)
	}
}

type approvedImageStub struct {
	image string
	err   error
}

func (s approvedImageStub) ApprovedImage(string) (string, error) { return s.image, s.err }

type managementRuntime struct {
	statuses      map[string]containerruntime.ContainerStatus
	errors        map[string]error
	lastOperation string
}

func (m *managementRuntime) EnsureNetwork(context.Context) error { return nil }
func (m *managementRuntime) Pull(context.Context, string) error  { return nil }
func (m *managementRuntime) Create(context.Context, containerruntime.ContainerSpec) error {
	return nil
}
func (m *managementRuntime) Inspect(_ context.Context, id string) (containerruntime.ContainerStatus, error) {
	if err := m.errors[id]; err != nil {
		return containerruntime.ContainerStatus{}, err
	}
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

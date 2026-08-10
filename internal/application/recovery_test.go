package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestRecoveryServiceStartsOnlyExistingStoppedContainersInDependencyOrder(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &recoverySetup{status: SetupStatus{
		Applications: []string{"radarr", "prowlarr", "qbittorrent"},
		StartAtLogin: true, StartAtLoginSupported: true,
	}}
	runtime := &recoveryRuntime{states: map[string]containerruntime.ContainerState{
		"prowlarr":    containerruntime.ContainerStateStopped,
		"qbittorrent": containerruntime.ContainerStateRunning,
		"radarr":      containerruntime.ContainerStateStopped,
	}}
	service := NewRecoveryService(setup, NewCatalog(registry), runtime)

	result, err := service.Recover(context.Background())
	if err != nil {
		t.Fatalf("recover applications: %v", err)
	}
	if !result.Complete {
		t.Fatalf("expected complete recovery %#v", result)
	}
	wantInspected := []string{"prowlarr", "qbittorrent", "radarr"}
	if !reflect.DeepEqual(runtime.inspected, wantInspected) {
		t.Fatalf("unexpected inspection order\nwant: %v\n got: %v", wantInspected, runtime.inspected)
	}
	if !reflect.DeepEqual(runtime.started, []string{"prowlarr", "radarr"}) {
		t.Fatalf("unexpected starts %v", runtime.started)
	}
}

func TestRecoveryServiceSkipsMissingContainersWithoutInstalling(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &recoverySetup{status: SetupStatus{
		Applications: []string{"jellyfin"}, StartAtLogin: true, StartAtLoginSupported: true,
	}}
	runtime := &recoveryRuntime{inspectErrors: map[string]error{
		"jellyfin": containerruntime.ErrResourceNotFound,
	}}
	service := NewRecoveryService(setup, NewCatalog(registry), runtime)

	result, err := service.Recover(context.Background())
	if err != nil || !result.Complete || len(runtime.started) != 0 {
		t.Fatalf("unexpected missing-container recovery %#v, %v, starts=%v", result, err, runtime.started)
	}
	if len(result.Items) != 1 || !result.Items[0].Skipped {
		t.Fatalf("expected skipped item, got %#v", result.Items)
	}
}

func TestRecoveryServiceRequiresEnabledApprovedPreference(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	for _, status := range []SetupStatus{
		{},
		{StartAtLogin: true, StartAtLoginSupported: true, StartAtLoginRequiresApproval: true},
	} {
		runtime := &recoveryRuntime{}
		service := NewRecoveryService(&recoverySetup{status: status}, NewCatalog(registry), runtime)
		if _, err := service.Recover(context.Background()); !errors.Is(err, ErrBackgroundRecoveryDisabled) {
			t.Fatalf("expected disabled recovery for %#v, got %v", status, err)
		}
		if len(runtime.inspected) != 0 {
			t.Fatalf("disabled recovery accessed runtime: %v", runtime.inspected)
		}
	}
}

type recoverySetup struct {
	status SetupStatus
	err    error
}

func (s *recoverySetup) Load() (SetupStatus, error) { return s.status, s.err }

type recoveryRuntime struct {
	states        map[string]containerruntime.ContainerState
	inspectErrors map[string]error
	startErrors   map[string]error
	inspected     []string
	started       []string
}

func (r *recoveryRuntime) Inspect(
	_ context.Context,
	applicationID string,
) (containerruntime.ContainerStatus, error) {
	r.inspected = append(r.inspected, applicationID)
	if err := r.inspectErrors[applicationID]; err != nil {
		return containerruntime.ContainerStatus{}, err
	}
	return containerruntime.ContainerStatus{
		ApplicationID: applicationID,
		State:         r.states[applicationID],
	}, nil
}

func (r *recoveryRuntime) Start(_ context.Context, applicationID string) error {
	r.started = append(r.started, applicationID)
	return r.startErrors[applicationID]
}

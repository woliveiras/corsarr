package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestConfigurationReconcilerProvisionsOnlyRunningSelectedApplications(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &recoverySetup{status: SetupStatus{
		StoragePath: "/media", Applications: []string{"radarr", "prowlarr", "qbittorrent"},
		OnboardingCompleted: true,
	}}
	runtime := &recoveryRuntime{states: map[string]containerruntime.ContainerState{
		"prowlarr":    containerruntime.ContainerStateRunning,
		"radarr":      containerruntime.ContainerStateStopped,
		"qbittorrent": containerruntime.ContainerStateRunning,
	}}
	provisioner := &reconciliationProvisioner{}
	service := NewConfigurationReconciler(setup, NewCatalog(registry), runtime, provisioner)

	result, err := service.Reconcile(context.Background())
	if err != nil || !result.Complete {
		t.Fatalf("reconcile configuration: result=%#v err=%v", result, err)
	}
	want := []string{"prowlarr", "qbittorrent"}
	if !reflect.DeepEqual(provisioner.applications, want) {
		t.Fatalf("unexpected reconciled applications\nwant: %v\n got: %v", want, provisioner.applications)
	}
	if provisioner.rootPath != "/media/Corsarr" ||
		!reflect.DeepEqual(provisioner.selected, setup.status.Applications) {
		t.Fatalf("unexpected reviewed setup %#v", provisioner)
	}
}

func TestConfigurationReconcilerContinuesAfterOneApplicationFails(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &recoverySetup{status: SetupStatus{
		StoragePath: "/media", Applications: []string{"radarr", "prowlarr"},
		OnboardingCompleted: true,
	}}
	runtime := &recoveryRuntime{states: map[string]containerruntime.ContainerState{
		"prowlarr": containerruntime.ContainerStateRunning,
		"radarr":   containerruntime.ContainerStateRunning,
	}}
	provisioner := &reconciliationProvisioner{errors: map[string]error{
		"prowlarr": errors.New("private detail"),
	}}
	service := NewConfigurationReconciler(setup, NewCatalog(registry), runtime, provisioner)

	result, err := service.Reconcile(context.Background())
	if err != nil {
		t.Fatalf("reconcile with item failure: %v", err)
	}
	if result.Complete || len(result.Items) != 2 || result.Items[0].Failed != true {
		t.Fatalf("expected bounded partial result %#v", result)
	}
	if !reflect.DeepEqual(provisioner.applications, []string{"prowlarr", "radarr"}) {
		t.Fatalf("expected reconciliation to continue, got %v", provisioner.applications)
	}
}

func TestConfigurationReconcilerRequiresCompletedOnboarding(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	provisioner := &reconciliationProvisioner{}
	service := NewConfigurationReconciler(
		&recoverySetup{status: SetupStatus{StoragePath: "/media", Applications: []string{"radarr"}}},
		NewCatalog(registry),
		&recoveryRuntime{},
		provisioner,
	)

	if _, err := service.Reconcile(context.Background()); !errors.Is(err, ErrConfigurationReconciliationDisabled) {
		t.Fatalf("expected incomplete onboarding rejection, got %v", err)
	}
	if len(provisioner.applications) != 0 {
		t.Fatal("incomplete onboarding reached provisioning")
	}
}

type reconciliationProvisioner struct {
	applications []string
	rootPath     string
	selected     []string
	errors       map[string]error
}

func (p *reconciliationProvisioner) Provision(
	_ context.Context,
	rootPath string,
	applicationID string,
	selected []string,
) error {
	p.applications = append(p.applications, applicationID)
	p.rootPath = rootPath
	p.selected = append([]string(nil), selected...)
	return p.errors[applicationID]
}

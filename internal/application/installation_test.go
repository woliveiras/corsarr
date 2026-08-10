package application

import (
	"context"
	"reflect"
	"testing"

	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
	"github.com/woliveiras/corsarr/internal/storage"
)

func TestInstallationServiceRequiresBackendConsent(t *testing.T) {
	service := &InstallationService{setup: &installationSetup{status: SetupStatus{
		StoragePath: "/Users/test/Media", Applications: []string{"jellyfin"}, CanPrepare: true,
	}}}

	_, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err == nil {
		t.Fatal("expected installation without consent to be rejected")
	}
}

func TestInstallationServicePreparesAndInstallsInDependencyOrder(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &installationSetup{status: SetupStatus{
		StoragePath: "/Users/test/Media", Applications: []string{"radarr", "qbittorrent", "prowlarr"},
		CanPrepare: true, CanInstall: true, TermsAccepted: true,
	}}
	layout := &installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Media/Corsarr"}}
	installer := &recordingInstaller{}
	provisioner := &recordingProvisioner{}
	service := NewInstallationService(setup, layout, NewCatalog(registry), installer, provisioner)

	result, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("install selected applications: %v", err)
	}
	if !result.Complete {
		t.Fatalf("expected complete result, got %#v", result)
	}
	want := []string{"prowlarr", "qbittorrent", "radarr"}
	if !reflect.DeepEqual(installer.applicationIDs, want) {
		t.Fatalf("unexpected installation order\nwant: %v\n got: %v", want, installer.applicationIDs)
	}
	if !reflect.DeepEqual(provisioner.applicationIDs, want) {
		t.Fatalf("unexpected provisioning order\nwant: %v\n got: %v", want, provisioner.applicationIDs)
	}
}

type installationSetup struct{ status SetupStatus }

func (s *installationSetup) Load() (SetupStatus, error) { return s.status, nil }

type installationLayout struct{ status storage.LayoutStatus }

func (l *installationLayout) Prepare(string, []string) (storage.LayoutStatus, error) {
	return l.status, nil
}

type recordingInstaller struct{ applicationIDs []string }

type recordingProvisioner struct{ applicationIDs []string }

func (p *recordingProvisioner) Provision(
	_ context.Context,
	_ string,
	applicationID string,
) error {
	p.applicationIDs = append(p.applicationIDs, applicationID)
	return nil
}

func (i *recordingInstaller) Install(
	_ context.Context,
	applicationID string,
	_ string,
	_ runtimecatalog.RuntimeOptions,
) (containerruntime.ContainerStatus, error) {
	i.applicationIDs = append(i.applicationIDs, applicationID)
	return containerruntime.ContainerStatus{
		ApplicationID: applicationID,
		State:         containerruntime.ContainerStateRunning,
	}, nil
}

package application

import (
	"context"
	"errors"
	"reflect"
	"strings"
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

func TestInstallationServiceReportsBoundedProgress(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &installationSetup{status: SetupStatus{
		StoragePath: "/Users/test/Media", Applications: []string{"jellyfin"},
		CanPrepare: true, CanInstall: true, TermsAccepted: true,
	}}
	service := NewInstallationService(
		setup,
		&installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Media/Corsarr"}},
		NewCatalog(registry),
		&recordingInstaller{},
		&recordingProvisioner{},
	)
	var progress []InstallationProgress

	result, err := service.InstallSelectedWithProgress(
		context.Background(),
		runtimecatalog.RuntimeOptions{},
		func(event InstallationProgress) { progress = append(progress, event) },
	)
	if err != nil || !result.Complete {
		t.Fatalf("install with progress: result=%#v err=%v", result, err)
	}
	want := []InstallationProgress{
		{ApplicationID: "jellyfin", Stage: InstallationStageInstalling, Position: 1, Total: 1},
		{ApplicationID: "jellyfin", Stage: InstallationStageProvisioning, Position: 1, Total: 1},
		{ApplicationID: "jellyfin", Stage: InstallationStageReady, Position: 1, Total: 1},
	}
	if !reflect.DeepEqual(progress, want) {
		t.Fatalf("unexpected progress\nwant: %#v\n got: %#v", want, progress)
	}
}

func TestInstallationServiceSkipsLegacyApplicationWithoutAutomatedSetup(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	setup := &installationSetup{status: SetupStatus{
		StoragePath: "/Users/test/Media", Applications: []string{"fileflows", "jellyfin"},
		CanPrepare: true, CanInstall: true, TermsAccepted: true,
	}}
	installer := &recordingInstaller{}
	provisioner := &recordingProvisioner{}
	service := NewInstallationService(
		setup,
		&installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Media/Corsarr"}},
		NewCatalog(registry),
		installer,
		provisioner,
	)

	result, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err != nil || !result.Complete {
		t.Fatalf("install automated subset: result=%#v err=%v", result, err)
	}
	if !reflect.DeepEqual(installer.applicationIDs, []string{"jellyfin"}) ||
		!reflect.DeepEqual(provisioner.applicationIDs, []string{"jellyfin"}) {
		t.Fatalf(
			"application without automated setup entered installation: installer=%v provisioner=%v",
			installer.applicationIDs,
			provisioner.applicationIDs,
		)
	}
}

func TestInstallationServiceReturnsOnlyBoundedFailureStateToDesktop(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	service := NewInstallationService(
		&installationSetup{status: SetupStatus{
			StoragePath: "/Users/test/Media", Applications: []string{"jellyfin"},
			CanPrepare: true, CanInstall: true, TermsAccepted: true,
		}},
		&installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Media/Corsarr"}},
		NewCatalog(registry),
		&recordingInstaller{err: errors.New("private runtime detail")},
		&recordingProvisioner{},
	)

	result, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("return structured install failure: %v", err)
	}
	if result.Complete || len(result.Items) != 1 || !result.Items[0].Failed || result.Items[0].Error == "" {
		t.Fatalf("unexpected bounded install failure %#v", result)
	}
	if result.Items[0].Issue == nil || result.Items[0].Issue.Code != "application_install_failed" {
		t.Fatalf("expected actionable installation issue, got %#v", result.Items[0].Issue)
	}
	if strings.Contains(result.Items[0].Issue.Summary, "private runtime detail") ||
		strings.Contains(result.Items[0].Issue.NextAction, "private runtime detail") {
		t.Fatalf("user-facing issue exposed backend detail: %#v", result.Items[0].Issue)
	}
}

func TestInstallationServiceExplainsRuntimeStorageAccessFailure(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	service := NewInstallationService(
		&installationSetup{status: SetupStatus{
			StoragePath: "/Users/test/Downloads", Applications: []string{"qbittorrent"},
			CanPrepare: true, CanInstall: true, TermsAccepted: true,
		}},
		&installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Downloads/Corsarr"}},
		NewCatalog(registry),
		&recordingInstaller{err: containerruntime.ErrBindMountAccessDenied},
		&recordingProvisioner{},
	)

	result, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("return structured storage access failure: %v", err)
	}
	issue := result.Items[0].Issue
	if issue == nil || issue.Code != "runtime_storage_access_denied" {
		t.Fatalf("expected actionable storage access issue, got %#v", issue)
	}
	if !strings.Contains(issue.NextAction, "outra pasta") {
		t.Fatalf("expected safe folder guidance, got %#v", issue)
	}
}

func TestInstallationServiceDistinguishesProvisioningFailure(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	service := NewInstallationService(
		&installationSetup{status: SetupStatus{
			StoragePath: "/Users/test/Media", Applications: []string{"jellyfin"},
			CanPrepare: true, CanInstall: true, TermsAccepted: true,
		}},
		&installationLayout{status: storage.LayoutStatus{RootPath: "/Users/test/Media/Corsarr"}},
		NewCatalog(registry),
		&recordingInstaller{},
		&recordingProvisioner{err: errors.New("private provisioning detail")},
	)

	result, err := service.InstallSelected(context.Background(), runtimecatalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("return structured provisioning failure: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Issue == nil ||
		result.Items[0].Issue.Code != "application_configuration_failed" {
		t.Fatalf("expected actionable provisioning issue, got %#v", result)
	}
}

type installationSetup struct{ status SetupStatus }

func (s *installationSetup) Load() (SetupStatus, error) { return s.status, nil }

type installationLayout struct{ status storage.LayoutStatus }

func (l *installationLayout) Prepare(string, []string) (storage.LayoutStatus, error) {
	return l.status, nil
}

type recordingInstaller struct {
	applicationIDs []string
	err            error
}

type recordingProvisioner struct {
	applicationIDs []string
	err            error
}

func (p *recordingProvisioner) Provision(
	_ context.Context,
	_ string,
	applicationID string,
	_ []string,
) error {
	p.applicationIDs = append(p.applicationIDs, applicationID)
	return p.err
}

func (i *recordingInstaller) Install(
	_ context.Context,
	applicationID string,
	_ string,
	_ runtimecatalog.RuntimeOptions,
) (containerruntime.ContainerStatus, error) {
	i.applicationIDs = append(i.applicationIDs, applicationID)
	if i.err != nil {
		return containerruntime.ContainerStatus{}, i.err
	}
	return containerruntime.ContainerStatus{
		ApplicationID: applicationID,
		State:         containerruntime.ContainerStateRunning,
	}, nil
}

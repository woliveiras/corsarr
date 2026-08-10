package application

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/orchestrator"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

func TestUpdateServiceUsesReviewedSetupAndProvisionsSuccessfulUpdate(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	setup := &updateSetup{status: SetupStatus{
		StoragePath: "/Users/test/Media", CanInstall: true, TermsAccepted: true,
	}}
	executor := &fakeUpdateExecutor{result: orchestrator.UpdateResult{
		ApplicationID: "radarr", Updated: true,
		Status: containerruntime.ContainerStatus{ApplicationID: "radarr", State: containerruntime.ContainerStateRunning},
	}}
	provisioner := &recordingProvisioner{}
	service := NewUpdateService(setup, NewCatalog(registry), executor, provisioner)

	result, err := service.Update(context.Background(), "radarr", catalog.RuntimeOptions{PUID: 1000, PGID: 1000})
	if err != nil {
		t.Fatalf("update application: %v", err)
	}
	if !result.Updated || result.Error != "" {
		t.Fatalf("unexpected result %#v", result)
	}
	if executor.rootPath != "/Users/test/Media/Corsarr" || executor.applicationID != "radarr" {
		t.Fatalf("unexpected executor call %#v", executor)
	}
	if len(provisioner.applicationIDs) != 1 || provisioner.applicationIDs[0] != "radarr" {
		t.Fatalf("expected provisioning after update, got %v", provisioner.applicationIDs)
	}
}

func TestUpdateServiceReturnsRollbackOutcomeWithoutProvisioning(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeUpdateExecutor{
		result: orchestrator.UpdateResult{ApplicationID: "sonarr", RolledBack: true},
		err:    errors.New("new image not ready"),
	}
	provisioner := &recordingProvisioner{}
	service := NewUpdateService(&updateSetup{status: SetupStatus{
		StoragePath: "/tmp", CanInstall: true, TermsAccepted: true,
	}}, NewCatalog(registry), executor, provisioner)

	result, err := service.Update(context.Background(), "sonarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("expected structured update failure, got %v", err)
	}
	if !result.RolledBack || result.RequiresAttention || result.Error == "" || len(provisioner.applicationIDs) != 0 {
		t.Fatalf("unexpected rollback result %#v", result)
	}
	if result.Issue == nil || result.Issue.Code != "application_update_rolled_back" {
		t.Fatalf("expected actionable rollback issue, got %#v", result.Issue)
	}
}

func TestUpdateServiceReportsAttentionWithoutExposingDetailsToDesktop(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeUpdateExecutor{err: errors.New("private rollback detail")}
	service := NewUpdateService(&updateSetup{status: SetupStatus{
		StoragePath: "/tmp", CanInstall: true, TermsAccepted: true,
	}}, NewCatalog(registry), executor, &recordingProvisioner{})

	result, err := service.Update(context.Background(), "sonarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("return structured update failure: %v", err)
	}
	if !result.RequiresAttention || result.Error == "" || result.RolledBack {
		t.Fatalf("unexpected attention result %#v", result)
	}
	if result.Issue == nil || result.Issue.Code != "application_update_failed" {
		t.Fatalf("expected actionable update issue, got %#v", result.Issue)
	}
}

func TestUpdateServiceDistinguishesProvisioningFailure(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeUpdateExecutor{result: orchestrator.UpdateResult{
		ApplicationID: "radarr", Updated: true,
	}}
	service := NewUpdateService(&updateSetup{status: SetupStatus{
		StoragePath: "/tmp", TermsAccepted: true,
	}}, NewCatalog(registry), executor, &recordingProvisioner{err: errors.New("private detail")})

	result, err := service.Update(context.Background(), "radarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("update application: %v", err)
	}
	if !result.RequiresAttention || result.Issue == nil ||
		result.Issue.Code != "application_configuration_failed" {
		t.Fatalf("expected configuration issue after update, got %#v", result)
	}
}

func TestUpdateServiceRejectsMissingConsentBeforeExecutor(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeUpdateExecutor{}
	service := NewUpdateService(&updateSetup{status: SetupStatus{StoragePath: "/tmp"}}, NewCatalog(registry), executor, &recordingProvisioner{})

	if _, err := service.Update(context.Background(), "radarr", catalog.RuntimeOptions{}); !errors.Is(err, ErrTermsNotAccepted) {
		t.Fatalf("expected consent error, got %v", err)
	}
	if executor.calls != 0 {
		t.Fatalf("expected no update execution, got %d", executor.calls)
	}
}

func TestUpdateServiceAllowsInstalledApplicationAfterSelectionWasCleared(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeUpdateExecutor{result: orchestrator.UpdateResult{ApplicationID: "radarr"}}
	service := NewUpdateService(&updateSetup{status: SetupStatus{
		StoragePath: "/tmp", TermsAccepted: true, CanInstall: false,
	}}, NewCatalog(registry), executor, &recordingProvisioner{})

	if _, err := service.Update(context.Background(), "radarr", catalog.RuntimeOptions{}); err != nil {
		t.Fatalf("update previously installed application: %v", err)
	}
	if executor.calls != 1 {
		t.Fatalf("expected one update execution, got %d", executor.calls)
	}
}

type updateSetup struct {
	status SetupStatus
	err    error
}

func (s *updateSetup) Load() (SetupStatus, error) { return s.status, s.err }

type fakeUpdateExecutor struct {
	result        orchestrator.UpdateResult
	err           error
	calls         int
	applicationID string
	rootPath      string
}

func (e *fakeUpdateExecutor) Update(
	_ context.Context,
	applicationID string,
	rootPath string,
	_ catalog.RuntimeOptions,
) (orchestrator.UpdateResult, error) {
	e.calls++
	e.applicationID = applicationID
	e.rootPath = rootPath
	return e.result, e.err
}

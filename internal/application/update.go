package application

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/orchestrator"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

type UpdateExecutor interface {
	Update(
		ctx context.Context,
		applicationID string,
		rootPath string,
		options catalog.RuntimeOptions,
	) (orchestrator.UpdateResult, error)
}

type ApplicationUpdateResult struct {
	ApplicationID string                           `json:"applicationId"`
	PreviousImage string                           `json:"previousImage"`
	ApprovedImage string                           `json:"approvedImage"`
	Backup        storage.BackupResult             `json:"backup"`
	Status        containerruntime.ContainerStatus `json:"status"`
	Updated       bool                             `json:"updated"`
	RolledBack    bool                             `json:"rolledBack"`
	Error         string                           `json:"error,omitempty"`
}

type UpdateService struct {
	setup       InstallationSetup
	catalog     *Catalog
	executor    UpdateExecutor
	provisioner ApplicationProvisioner
	mu          sync.Mutex
}

func NewUpdateService(
	setup InstallationSetup,
	catalog *Catalog,
	executor UpdateExecutor,
	provisioner ApplicationProvisioner,
) *UpdateService {
	return &UpdateService{
		setup: setup, catalog: catalog, executor: executor, provisioner: provisioner,
	}
}

// Update accepts only a catalog application ID and derives storage and runtime
// configuration from persisted, consented setup.
func (s *UpdateService) Update(
	ctx context.Context,
	applicationID string,
	options catalog.RuntimeOptions,
) (ApplicationUpdateResult, error) {
	if _, exists := s.catalog.byID[applicationID]; !exists {
		return ApplicationUpdateResult{}, fmt.Errorf(
			"application is not available in the desktop catalog: %s",
			applicationID,
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	setup, err := s.setup.Load()
	if err != nil {
		return ApplicationUpdateResult{}, fmt.Errorf("load reviewed setup: %w", err)
	}
	if !setup.TermsAccepted {
		return ApplicationUpdateResult{}, ErrTermsNotAccepted
	}
	if setup.StoragePath == "" {
		return ApplicationUpdateResult{}, fmt.Errorf("reviewed storage path is not configured")
	}

	execution, updateErr := s.executor.Update(
		ctx,
		applicationID,
		filepath.Join(setup.StoragePath, "Corsarr"),
		options,
	)
	result := applicationUpdateResult(execution)
	if updateErr != nil {
		result.Error = updateErr.Error()
		return result, nil
	}
	if !result.Updated {
		return result, nil
	}
	if err := s.provisioner.Provision(ctx, filepath.Join(setup.StoragePath, "Corsarr"), applicationID); err != nil {
		result.Error = fmt.Sprintf("reconcile application configuration after update: %v", err)
	}
	return result, nil
}

func applicationUpdateResult(execution orchestrator.UpdateResult) ApplicationUpdateResult {
	return ApplicationUpdateResult{
		ApplicationID: execution.ApplicationID,
		PreviousImage: execution.PreviousImage,
		ApprovedImage: execution.ApprovedImage,
		Backup:        execution.Backup,
		Status:        execution.Status,
		Updated:       execution.Updated,
		RolledBack:    execution.RolledBack,
	}
}

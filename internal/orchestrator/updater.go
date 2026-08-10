package orchestrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

type ConfigurationBackup interface {
	Backup(rootPath, applicationID string) (storage.BackupResult, error)
}

type UpdateResult struct {
	ApplicationID string                           `json:"applicationId"`
	PreviousImage string                           `json:"previousImage"`
	ApprovedImage string                           `json:"approvedImage"`
	Backup        storage.BackupResult             `json:"backup"`
	Status        containerruntime.ContainerStatus `json:"status"`
	Updated       bool                             `json:"updated"`
	RolledBack    bool                             `json:"rolledBack"`
}

type Updater struct {
	runtime   containerruntime.Manager
	resolver  SpecResolver
	readiness ReadinessWaiter
	backup    ConfigurationBackup
}

func NewUpdater(
	runtime containerruntime.Manager,
	resolver SpecResolver,
	readiness ReadinessWaiter,
	backup ConfigurationBackup,
) *Updater {
	return &Updater{runtime: runtime, resolver: resolver, readiness: readiness, backup: backup}
}

// Update replaces an owned container with the catalog-approved image. The
// previous immutable image is recreated if replacement or readiness fails.
func (u *Updater) Update(
	ctx context.Context,
	applicationID string,
	rootPath string,
	options catalog.RuntimeOptions,
) (UpdateResult, error) {
	result := UpdateResult{ApplicationID: applicationID}
	approvedSpec, err := u.resolver.Resolve(applicationID, rootPath, options)
	if err != nil {
		return result, fmt.Errorf("resolve approved application manifest: %w", err)
	}
	if approvedSpec.ApplicationID != applicationID {
		return result, fmt.Errorf(
			"resolved application mismatch: requested %s, got %s",
			applicationID,
			approvedSpec.ApplicationID,
		)
	}
	if err := approvedSpec.Validate(); err != nil {
		return result, fmt.Errorf("validate approved application manifest: %w", err)
	}
	approvedContract, err := approvedSpec.ContractFingerprint()
	if err != nil {
		return result, fmt.Errorf("fingerprint approved application manifest: %w", err)
	}
	result.ApprovedImage = approvedSpec.Image

	if err := u.runtime.EnsureNetwork(ctx); err != nil {
		return result, fmt.Errorf("prepare runtime network: %w", err)
	}
	previousStatus, err := u.runtime.Inspect(ctx, applicationID)
	if err != nil {
		return result, fmt.Errorf("inspect installed application: %w", err)
	}
	result.PreviousImage = previousStatus.Image
	result.Status = previousStatus
	if previousStatus.Image == approvedSpec.Image {
		return result, nil
	}
	if previousStatus.ContractFingerprint != approvedContract {
		return result, fmt.Errorf(
			"installed container contract differs from the approved image-only update contract",
		)
	}
	if previousStatus.State != containerruntime.ContainerStateRunning &&
		previousStatus.State != containerruntime.ContainerStateStopped &&
		previousStatus.State != containerruntime.ContainerStateCreated {
		return result, fmt.Errorf("application cannot be safely updated from state %s", previousStatus.State)
	}

	previousSpec := approvedSpec
	previousSpec.Image = previousStatus.Image
	if err := previousSpec.Validate(); err != nil {
		return result, fmt.Errorf("previous image cannot be safely restored: %w", err)
	}

	backup, err := u.backup.Backup(rootPath, applicationID)
	if err != nil {
		return result, fmt.Errorf("back up application configuration: %w", err)
	}
	result.Backup = backup
	if err := u.runtime.Pull(ctx, approvedSpec.Image); err != nil {
		return result, fmt.Errorf("download approved application image: %w", err)
	}

	wasRunning := previousStatus.State == containerruntime.ContainerStateRunning
	if wasRunning {
		if err := u.runtime.Stop(ctx, applicationID); err != nil {
			return result, fmt.Errorf("stop application before update: %w", err)
		}
	}
	if err := u.runtime.Remove(ctx, applicationID); err != nil {
		updateErr := fmt.Errorf("remove previous application container: %w", err)
		if wasRunning {
			if restartErr := u.runtime.Start(context.WithoutCancel(ctx), applicationID); restartErr != nil {
				updateErr = errors.Join(updateErr, fmt.Errorf("restore previous running state: %w", restartErr))
			} else {
				result.RolledBack = true
			}
		}
		return result, updateErr
	}

	newCreated := false
	if err := u.runtime.Create(ctx, approvedSpec); err != nil {
		return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
			fmt.Errorf("create updated application container: %w", err))
	}
	newCreated = true
	if err := u.runtime.Start(ctx, applicationID); err != nil {
		return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
			fmt.Errorf("start updated application container: %w", err))
	}
	updatedStatus, err := u.runtime.Inspect(ctx, applicationID)
	if err != nil {
		return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
			fmt.Errorf("inspect updated application container: %w", err))
	}
	if updatedStatus.State != containerruntime.ContainerStateRunning {
		return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
			fmt.Errorf("updated application did not reach running state: %s", updatedStatus.State))
	}
	if err := u.readiness.Wait(ctx, applicationID); err != nil {
		return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
			fmt.Errorf("wait for updated application readiness: %w", err))
	}

	if !wasRunning {
		if err := u.runtime.Stop(ctx, applicationID); err != nil {
			return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
				fmt.Errorf("restore stopped application state: %w", err))
		}
		updatedStatus, err = u.runtime.Inspect(ctx, applicationID)
		if err != nil {
			return u.rollback(ctx, result, previousSpec, wasRunning, newCreated,
				fmt.Errorf("verify stopped application state: %w", err))
		}
	}

	result.Status = updatedStatus
	result.Updated = true
	return result, nil
}

func (u *Updater) rollback(
	ctx context.Context,
	result UpdateResult,
	previousSpec containerruntime.ContainerSpec,
	wasRunning bool,
	newCreated bool,
	updateErr error,
) (UpdateResult, error) {
	rollbackContext := context.WithoutCancel(ctx)
	if newCreated {
		if err := u.runtime.Remove(rollbackContext, result.ApplicationID); err != nil {
			return result, errors.Join(updateErr, fmt.Errorf("remove failed update container: %w", err))
		}
	}
	if err := u.runtime.Create(rollbackContext, previousSpec); err != nil {
		return result, errors.Join(updateErr, fmt.Errorf("recreate previous application container: %w", err))
	}
	if wasRunning {
		if err := u.runtime.Start(rollbackContext, result.ApplicationID); err != nil {
			return result, errors.Join(updateErr, fmt.Errorf("restart previous application container: %w", err))
		}
	}
	status, err := u.runtime.Inspect(rollbackContext, result.ApplicationID)
	if err != nil {
		return result, errors.Join(updateErr, fmt.Errorf("verify previous application container: %w", err))
	}
	if wasRunning {
		if status.State != containerruntime.ContainerStateRunning {
			return result, errors.Join(updateErr, fmt.Errorf(
				"previous application did not return to running state: %s", status.State,
			))
		}
		if err := u.readiness.Wait(rollbackContext, result.ApplicationID); err != nil {
			return result, errors.Join(updateErr, fmt.Errorf("verify previous application readiness: %w", err))
		}
	}
	result.Status = status
	result.RolledBack = true
	return result, updateErr
}

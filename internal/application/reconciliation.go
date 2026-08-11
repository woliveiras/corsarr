package application

import (
	"context"
	"errors"
	"fmt"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

var ErrConfigurationReconciliationDisabled = errors.New(
	"configuration reconciliation is not enabled",
)

type ConfigurationRuntime interface {
	Inspect(ctx context.Context, applicationID string) (containerruntime.ContainerStatus, error)
}

type ConfigurationReconciliationItem struct {
	ApplicationID string
	Skipped       bool
	Failed        bool
}

type ConfigurationReconciliationResult struct {
	Items    []ConfigurationReconciliationItem
	Complete bool
}

type ConfigurationReconciler struct {
	setup       RecoverySetup
	catalog     *Catalog
	runtime     ConfigurationRuntime
	provisioner ApplicationProvisioner
}

func NewConfigurationReconciler(
	setup RecoverySetup,
	catalog *Catalog,
	runtime ConfigurationRuntime,
	provisioner ApplicationProvisioner,
) *ConfigurationReconciler {
	return &ConfigurationReconciler{
		setup: setup, catalog: catalog, runtime: runtime, provisioner: provisioner,
	}
}

func (r *ConfigurationReconciler) Reconcile(
	ctx context.Context,
) (ConfigurationReconciliationResult, error) {
	setup, err := r.setup.Load()
	if err != nil {
		return ConfigurationReconciliationResult{}, fmt.Errorf(
			"load setup for configuration reconciliation: %w",
			err,
		)
	}
	if !setup.OnboardingCompleted || setup.StoragePath == "" || len(setup.Applications) == 0 {
		return ConfigurationReconciliationResult{}, ErrConfigurationReconciliationDisabled
	}
	ordered, err := r.catalog.InstallationOrder(setup.Applications)
	if err != nil {
		return ConfigurationReconciliationResult{}, fmt.Errorf(
			"order configuration reconciliation: %w",
			err,
		)
	}

	result := ConfigurationReconciliationResult{
		Items: make([]ConfigurationReconciliationItem, 0, len(ordered)), Complete: true,
	}
	for _, applicationID := range ordered {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		item := ConfigurationReconciliationItem{ApplicationID: applicationID}
		status, inspectErr := r.runtime.Inspect(ctx, applicationID)
		if errors.Is(inspectErr, containerruntime.ErrResourceNotFound) {
			item.Skipped = true
			result.Items = append(result.Items, item)
			continue
		}
		if inspectErr != nil {
			item.Failed = true
			result.Complete = false
			result.Items = append(result.Items, item)
			continue
		}
		if status.State != containerruntime.ContainerStateRunning {
			item.Skipped = true
			result.Items = append(result.Items, item)
			continue
		}
		if err := r.provisioner.Provision(
			ctx,
			setup.StoragePath,
			applicationID,
			setup.Applications,
		); err != nil {
			item.Failed = true
			result.Complete = false
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

package orchestrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type SpecResolver interface {
	Resolve(
		applicationID string,
		rootPath string,
		options catalog.RuntimeOptions,
	) (containerruntime.ContainerSpec, error)
}

type Installer struct {
	runtime  containerruntime.Manager
	resolver SpecResolver
}

func NewInstaller(runtime containerruntime.Manager, resolver SpecResolver) *Installer {
	return &Installer{runtime: runtime, resolver: resolver}
}

func (i *Installer) Install(
	ctx context.Context,
	applicationID string,
	rootPath string,
	options catalog.RuntimeOptions,
) (containerruntime.ContainerStatus, error) {
	spec, err := i.resolver.Resolve(applicationID, rootPath, options)
	if err != nil {
		return containerruntime.ContainerStatus{}, fmt.Errorf("resolve application manifest: %w", err)
	}
	if spec.ApplicationID != applicationID {
		return containerruntime.ContainerStatus{}, fmt.Errorf(
			"resolved application mismatch: requested %s, got %s",
			applicationID,
			spec.ApplicationID,
		)
	}
	if err := spec.Validate(); err != nil {
		return containerruntime.ContainerStatus{}, fmt.Errorf("validate application manifest: %w", err)
	}
	if err := i.runtime.EnsureNetwork(ctx); err != nil {
		return containerruntime.ContainerStatus{}, fmt.Errorf("prepare runtime network: %w", err)
	}
	existing, inspectErr := i.runtime.Inspect(ctx, applicationID)
	if inspectErr == nil {
		if existing.Image != spec.Image {
			return containerruntime.ContainerStatus{}, fmt.Errorf(
				"installed image differs from approved image; use the update workflow",
			)
		}
		if existing.State == containerruntime.ContainerStateRunning {
			return existing, nil
		}
		if err := i.runtime.Start(ctx, applicationID); err != nil {
			return containerruntime.ContainerStatus{}, fmt.Errorf("start existing application container: %w", err)
		}
		status, err := i.runtime.Inspect(ctx, applicationID)
		if err != nil {
			return containerruntime.ContainerStatus{}, fmt.Errorf("verify existing application container: %w", err)
		}
		if status.State != containerruntime.ContainerStateRunning {
			return containerruntime.ContainerStatus{}, fmt.Errorf(
				"existing application container did not reach running state: %s",
				status.State,
			)
		}
		return status, nil
	}
	if !errors.Is(inspectErr, containerruntime.ErrResourceNotFound) {
		return containerruntime.ContainerStatus{}, fmt.Errorf("inspect existing application: %w", inspectErr)
	}
	if err := i.runtime.Pull(ctx, spec.Image); err != nil {
		return containerruntime.ContainerStatus{}, fmt.Errorf("download application image: %w", err)
	}
	if err := i.runtime.Create(ctx, spec); err != nil {
		return containerruntime.ContainerStatus{}, fmt.Errorf("create application container: %w", err)
	}

	if err := i.runtime.Start(ctx, applicationID); err != nil {
		return containerruntime.ContainerStatus{}, i.cleanupIncomplete(
			ctx,
			applicationID,
			fmt.Errorf("start application container: %w", err),
		)
	}
	status, err := i.runtime.Inspect(ctx, applicationID)
	if err != nil {
		return containerruntime.ContainerStatus{}, i.cleanupIncomplete(
			ctx,
			applicationID,
			fmt.Errorf("verify application container: %w", err),
		)
	}
	if status.State != containerruntime.ContainerStateRunning {
		return containerruntime.ContainerStatus{}, i.cleanupIncomplete(
			ctx,
			applicationID,
			fmt.Errorf("application container did not reach running state: %s", status.State),
		)
	}
	return status, nil
}

func (i *Installer) cleanupIncomplete(
	ctx context.Context,
	applicationID string,
	installErr error,
) error {
	cleanupErr := i.runtime.Remove(context.WithoutCancel(ctx), applicationID)
	if cleanupErr == nil {
		return installErr
	}
	return errors.Join(installErr, fmt.Errorf("remove incomplete container: %w", cleanupErr))
}

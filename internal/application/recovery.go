package application

import (
	"context"
	"errors"
	"fmt"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

var ErrBackgroundRecoveryDisabled = errors.New("background recovery is not enabled")

type RecoverySetup interface {
	Load() (SetupStatus, error)
}

type RecoveryRuntime interface {
	Inspect(ctx context.Context, applicationID string) (containerruntime.ContainerStatus, error)
	Start(ctx context.Context, applicationID string) error
}

type RecoveryItem struct {
	ApplicationID string `json:"applicationId"`
	Started       bool   `json:"started"`
	Skipped       bool   `json:"skipped"`
	Error         string `json:"error,omitempty"`
}

type RecoveryResult struct {
	Items    []RecoveryItem `json:"items"`
	Complete bool           `json:"complete"`
}

// RecoveryService starts only selected containers that already exist and are
// stopped. It never pulls an image, creates a container, provisions an app, or
// removes a resource.
type RecoveryService struct {
	setup   RecoverySetup
	catalog *Catalog
	runtime RecoveryRuntime
}

func NewRecoveryService(
	setup RecoverySetup,
	catalog *Catalog,
	runtime RecoveryRuntime,
) *RecoveryService {
	return &RecoveryService{setup: setup, catalog: catalog, runtime: runtime}
}

func (s *RecoveryService) Recover(ctx context.Context) (RecoveryResult, error) {
	setup, err := s.setup.Load()
	if err != nil {
		return RecoveryResult{}, fmt.Errorf("load setup for background recovery: %w", err)
	}
	if !setup.StartAtLogin || !setup.StartAtLoginSupported ||
		setup.StartAtLoginRequiresApproval {
		return RecoveryResult{}, ErrBackgroundRecoveryDisabled
	}
	ordered, err := s.catalog.InstallationOrder(setup.Applications)
	if err != nil {
		return RecoveryResult{}, fmt.Errorf("order background recovery: %w", err)
	}

	result := RecoveryResult{Items: make([]RecoveryItem, 0, len(ordered)), Complete: true}
	for _, applicationID := range ordered {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		item := RecoveryItem{ApplicationID: applicationID}
		status, inspectErr := s.runtime.Inspect(ctx, applicationID)
		if errors.Is(inspectErr, containerruntime.ErrResourceNotFound) {
			item.Skipped = true
			result.Items = append(result.Items, item)
			continue
		}
		if inspectErr != nil {
			item.Error = inspectErr.Error()
			result.Complete = false
			result.Items = append(result.Items, item)
			continue
		}

		switch status.State {
		case containerruntime.ContainerStateRunning:
			// Already reconciled by the runtime restart policy.
		case containerruntime.ContainerStateCreated, containerruntime.ContainerStateStopped:
			if startErr := s.runtime.Start(ctx, applicationID); startErr != nil {
				item.Error = startErr.Error()
				result.Complete = false
			} else {
				item.Started = true
			}
		default:
			item.Error = fmt.Sprintf("container is not startable from state %q", status.State)
			result.Complete = false
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

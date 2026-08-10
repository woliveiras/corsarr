package application

import (
	"context"
	"errors"
	"fmt"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type ManagedState string

const (
	ManagedStateNotInstalled ManagedState = "not_installed"
	ManagedStateRunning      ManagedState = "running"
	ManagedStateStopped      ManagedState = "stopped"
	ManagedStateAttention    ManagedState = "attention"
)

type ManagedApplicationStatus struct {
	ApplicationID   string       `json:"applicationId"`
	State           ManagedState `json:"state"`
	Health          string       `json:"health,omitempty"`
	Image           string       `json:"image,omitempty"`
	TechnicalDetail string       `json:"technicalDetail,omitempty"`
}

type ManagementService struct {
	catalog *Catalog
	runtime containerruntime.Manager
}

func NewManagementService(catalog *Catalog, runtime containerruntime.Manager) *ManagementService {
	return &ManagementService{catalog: catalog, runtime: runtime}
}

func (s *ManagementService) ListStatuses(ctx context.Context) []ManagedApplicationStatus {
	applications := s.catalog.ListApplications()
	statuses := make([]ManagedApplicationStatus, 0, len(applications))
	for _, application := range applications {
		container, err := s.runtime.Inspect(ctx, application.ID)
		if errors.Is(err, containerruntime.ErrResourceNotFound) {
			statuses = append(statuses, ManagedApplicationStatus{
				ApplicationID: application.ID,
				State:         ManagedStateNotInstalled,
			})
			continue
		}
		if err != nil {
			statuses = append(statuses, ManagedApplicationStatus{
				ApplicationID:   application.ID,
				State:           ManagedStateAttention,
				TechnicalDetail: err.Error(),
			})
			continue
		}
		statuses = append(statuses, ManagedApplicationStatus{
			ApplicationID: application.ID,
			State:         managedState(container.State),
			Health:        container.Health,
			Image:         container.Image,
		})
	}
	return statuses
}

func (s *ManagementService) Start(ctx context.Context, applicationID string) error {
	if err := s.validateTarget(applicationID); err != nil {
		return err
	}
	return s.runtime.Start(ctx, applicationID)
}

func (s *ManagementService) Stop(ctx context.Context, applicationID string) error {
	if err := s.validateTarget(applicationID); err != nil {
		return err
	}
	return s.runtime.Stop(ctx, applicationID)
}

func (s *ManagementService) Restart(ctx context.Context, applicationID string) error {
	if err := s.validateTarget(applicationID); err != nil {
		return err
	}
	return s.runtime.Restart(ctx, applicationID)
}

// Remove removes only the owned runtime container. Bind-mounted data is preserved.
func (s *ManagementService) Remove(ctx context.Context, applicationID string) error {
	if err := s.validateTarget(applicationID); err != nil {
		return err
	}
	return s.runtime.Remove(ctx, applicationID)
}

func (s *ManagementService) validateTarget(applicationID string) error {
	if _, exists := s.catalog.byID[applicationID]; !exists {
		return fmt.Errorf("application is not available in the desktop catalog: %s", applicationID)
	}
	return nil
}

func managedState(state containerruntime.ContainerState) ManagedState {
	switch state {
	case containerruntime.ContainerStateRunning:
		return ManagedStateRunning
	case containerruntime.ContainerStateCreated, containerruntime.ContainerStateStopped:
		return ManagedStateStopped
	default:
		return ManagedStateAttention
	}
}

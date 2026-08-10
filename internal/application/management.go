package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

var ErrApplicationRequired = errors.New("application is required by installed dependents")

type ManagedState string

const (
	ManagedStateNotInstalled ManagedState = "not_installed"
	ManagedStateRunning      ManagedState = "running"
	ManagedStateStopped      ManagedState = "stopped"
	ManagedStateAttention    ManagedState = "attention"
)

type ManagedApplicationStatus struct {
	ApplicationID    string       `json:"applicationId"`
	State            ManagedState `json:"state"`
	Health           string       `json:"health,omitempty"`
	Image            string       `json:"image,omitempty"`
	ApprovedImage    string       `json:"approvedImage,omitempty"`
	UpdateAvailable  bool         `json:"updateAvailable"`
	TechnicalDetail  string       `json:"technicalDetail,omitempty"`
	RemovalBlockedBy []string     `json:"removalBlockedBy,omitempty"`
}

type ApprovedImageResolver interface {
	ApprovedImage(applicationID string) (string, error)
}

type ManagementService struct {
	catalog        *Catalog
	runtime        containerruntime.Manager
	approvedImages ApprovedImageResolver
}

func NewManagementService(
	catalog *Catalog,
	runtime containerruntime.Manager,
	approvedImages ...ApprovedImageResolver,
) *ManagementService {
	service := &ManagementService{catalog: catalog, runtime: runtime}
	if len(approvedImages) > 0 {
		service.approvedImages = approvedImages[0]
	}
	return service
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
		status := ManagedApplicationStatus{
			ApplicationID: application.ID,
			State:         managedState(container.State),
			Health:        container.Health,
			Image:         container.Image,
		}
		if s.approvedImages != nil {
			approvedImage, resolveErr := s.approvedImages.ApprovedImage(application.ID)
			if resolveErr != nil {
				status.State = ManagedStateAttention
				status.TechnicalDetail = resolveErr.Error()
			} else {
				status.ApprovedImage = approvedImage
				status.UpdateAvailable = container.Image != approvedImage
			}
		}
		statuses = append(statuses, status)
	}
	installed := make(map[string]bool, len(statuses))
	for _, status := range statuses {
		installed[status.ApplicationID] = status.State != ManagedStateNotInstalled
	}
	for index := range statuses {
		if !installed[statuses[index].ApplicationID] {
			continue
		}
		statuses[index].RemovalBlockedBy = s.catalogDependents(
			statuses[index].ApplicationID,
			installed,
		)
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
	dependents, err := s.installedDependents(ctx, applicationID)
	if err != nil {
		return err
	}
	if len(dependents) > 0 {
		return fmt.Errorf("%w: %s", ErrApplicationRequired, strings.Join(dependents, ", "))
	}
	return s.runtime.Remove(ctx, applicationID)
}

func (s *ManagementService) installedDependents(
	ctx context.Context,
	applicationID string,
) ([]string, error) {
	installed := make(map[string]bool, len(s.catalog.applications))
	for _, application := range s.catalog.applications {
		if application.ID == applicationID || !dependsOn(application, applicationID) {
			continue
		}
		_, err := s.runtime.Inspect(ctx, application.ID)
		switch {
		case err == nil:
			installed[application.ID] = true
		case errors.Is(err, containerruntime.ErrResourceNotFound):
			installed[application.ID] = false
		default:
			return nil, fmt.Errorf("inspect dependent application %s: %w", application.ID, err)
		}
	}
	return s.catalogDependents(applicationID, installed), nil
}

func dependsOn(application ApplicationSummary, dependencyID string) bool {
	for _, candidate := range application.Dependencies {
		if candidate == dependencyID {
			return true
		}
	}
	return false
}

func (s *ManagementService) catalogDependents(
	applicationID string,
	installed map[string]bool,
) []string {
	dependents := make([]string, 0)
	for _, application := range s.catalog.applications {
		if !installed[application.ID] {
			continue
		}
		if dependsOn(application, applicationID) {
			dependents = append(dependents, application.ID)
		}
	}
	sort.Strings(dependents)
	return dependents
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

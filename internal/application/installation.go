package application

import (
	"context"
	"errors"
	"fmt"

	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

var ErrTermsNotAccepted = errors.New("current runtime terms have not been accepted")

type InstallationSetup interface {
	Load() (SetupStatus, error)
}

type InstallationLayout interface {
	Prepare(basePath string, applicationIDs []string) (storage.LayoutStatus, error)
}

type ApplicationInstaller interface {
	Install(
		ctx context.Context,
		applicationID string,
		rootPath string,
		options runtimecatalog.RuntimeOptions,
	) (containerruntime.ContainerStatus, error)
}

type ApplicationProvisioner interface {
	Provision(ctx context.Context, rootPath, applicationID string, selected []string) error
}

type InstallationItem struct {
	ApplicationID string                           `json:"applicationId"`
	Status        containerruntime.ContainerStatus `json:"-"`
	Failed        bool                             `json:"failed"`
	Issue         *OperationIssue                  `json:"issue,omitempty"`
	Error         string                           `json:"-"`
}

type InstallationResult struct {
	Items    []InstallationItem `json:"items"`
	Complete bool               `json:"complete"`
}

type InstallationStage string

const (
	InstallationStageInstalling   InstallationStage = "installing"
	InstallationStageProvisioning InstallationStage = "provisioning"
	InstallationStageReady        InstallationStage = "ready"
	InstallationStageFailed       InstallationStage = "failed"
)

// InstallationProgress contains only bounded catalog state suitable for a
// user-facing progress event. Runtime output and provisioning errors are not
// included because they may contain sensitive implementation details.
type InstallationProgress struct {
	ApplicationID string            `json:"applicationId"`
	Stage         InstallationStage `json:"stage"`
	Position      int               `json:"position"`
	Total         int               `json:"total"`
}

type InstallationProgressObserver func(InstallationProgress)

type InstallationService struct {
	setup       InstallationSetup
	layout      InstallationLayout
	catalog     *Catalog
	installer   ApplicationInstaller
	provisioner ApplicationProvisioner
}

func NewInstallationService(
	setup InstallationSetup,
	layout InstallationLayout,
	catalog *Catalog,
	installer ApplicationInstaller,
	provisioner ApplicationProvisioner,
) *InstallationService {
	return &InstallationService{
		setup: setup, layout: layout, catalog: catalog,
		installer: installer, provisioner: provisioner,
	}
}

func (s *InstallationService) InstallSelected(
	ctx context.Context,
	options runtimecatalog.RuntimeOptions,
) (InstallationResult, error) {
	return s.InstallSelectedWithProgress(ctx, options, nil)
}

func (s *InstallationService) InstallSelectedWithProgress(
	ctx context.Context,
	options runtimecatalog.RuntimeOptions,
	observer InstallationProgressObserver,
) (InstallationResult, error) {
	setup, err := s.setup.Load()
	if err != nil {
		return InstallationResult{}, fmt.Errorf("load reviewed setup: %w", err)
	}
	if !setup.CanInstall || !setup.TermsAccepted {
		return InstallationResult{}, ErrTermsNotAccepted
	}
	layout, err := s.layout.Prepare(setup.StoragePath, setup.Applications)
	if err != nil {
		return InstallationResult{}, fmt.Errorf("prepare reviewed storage: %w", err)
	}
	ordered, err := s.catalog.InstallationOrder(setup.Applications)
	if err != nil {
		return InstallationResult{}, fmt.Errorf("order selected applications: %w", err)
	}

	result := InstallationResult{Items: make([]InstallationItem, 0, len(ordered))}
	for index, applicationID := range ordered {
		notifyInstallationProgress(observer, applicationID, InstallationStageInstalling, index+1, len(ordered))
		status, installErr := s.installer.Install(ctx, applicationID, layout.RootPath, options)
		item := InstallationItem{ApplicationID: applicationID, Status: status}
		if installErr != nil {
			item.Error = installErr.Error()
			item.Failed = true
			item.Issue = installationIssue()
			result.Items = append(result.Items, item)
			notifyInstallationProgress(observer, applicationID, InstallationStageFailed, index+1, len(ordered))
			return result, nil
		}
		notifyInstallationProgress(observer, applicationID, InstallationStageProvisioning, index+1, len(ordered))
		if provisionErr := s.provisioner.Provision(
			ctx,
			layout.RootPath,
			applicationID,
			setup.Applications,
		); provisionErr != nil {
			item.Error = provisionErr.Error()
			item.Failed = true
			item.Issue = configurationIssue()
			result.Items = append(result.Items, item)
			notifyInstallationProgress(observer, applicationID, InstallationStageFailed, index+1, len(ordered))
			return result, nil
		}
		result.Items = append(result.Items, item)
		notifyInstallationProgress(observer, applicationID, InstallationStageReady, index+1, len(ordered))
	}
	result.Complete = true
	return result, nil
}

func notifyInstallationProgress(
	observer InstallationProgressObserver,
	applicationID string,
	stage InstallationStage,
	position int,
	total int,
) {
	if observer == nil {
		return
	}
	observer(InstallationProgress{
		ApplicationID: applicationID,
		Stage:         stage,
		Position:      position,
		Total:         total,
	})
}

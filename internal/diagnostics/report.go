package diagnostics

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/woliveiras/corsarr/internal/application"
	"github.com/woliveiras/corsarr/internal/storage"
)

const (
	CurrentSchemaVersion         = 1
	RedactedValue                = "<redacted>"
	maximumTechnicalDetailLength = 1_000
)

var (
	credentialPattern = regexp.MustCompile(
		`(?i)(password|passwd|api[-_ ]?key|token|authorization|cookie)(\s*[:=]\s*)([^\s,;]+)`,
	)
	urlCredentialPattern = regexp.MustCompile(`://([^:/@\s]+):([^@\s]+)@`)
)

type EnvironmentReader interface {
	Status(ctx context.Context) application.EnvironmentStatus
}

type SetupReader interface {
	Load() (application.SetupStatus, error)
}

type ApplicationReader interface {
	ListStatuses(ctx context.Context) []application.ManagedApplicationStatus
}

type StorageReader interface {
	Inspect(path string) storage.Status
}

type SetupReport struct {
	StoragePath                  string   `json:"storagePath,omitempty"`
	Applications                 []string `json:"applications"`
	TermsVersion                 string   `json:"termsVersion"`
	TermsAccepted                bool     `json:"termsAccepted"`
	StartAtLogin                 bool     `json:"startAtLogin"`
	StartAtLoginRequiresApproval bool     `json:"startAtLoginRequiresApproval"`
	JellyfinLANEnabled           bool     `json:"jellyfinLanEnabled"`
}

// ApplicationStatus is an explicit diagnostic-file projection. The desktop
// status payload omits raw technical detail; an export includes only this
// separately sanitized copy after the user selects a destination.
type ApplicationStatus struct {
	ApplicationID    string                      `json:"applicationId"`
	State            application.ManagedState    `json:"state"`
	Health           string                      `json:"health,omitempty"`
	Image            string                      `json:"image,omitempty"`
	ApprovedImage    string                      `json:"approvedImage,omitempty"`
	UpdateAvailable  bool                        `json:"updateAvailable"`
	TechnicalDetail  string                      `json:"technicalDetail,omitempty"`
	Issue            *application.OperationIssue `json:"issue,omitempty"`
	RemovalBlockedBy []string                    `json:"removalBlockedBy,omitempty"`
}

type Report struct {
	SchemaVersion     int                           `json:"schemaVersion"`
	GeneratedAt       string                        `json:"generatedAt"`
	CorsarrVersion    string                        `json:"corsarrVersion"`
	CatalogVerifiedAt string                        `json:"catalogVerifiedAt"`
	Environment       application.EnvironmentStatus `json:"environment"`
	Setup             SetupReport                   `json:"setup"`
	Storage           *storage.Status               `json:"storage,omitempty"`
	Applications      []ApplicationStatus           `json:"applications"`
}

// Reporter captures only bounded diagnostic state. It deliberately excludes
// application logs, credentials, cookies, request bodies, and runtime sockets.
type Reporter struct {
	environment       EnvironmentReader
	setup             SetupReader
	applications      ApplicationReader
	storage           StorageReader
	corsarrVersion    string
	catalogVerifiedAt string
	now               func() time.Time
}

func NewReporter(
	environment EnvironmentReader,
	setup SetupReader,
	applications ApplicationReader,
	storageReader StorageReader,
	corsarrVersion string,
	catalogVerifiedAt string,
) *Reporter {
	return &Reporter{
		environment: environment, setup: setup, applications: applications,
		storage: storageReader, corsarrVersion: corsarrVersion,
		catalogVerifiedAt: catalogVerifiedAt, now: time.Now,
	}
}

func (r *Reporter) Build(ctx context.Context) (Report, error) {
	setupStatus, err := r.setup.Load()
	if err != nil {
		return Report{}, fmt.Errorf("load setup for diagnostics: %w", err)
	}
	environmentStatus := r.environment.Status(ctx)
	environmentStatus.Runtime.TechnicalDetail = sanitizedDetail(
		environmentStatus.Runtime.TechnicalDetail,
	)

	managedStatuses := r.applications.ListStatuses(ctx)
	applicationStatuses := make([]ApplicationStatus, 0, len(managedStatuses))
	for _, status := range managedStatuses {
		applicationStatuses = append(applicationStatuses, ApplicationStatus{
			ApplicationID:    status.ApplicationID,
			State:            status.State,
			Health:           status.Health,
			Image:            status.Image,
			ApprovedImage:    status.ApprovedImage,
			UpdateAvailable:  status.UpdateAvailable,
			TechnicalDetail:  sanitizedDetail(status.TechnicalDetail),
			Issue:            status.Issue,
			RemovalBlockedBy: append([]string(nil), status.RemovalBlockedBy...),
		})
	}

	report := Report{
		SchemaVersion:     CurrentSchemaVersion,
		GeneratedAt:       r.now().UTC().Format(time.RFC3339),
		CorsarrVersion:    r.corsarrVersion,
		CatalogVerifiedAt: r.catalogVerifiedAt,
		Environment:       environmentStatus,
		Setup: SetupReport{
			StoragePath:                  setupStatus.StoragePath,
			Applications:                 append([]string(nil), setupStatus.Applications...),
			TermsVersion:                 setupStatus.TermsVersion,
			TermsAccepted:                setupStatus.TermsAccepted,
			StartAtLogin:                 setupStatus.StartAtLogin,
			StartAtLoginRequiresApproval: setupStatus.StartAtLoginRequiresApproval,
			JellyfinLANEnabled:           setupStatus.JellyfinLANEnabled,
		},
		Applications: applicationStatuses,
	}
	if strings.TrimSpace(setupStatus.StoragePath) != "" {
		storageStatus := r.storage.Inspect(setupStatus.StoragePath)
		storageStatus.TechnicalDetail = sanitizedDetail(storageStatus.TechnicalDetail)
		report.Storage = &storageStatus
	}
	return report, nil
}

func sanitizedDetail(detail string) string {
	if detail == "" {
		return ""
	}
	redacted := credentialPattern.ReplaceAllString(detail, `${1}${2}`+RedactedValue)
	redacted = urlCredentialPattern.ReplaceAllString(redacted, `://$1:`+RedactedValue+`@`)
	redacted = strings.TrimSpace(redacted)
	if len(redacted) <= maximumTechnicalDetailLength {
		return redacted
	}
	return redacted[:maximumTechnicalDetailLength] + "…"
}

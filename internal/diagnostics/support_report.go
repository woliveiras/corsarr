package diagnostics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/woliveiras/corsarr/internal/application"
)

var (
	unixHomePathPattern    = regexp.MustCompile(`/Users/[^/\s]+`)
	windowsHomePathPattern = regexp.MustCompile(`(?i)[A-Z]:\\Users\\[^\\\s]+`)
)

type installationFailure struct {
	ApplicationID   string                      `json:"applicationId"`
	Issue           *application.OperationIssue `json:"issue,omitempty"`
	TechnicalDetail string                      `json:"technicalDetail"`
}

type installationSupportReport struct {
	Diagnostics Report              `json:"diagnostics"`
	Failure     installationFailure `json:"failure"`
}

// FormatInstallationSupportReport creates a bounded diagnostic artifact for an
// explicit copy action. Credentials and user-selected filesystem paths are
// removed before the report can cross the desktop bridge.
func FormatInstallationSupportReport(
	report Report,
	applicationID string,
	issue *application.OperationIssue,
	technicalDetail string,
) (string, error) {
	applicationID = strings.TrimSpace(applicationID)
	if applicationID == "" {
		return "", fmt.Errorf("support report application ID is empty")
	}
	privatePaths := []string{report.Setup.StoragePath}
	if report.Storage != nil {
		privatePaths = append(privatePaths, report.Storage.Path)
	}
	if report.Setup.StoragePath != "" {
		report.Setup.StoragePath = RedactedValue
	}
	if report.Storage != nil {
		storageCopy := *report.Storage
		if storageCopy.Path != "" {
			storageCopy.Path = RedactedValue
		}
		report.Storage = &storageCopy
	}
	report.Environment.Runtime.TechnicalDetail = redactSupportPaths(
		report.Environment.Runtime.TechnicalDetail,
	)
	for index := range report.Applications {
		report.Applications[index].TechnicalDetail = redactSupportPaths(
			report.Applications[index].TechnicalDetail,
		)
	}
	payload := installationSupportReport{
		Diagnostics: report,
		Failure: installationFailure{
			ApplicationID: applicationID,
			Issue:         issue,
			TechnicalDetail: redactSupportPaths(
				redactSelectedPaths(sanitizedDetail(technicalDetail), privatePaths),
			),
		},
	}

	var contents bytes.Buffer
	encoder := json.NewEncoder(&contents)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return "", fmt.Errorf("encode installation support report: %w", err)
	}
	return strings.TrimSuffix(contents.String(), "\n"), nil
}

func redactSelectedPaths(detail string, paths []string) string {
	for _, path := range paths {
		if strings.TrimSpace(path) != "" {
			detail = strings.ReplaceAll(detail, path, RedactedValue)
		}
	}
	return detail
}

func redactSupportPaths(detail string) string {
	detail = unixHomePathPattern.ReplaceAllString(detail, "$HOME")
	return windowsHomePathPattern.ReplaceAllString(detail, `%USERPROFILE%`)
}

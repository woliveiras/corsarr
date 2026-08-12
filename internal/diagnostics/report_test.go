package diagnostics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/application"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

func TestReporterBuildsBoundedRedactedSnapshot(t *testing.T) {
	reporter := NewReporter(
		&diagnosticEnvironment{status: application.EnvironmentStatus{
			Platform: "darwin", Architecture: "arm64",
			Runtime: containerruntime.Status{
				Provider: containerruntime.ProviderDocker,
				State:    containerruntime.StateError,
				Version:  "28.3.2",
				TechnicalDetail: "connection failed password=super-secret " +
					strings.Repeat("x", 2_000),
			},
		}},
		&diagnosticSetup{status: application.SetupStatus{
			StoragePath: "/Users/test/Media", Applications: []string{"radarr"},
			OnboardingCompleted: true, OnboardingStep: application.OnboardingStepComplete,
			TermsVersion: "2026-08-10.2", TermsAccepted: true,
		}},
		&diagnosticApplications{statuses: []application.ManagedApplicationStatus{{
			ApplicationID:   "radarr",
			State:           application.ManagedStateAttention,
			Image:           "lscr.io/linuxserver/radarr@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TechnicalDetail: "api_key: abcdef123456",
		}}},
		&diagnosticStorage{status: storage.Status{
			Path: "/Users/test/Media", State: storage.StateReady, Writable: true, Hardlinks: true,
		}},
		"0.1.0-test",
		"2026-08-10",
	)
	reporter.now = func() time.Time {
		return time.Date(2026, 8, 10, 21, 30, 0, 0, time.UTC)
	}

	report, err := reporter.Build(context.Background())
	if err != nil {
		t.Fatalf("build diagnostics: %v", err)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("encode diagnostics: %v", err)
	}
	text := string(encoded)
	if strings.Contains(text, "super-secret") || strings.Contains(text, "abcdef123456") {
		t.Fatalf("diagnostics leaked a credential: %s", text)
	}
	if !strings.Contains(text, `"technicalDetail":"api_key: \u003credacted\u003e"`) {
		t.Fatalf("explicit diagnostic export omitted sanitized application detail: %s", text)
	}
	if !strings.Contains(report.Environment.Runtime.TechnicalDetail, RedactedValue) ||
		!strings.Contains(report.Applications[0].TechnicalDetail, RedactedValue) {
		t.Fatalf("expected redaction markers in %#v", report)
	}
	if report.GeneratedAt != "2026-08-10T21:30:00Z" || report.CorsarrVersion != "0.1.0-test" {
		t.Fatalf("unexpected metadata %#v", report)
	}
	if report.Storage == nil || !report.Storage.Hardlinks || report.Storage.Path != "/Users/test/Media" {
		t.Fatalf("unexpected storage report %#v", report.Storage)
	}
	if !report.Setup.OnboardingCompleted || report.Setup.OnboardingStep != application.OnboardingStepComplete {
		t.Fatalf("unexpected onboarding report %#v", report.Setup)
	}
	if len(report.Environment.Runtime.TechnicalDetail) > maximumTechnicalDetailLength+len("…") {
		t.Fatalf("runtime detail was not bounded: %d", len(report.Environment.Runtime.TechnicalDetail))
	}
}

func TestReporterDoesNotInspectStorageWhenSetupHasNoPath(t *testing.T) {
	inspector := &diagnosticStorage{}
	reporter := NewReporter(
		&diagnosticEnvironment{},
		&diagnosticSetup{},
		&diagnosticApplications{},
		inspector,
		"development",
		"2026-08-10",
	)

	report, err := reporter.Build(context.Background())
	if err != nil {
		t.Fatalf("build diagnostics: %v", err)
	}
	if report.Storage != nil || inspector.calls != 0 {
		t.Fatalf("expected no storage inspection, report=%#v calls=%d", report.Storage, inspector.calls)
	}
}

func TestInstallationSupportReportIncludesFailureAndRedactsPrivateData(t *testing.T) {
	report := Report{
		SchemaVersion:  CurrentSchemaVersion,
		GeneratedAt:    "2026-08-11T10:33:00Z",
		CorsarrVersion: "development",
		Setup: SetupReport{
			StoragePath:  "/Volumes/Media Library",
			Applications: []string{"jellyseerr"},
		},
		Storage: &storage.Status{Path: "/Volumes/Media Library", State: storage.StateReady},
	}
	issue := &application.OperationIssue{Code: "application_configuration_failed"}

	contents, err := FormatInstallationSupportReport(
		report,
		"jellyseerr",
		issue,
		"ensure Seerr setup from /Volumes/Media Library: password=private-value unexpected HTTP status 500",
	)
	if err != nil {
		t.Fatalf("format support report: %v", err)
	}
	if !strings.Contains(contents, `"applicationId": "jellyseerr"`) ||
		!strings.Contains(contents, "unexpected HTTP status 500") {
		t.Fatalf("support report omitted failure context: %s", contents)
	}
	for _, private := range []string{"private-value", "/Volumes/Media Library"} {
		if strings.Contains(contents, private) {
			t.Fatalf("support report leaked %q: %s", private, contents)
		}
	}
	if !strings.Contains(contents, RedactedValue) {
		t.Fatalf("support report omitted redaction marker: %s", contents)
	}
}

func TestFileWriterCreatesPrivateAtomicJSON(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "corsarr-diagnostics.json")
	report := Report{SchemaVersion: CurrentSchemaVersion, GeneratedAt: "2026-08-10T21:30:00Z"}

	if err := NewFileWriter().Write(destination, report); err != nil {
		t.Fatalf("write diagnostics: %v", err)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stat diagnostics: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected private diagnostics, got %04o", info.Mode().Perm())
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read diagnostics: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") || !json.Valid(data) {
		t.Fatalf("expected formatted JSON, got %q", string(data))
	}
}

func TestFileWriterRejectsRelativeAndSymlinkTargets(t *testing.T) {
	writer := NewFileWriter()
	if err := writer.Write("diagnostics.json", Report{}); err == nil {
		t.Fatal("expected relative destination rejection")
	}
	directory := t.TempDir()
	target := filepath.Join(directory, "actual.json")
	if err := os.WriteFile(target, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(directory, "diagnostics.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	if err := writer.Write(link, Report{}); err == nil {
		t.Fatal("expected symlink destination rejection")
	}
}

type diagnosticEnvironment struct {
	status application.EnvironmentStatus
}

func (f *diagnosticEnvironment) Status(context.Context) application.EnvironmentStatus {
	return f.status
}

type diagnosticSetup struct {
	status application.SetupStatus
	err    error
}

func (f *diagnosticSetup) Load() (application.SetupStatus, error) {
	return f.status, f.err
}

type diagnosticApplications struct {
	statuses []application.ManagedApplicationStatus
}

func (f *diagnosticApplications) ListStatuses(context.Context) []application.ManagedApplicationStatus {
	return f.statuses
}

type diagnosticStorage struct {
	status storage.Status
	calls  int
}

func (f *diagnosticStorage) Inspect(string) storage.Status {
	f.calls++
	return f.status
}

package quality

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestSyncerPreviewsBeforeApplyingPinnedPresetWithoutPersistingAPIKeys(t *testing.T) {
	rootPath := t.TempDir()
	runner := &recordingRunner{}
	syncer := NewSyncer(runner, fixedCredentialReader{
		"radarr": strings.Repeat("a", 32),
		"sonarr": strings.Repeat("b", 32),
	})

	result, err := syncer.Apply(context.Background(), Request{
		RootPath: rootPath, Applications: []string{"radarr", "sonarr"},
		Preset: PresetBalanced1080p, PUID: 501, PGID: 20,
	})
	if err != nil {
		t.Fatalf("apply quality preset: %v", err)
	}
	if !result.Previewed || !result.Applied || result.ProfileName != "Corsarr - Equilibrado 1080p" {
		t.Fatalf("unexpected quality result %#v", result)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("expected pull, preview and apply, got %#v", runner.calls)
	}
	if !reflect.DeepEqual(runner.calls[0].arguments, []string{"pull", RecyclarrImage}) {
		t.Fatalf("expected immutable image pull first, got %#v", runner.calls[0].arguments)
	}
	if !containsArgument(runner.calls[1].arguments, "--preview") ||
		containsArgument(runner.calls[2].arguments, "--preview") {
		t.Fatalf("expected preview before apply, got %#v", runner.calls)
	}
	for _, call := range runner.calls {
		joined := strings.Join(call.arguments, " ")
		if strings.Contains(joined, strings.Repeat("a", 32)) || strings.Contains(joined, strings.Repeat("b", 32)) {
			t.Fatalf("API key leaked into command arguments: %s", joined)
		}
	}
	if runner.calls[1].environment["RADARR_API_KEY"] != strings.Repeat("a", 32) ||
		runner.calls[1].environment["SONARR_API_KEY"] != strings.Repeat("b", 32) {
		t.Fatalf("expected runtime-only credentials, got %#v", runner.calls[1].environment)
	}

	configPath := filepath.Join(rootPath, "config", "recyclarr", "recyclarr.yml")
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	contents := string(config)
	for _, expected := range []string{
		"# Corsarr quality policy " + PresetCatalogVersion,
		"d1d67249d3890e49bc12e275d989a7e9",
		"9d142234e45d6143785ac55f5a9e8dc9",
		"!env_var RADARR_API_KEY",
		"!env_var SONARR_API_KEY",
		"movie:\n        rename: true\n        standard: standard",
		"episodes:\n        rename: true\n        standard: default",
	} {
		if !strings.Contains(contents, expected) {
			t.Fatalf("generated config does not contain %q:\n%s", expected, contents)
		}
	}
	if strings.Contains(contents, strings.Repeat("a", 32)) || strings.Contains(contents, strings.Repeat("b", 32)) {
		t.Fatalf("generated config persisted an API key:\n%s", contents)
	}
	configInfo, err := os.Stat(configPath)
	if err != nil || configInfo.Mode().Perm() != 0o600 {
		t.Fatalf("expected private generated config: info=%v err=%v", configInfo, err)
	}
	directoryInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil || directoryInfo.Mode().Perm() != 0o700 {
		t.Fatalf("expected private Recyclarr directory: info=%v err=%v", directoryInfo, err)
	}

	settings, err := os.ReadFile(filepath.Join(rootPath, "config", "recyclarr", "settings.yml"))
	if err != nil {
		t.Fatalf("read generated settings: %v", err)
	}
	if !strings.Contains(string(settings), TrashGuidesCommit) ||
		!strings.Contains(string(settings), "replace_default: true") {
		t.Fatalf("TRaSH Guides source is not pinned:\n%s", settings)
	}
}

func TestSyncerDoesNotApplyWhenPreviewFails(t *testing.T) {
	runner := &recordingRunner{failCall: 2}
	syncer := NewSyncer(runner, fixedCredentialReader{"radarr": strings.Repeat("a", 32)})

	result, err := syncer.Apply(context.Background(), Request{
		RootPath: t.TempDir(), Applications: []string{"radarr"}, Preset: PresetBalanced1080p,
	})
	if !errors.Is(err, ErrSyncFailed) {
		t.Fatalf("expected sanitized preview failure, got result=%#v err=%v", result, err)
	}
	if result.Previewed || result.Applied || len(runner.calls) != 2 {
		t.Fatalf("preview failure reached apply: result=%#v calls=%#v", result, runner.calls)
	}
}

func TestSyncerPreservesFailureDetailWithoutAPIKeys(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	runner := &recordingRunner{failCall: 2, errorText: "preview failed for " + secret}
	syncer := NewSyncer(runner, fixedCredentialReader{"radarr": secret})

	_, err := syncer.Apply(context.Background(), Request{
		RootPath: t.TempDir(), Applications: []string{"radarr"}, Preset: PresetBalanced1080p,
	})
	if !errors.Is(err, ErrSyncFailed) || !strings.Contains(err.Error(), "preview failed") {
		t.Fatalf("expected useful sync failure, got %v", err)
	}
	if strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("API key was not redacted from sync failure: %v", err)
	}
}

func TestRecyclarrArgumentsQuoteMountPathsContainingCSVDelimiters(t *testing.T) {
	arguments := recyclarrArguments("/Users/test/Corsarr, Media/config/recyclarr", 501, 20, []string{"radarr"})
	want := `type=bind,src="/Users/test/Corsarr, Media/config/recyclarr",dst=/config`
	if !containsArgument(arguments, want) {
		t.Fatalf("expected safely quoted mount %q, got %#v", want, arguments)
	}
}

func TestSyncerRunsOnlyForSelectedARRApplications(t *testing.T) {
	runner := &recordingRunner{}
	syncer := NewSyncer(runner, fixedCredentialReader{"radarr": strings.Repeat("a", 32)})

	result, err := syncer.Apply(context.Background(), Request{
		RootPath: t.TempDir(), Applications: []string{"jellyfin"}, Preset: PresetBalanced1080p,
	})
	if err != nil {
		t.Fatalf("skip quality sync: %v", err)
	}
	if result.Applied || len(runner.calls) != 0 {
		t.Fatalf("expected no Arr sync, result=%#v calls=%#v", result, runner.calls)
	}
}

func TestSyncerRejectsUnknownPresetBeforeReadingCredentialsOrRunningContainer(t *testing.T) {
	runner := &recordingRunner{}
	reader := &countingCredentialReader{}
	syncer := NewSyncer(runner, reader)

	_, err := syncer.Apply(context.Background(), Request{
		RootPath: t.TempDir(), Applications: []string{"radarr"}, Preset: PresetID("unknown"),
	})
	if err == nil {
		t.Fatal("expected unknown preset to be rejected")
	}
	if reader.calls != 0 || len(runner.calls) != 0 {
		t.Fatalf("rejected preset crossed an external boundary: reads=%d runs=%d", reader.calls, len(runner.calls))
	}
}

type recordedCall struct {
	environment map[string]string
	arguments   []string
}

type recordingRunner struct {
	calls     []recordedCall
	failCall  int
	errorText string
}

func (r *recordingRunner) Run(_ context.Context, environment map[string]string, arguments ...string) error {
	copyEnvironment := make(map[string]string, len(environment))
	for key, value := range environment {
		copyEnvironment[key] = value
	}
	r.calls = append(r.calls, recordedCall{copyEnvironment, append([]string(nil), arguments...)})
	if r.failCall == len(r.calls) {
		if r.errorText != "" {
			return errors.New(r.errorText)
		}
		return errors.New("sensitive runtime output")
	}
	return nil
}

type fixedCredentialReader map[string]string

func (r fixedCredentialReader) Read(_ string, applicationID string) (credentials.Secret, error) {
	return credentials.NewSecret(r[applicationID]), nil
}

type countingCredentialReader struct{ calls int }

func (r *countingCredentialReader) Read(string, string) (credentials.Secret, error) {
	r.calls++
	return credentials.Secret{}, nil
}

func containsArgument(arguments []string, expected string) bool {
	for _, argument := range arguments {
		if argument == expected {
			return true
		}
	}
	return false
}

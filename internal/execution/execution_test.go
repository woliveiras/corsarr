package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

func TestConfigurationPersistenceAndValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "execution.json")
	s, err := Open(path, "linux")
	if err != nil {
		t.Fatal(err)
	}
	if s.Config().Kind != Desktop {
		t.Fatal("legacy default changed")
	}
	want := Config{Kind: Existing, Context: "colima"}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Open(path, "linux")
	if err != nil || reloaded.Config() != want {
		t.Fatalf("persisted selection: %v, %v", reloaded, err)
	}
	for _, bad := range []Config{{Kind: "remote"}, {Kind: Existing, Context: "--host=evil"}, {Kind: Engine, Context: "production"}} {
		if err := s.Save(bad); err == nil {
			t.Fatalf("accepted invalid selection: %#v", bad)
		}
	}
	if s.Config() != want {
		t.Fatal("failed save changed active selection")
	}
	if err := (Config{Kind: Engine}).Validate("windows"); err == nil {
		t.Fatal("accepted native Engine on Windows")
	}
	if err := os.WriteFile(path, []byte(`{"kind":"remote"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path, "linux"); err == nil {
		t.Fatal("silently fell back from invalid persisted config")
	}
}

type recordedCall struct {
	args []string
	env  map[string]string
}
type fakeRunner struct {
	endpoint, osType string
	calls            []recordedCall
	fail             bool
}

func (*fakeRunner) LookPath(string) (string, error) { return "/bin/docker", nil }
func (r *fakeRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return r.RunWithEnvironment(ctx, nil, name, args...)
}
func (r *fakeRunner) RunWithEnvironment(_ context.Context, env map[string]string, _ string, args ...string) (string, error) {
	r.calls = append(r.calls, recordedCall{append([]string(nil), args...), env})
	if r.fail {
		return "", fmt.Errorf("unavailable")
	}
	if len(args) > 0 && args[0] == "context" {
		return r.endpoint, nil
	}
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--version") {
		return "Docker version 28.3.2", nil
	}
	if strings.Contains(joined, ".OSType") {
		return r.osType, nil
	}
	return "28.3.2", nil
}

func selectionFor(t *testing.T, c Config) *Selection {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "execution.json"), "linux")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(c); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestExistingContextCannotRedirectOperationsToRemoteHost(t *testing.T) {
	base := &fakeRunner{endpoint: "unix:///run/user/1000/docker.sock"}
	runner := Runner{Selection: selectionFor(t, Config{Kind: Existing, Context: "work"}), Base: base}
	if _, err := runner.RunWithEnvironment(context.Background(), map[string]string{"RADARR_API_KEY": "secret"}, "docker", "start", "corsarr-radarr"); err != nil {
		t.Fatal(err)
	}
	call := base.calls[1]
	if !reflect.DeepEqual(call.args, []string{"--host", base.endpoint, "start", "corsarr-radarr"}) {
		t.Fatalf("wrong destination: %v", call.args)
	}
	if call.env["RADARR_API_KEY"] != "secret" || call.env["DOCKER_CONTEXT"] != "" || call.env["DOCKER_HOST"] != "" {
		t.Fatal("incorrect process environment")
	}
	for _, endpoint := range []string{"ssh://server", "tcp://127.0.0.1:2375", "npipe:////server/pipe/docker_engine", ""} {
		base.endpoint = endpoint
		base.calls = nil
		if _, err := runner.Run(context.Background(), "docker", "rm", "corsarr-radarr"); err == nil {
			t.Fatalf("accepted %q", endpoint)
		}
		if len(base.calls) != 1 {
			t.Fatal("container mutation ran against remote context")
		}
	}
}

func TestEnginePinsSystemSocketAndDoesNotInspectDefaultContext(t *testing.T) {
	base := &fakeRunner{}
	runner := Runner{Selection: selectionFor(t, Config{Kind: Engine}), Base: base}
	if _, err := runner.Run(context.Background(), "docker", "compose", "version"); err != nil {
		t.Fatal(err)
	}
	if len(base.calls) != 1 || !reflect.DeepEqual(base.calls[0].args, []string{"--host", "unix:///var/run/docker.sock", "compose", "version"}) {
		t.Fatalf("wrong command: %v", base.calls)
	}
}

func TestProbeRejectsWindowsContainers(t *testing.T) {
	for _, osType := range []string{"linux", "windows"} {
		base := &fakeRunner{osType: osType}
		env := New(selectionFor(t, Config{Kind: Engine}), base, "linux", "amd64")
		status := env.Probe.Check(context.Background())
		if (status.State == containerruntime.StateReady) != (osType == "linux") {
			t.Fatalf("accepted incorrect server type: %#v", status)
		}
	}
}

func TestExistingPreparationNeverInstallsOrStartsRuntime(t *testing.T) {
	base := &fakeRunner{endpoint: "unix:///run/docker.sock", osType: "linux"}
	env := New(selectionFor(t, Config{Kind: Existing, Context: "default"}), base, "linux", "amd64")
	result, err := env.Prepare(context.Background())
	if err != nil || !result.Ready || result.Installed || result.Started {
		t.Fatalf("existing runtime: %#v %v", result, err)
	}
	base.fail = true
	if _, err := env.Recover(context.Background()); err == nil {
		t.Fatal("unavailable runtime reported ready")
	}
	for _, call := range base.calls {
		for _, arg := range call.args {
			if arg == "start" || arg == "install" {
				t.Fatal("unexpected mutation")
			}
		}
	}
}

func TestSupportedLinuxReleaseDoesNotAcceptDerivatives(t *testing.T) {
	for _, tc := range []struct {
		release   string
		supported bool
	}{
		{"ID=ubuntu\nVERSION_CODENAME=noble", true},
		{"ID=debian\nVERSION_CODENAME=\"trixie\"", true},
		{"ID=linuxmint\nID_LIKE=ubuntu\nVERSION_CODENAME=noble", false},
		{"ID=ubuntu\nVERSION_CODENAME=future", false},
	} {
		if SupportedLinuxRelease(tc.release) != tc.supported {
			t.Fatalf("incorrect release support: %s", tc.release)
		}
	}
}

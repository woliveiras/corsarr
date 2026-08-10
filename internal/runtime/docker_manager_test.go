package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDockerManagerCreatesValidatedOwnedContainer(t *testing.T) {
	runner := &recordingCommandRunner{path: "/usr/local/bin/docker"}
	manager := NewDockerManager(runner, time.Second)
	spec := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "lscr.io/linuxserver/radarr@" + testImageDigest,
		Ports: []PortBinding{{
			HostPort:      7878,
			ContainerPort: 7878,
			Protocol:      ProtocolTCP,
			Exposure:      ExposureLoopback,
		}},
		Mounts: []BindMount{
			{HostPath: "/Users/test/Media/Corsarr/media", ContainerPath: "/data"},
			{HostPath: "/Users/test/Media/Corsarr/config/radarr", ContainerPath: "/config"},
		},
		Environment: map[string]string{"TZ": "Europe/Madrid", "UMASK": "002"},
	}

	if err := manager.Create(context.Background(), spec); err != nil {
		t.Fatalf("create container: %v", err)
	}
	want := commandCall{
		name: "/usr/local/bin/docker",
		args: []string{
			"create",
			"--name", "corsarr-radarr",
			"--label", "io.corsarr.managed=true",
			"--label", "io.corsarr.application=radarr",
			"--network", "corsarr",
			"--network-alias", "radarr",
			"--restart", "unless-stopped",
			"--publish", "127.0.0.1:7878:7878/tcp",
			"--mount", "type=bind,src=/Users/test/Media/Corsarr/config/radarr,dst=/config",
			"--mount", "type=bind,src=/Users/test/Media/Corsarr/media,dst=/data",
			"--env", "TZ=Europe/Madrid",
			"--env", "UMASK=002",
			"lscr.io/linuxserver/radarr@" + testImageDigest,
		},
	}
	if !reflect.DeepEqual(runner.calls, []commandCall{want}) {
		t.Fatalf("unexpected Docker command\nwant: %#v\n got: %#v", []commandCall{want}, runner.calls)
	}
}

func TestDockerManagerRejectsInvalidSpecBeforeRuntimeAccess(t *testing.T) {
	runner := &recordingCommandRunner{path: "/usr/local/bin/docker"}
	manager := NewDockerManager(runner, time.Second)

	err := manager.Create(context.Background(), ContainerSpec{
		ApplicationID: "../foreign",
		Image:         "example.invalid/app:latest",
	})
	if err == nil {
		t.Fatal("expected invalid spec to be rejected")
	}
	if len(runner.calls) != 0 || runner.lookPathCalls != 0 {
		t.Fatalf("expected no runtime access, got lookups=%d calls=%v", runner.lookPathCalls, runner.calls)
	}
}

func TestDockerManagerCreatesOwnedNetworkWhenMissing(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/usr/local/bin/docker",
		results: []managerCommandResult{
			{err: errors.New("Error: No such network: corsarr")},
			{output: "corsarr"},
		},
	}
	manager := NewDockerManager(runner, time.Second)

	if err := manager.EnsureNetwork(context.Background()); err != nil {
		t.Fatalf("ensure network: %v", err)
	}
	want := []commandCall{
		{name: runner.path, args: []string{"network", "inspect", "corsarr", "--format", `{{index .Labels "io.corsarr.managed"}}`}},
		{name: runner.path, args: []string{"network", "create", "--driver", "bridge", "--label", "io.corsarr.managed=true", "corsarr"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected network commands\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestDockerManagerRefusesForeignNetwork(t *testing.T) {
	runner := &recordingCommandRunner{
		path:    "/usr/local/bin/docker",
		results: []managerCommandResult{{output: ""}},
	}
	manager := NewDockerManager(runner, time.Second)

	err := manager.EnsureNetwork(context.Background())
	if !errors.Is(err, ErrResourceNotOwned) {
		t.Fatalf("expected foreign network error, got %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected inspect only, got %v", runner.calls)
	}
}

func TestDockerManagerVerifiesOwnershipBeforeRemoval(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/usr/local/bin/docker",
		results: []managerCommandResult{
			{output: "true"},
			{output: "corsarr-radarr"},
		},
	}
	manager := NewDockerManager(runner, time.Second)

	if err := manager.Remove(context.Background(), "radarr"); err != nil {
		t.Fatalf("remove owned container: %v", err)
	}
	want := []commandCall{
		{name: runner.path, args: []string{"inspect", "corsarr-radarr", "--format", `{{index .Config.Labels "io.corsarr.managed"}}`}},
		{name: runner.path, args: []string{"rm", "--force", "corsarr-radarr"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected removal commands\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestDockerManagerRefusesToRemoveForeignContainer(t *testing.T) {
	runner := &recordingCommandRunner{
		path:    "/usr/local/bin/docker",
		results: []managerCommandResult{{output: "false"}},
	}
	manager := NewDockerManager(runner, time.Second)

	err := manager.Remove(context.Background(), "radarr")
	if !errors.Is(err, ErrResourceNotOwned) {
		t.Fatalf("expected foreign container error, got %v", err)
	}
	if len(runner.calls) != 1 || strings.Contains(strings.Join(runner.calls[0].args, " "), " rm ") {
		t.Fatalf("expected inspect without removal, got %v", runner.calls)
	}
}

func TestDockerManagerInspectsOwnedContainerState(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/usr/local/bin/docker",
		results: []managerCommandResult{{output: `[
  {
    "Config": {
      "Image": "lscr.io/linuxserver/radarr@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "Labels": {"io.corsarr.managed": "true", "io.corsarr.application": "radarr"}
    },
    "State": {"Status": "running", "Health": {"Status": "healthy"}}
  }
]`}},
	}
	manager := NewDockerManager(runner, time.Second)

	status, err := manager.Inspect(context.Background(), "radarr")
	if err != nil {
		t.Fatalf("inspect owned container: %v", err)
	}
	want := ContainerStatus{
		ApplicationID: "radarr",
		State:         ContainerStateRunning,
		Health:        "healthy",
		Image:         "lscr.io/linuxserver/radarr@" + testImageDigest,
	}
	if !reflect.DeepEqual(status, want) {
		t.Fatalf("unexpected container status\nwant: %#v\n got: %#v", want, status)
	}
}

func TestDockerManagerInspectRejectsMismatchedApplicationLabel(t *testing.T) {
	runner := &recordingCommandRunner{
		path:    "/usr/local/bin/docker",
		results: []managerCommandResult{{output: `[{"Config":{"Labels":{"io.corsarr.managed":"true","io.corsarr.application":"sonarr"}},"State":{"Status":"running"}}]`}},
	}
	manager := NewDockerManager(runner, time.Second)

	_, err := manager.Inspect(context.Background(), "radarr")
	if !errors.Is(err, ErrResourceNotOwned) {
		t.Fatalf("expected mismatched label to be rejected, got %v", err)
	}
}

type commandCall struct {
	name string
	args []string
}

type managerCommandResult struct {
	output string
	err    error
}

type recordingCommandRunner struct {
	path          string
	lookPathErr   error
	lookPathCalls int
	calls         []commandCall
	results       []managerCommandResult
}

func (r *recordingCommandRunner) LookPath(string) (string, error) {
	r.lookPathCalls++
	return r.path, r.lookPathErr
}

func (r *recordingCommandRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	r.calls = append(r.calls, commandCall{name: name, args: append([]string(nil), args...)})
	if len(r.results) == 0 {
		return "", nil
	}
	result := r.results[0]
	r.results = r.results[1:]
	return result.output, result.err
}

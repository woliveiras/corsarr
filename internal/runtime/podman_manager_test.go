package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestPodmanManagerImplementsRuntimeContract(t *testing.T) {
	var _ Manager = NewPodmanManager(&recordingCommandRunner{}, time.Second)
}

func TestPodmanManagerCreatesValidatedOwnedContainerDirectly(t *testing.T) {
	runner := &recordingCommandRunner{path: "/opt/homebrew/bin/podman"}
	manager := NewPodmanManager(runner, time.Second)
	spec := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "lscr.io/linuxserver/radarr@" + testImageDigest,
		Init:          true,
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
	want := []commandCall{{
		name: runner.path,
		args: []string{
			"create",
			"--name", "corsarr-radarr",
			"--label", "io.corsarr.managed=true",
			"--label", "io.corsarr.application=radarr",
			"--network", "corsarr",
			"--network-alias", "radarr",
			"--restart", "unless-stopped",
			"--init",
			"--publish", "127.0.0.1:7878:7878/tcp",
			"--mount", "type=bind,src=/Users/test/Media/Corsarr/config/radarr,dst=/config",
			"--mount", "type=bind,src=/Users/test/Media/Corsarr/media,dst=/data",
			"--env", "TZ=Europe/Madrid",
			"--env", "UMASK=002",
			"lscr.io/linuxserver/radarr@" + testImageDigest,
		},
	}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected Podman command\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestPodmanManagerCreatesOwnedNetworkWhenMissing(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/opt/homebrew/bin/podman",
		results: []managerCommandResult{
			{err: errors.New("Error: network corsarr not found")},
			{output: "corsarr"},
		},
	}
	manager := NewPodmanManager(runner, time.Second)

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

func TestPodmanManagerPublishesExplicitLANPort(t *testing.T) {
	runner := &recordingCommandRunner{path: "/opt/homebrew/bin/podman"}
	manager := NewPodmanManager(runner, time.Second)

	if err := manager.Create(context.Background(), ContainerSpec{
		ApplicationID: "jellyfin",
		Image:         "lscr.io/linuxserver/jellyfin@" + testImageDigest,
		Ports: []PortBinding{{
			HostPort: 8096, ContainerPort: 8096, Protocol: ProtocolTCP, Exposure: ExposureLAN,
		}},
	}); err != nil {
		t.Fatalf("create LAN container: %v", err)
	}
	if !containsArguments(runner.calls[0].args, "--publish", "0.0.0.0:8096:8096/tcp") {
		t.Fatalf("expected explicit LAN publication, got %v", runner.calls[0].args)
	}
}

func TestPodmanManagerInspectsOwnedContainerState(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/opt/homebrew/bin/podman",
		results: []managerCommandResult{{output: `[{
  "Config": {
    "Image": "lscr.io/linuxserver/radarr@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "Labels": {"io.corsarr.managed": "true", "io.corsarr.application": "radarr"}
  },
  "State": {"Status": "running", "Health": {"Status": "healthy"}}
}]`}},
	}
	manager := NewPodmanManager(runner, time.Second)

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

func TestPodmanManagerVerifiesOwnershipBeforeRemoval(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/opt/homebrew/bin/podman",
		results: []managerCommandResult{
			{output: "true"},
			{output: "corsarr-radarr"},
		},
	}
	manager := NewPodmanManager(runner, time.Second)

	if err := manager.Remove(context.Background(), "radarr"); err != nil {
		t.Fatalf("remove owned container: %v", err)
	}
	want := []commandCall{
		{name: runner.path, args: []string{"inspect", "corsarr-radarr", "--format", containerOwnershipFormat}},
		{name: runner.path, args: []string{"rm", "--force", "corsarr-radarr"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected removal commands\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestPodmanManagerRefusesForeignContainer(t *testing.T) {
	runner := &recordingCommandRunner{
		path:    "/opt/homebrew/bin/podman",
		results: []managerCommandResult{{output: "false"}},
	}
	manager := NewPodmanManager(runner, time.Second)

	err := manager.Remove(context.Background(), "radarr")
	if !errors.Is(err, ErrResourceNotOwned) {
		t.Fatalf("expected foreign container error, got %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected inspect without removal, got %v", runner.calls)
	}
}

func TestPodmanManagerMapsMissingContainer(t *testing.T) {
	runner := &recordingCommandRunner{
		path:    "/opt/homebrew/bin/podman",
		results: []managerCommandResult{{err: errors.New("Error: no container with name or ID corsarr-radarr found")}},
	}
	manager := NewPodmanManager(runner, time.Second)

	_, err := manager.Inspect(context.Background(), "radarr")
	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected missing resource error, got %v", err)
	}
}

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	CorsarrNetworkName     = "corsarr"
	managedLabelName       = "io.corsarr.managed"
	applicationLabelName   = "io.corsarr.application"
	managedLabelValue      = "true"
	containerNamePrefix    = "corsarr-"
	dockerOwnershipFormat  = `{{index .Config.Labels "io.corsarr.managed"}}`
	networkOwnershipFormat = `{{index .Labels "io.corsarr.managed"}}`
)

var (
	ErrResourceNotFound = errors.New("runtime resource not found")
	ErrResourceNotOwned = errors.New("runtime resource is not owned by Corsarr")
)

type Manager interface {
	EnsureNetwork(ctx context.Context) error
	Pull(ctx context.Context, image string) error
	Create(ctx context.Context, spec ContainerSpec) error
	Inspect(ctx context.Context, applicationID string) (ContainerStatus, error)
	Start(ctx context.Context, applicationID string) error
	Stop(ctx context.Context, applicationID string) error
	Restart(ctx context.Context, applicationID string) error
	Remove(ctx context.Context, applicationID string) error
	Logs(ctx context.Context, applicationID string, tail int) (string, error)
}

type ContainerState string

const (
	ContainerStateCreated    ContainerState = "created"
	ContainerStateRunning    ContainerState = "running"
	ContainerStateStopped    ContainerState = "stopped"
	ContainerStateRestarting ContainerState = "restarting"
	ContainerStatePaused     ContainerState = "paused"
	ContainerStateUnknown    ContainerState = "unknown"
)

type ContainerStatus struct {
	ApplicationID string         `json:"applicationId"`
	State         ContainerState `json:"state"`
	Health        string         `json:"health,omitempty"`
	Image         string         `json:"image,omitempty"`
}

type DockerManager struct {
	runner  CommandRunner
	timeout time.Duration
}

func NewDockerManager(runner CommandRunner, timeout time.Duration) *DockerManager {
	return &DockerManager{runner: runner, timeout: timeout}
}

func (m *DockerManager) EnsureNetwork(ctx context.Context) error {
	output, err := m.run(ctx, "network", "inspect", CorsarrNetworkName, "--format", networkOwnershipFormat)
	if err == nil {
		if strings.TrimSpace(output) != managedLabelValue {
			return fmt.Errorf("network %s: %w", CorsarrNetworkName, ErrResourceNotOwned)
		}
		return nil
	}
	if !indicatesMissingResource(err.Error(), "network") {
		return fmt.Errorf("inspect Corsarr network: %w", err)
	}

	_, err = m.run(
		ctx,
		"network", "create",
		"--driver", "bridge",
		"--label", managedLabelName+"="+managedLabelValue,
		CorsarrNetworkName,
	)
	if err != nil {
		return fmt.Errorf("create Corsarr network: %w", err)
	}
	return nil
}

func (m *DockerManager) Pull(ctx context.Context, image string) error {
	if err := validateImageReference(image); err != nil {
		return err
	}
	if _, err := m.run(ctx, "pull", image); err != nil {
		return fmt.Errorf("pull approved image: %w", err)
	}
	return nil
}

func (m *DockerManager) Create(ctx context.Context, spec ContainerSpec) error {
	if err := spec.Validate(); err != nil {
		return err
	}

	arguments := []string{
		"create",
		"--name", containerName(spec.ApplicationID),
		"--label", managedLabelName + "=" + managedLabelValue,
		"--label", applicationLabelName + "=" + spec.ApplicationID,
		"--network", CorsarrNetworkName,
		"--network-alias", spec.ApplicationID,
		"--restart", "unless-stopped",
	}

	ports := append([]PortBinding(nil), spec.Ports...)
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].HostPort == ports[j].HostPort {
			return ports[i].Protocol < ports[j].Protocol
		}
		return ports[i].HostPort < ports[j].HostPort
	})
	for _, binding := range ports {
		host := "127.0.0.1"
		if binding.Exposure == ExposureLAN {
			host = "0.0.0.0"
		}
		publishedPort := fmt.Sprintf(
			"%s:%d:%d/%s",
			host,
			binding.HostPort,
			binding.ContainerPort,
			binding.Protocol,
		)
		arguments = append(arguments, "--publish", publishedPort)
	}

	mounts := append([]BindMount(nil), spec.Mounts...)
	sort.Slice(mounts, func(i, j int) bool {
		return path.Clean(mounts[i].ContainerPath) < path.Clean(mounts[j].ContainerPath)
	})
	for _, mount := range mounts {
		mountValue := fmt.Sprintf(
			"type=bind,src=%s,dst=%s",
			dockerMountPath(filepath.Clean(mount.HostPath)),
			path.Clean(mount.ContainerPath),
		)
		if mount.ReadOnly {
			mountValue += ",readonly"
		}
		arguments = append(arguments, "--mount", mountValue)
	}

	environmentNames := make([]string, 0, len(spec.Environment))
	for name := range spec.Environment {
		environmentNames = append(environmentNames, name)
	}
	sort.Strings(environmentNames)
	for _, name := range environmentNames {
		arguments = append(arguments, "--env", name+"="+spec.Environment[name])
	}
	arguments = append(arguments, spec.Image)

	if _, err := m.run(ctx, arguments...); err != nil {
		return fmt.Errorf("create container for %s: %w", spec.ApplicationID, err)
	}
	return nil
}

func (m *DockerManager) Inspect(
	ctx context.Context,
	applicationID string,
) (ContainerStatus, error) {
	if !runtimeApplicationIDPattern.MatchString(applicationID) {
		return ContainerStatus{}, fmt.Errorf("unsafe application ID: %q", applicationID)
	}
	output, err := m.run(ctx, "inspect", containerName(applicationID))
	if err != nil {
		if indicatesMissingResource(err.Error(), "container") {
			return ContainerStatus{}, fmt.Errorf(
				"container for %s: %w",
				applicationID,
				ErrResourceNotFound,
			)
		}
		return ContainerStatus{}, fmt.Errorf("inspect container for %s: %w", applicationID, err)
	}

	var containers []struct {
		Config struct {
			Image  string            `json:"Image"`
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		State struct {
			Status string `json:"Status"`
			Health *struct {
				Status string `json:"Status"`
			} `json:"Health"`
		} `json:"State"`
	}
	if err := json.Unmarshal([]byte(output), &containers); err != nil {
		return ContainerStatus{}, fmt.Errorf("decode container status for %s: %w", applicationID, err)
	}
	if len(containers) != 1 {
		return ContainerStatus{}, fmt.Errorf("unexpected container result for %s", applicationID)
	}
	container := containers[0]
	if container.Config.Labels[managedLabelName] != managedLabelValue ||
		container.Config.Labels[applicationLabelName] != applicationID {
		return ContainerStatus{}, fmt.Errorf("container for %s: %w", applicationID, ErrResourceNotOwned)
	}

	status := ContainerStatus{
		ApplicationID: applicationID,
		State:         normalizedContainerState(container.State.Status),
		Image:         container.Config.Image,
	}
	if container.State.Health != nil {
		status.Health = container.State.Health.Status
	}
	return status, nil
}

func (m *DockerManager) Start(ctx context.Context, applicationID string) error {
	return m.ownedLifecycle(ctx, applicationID, "start")
}

func (m *DockerManager) Stop(ctx context.Context, applicationID string) error {
	return m.ownedLifecycle(ctx, applicationID, "stop")
}

func (m *DockerManager) Restart(ctx context.Context, applicationID string) error {
	return m.ownedLifecycle(ctx, applicationID, "restart")
}

func (m *DockerManager) Remove(ctx context.Context, applicationID string) error {
	if err := m.verifyOwnedContainer(ctx, applicationID); err != nil {
		return err
	}
	if _, err := m.run(ctx, "rm", "--force", containerName(applicationID)); err != nil {
		return fmt.Errorf("remove container for %s: %w", applicationID, err)
	}
	return nil
}

// Logs returns only a bounded tail from an owned container. It is intentionally
// backend-only because application logs may contain bootstrap credentials.
func (m *DockerManager) Logs(
	ctx context.Context,
	applicationID string,
	tail int,
) (string, error) {
	if tail < 1 || tail > 500 {
		return "", fmt.Errorf("log tail must be between 1 and 500 lines")
	}
	if err := m.verifyOwnedContainer(ctx, applicationID); err != nil {
		return "", err
	}
	output, err := m.run(
		ctx,
		"logs", "--tail", strconv.Itoa(tail), containerName(applicationID),
	)
	if err != nil {
		return "", fmt.Errorf("read container logs for %s: %w", applicationID, err)
	}
	return output, nil
}

func (m *DockerManager) ownedLifecycle(
	ctx context.Context,
	applicationID string,
	operation string,
) error {
	if err := m.verifyOwnedContainer(ctx, applicationID); err != nil {
		return err
	}
	if _, err := m.run(ctx, operation, containerName(applicationID)); err != nil {
		return fmt.Errorf("%s container for %s: %w", operation, applicationID, err)
	}
	return nil
}

func (m *DockerManager) verifyOwnedContainer(ctx context.Context, applicationID string) error {
	if !runtimeApplicationIDPattern.MatchString(applicationID) {
		return fmt.Errorf("unsafe application ID: %q", applicationID)
	}
	output, err := m.run(
		ctx,
		"inspect", containerName(applicationID),
		"--format", dockerOwnershipFormat,
	)
	if err != nil {
		if indicatesMissingResource(err.Error(), "container") {
			return fmt.Errorf("container for %s: %w", applicationID, ErrResourceNotFound)
		}
		return fmt.Errorf("inspect container for %s: %w", applicationID, err)
	}
	if strings.TrimSpace(output) != managedLabelValue {
		return fmt.Errorf("container for %s: %w", applicationID, ErrResourceNotOwned)
	}
	return nil
}

func (m *DockerManager) run(ctx context.Context, arguments ...string) (string, error) {
	dockerPath, err := m.runner.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("find Docker client: %w", err)
	}
	operationContext, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	return m.runner.Run(operationContext, dockerPath, arguments...)
}

func containerName(applicationID string) string {
	return containerNamePrefix + applicationID
}

func indicatesMissingResource(detail string, resource string) bool {
	normalized := strings.ToLower(detail)
	return strings.Contains(normalized, "no such "+resource) ||
		(resource == "container" && strings.Contains(normalized, "no such object")) ||
		strings.Contains(normalized, resource+" not found")
}

func dockerMountPath(hostPath string) string {
	if strings.ContainsAny(hostPath, ",\"") {
		return strconv.Quote(hostPath)
	}
	return hostPath
}

func normalizedContainerState(state string) ContainerState {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "created":
		return ContainerStateCreated
	case "running":
		return ContainerStateRunning
	case "restarting":
		return ContainerStateRestarting
	case "paused":
		return ContainerStatePaused
	case "exited", "dead", "removing":
		return ContainerStateStopped
	default:
		return ContainerStateUnknown
	}
}

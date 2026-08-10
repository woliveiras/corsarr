package onboarding

import (
	"context"
	"fmt"
	"strings"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

const (
	podmanDefaultMachineName = "podman-machine-default"
	podmanMachineCPUs        = "4"
	podmanMachineMemoryMiB   = "4096"
	podmanMachineDiskGiB     = "100"
)

type PodmanMachinePreparationResult struct {
	Ready       bool   `json:"ready"`
	Initialized bool   `json:"initialized"`
	Started     bool   `json:"started"`
	Version     string `json:"version,omitempty"`
}

// PodmanMachineService initializes or starts only Corsarr's fixed default
// rootless machine on macOS and Windows. It never installs the Podman client,
// removes a machine, changes providers, or runs a machine command on Linux.
type PodmanMachineService struct {
	probe            containerruntime.Probe
	runner           containerruntime.CommandRunner
	platform         string
	pollInterval     time.Duration
	timeout          time.Duration
	operationTimeout time.Duration
}

func NewPodmanMachineService(
	probe containerruntime.Probe,
	runner containerruntime.CommandRunner,
	platform string,
) *PodmanMachineService {
	return &PodmanMachineService{
		probe:            probe,
		runner:           runner,
		platform:         platform,
		pollInterval:     2 * time.Second,
		timeout:          3 * time.Minute,
		operationTimeout: 10 * time.Minute,
	}
}

func (s *PodmanMachineService) Prepare(
	ctx context.Context,
) (PodmanMachinePreparationResult, error) {
	status := s.probe.Check(ctx)
	if status.State == containerruntime.StateReady {
		return PodmanMachinePreparationResult{Ready: true, Version: status.Version}, nil
	}
	if s.platform != "darwin" && s.platform != "windows" {
		return PodmanMachinePreparationResult{}, fmt.Errorf(
			"Podman Machine is not used for native runtime preparation on %s",
			s.platform,
		)
	}

	podmanPath, err := s.runner.LookPath("podman")
	if err != nil {
		return PodmanMachinePreparationResult{}, fmt.Errorf("find Podman client: %w", err)
	}
	result := PodmanMachinePreparationResult{}
	stateOutput, err := s.run(
		ctx,
		podmanPath,
		"machine", "inspect", podmanDefaultMachineName,
		"--format", "{{.State}}",
	)
	if err != nil {
		if !indicatesMissingPodmanMachine(err.Error()) {
			return result, fmt.Errorf("inspect Podman machine: %w", err)
		}
		if _, initErr := s.run(
			ctx,
			podmanPath,
			"machine", "init",
			"--cpus", podmanMachineCPUs,
			"--memory", podmanMachineMemoryMiB,
			"--disk-size", podmanMachineDiskGiB,
			"--update-connection=true",
			podmanDefaultMachineName,
		); initErr != nil {
			return result, fmt.Errorf("initialize Podman machine: %w", initErr)
		}
		result.Initialized = true
		stateOutput = "stopped"
	}

	switch strings.ToLower(strings.TrimSpace(stateOutput)) {
	case "stopped":
		if _, err := s.run(
			ctx,
			podmanPath,
			"machine", "start", "--quiet", "--update-connection=true",
			podmanDefaultMachineName,
		); err != nil {
			return result, fmt.Errorf("start Podman machine: %w", err)
		}
		result.Started = true
	case "running", "starting", "stopping":
		// Readiness polling below handles an in-progress transition.
	default:
		return result, fmt.Errorf("unsupported Podman machine state: %q", strings.TrimSpace(stateOutput))
	}

	waitContext, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		status = s.probe.Check(waitContext)
		if status.State == containerruntime.StateReady {
			result.Ready = true
			result.Version = status.Version
			return result, nil
		}
		select {
		case <-waitContext.Done():
			return result, fmt.Errorf("Podman runtime did not become ready: %w", waitContext.Err())
		case <-ticker.C:
		}
	}
}

func (s *PodmanMachineService) run(
	ctx context.Context,
	podmanPath string,
	arguments ...string,
) (string, error) {
	operationContext, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	return s.runner.Run(operationContext, podmanPath, arguments...)
}

func indicatesMissingPodmanMachine(detail string) bool {
	normalized := strings.ToLower(detail)
	markers := []string{
		"vm does not exist",
		"machine does not exist",
		"no such machine",
		"no vm with name",
		"machine not found",
	}
	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

package onboarding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type DockerInstaller interface {
	Install(ctx context.Context) error
}

type PreparationResult struct {
	Ready     bool   `json:"ready"`
	Installed bool   `json:"installed"`
	Started   bool   `json:"started"`
	Version   string `json:"version,omitempty"`
}

type DockerService struct {
	probe        containerruntime.Probe
	runner       containerruntime.CommandRunner
	installer    DockerInstaller
	platform     string
	dockerApp    string
	pollInterval time.Duration
	timeout      time.Duration
	exists       func(string) bool
}

func NewDockerService(
	probe containerruntime.Probe,
	runner containerruntime.CommandRunner,
	installer DockerInstaller,
	platform string,
) *DockerService {
	return &DockerService{
		probe: probe, runner: runner, installer: installer, platform: platform,
		dockerApp: "/Applications/Docker.app", pollInterval: 2 * time.Second,
		timeout: 3 * time.Minute,
		exists: func(path string) bool {
			info, err := os.Stat(path)
			return err == nil && info.IsDir()
		},
	}
}

func (s *DockerService) Prepare(ctx context.Context) (PreparationResult, error) {
	status := s.probe.Check(ctx)
	if status.State == containerruntime.StateReady {
		return PreparationResult{Ready: true, Version: status.Version}, nil
	}
	if s.platform != "darwin" {
		return PreparationResult{}, fmt.Errorf("automatic runtime preparation is not implemented for %s", s.platform)
	}

	result := PreparationResult{}
	if !s.exists(s.dockerApp) {
		if s.installer == nil {
			return result, fmt.Errorf("Docker Desktop installer is unavailable")
		}
		if err := s.installer.Install(ctx); err != nil {
			return result, err
		}
		result.Installed = true
	}
	if err := s.start(ctx); err != nil {
		return result, err
	}
	result.Started = true
	return s.waitForReady(ctx, result)
}

// Recover starts only an already installed runtime. It is safe for login-time
// recovery because it never invokes the installer or requests elevation.
func (s *DockerService) Recover(ctx context.Context) (PreparationResult, error) {
	status := s.probe.Check(ctx)
	if status.State == containerruntime.StateReady {
		return PreparationResult{Ready: true, Version: status.Version}, nil
	}
	if s.platform != "darwin" {
		return PreparationResult{}, fmt.Errorf(
			"automatic runtime recovery is not implemented for %s",
			s.platform,
		)
	}
	if !s.exists(s.dockerApp) {
		return PreparationResult{}, fmt.Errorf("installed Docker Desktop was not found")
	}
	if err := s.start(ctx); err != nil {
		return PreparationResult{}, err
	}
	return s.waitForReady(ctx, PreparationResult{Started: true})
}

func (s *DockerService) waitForReady(
	ctx context.Context,
	result PreparationResult,
) (PreparationResult, error) {
	waitContext, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		status := s.probe.Check(waitContext)
		if status.State == containerruntime.StateReady {
			result.Ready = true
			result.Version = status.Version
			return result, nil
		}
		select {
		case <-waitContext.Done():
			return result, fmt.Errorf("runtime did not become ready: %w", waitContext.Err())
		case <-ticker.C:
		}
	}
}

func (s *DockerService) start(ctx context.Context) error {
	dockerPath, pathErr := s.runner.LookPath("docker")
	if pathErr == nil {
		if _, err := s.runner.Run(ctx, dockerPath, "desktop", "start"); err == nil {
			return nil
		}
	}
	if _, err := s.runner.Run(ctx, "/usr/bin/open", "-a", s.dockerApp); err != nil {
		if pathErr != nil {
			return errors.Join(fmt.Errorf("find Docker client: %w", pathErr), fmt.Errorf("open Docker Desktop: %w", err))
		}
		return fmt.Errorf("open Docker Desktop: %w", err)
	}
	return nil
}

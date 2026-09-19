package onboarding

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

//go:embed linux_engine_install.sh
var linuxEngineInstallScript string

// LinuxEngineInstaller installs signed official APT packages. Permission to use
// the Docker socket is deliberately not granted by changing account groups.
type LinuxEngineInstaller struct {
	Runner   containerruntime.CommandRunner
	Platform string
}

func (i LinuxEngineInstaller) Install(ctx context.Context) error {
	if i.Platform != "linux" {
		return fmt.Errorf("native Docker Engine installation requires Linux")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	_, err := runLinuxPrivileged(ctx, i.Runner, "/bin/sh", "-c", linuxEngineInstallScript)
	return err
}

func runLinuxPrivileged(ctx context.Context, runner containerruntime.CommandRunner, command string, args ...string) (string, error) {
	if os.Geteuid() == 0 {
		return runner.Run(ctx, command, args...)
	}
	pkexec, err := runner.LookPath("pkexec")
	if err != nil {
		return "", fmt.Errorf("administrator authorization is unavailable: run the CLI installer with sudo, or install a PolicyKit authentication agent")
	}
	return runner.Run(ctx, pkexec, append([]string{command}, args...)...)
}

type LinuxEngineService struct {
	Probe     containerruntime.Probe
	Runner    containerruntime.CommandRunner
	Installer DockerInstaller
}

func (s *LinuxEngineService) Prepare(ctx context.Context) (PreparationResult, error) {
	status := s.Probe.Check(ctx)
	if status.State == containerruntime.StateReady {
		return PreparationResult{Ready: true, Version: status.Version}, nil
	}
	result := PreparationResult{}
	switch status.State {
	case containerruntime.StateUnavailable:
		if s.Installer == nil {
			return result, fmt.Errorf("native Docker Engine installer is unavailable")
		}
		if err := s.Installer.Install(ctx); err != nil {
			return result, err
		}
		result.Installed = true
	case containerruntime.StateStopped:
		startCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		if _, err := runLinuxPrivileged(startCtx, s.Runner, "/usr/bin/systemctl", "start", "docker"); err != nil {
			return result, err
		}
		result.Started = true
	default:
		return result, fmt.Errorf("cannot access Docker Engine: %s. Ask your administrator to grant Docker socket access, then sign out and back in", status.TechnicalDetail)
	}
	status = s.Probe.Check(ctx)
	result.Ready = status.State == containerruntime.StateReady
	result.Version = status.Version
	if !result.Ready {
		result.Message = "Docker Engine was prepared but is not accessible to this account. Ask your administrator to grant Docker socket access, then sign out and back in. Adding an account to the docker group grants root-equivalent privileges."
	}
	return result, nil
}

// Recover never installs software or requests administrator authorization.
func (s *LinuxEngineService) Recover(ctx context.Context) (PreparationResult, error) {
	status := s.Probe.Check(ctx)
	if status.State != containerruntime.StateReady {
		return PreparationResult{}, fmt.Errorf("native Docker Engine is not ready; start its system service or check socket permissions")
	}
	return PreparationResult{Ready: true, Version: status.Version}, nil
}

package execution

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/woliveiras/corsarr/internal/hostreadiness"
	"github.com/woliveiras/corsarr/internal/onboarding"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type Lifecycle interface {
	Prepare(context.Context) (onboarding.PreparationResult, error)
	Recover(context.Context) (onboarding.PreparationResult, error)
}

type Environment struct {
	Selection              *Selection
	Runner                 *Runner
	Probe                  containerruntime.Probe
	Manager                containerruntime.Manager
	Desktop                Lifecycle
	DesktopHost            hostreadiness.Checker
	Platform, Architecture string
}

func New(selection *Selection, base EnvironmentRunner, platform, architecture string) *Environment {
	runner := &Runner{Selection: selection, Base: base}
	return &Environment{Selection: selection, Runner: runner,
		Probe:    linuxProbe{runner: runner, docker: containerruntime.NewDockerDetector(runner, 5*time.Second)},
		Manager:  containerruntime.NewDockerManager(runner, 10*time.Minute),
		Platform: platform, Architecture: architecture}
}

func (e *Environment) Prepare(ctx context.Context) (onboarding.PreparationResult, error) {
	switch e.Selection.Config().Kind {
	case Desktop:
		if e.Desktop == nil {
			return onboarding.PreparationResult{}, fmt.Errorf("automatic Docker Desktop preparation is unavailable on this platform")
		}
		return e.Desktop.Prepare(ctx)
	case Engine:
		if status := e.Check(ctx); !status.Ready {
			return onboarding.PreparationResult{}, fmt.Errorf("%s", strings.Join(status.Issues, "; "))
		}
		return e.engine().Prepare(ctx)
	default:
		return e.existing(ctx)
	}
}

func (e *Environment) Recover(ctx context.Context) (onboarding.PreparationResult, error) {
	if e.Selection.Config().Kind == Desktop && e.Desktop != nil {
		return e.Desktop.Recover(ctx)
	}
	return e.existing(ctx)
}

func (e *Environment) existing(ctx context.Context) (onboarding.PreparationResult, error) {
	status := e.Probe.Check(ctx)
	if status.State != containerruntime.StateReady {
		return onboarding.PreparationResult{}, fmt.Errorf("selected Docker environment is not ready; start it externally and check permissions: %s", status.TechnicalDetail)
	}
	return onboarding.PreparationResult{Ready: true, Version: status.Version}, nil
}

func (e *Environment) engine() *onboarding.LinuxEngineService {
	return &onboarding.LinuxEngineService{Probe: e.Probe, Runner: e.Runner.Base,
		Installer: onboarding.LinuxEngineInstaller{Runner: e.Runner.Base, Platform: e.Platform}}
}

// Check assesses prerequisites for the selected environment, rather than
// applying Docker Desktop's macOS installation requirements to every runtime.
func (e *Environment) Check(ctx context.Context) hostreadiness.Status {
	if e.Selection.Config().Kind == Desktop && e.DesktopHost != nil {
		return e.DesktopHost.Check(ctx)
	}
	status := hostreadiness.Status{Ready: true, Issues: []string{}}
	if e.Architecture != "amd64" && e.Architecture != "arm64" {
		status.Issues = append(status.Issues, "unsupported architecture")
	}
	if e.Selection.Config().Kind == Engine {
		data, err := os.ReadFile("/etc/os-release")
		if err != nil || !SupportedLinuxRelease(string(data)) {
			status.Issues = append(status.Issues, "automatic installation supports Ubuntu 22.04/24.04/26.04 and Debian 12/13")
		}
		if _, err := os.Stat("/run/systemd/system"); err != nil {
			status.Issues = append(status.Issues, "Docker Engine preparation requires a running systemd system")
		}
	}
	status.Ready = len(status.Issues) == 0
	return status
}

func SupportedLinuxRelease(contents string) bool {
	values := map[string]string{}
	for _, line := range strings.Split(contents, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[key] = strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}
	switch values["ID"] + ":" + values["VERSION_CODENAME"] {
	case "ubuntu:jammy", "ubuntu:noble", "ubuntu:resolute", "debian:bookworm", "debian:trixie":
		return true
	default:
		return false
	}
}

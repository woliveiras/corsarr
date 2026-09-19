package onboarding

import (
	"context"
	"fmt"
	"testing"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type engineProbe struct {
	states []containerruntime.State
	count  int
}

func (p *engineProbe) Check(context.Context) containerruntime.Status {
	index := p.count
	if index >= len(p.states) {
		index = len(p.states) - 1
	}
	p.count++
	return containerruntime.Status{Provider: containerruntime.ProviderDocker, State: p.states[index], TechnicalDetail: "permission denied"}
}

type engineInstaller struct {
	count int
	err   error
}

func (i *engineInstaller) Install(context.Context) error { i.count++; return i.err }

func TestEngineInstallationReportsPermissionGapWithoutReinstalling(t *testing.T) {
	probe := &engineProbe{states: []containerruntime.State{containerruntime.StateUnavailable, containerruntime.StateError}}
	installer := &engineInstaller{}
	service := LinuxEngineService{Probe: probe, Installer: installer}
	result, err := service.Prepare(context.Background())
	if err != nil || !result.Installed || result.Ready || result.Message == "" {
		t.Fatalf("permission gap lost: %#v %v", result, err)
	}
	if _, err := service.Prepare(context.Background()); err == nil {
		t.Fatal("expected actionable access error")
	}
	if installer.count != 1 {
		t.Fatal("reinstalled Docker to fix account permissions")
	}
}

func TestEngineRecoveryCannotInstallOrElevate(t *testing.T) {
	installer := &engineInstaller{}
	service := LinuxEngineService{Probe: &engineProbe{states: []containerruntime.State{containerruntime.StateUnavailable}}, Installer: installer}
	if _, err := service.Recover(context.Background()); err == nil {
		t.Fatal("expected stopped error")
	}
	if installer.count != 0 {
		t.Fatal("background recovery installed Docker")
	}
}

func TestEngineDoesNotClaimInstallationSuccessOnFailure(t *testing.T) {
	service := LinuxEngineService{Probe: &engineProbe{states: []containerruntime.State{containerruntime.StateUnavailable}}, Installer: &engineInstaller{err: fmt.Errorf("authorization canceled")}}
	result, err := service.Prepare(context.Background())
	if err == nil || result.Installed || result.Ready {
		t.Fatalf("failed installation claimed success: %#v %v", result, err)
	}
}

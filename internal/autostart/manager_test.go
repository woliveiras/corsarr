package autostart

import (
	"errors"
	"testing"
)

func TestMainAppManagerRegistersAndUnregistersIdempotently(t *testing.T) {
	service := &fakeNativeService{status: nativeStatusNotRegistered}
	manager := &mainAppManager{service: service}

	enabled, err := manager.SetEnabled(true)
	if err != nil || !enabled.Enabled || !enabled.Supported {
		t.Fatalf("enable login item: %#v, %v", enabled, err)
	}
	if service.registerCalls != 1 {
		t.Fatalf("expected one registration, got %d", service.registerCalls)
	}
	if _, err := manager.SetEnabled(true); err != nil {
		t.Fatalf("idempotent enable: %v", err)
	}
	if service.registerCalls != 1 {
		t.Fatalf("idempotent enable registered again: %d", service.registerCalls)
	}

	disabled, err := manager.SetEnabled(false)
	if err != nil || disabled.Enabled || !disabled.Supported {
		t.Fatalf("disable login item: %#v, %v", disabled, err)
	}
	if service.unregisterCalls != 1 {
		t.Fatalf("expected one unregistration, got %d", service.unregisterCalls)
	}
}

func TestMainAppManagerReportsApprovalRequired(t *testing.T) {
	service := &fakeNativeService{
		status:              nativeStatusNotRegistered,
		statusAfterRegister: nativeStatusRequiresApproval,
	}
	manager := &mainAppManager{service: service}

	status, err := manager.SetEnabled(true)
	if err != nil {
		t.Fatalf("register approval-pending login item: %v", err)
	}
	if !status.Supported || status.Enabled || !status.RequiresApproval {
		t.Fatalf("unexpected approval status %#v", status)
	}
	if err := manager.OpenSystemSettings(); err != nil {
		t.Fatalf("open login item settings: %v", err)
	}
	if service.openSettingsCalls != 1 {
		t.Fatalf("expected one System Settings request, got %d", service.openSettingsCalls)
	}
}

func TestMainAppManagerDoesNotHideRegistrationFailure(t *testing.T) {
	service := &fakeNativeService{
		status:      nativeStatusNotRegistered,
		registerErr: errors.New("registration denied"),
	}
	manager := &mainAppManager{service: service}

	status, err := manager.SetEnabled(true)
	if err == nil || status.Enabled {
		t.Fatalf("expected registration failure, status=%#v err=%v", status, err)
	}
}

func TestPlatformManagerIsUnsupportedOutsideMacOS(t *testing.T) {
	manager := NewPlatformManager("linux")
	status, err := manager.Status()
	if err != nil || status.Supported {
		t.Fatalf("unexpected unsupported status %#v, %v", status, err)
	}
	if _, err := manager.SetEnabled(true); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

type fakeNativeService struct {
	status              nativeStatus
	statusAfterRegister nativeStatus
	registerErr         error
	unregisterErr       error
	registerCalls       int
	unregisterCalls     int
	openSettingsCalls   int
}

func (f *fakeNativeService) Status() nativeStatus { return f.status }

func (f *fakeNativeService) Register() error {
	f.registerCalls++
	if f.registerErr == nil {
		if f.statusAfterRegister != 0 {
			f.status = f.statusAfterRegister
		} else {
			f.status = nativeStatusEnabled
		}
	}
	return f.registerErr
}

func (f *fakeNativeService) Unregister() error {
	f.unregisterCalls++
	if f.unregisterErr == nil {
		f.status = nativeStatusNotRegistered
	}
	return f.unregisterErr
}

func (f *fakeNativeService) OpenSystemSettings() { f.openSettingsCalls++ }

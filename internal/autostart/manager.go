package autostart

import (
	"errors"
	"fmt"
)

var ErrUnsupported = errors.New("start at login is not supported on this platform")

type Status struct {
	Supported        bool `json:"supported"`
	Enabled          bool `json:"enabled"`
	RequiresApproval bool `json:"requiresApproval"`
}

type Manager interface {
	Status() (Status, error)
	SetEnabled(enabled bool) (Status, error)
	OpenSystemSettings() error
}

type nativeStatus int

const (
	nativeStatusUnsupported      nativeStatus = -1
	nativeStatusNotRegistered    nativeStatus = 0
	nativeStatusEnabled          nativeStatus = 1
	nativeStatusRequiresApproval nativeStatus = 2
	nativeStatusNotFound         nativeStatus = 3
)

type nativeMainAppService interface {
	Status() nativeStatus
	Register() error
	Unregister() error
	OpenSystemSettings()
}

type mainAppManager struct {
	service nativeMainAppService
}

func NewPlatformManager(platform string) Manager {
	if platform != "darwin" {
		return unsupportedManager{}
	}
	return &mainAppManager{service: newNativeMainAppService()}
}

func (m *mainAppManager) Status() (Status, error) {
	return statusFromNative(m.service.Status())
}

func (m *mainAppManager) SetEnabled(enabled bool) (Status, error) {
	before, err := m.Status()
	if err != nil {
		return Status{}, err
	}
	if !before.Supported {
		return before, ErrUnsupported
	}
	if enabled && (before.Enabled || before.RequiresApproval) {
		return before, nil
	}
	if !enabled && !before.Enabled && !before.RequiresApproval {
		return before, nil
	}

	if enabled {
		err = m.service.Register()
	} else {
		err = m.service.Unregister()
	}
	after, statusErr := m.Status()
	if err != nil {
		if statusErr == nil && after.Enabled == enabled && !after.RequiresApproval {
			return after, nil
		}
		return after, fmt.Errorf("update start-at-login registration: %w", err)
	}
	if statusErr != nil {
		return Status{}, statusErr
	}
	return after, nil
}

func (m *mainAppManager) OpenSystemSettings() error {
	status, err := m.Status()
	if err != nil {
		return err
	}
	if !status.Supported {
		return ErrUnsupported
	}
	m.service.OpenSystemSettings()
	return nil
}

func statusFromNative(value nativeStatus) (Status, error) {
	switch value {
	case nativeStatusUnsupported:
		return Status{}, nil
	case nativeStatusNotRegistered:
		return Status{Supported: true}, nil
	case nativeStatusEnabled:
		return Status{Supported: true, Enabled: true}, nil
	case nativeStatusRequiresApproval:
		return Status{Supported: true, RequiresApproval: true}, nil
	case nativeStatusNotFound:
		return Status{Supported: true}, nil
	default:
		return Status{}, fmt.Errorf("unexpected macOS login item status: %d", value)
	}
}

type unsupportedManager struct{}

func (unsupportedManager) Status() (Status, error) { return Status{}, nil }

func (unsupportedManager) SetEnabled(bool) (Status, error) {
	return Status{}, ErrUnsupported
}

func (unsupportedManager) OpenSystemSettings() error { return ErrUnsupported }

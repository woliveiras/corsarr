package main

import (
	"fmt"

	"github.com/woliveiras/corsarr/internal/execution"
)

func (a *App) GetExecutionConfiguration() execution.Config {
	if a.executionEnvironment == nil {
		return execution.Config{Kind: execution.Desktop}
	}
	return a.executionEnvironment.Selection.Config()
}

// SaveExecutionConfiguration is only available before storage selection. It
// cannot silently move an existing library or adopt another server's state.
func (a *App) SaveExecutionConfiguration(config execution.Config) (execution.Config, error) {
	release, err := a.beginChange()
	if err != nil {
		return execution.Config{}, err
	}
	defer release()
	if a.executionEnvironment == nil {
		return execution.Config{}, fmt.Errorf("execution configuration is unavailable")
	}
	current := a.executionEnvironment.Selection.Config()
	if current == config {
		return current, nil
	}
	setup, err := a.setup.Load()
	if err != nil {
		return execution.Config{}, err
	}
	if setup.StoragePath != "" || len(setup.Applications) > 0 || setup.OnboardingCompleted {
		return execution.Config{}, fmt.Errorf("execution environment is locked after storage or applications have been selected")
	}
	if err := config.Validate(a.executionEnvironment.Platform); err != nil {
		return execution.Config{}, err
	}
	if err := a.setup.ClearRuntimeConsent(); err != nil {
		return execution.Config{}, err
	}
	if err := a.executionEnvironment.Selection.Save(config); err != nil {
		return execution.Config{}, err
	}
	return config, nil
}

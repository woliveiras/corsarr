package main

import (
	"path/filepath"
	"testing"

	"github.com/woliveiras/corsarr/internal/application"
	"github.com/woliveiras/corsarr/internal/execution"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

func TestDesktopExecutionSelectionUpdatesSharedRunnerAndLocksAfterStorage(t *testing.T) {
	selection, err := execution.Open(filepath.Join(t.TempDir(), "execution.json"), "linux")
	if err != nil {
		t.Fatal(err)
	}
	setup := &desktopSetupManager{}
	app := &App{setup: setup, executionEnvironment: execution.New(selection, containerruntime.OSCommandRunner{}, "linux", "amd64")}
	want := execution.Config{Kind: execution.Existing, Context: "rootless"}
	if _, err := app.SaveExecutionConfiguration(want); err != nil {
		t.Fatal(err)
	}
	if app.executionEnvironment.Runner.Selection.Config() != want {
		t.Fatal("runner did not follow selection")
	}
	setup.status = application.SetupStatus{StoragePath: "/data"}
	if _, err := app.SaveExecutionConfiguration(execution.Config{Kind: execution.Engine}); err == nil {
		t.Fatal("changed environment after storage selection")
	}
	if app.GetExecutionConfiguration() != want {
		t.Fatal("failed change modified active environment")
	}
}

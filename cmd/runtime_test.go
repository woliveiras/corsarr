package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/woliveiras/corsarr/internal/execution"
	"github.com/woliveiras/corsarr/internal/state"
)

func TestRuntimeInstallRequiresExplicitConsent(t *testing.T) {
	command := newRuntimeCommand()
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})
	command.SetArgs([]string{"install"})
	if err := command.Execute(); err == nil {
		t.Fatal("installer accepted missing consent")
	}
}

func TestRuntimeSelectionCannotMoveExistingDesktopStorage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "desktop-state.json")
	if err := state.NewFileStore(path).Save(state.DesktopState{StoragePath: filepath.Join(dir, "media")}); err != nil {
		t.Fatal(err)
	}
	err := prepareExecutionChange(path, execution.Config{Kind: execution.Desktop}, execution.Config{Kind: execution.Existing, Context: "default"})
	if err == nil {
		t.Fatal("switched destination underneath an existing library")
	}
}

func TestRuntimeChangeClearsPriorConsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "desktop-state.json")
	store := state.NewFileStore(path)
	if err := store.Save(state.DesktopState{Language: "it", RuntimeConsentVersion: "old", RuntimeConsentAcceptedAt: "yesterday"}); err != nil {
		t.Fatal(err)
	}
	if err := prepareExecutionChange(path, execution.Config{Kind: execution.Desktop}, execution.Config{Kind: execution.Existing, Context: "default"}); err != nil {
		t.Fatal(err)
	}
	saved, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if saved.RuntimeConsentVersion != "" || saved.RuntimeConsentAcceptedAt != "" || saved.Language != "it" {
		t.Fatalf("incorrect consent reset: %#v", saved)
	}
}

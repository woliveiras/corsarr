package main

import "testing"

func TestNewAppExposesCatalogWithoutArbitraryURLs(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("create desktop app: %v", err)
	}

	if applications := app.ListApplications(); len(applications) == 0 {
		t.Fatal("expected desktop application catalog")
	}

	if err := app.OpenApplication("https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary URL to be rejected before reaching Wails runtime")
	}
}

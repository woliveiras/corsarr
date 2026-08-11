package buildinfo

import "testing"

func TestCurrentUsesInjectedVersion(t *testing.T) {
	previous := Version
	Version = "v1.2.0"
	t.Cleanup(func() { Version = previous })

	if got := Current(); got != "1.2.0" {
		t.Fatalf("Current() = %q, want %q", got, "1.2.0")
	}
}

func TestCurrentIdentifiesLocalBuild(t *testing.T) {
	previous := Version
	Version = "development"
	t.Cleanup(func() { Version = previous })

	if got := Current(); got != "development" {
		t.Fatalf("Current() = %q, want %q", got, "development")
	}
}

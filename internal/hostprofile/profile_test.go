package hostprofile

import (
	"errors"
	"testing"
)

func TestProfilerUsesValidatedTZEnvironment(t *testing.T) {
	profiler := testProfiler()
	profiler.getenv = func(name string) string {
		if name == "TZ" {
			return "America/Sao_Paulo"
		}
		return ""
	}

	profile := profiler.Current("darwin")
	if profile.Timezone != "America/Sao_Paulo" {
		t.Fatalf("unexpected timezone %q", profile.Timezone)
	}
}

func TestProfilerFindsIANAZoneFromLocaltimeSymlink(t *testing.T) {
	profiler := testProfiler()
	profiler.evalSymlinks = func(string) (string, error) {
		return "/var/db/timezone/zoneinfo/Europe/Madrid", nil
	}

	profile := profiler.Current("darwin")
	if profile.Timezone != "Europe/Madrid" {
		t.Fatalf("unexpected timezone %q", profile.Timezone)
	}
}

func TestProfilerUsesCurrentLinuxIdentityForEveryContainer(t *testing.T) {
	profiler := testProfiler()
	profiler.currentUser = func() (string, string, error) { return "1001", "1002", nil }

	profile := profiler.Current("linux")
	if profile.PUID != 1001 || profile.PGID != 1002 {
		t.Fatalf("unexpected Linux identity %#v", profile)
	}
}

func TestProfilerUsesStableVMIdentityOutsideNativeLinux(t *testing.T) {
	profiler := testProfiler()
	profiler.currentUser = func() (string, string, error) {
		return "502", "20", errors.New("must not be called")
	}

	for _, platform := range []string{"darwin", "windows"} {
		profile := profiler.Current(platform)
		if profile.PUID != 1000 || profile.PGID != 1000 {
			t.Fatalf("unexpected %s VM identity %#v", platform, profile)
		}
	}
}

func TestProfilerFallsBackSafelyWhenHostMetadataIsUnavailable(t *testing.T) {
	profiler := testProfiler()
	profiler.getenv = func(string) string { return "../../unsafe" }
	profiler.evalSymlinks = func(string) (string, error) { return "", errors.New("missing") }
	profiler.currentUser = func() (string, string, error) { return "invalid", "invalid", nil }

	profile := profiler.Current("linux")
	if profile.Timezone != "Etc/UTC" || profile.PUID != 1000 || profile.PGID != 1000 {
		t.Fatalf("unexpected fallback profile %#v", profile)
	}
}

func testProfiler() *Profiler {
	profiler := NewProfiler()
	profiler.getenv = func(string) string { return "" }
	profiler.evalSymlinks = func(string) (string, error) { return "", errors.New("missing") }
	profiler.currentUser = func() (string, string, error) { return "1000", "1000", nil }
	return profiler
}

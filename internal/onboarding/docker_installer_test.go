package onboarding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacDockerInstallerVerifiesAndRunsOfficialInstaller(t *testing.T) {
	content := []byte("signed docker disk image")
	digest := sha256.Sum256(content)
	runner := &installerRunner{t: t}
	installer := &MacDockerInstaller{
		runner: runner, downloader: installerDownloader{content: content},
		artifact:  DockerArtifact{URL: "https://desktop.docker.com/Docker.dmg", SHA256: hex.EncodeToString(digest[:])},
		cacheRoot: filepath.Join(t.TempDir(), "cache"), username: "william",
		dockerApp:   filepath.Join(t.TempDir(), "Applications", "Docker.app"),
		hdiutilPath: "hdiutil", codesign: "codesign", spctl: "spctl", osascript: "osascript",
	}

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install Docker Desktop: %v", err)
	}
	joined := strings.Join(runner.operations, "\n")
	for _, expected := range []string{"hdiutil attach", "codesign -dv", "codesign --verify", "spctl --assess", "osascript -e", "--accept-license", "--user='william'", "hdiutil detach"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected %q in operations:\n%s", expected, joined)
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(installer.cacheRoot, "docker-desktop-*")); len(matches) != 0 {
		t.Fatalf("expected installer workspace cleanup, got %v", matches)
	}
}

func TestMacDockerInstallerRejectsChecksumBeforeMount(t *testing.T) {
	runner := &installerRunner{t: t}
	installer := &MacDockerInstaller{
		runner: runner, downloader: installerDownloader{content: []byte("tampered")},
		artifact:  DockerArtifact{URL: "https://desktop.docker.com/Docker.dmg", SHA256: strings.Repeat("a", 64)},
		cacheRoot: filepath.Join(t.TempDir(), "cache"), username: "william",
		dockerApp: filepath.Join(t.TempDir(), "Applications", "Docker.app"),
	}

	if err := installer.Install(context.Background()); err == nil {
		t.Fatal("expected checksum rejection")
	}
	if len(runner.operations) != 0 {
		t.Fatalf("expected no mounted or executed content, got %v", runner.operations)
	}
}

type installerDownloader struct{ content []byte }

func (d installerDownloader) Download(_ context.Context, _ string, destination string) error {
	return os.WriteFile(destination, d.content, 0o600)
}

type installerRunner struct {
	t          *testing.T
	operations []string
}

func (r *installerRunner) LookPath(file string) (string, error) { return file, nil }

func (r *installerRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	operation := strings.Join(append([]string{name}, args...), " ")
	r.operations = append(r.operations, operation)
	if name == "hdiutil" && len(args) > 0 && args[0] == "attach" {
		mountPath := args[len(args)-1]
		installer := filepath.Join(mountPath, "Docker.app", "Contents", "MacOS", "install")
		if err := os.MkdirAll(filepath.Dir(installer), 0o700); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(installer, []byte("installer"), 0o700); err != nil {
			r.t.Fatal(err)
		}
	}
	if name == "codesign" && len(args) > 0 && args[0] == "-dv" {
		return "Authority=Developer ID Application: Docker Inc (9BNSXJN65R)\nTeamIdentifier=9BNSXJN65R", nil
	}
	return "", nil
}

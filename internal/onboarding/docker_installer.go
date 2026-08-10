package onboarding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

const (
	DockerDesktopVersion = "4.86.0"
	dockerTeamIdentifier = "9BNSXJN65R"
	maxInstallerBytes    = 2 << 30
)

type DockerArtifact struct {
	URL    string
	SHA256 string
}

var DockerMacArtifacts = map[string]DockerArtifact{
	"arm64": {
		URL:    "https://desktop.docker.com/mac/main/arm64/236216/Docker.dmg",
		SHA256: "4b662c088e9cb4dfcfb59a64c6ea9ff39063af303a660c604e6f4a5eec5c5fe4",
	},
	"amd64": {
		URL:    "https://desktop.docker.com/mac/main/amd64/236216/Docker.dmg",
		SHA256: "3acbd44ef3e2f95deab09fbddac82b977af453f5762986baed4bd47151ae348b",
	},
}

type InstallerDownloader interface {
	Download(ctx context.Context, sourceURL, destination string) error
}

type HTTPDownloader struct {
	client *http.Client
}

func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{client: &http.Client{Timeout: 30 * time.Minute}}
}

func (d *HTTPDownloader) Download(ctx context.Context, sourceURL, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("official installer download returned %s", response.Status)
	}
	if response.Request.URL.Scheme != "https" || response.Request.URL.Hostname() != "desktop.docker.com" {
		return fmt.Errorf("official installer download redirected outside desktop.docker.com")
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, maxInstallerBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written > maxInstallerBytes {
		return fmt.Errorf("official installer exceeds the maximum accepted size")
	}
	return nil
}

type MacDockerInstaller struct {
	runner      containerruntime.CommandRunner
	downloader  InstallerDownloader
	artifact    DockerArtifact
	cacheRoot   string
	username    string
	dockerApp   string
	hdiutilPath string
	codesign    string
	spctl       string
	osascript   string
}

func NewMacDockerInstaller(
	runner containerruntime.CommandRunner,
	downloader InstallerDownloader,
	architecture string,
	cacheRoot string,
	username string,
) (*MacDockerInstaller, error) {
	artifact, exists := DockerMacArtifacts[architecture]
	if !exists {
		return nil, fmt.Errorf("Docker Desktop has no approved macOS artifact for %s", architecture)
	}
	if !regexp.MustCompile(`^[A-Za-z0-9._-]+$`).MatchString(username) {
		return nil, fmt.Errorf("macOS username contains unsupported characters")
	}
	return &MacDockerInstaller{
		runner: runner, downloader: downloader, artifact: artifact,
		cacheRoot: cacheRoot, username: username, dockerApp: "/Applications/Docker.app",
		hdiutilPath: "/usr/bin/hdiutil", codesign: "/usr/bin/codesign",
		spctl: "/usr/sbin/spctl", osascript: "/usr/bin/osascript",
	}, nil
}

func (i *MacDockerInstaller) Install(ctx context.Context) (resultErr error) {
	if info, err := os.Stat(i.dockerApp); err == nil && info.IsDir() {
		return nil
	}
	if err := os.MkdirAll(i.cacheRoot, 0o700); err != nil {
		return fmt.Errorf("create private installer cache: %w", err)
	}
	workPath, err := os.MkdirTemp(i.cacheRoot, "docker-desktop-")
	if err != nil {
		return fmt.Errorf("create installer workspace: %w", err)
	}
	defer os.RemoveAll(workPath)
	diskImage := filepath.Join(workPath, "Docker.dmg")
	if err := i.downloader.Download(ctx, i.artifact.URL, diskImage); err != nil {
		return fmt.Errorf("download official Docker Desktop installer: %w", err)
	}
	if err := verifyFileSHA256(diskImage, i.artifact.SHA256); err != nil {
		return fmt.Errorf("verify Docker Desktop installer checksum: %w", err)
	}

	mountPath := filepath.Join(workPath, "mount")
	if err := os.Mkdir(mountPath, 0o700); err != nil {
		return fmt.Errorf("create installer mount point: %w", err)
	}
	if _, err := i.runner.Run(ctx, i.hdiutilPath, "attach", diskImage, "-nobrowse", "-readonly", "-mountpoint", mountPath); err != nil {
		return fmt.Errorf("mount Docker Desktop installer: %w", err)
	}
	defer func() {
		detachContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Minute)
		defer cancel()
		if _, err := i.runner.Run(detachContext, i.hdiutilPath, "detach", mountPath); err != nil {
			resultErr = errorsJoin(resultErr, fmt.Errorf("detach Docker Desktop installer: %w", err))
		}
	}()

	mountedApp := filepath.Join(mountPath, "Docker.app")
	info, err := os.Lstat(mountedApp)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("mounted installer does not contain a regular Docker.app")
	}
	identity, err := i.runner.Run(ctx, i.codesign, "-dv", "--verbose=4", mountedApp)
	if err != nil {
		return fmt.Errorf("verify Docker Desktop code signature: %w", err)
	}
	if !strings.Contains(identity, "TeamIdentifier="+dockerTeamIdentifier) {
		return fmt.Errorf("Docker Desktop signer identity is not approved")
	}
	if _, err := i.runner.Run(ctx, i.codesign, "--verify", "--deep", "--strict", mountedApp); err != nil {
		return fmt.Errorf("validate Docker Desktop signed contents: %w", err)
	}
	if _, err := i.runner.Run(ctx, i.spctl, "--assess", "--type", "execute", mountedApp); err != nil {
		return fmt.Errorf("validate Docker Desktop with macOS security policy: %w", err)
	}

	installer := filepath.Join(mountedApp, "Contents", "MacOS", "install")
	installerInfo, err := os.Lstat(installer)
	if err != nil || installerInfo.Mode()&os.ModeSymlink != 0 || !installerInfo.Mode().IsRegular() {
		return fmt.Errorf("Docker Desktop installer executable is missing")
	}
	command := shellQuote(installer) + " --accept-license --user=" + shellQuote(i.username)
	appleScript := "do shell script " + strconv.Quote(command) + " with administrator privileges"
	if _, err := i.runner.Run(ctx, i.osascript, "-e", appleScript); err != nil {
		return fmt.Errorf("install Docker Desktop with macOS authorization: %w", err)
	}
	return nil
}

func verifyFileSHA256(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func errorsJoin(primary, secondary error) error {
	if primary == nil {
		return secondary
	}
	return fmt.Errorf("%v; %w", primary, secondary)
}

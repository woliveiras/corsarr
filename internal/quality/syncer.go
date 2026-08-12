package quality

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/woliveiras/corsarr/internal/credentials"
	"github.com/woliveiras/corsarr/internal/provisioning"
)

const (
	RecyclarrImage    = "ghcr.io/recyclarr/recyclarr@sha256:2d6107f758d882a59fe9d646aa54fa8a5a4fb7a40995125fade575652a3f7871"
	RecyclarrVersion  = "8.7.0"
	TrashGuidesCommit = "0943b2677a0454d7d69fc9697a8ddcdb2eebd8d9"
)

var ErrSyncFailed = errors.New("quality profile synchronization failed")

type Runner interface {
	Run(ctx context.Context, environment map[string]string, arguments ...string) error
}

type CredentialSource interface {
	Read(rootPath, applicationID string) (credentials.Secret, error)
}

type APIKeyReader interface {
	Read(rootPath, applicationID string) (provisioning.APIKey, error)
}

type ARRCredentialSource struct{ reader APIKeyReader }

func NewARRCredentialSource(reader APIKeyReader) *ARRCredentialSource {
	return &ARRCredentialSource{reader: reader}
}

func (s *ARRCredentialSource) Read(rootPath, applicationID string) (credentials.Secret, error) {
	key, err := s.reader.Read(rootPath, applicationID)
	if err != nil {
		return credentials.Secret{}, err
	}
	return credentials.NewSecret(key.Reveal()), nil
}

type Request struct {
	RootPath     string
	Applications []string
	Preset       PresetID
	PUID         int
	PGID         int
}

type Result struct {
	Preset       PresetID `json:"preset"`
	ProfileName  string   `json:"profileName,omitempty"`
	Applications []string `json:"applications"`
	Previewed    bool     `json:"previewed"`
	Applied      bool     `json:"applied"`
}

type Syncer struct {
	runner      Runner
	credentials CredentialSource
}

func NewSyncer(runner Runner, credentialSource CredentialSource) *Syncer {
	return &Syncer{runner: runner, credentials: credentialSource}
}

func (s *Syncer) Apply(ctx context.Context, request Request) (Result, error) {
	preset, found := FindPreset(request.Preset)
	if !found {
		return Result{}, fmt.Errorf("unsupported quality preset: %s", request.Preset)
	}
	applications := selectedARRApplications(request.Applications)
	result := Result{
		Preset: request.Preset, ProfileName: ProfileName(request.Preset), Applications: applications,
	}
	if len(applications) == 0 || request.Preset == PresetUnmanaged {
		return result, nil
	}
	if !filepath.IsAbs(request.RootPath) {
		return Result{}, fmt.Errorf("corsarr root path must be absolute")
	}

	configDirectory := filepath.Join(request.RootPath, "config", "recyclarr")
	if err := prepareConfigDirectory(configDirectory); err != nil {
		return Result{}, err
	}
	if err := writePrivateFile(filepath.Join(configDirectory, "recyclarr.yml"), renderConfig(preset, applications)); err != nil {
		return Result{}, err
	}
	if err := writePrivateFile(filepath.Join(configDirectory, "settings.yml"), renderSettings()); err != nil {
		return Result{}, err
	}

	environment := make(map[string]string, len(applications))
	for _, applicationID := range applications {
		key, err := s.credentials.Read(request.RootPath, applicationID)
		if err != nil {
			return Result{}, fmt.Errorf("read %s credential for quality profile sync: %w", applicationID, err)
		}
		environment[strings.ToUpper(applicationID)+"_API_KEY"] = key.Reveal()
	}
	if err := s.runner.Run(ctx, nil, "pull", RecyclarrImage); err != nil {
		return Result{}, syncFailure(err, nil)
	}
	baseArguments := recyclarrArguments(configDirectory, request.PUID, request.PGID, applications)
	previewArguments := append(append([]string(nil), baseArguments...), "--preview")
	if err := s.runner.Run(ctx, environment, previewArguments...); err != nil {
		return Result{}, syncFailure(err, environment)
	}
	result.Previewed = true
	if err := s.runner.Run(ctx, environment, baseArguments...); err != nil {
		return result, syncFailure(err, environment)
	}
	result.Applied = true
	return result, nil
}

func syncFailure(err error, environment map[string]string) error {
	if err == nil {
		return ErrSyncFailed
	}
	if errors.Is(err, ErrSyncFailed) {
		return err
	}
	return fmt.Errorf("%w: %s", ErrSyncFailed, redactEnvironmentValues(err.Error(), environment))
}

func selectedARRApplications(selected []string) []string {
	set := make(map[string]struct{}, 2)
	for _, applicationID := range selected {
		if applicationID == "radarr" || applicationID == "sonarr" {
			set[applicationID] = struct{}{}
		}
	}
	applications := make([]string, 0, len(set))
	for applicationID := range set {
		applications = append(applications, applicationID)
	}
	sort.Strings(applications)
	return applications
}

func prepareConfigDirectory(configDirectory string) error {
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		return fmt.Errorf("create Recyclarr configuration directory: %w", err)
	}
	info, err := os.Lstat(configDirectory)
	if err != nil {
		return fmt.Errorf("inspect Recyclarr configuration directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("Recyclarr configuration path is not a regular directory")
	}
	if err := os.Chmod(configDirectory, 0o700); err != nil {
		return fmt.Errorf("secure Recyclarr configuration directory: %w", err)
	}
	return nil
}

func writePrivateFile(path string, contents string) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".recyclarr-*")
	if err != nil {
		return fmt.Errorf("create temporary Recyclarr configuration: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.WriteString(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace Recyclarr configuration: %w", err)
	}
	return nil
}

func renderSettings() string {
	return "resource_providers:\n" +
		"  - name: corsarr-trash-guides\n" +
		"    type: trash-guides\n" +
		"    clone_url: https://github.com/TRaSH-Guides/Guides.git\n" +
		"    reference: " + TrashGuidesCommit + "\n" +
		"    replace_default: true\n"
}

func renderConfig(preset Preset, applications []string) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Corsarr quality policy %s\n", PresetCatalogVersion)
	for _, applicationID := range applications {
		trashID, qualityType, port := preset.RadarrTrashID, "movie", 7878
		if applicationID == "sonarr" {
			trashID, qualityType, port = preset.SonarrTrashID, "series", 8989
		}
		fmt.Fprintf(&builder, "%s:\n", applicationID)
		builder.WriteString("  corsarr:\n")
		fmt.Fprintf(&builder, "    base_url: http://%s:%d\n", applicationID, port)
		fmt.Fprintf(&builder, "    api_key: !env_var %s_API_KEY\n", strings.ToUpper(applicationID))
		builder.WriteString("    quality_definition:\n")
		fmt.Fprintf(&builder, "      type: %s\n", qualityType)
		builder.WriteString("    quality_profiles:\n")
		fmt.Fprintf(&builder, "      - trash_id: %s\n", trashID)
		fmt.Fprintf(&builder, "        name: %s\n", ProfileName(preset.ID))
		builder.WriteString("        reset_unmatched_scores:\n")
		builder.WriteString("          enabled: false\n")
		if applicationID == "radarr" {
			builder.WriteString("    media_naming:\n")
			builder.WriteString("      folder: default\n")
			builder.WriteString("      movie:\n")
			builder.WriteString("        rename: true\n")
			builder.WriteString("        standard: standard\n")
		} else {
			builder.WriteString("    media_naming:\n")
			builder.WriteString("      season: default\n")
			builder.WriteString("      series: default\n")
			builder.WriteString("      episodes:\n")
			builder.WriteString("        rename: true\n")
			builder.WriteString("        standard: default\n")
			builder.WriteString("        daily: default\n")
			builder.WriteString("        anime: default\n")
		}
	}
	return builder.String()
}

func recyclarrArguments(configDirectory string, puid, pgid int, applications []string) []string {
	user := "1000:1000"
	if puid > 0 && pgid > 0 {
		user = strconv.Itoa(puid) + ":" + strconv.Itoa(pgid)
	}
	arguments := []string{
		"run", "--rm", "--network", "corsarr", "--user", user,
		"--mount", "type=bind,src=" + containerCLIMountPath(filepath.Clean(configDirectory)) + ",dst=/config",
	}
	for _, applicationID := range applications {
		arguments = append(arguments, "--env", strings.ToUpper(applicationID)+"_API_KEY")
	}
	return append(arguments, RecyclarrImage, "sync", "--config", "/config/recyclarr.yml", "--log", "info")
}

func containerCLIMountPath(hostPath string) string {
	if strings.ContainsAny(hostPath, ",\"") {
		return strconv.Quote(hostPath)
	}
	return hostPath
}

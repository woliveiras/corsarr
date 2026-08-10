package catalog

import (
	"fmt"
	"path/filepath"
	"strconv"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
)

const RuntimeCatalogVerifiedAt = "2026-08-10"

type RuntimeOptions struct {
	Timezone string
	PUID     int
	PGID     int
}

type RuntimeManifest struct {
	ApplicationID       string
	Image               string
	HostPort            int
	ContainerPort       int
	ConfigTarget        string
	MediaTarget         string
	SupportsUserMapping bool
	RequiresInit        bool
	SourceURL           string
}

type RuntimeAttribution struct {
	ApprovedImage string
	ImageSource   string
}

type RuntimeCatalog struct {
	manifests map[string]RuntimeManifest
}

type approvedImage struct {
	repository          string
	digest              string
	configTarget        string
	mediaTarget         string
	containerPort       int
	supportsUserMapping bool
	requiresInit        bool
	sourceURL           string
}

var approvedImages = map[string]approvedImage{
	"qbittorrent": {
		repository: "lscr.io/linuxserver/qbittorrent", digest: "sha256:b6ab43fe86039e5bdd3cc0b59b946414fcff0c8183e93636e6cb438fdac45028",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-qbittorrent/",
	},
	"prowlarr": {
		repository: "lscr.io/linuxserver/prowlarr", digest: "sha256:1295cff29d10b486c0d8324d1559a552140a5932bf8b3d87e398654414f63f92",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-prowlarr/",
	},
	"lazylibrarian": {
		repository: "lscr.io/linuxserver/lazylibrarian", digest: "sha256:009eab5c1a7550f5406ea987ba66047d9023e4f58f01979761a389e065f8f99f",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-lazylibrarian/",
	},
	"lidarr": {
		repository: "ghcr.io/hotio/lidarr", digest: "sha256:a2d2774f84decf17e6405faf978bfefc7a6793f7d9fbc97c0ec8d3fbff37f51c",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://hotio.dev/containers/lidarr/",
	},
	"radarr": {
		repository: "lscr.io/linuxserver/radarr", digest: "sha256:a45b5ab0f850f39edb4cc9c95bbd967b52ddc3d4574a4dfb45561177db6c88f4",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-radarr/",
	},
	"sonarr": {
		repository: "lscr.io/linuxserver/sonarr", digest: "sha256:373159ba768e23a3a1c497d9f2b936addf8fd5b1fdce7dd6a14080ac928bfda0",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-sonarr/",
	},
	"jellyseerr": {
		repository: "ghcr.io/seerr-team/seerr", digest: "sha256:f4768de5f616248d723e05891f3345a1402123775d03bf0890dbfedc0831bda1",
		configTarget: "/app/config", mediaTarget: "/data",
		requiresInit: true, sourceURL: "https://docs.seerr.dev/getting-started/docker",
	},
	"jellyfin": {
		repository: "lscr.io/linuxserver/jellyfin", digest: "sha256:b8dcc7b71d0ea872b74314da4b995c0cf282b1778438c295996e7be88c70fdda",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://docs.linuxserver.io/images/docker-jellyfin/",
	},
	"bazarr": {
		repository: "ghcr.io/hotio/bazarr", digest: "sha256:b8513bdfa0807ed80c88aba26c2ca7e4e4b8f9040c4b12a8d978faeada4b5efd",
		configTarget: "/config", mediaTarget: "/data", supportsUserMapping: true,
		sourceURL: "https://hotio.dev/containers/bazarr/",
	},
	"fileflows": {
		repository: "revenz/fileflows", digest: "sha256:a9ce79d8ad21a37ff1579f59cfa8844aded4026fd094e5a893ff476ffe0eaf93",
		configTarget: "/app/Data", mediaTarget: "/media", containerPort: 5000, supportsUserMapping: true,
		sourceURL: "https://fileflows.com/docs/installation/docker/",
	},
}

func NewRuntimeCatalog(registry *services.Registry) (*RuntimeCatalog, error) {
	manifests := make(map[string]RuntimeManifest)
	for _, service := range registry.GetAllServices() {
		if service.WebUI == nil {
			continue
		}
		approved, exists := approvedImages[service.ID]
		if !exists {
			return nil, fmt.Errorf("application has no approved runtime image: %s", service.ID)
		}
		hostPort, err := strconv.Atoi(service.WebUI.Port)
		if err != nil {
			return nil, fmt.Errorf("invalid web UI port for %s: %w", service.ID, err)
		}
		containerPort := hostPort
		if approved.containerPort != 0 {
			containerPort = approved.containerPort
		}
		manifests[service.ID] = RuntimeManifest{
			ApplicationID: service.ID, Image: approved.repository + "@" + approved.digest,
			HostPort: hostPort, ContainerPort: containerPort,
			ConfigTarget: approved.configTarget, MediaTarget: approved.mediaTarget,
			SupportsUserMapping: approved.supportsUserMapping, SourceURL: approved.sourceURL,
			RequiresInit: approved.requiresInit,
		}
	}
	return &RuntimeCatalog{manifests: manifests}, nil
}

func (c *RuntimeCatalog) ApprovedImage(applicationID string) (string, error) {
	manifest, exists := c.manifests[applicationID]
	if !exists {
		return "", fmt.Errorf("application is not approved for installation: %s", applicationID)
	}
	return manifest.Image, nil
}

func (c *RuntimeCatalog) Attribution(applicationID string) (RuntimeAttribution, error) {
	manifest, exists := c.manifests[applicationID]
	if !exists {
		return RuntimeAttribution{}, fmt.Errorf("application is not approved for installation: %s", applicationID)
	}
	return RuntimeAttribution{
		ApprovedImage: manifest.Image,
		ImageSource:   manifest.SourceURL,
	}, nil
}

func (c *RuntimeCatalog) Resolve(
	applicationID string,
	rootPath string,
	options RuntimeOptions,
) (containerruntime.ContainerSpec, error) {
	manifest, exists := c.manifests[applicationID]
	if !exists {
		return containerruntime.ContainerSpec{}, fmt.Errorf("application is not approved for installation: %s", applicationID)
	}
	if !filepath.IsAbs(rootPath) {
		return containerruntime.ContainerSpec{}, fmt.Errorf("Corsarr root path must be absolute")
	}
	environment := make(map[string]string)
	if options.Timezone != "" {
		environment["TZ"] = options.Timezone
	}
	if manifest.SupportsUserMapping && options.PUID > 0 && options.PGID > 0 {
		environment["PUID"] = strconv.Itoa(options.PUID)
		environment["PGID"] = strconv.Itoa(options.PGID)
		environment["UMASK"] = "002"
	}
	if applicationID == "qbittorrent" {
		environment["WEBUI_PORT"] = strconv.Itoa(manifest.ContainerPort)
	}

	return containerruntime.ContainerSpec{
		ApplicationID: applicationID,
		Image:         manifest.Image,
		Init:          manifest.RequiresInit,
		Ports: []containerruntime.PortBinding{{
			HostPort: manifest.HostPort, ContainerPort: manifest.ContainerPort,
			Protocol: containerruntime.ProtocolTCP, Exposure: containerruntime.ExposureLoopback,
		}},
		Mounts: []containerruntime.BindMount{
			{HostPath: filepath.Join(rootPath, "config", applicationID), ContainerPath: manifest.ConfigTarget},
			{HostPath: filepath.Join(rootPath, "media"), ContainerPath: manifest.MediaTarget},
		},
		Environment: environment,
	}, nil
}

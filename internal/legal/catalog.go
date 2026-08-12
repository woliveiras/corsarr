package legal

import (
	"fmt"
	"net/url"
	"sort"

	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/quality"
	"github.com/woliveiras/corsarr/internal/services"
)

type ComponentType string

const (
	ComponentApplication ComponentType = "application"
	ComponentRuntime     ComponentType = "runtime"
)

const (
	LinkOfficial = "official"
	LinkSource   = "source"
	LinkLicense  = "license"
	LinkImage    = "image"
	LinkSupport  = "support"
)

type LinkSummary struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type Notice struct {
	ID                   string        `json:"id"`
	Name                 string        `json:"name"`
	Purpose              string        `json:"purpose"`
	ComponentType        ComponentType `json:"componentType"`
	License              string        `json:"license"`
	CopyrightNotice      string        `json:"copyrightNotice"`
	ImageMaintainer      string        `json:"imageMaintainer,omitempty"`
	ApprovedImage        string        `json:"approvedImage,omitempty"`
	AffiliationStatement string        `json:"affiliationStatement"`
	Links                []LinkSummary `json:"links"`
}

type legalMetadata struct {
	license         string
	officialURL     string
	sourceURL       string
	licenseURL      string
	supportURL      string
	imageMaintainer string
}

type Catalog struct {
	notices []Notice
	links   map[string]map[string]string
}

var applicationMetadata = map[string]legalMetadata{
	"qbittorrent": {
		license: "GNU GPL v2+; distribuições binárias GPL v3+", officialURL: "https://www.qbittorrent.org/",
		sourceURL:       "https://github.com/qbittorrent/qBittorrent",
		licenseURL:      "https://github.com/qbittorrent/qBittorrent/tree/master/COPYING",
		imageMaintainer: "LinuxServer.io",
	},
	"prowlarr": {
		license: "GNU GPL v3", officialURL: "https://prowlarr.com/",
		sourceURL:       "https://github.com/Prowlarr/Prowlarr",
		licenseURL:      "https://github.com/Prowlarr/Prowlarr/blob/develop/LICENSE",
		imageMaintainer: "LinuxServer.io",
	},
	"lazylibrarian": {
		license: "GNU GPL v3 ou posterior", officialURL: "https://lazylibrarian.gitlab.io/",
		sourceURL:       "https://gitlab.com/LazyLibrarian/LazyLibrarian",
		licenseURL:      "https://gitlab.com/LazyLibrarian/LazyLibrarian/-/blob/master/LICENSE",
		imageMaintainer: "LinuxServer.io",
	},
	"lidarr": {
		license: "GNU GPL v3", officialURL: "https://lidarr.audio/",
		sourceURL:       "https://github.com/Lidarr/Lidarr",
		licenseURL:      "https://github.com/Lidarr/Lidarr/blob/develop/LICENSE.md",
		imageMaintainer: "hotio",
	},
	"radarr": {
		license: "GNU GPL v3", officialURL: "https://radarr.video/",
		sourceURL:       "https://github.com/Radarr/Radarr",
		licenseURL:      "https://github.com/Radarr/Radarr/blob/develop/LICENSE",
		imageMaintainer: "LinuxServer.io",
	},
	"sonarr": {
		license: "GNU GPL v3", officialURL: "https://sonarr.tv/",
		sourceURL:       "https://github.com/Sonarr/Sonarr",
		licenseURL:      "https://github.com/Sonarr/Sonarr/blob/develop/LICENSE.md",
		imageMaintainer: "LinuxServer.io",
	},
	"jellyseerr": {
		license: "MIT", officialURL: "https://docs.seerr.dev/",
		sourceURL:       "https://github.com/seerr-team/seerr",
		licenseURL:      "https://github.com/seerr-team/seerr/blob/develop/LICENSE",
		imageMaintainer: "Seerr Team",
	},
	"jellyfin": {
		license: "GNU GPL v2", officialURL: "https://jellyfin.org/",
		sourceURL:       "https://github.com/jellyfin/jellyfin",
		licenseURL:      "https://github.com/jellyfin/jellyfin/blob/master/LICENSE",
		supportURL:      "https://opencollective.com/jellyfin",
		imageMaintainer: "LinuxServer.io",
	},
	"bazarr": {
		license: "GNU GPL v3", officialURL: "https://www.bazarr.media/",
		sourceURL:       "https://github.com/morpheus65535/bazarr",
		licenseURL:      "https://github.com/morpheus65535/bazarr/blob/master/LICENSE",
		imageMaintainer: "hotio",
	},
	"fileflows": {
		license:     "Termos comerciais; plano Personal Free disponível",
		officialURL: "https://fileflows.com/", licenseURL: "https://fileflows.com/pricing",
		imageMaintainer: "FileFlows",
	},
}

func NewCatalog(registry *services.Registry, runtime *runtimecatalog.RuntimeCatalog) (*Catalog, error) {
	return newCatalog(registry, runtime, nil)
}

func NewLocalizedCatalog(
	registry *services.Registry,
	runtime *runtimecatalog.RuntimeCatalog,
	translator *i18n.I18n,
) (*Catalog, error) {
	return newCatalog(registry, runtime, translator)
}

func newCatalog(
	registry *services.Registry,
	runtime *runtimecatalog.RuntimeCatalog,
	translator *i18n.I18n,
) (*Catalog, error) {
	catalog := &Catalog{links: make(map[string]map[string]string)}
	for _, service := range registry.GetAllServices() {
		if service.WebUI == nil {
			continue
		}
		metadata, exists := applicationMetadata[service.ID]
		if !exists {
			return nil, fmt.Errorf("application has no legal metadata: %s", service.ID)
		}
		attribution, err := runtime.Attribution(service.ID)
		if err != nil {
			return nil, err
		}
		links := map[string]string{
			LinkOfficial: metadata.officialURL,
			LinkLicense:  metadata.licenseURL,
			LinkImage:    attribution.ImageSource,
		}
		if metadata.sourceURL != "" {
			links[LinkSource] = metadata.sourceURL
		}
		if metadata.supportURL != "" {
			links[LinkSupport] = metadata.supportURL
		}
		name := service.Name
		purpose := service.Description
		if translator != nil {
			if translated := translator.T(service.GetNameKey()); translated != service.GetNameKey() {
				name = translated
			}
			if translated := translator.T(service.GetDescriptionKey()); translated != service.GetDescriptionKey() {
				purpose = translated
			}
		}
		if err := catalog.addNotice(Notice{
			ID: service.ID, Name: name, Purpose: purpose,
			ComponentType: ComponentApplication, License: metadata.license,
			CopyrightNotice: "Direitos autorais pertencem aos autores e contribuidores indicados pelo projeto oficial.",
			ImageMaintainer: metadata.imageMaintainer, ApprovedImage: attribution.ApprovedImage,
			AffiliationStatement: "O Corsarr não é afiliado, patrocinado nem endossado por este projeto.",
		}, links); err != nil {
			return nil, err
		}
	}

	if err := catalog.addNotice(Notice{
		ID: "runtime-docker", Name: "Docker Desktop", Purpose: "Runtime de containers suportado pelo MVP",
		ComponentType: ComponentRuntime, License: "Docker Subscription Service Agreement e licenças dos componentes open source",
		CopyrightNotice:      "Docker e o logotipo Docker são marcas da Docker, Inc.",
		AffiliationStatement: "O Corsarr não é afiliado, patrocinado nem endossado pela Docker, Inc.",
	}, map[string]string{
		LinkOfficial: "https://www.docker.com/products/docker-desktop/",
		LinkLicense:  "https://www.docker.com/legal/docker-subscription-service-agreement/",
		LinkSource:   "https://github.com/docker",
	}); err != nil {
		return nil, err
	}
	if err := catalog.addNotice(Notice{
		ID: "runtime-podman", Name: "Podman", Purpose: "Runtime open source em avaliação pelo Corsarr",
		ComponentType: ComponentRuntime, License: "Apache License 2.0",
		CopyrightNotice:      "Direitos autorais pertencem aos contribuidores do projeto Podman.",
		AffiliationStatement: "O Corsarr não é afiliado, patrocinado nem endossado pelo projeto Podman ou pela Red Hat.",
	}, map[string]string{
		LinkOfficial: "https://podman.io/",
		LinkLicense:  "https://github.com/containers/podman/blob/main/LICENSE",
		LinkSource:   "https://github.com/containers/podman",
	}); err != nil {
		return nil, err
	}
	if err := catalog.addNotice(Notice{
		ID: "runtime-recyclarr", Name: "Recyclarr", Purpose: "Sincronização efêmera de perfis de qualidade",
		ComponentType: ComponentRuntime, License: "MIT",
		CopyrightNotice:      "Direitos autorais pertencem aos autores e contribuidores do Recyclarr.",
		ImageMaintainer:      "Recyclarr project",
		ApprovedImage:        quality.RecyclarrImage,
		AffiliationStatement: "O Corsarr não é afiliado, patrocinado nem endossado pelo Recyclarr.",
	}, map[string]string{
		LinkOfficial: "https://recyclarr.dev/",
		LinkLicense:  "https://github.com/recyclarr/recyclarr/blob/master/LICENSE",
		LinkSource:   "https://github.com/recyclarr/recyclarr",
		LinkImage:    "https://github.com/recyclarr/recyclarr/pkgs/container/recyclarr",
	}); err != nil {
		return nil, err
	}
	if err := catalog.addNotice(Notice{
		ID: "guide-trash", Name: "TRaSH Guides", Purpose: "Fonte versionada das recomendações de qualidade",
		ComponentType: ComponentRuntime, License: "MIT",
		CopyrightNotice:      "Copyright (c) 2021 TRaSH e contribuidores.",
		AffiliationStatement: "O Corsarr não é afiliado, patrocinado nem endossado pelo TRaSH Guides.",
	}, map[string]string{
		LinkOfficial: "https://trash-guides.info/",
		LinkLicense:  "https://github.com/TRaSH-Guides/Guides/blob/master/LICENSE",
		LinkSource:   "https://github.com/TRaSH-Guides/Guides",
	}); err != nil {
		return nil, err
	}

	sort.Slice(catalog.notices, func(i, j int) bool {
		if catalog.notices[i].ComponentType == catalog.notices[j].ComponentType {
			return catalog.notices[i].Name < catalog.notices[j].Name
		}
		return catalog.notices[i].ComponentType < catalog.notices[j].ComponentType
	})
	return catalog, nil
}

func (c *Catalog) ListNotices() []Notice {
	result := make([]Notice, len(c.notices))
	copy(result, c.notices)
	for index := range result {
		result[index].Links = append([]LinkSummary(nil), result[index].Links...)
	}
	return result
}

func (c *Catalog) ResolveLink(componentID, kind string) (string, error) {
	links, exists := c.links[componentID]
	if !exists {
		return "", fmt.Errorf("unknown legal component: %s", componentID)
	}
	link, exists := links[kind]
	if !exists {
		return "", fmt.Errorf("unknown legal link for %s: %s", componentID, kind)
	}
	return link, nil
}

func (c *Catalog) addNotice(notice Notice, links map[string]string) error {
	if notice.ID == "" || notice.Name == "" || notice.License == "" || len(links) == 0 {
		return fmt.Errorf("incomplete legal notice: %s", notice.ID)
	}
	summaries := make([]LinkSummary, 0, len(links))
	for kind, value := range links {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("invalid official legal link for %s: %s", notice.ID, value)
		}
		summaries = append(summaries, LinkSummary{Kind: kind, Label: linkLabel(kind)})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Kind < summaries[j].Kind })
	notice.Links = summaries
	c.notices = append(c.notices, notice)
	c.links[notice.ID] = links
	return nil
}

func linkLabel(kind string) string {
	switch kind {
	case LinkOfficial:
		return "Site oficial"
	case LinkSource:
		return "Código-fonte"
	case LinkLicense:
		return "Licença completa"
	case LinkImage:
		return "Imagem do container"
	case LinkSupport:
		return "Apoiar o projeto"
	default:
		return kind
	}
}

package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const bazarrSettingsPath = "/api/system/settings"

type BazarrClient struct {
	api *ARRClient
}

func NewBazarrClient(resolver ApplicationEndpointResolver) *BazarrClient {
	return &BazarrClient{api: NewARRClient(resolver)}
}

func (c *BazarrClient) EnsureARRConnections(
	ctx context.Context,
	bazarrKey APIKey,
	radarrKey APIKey,
	sonarrKey APIKey,
) error {
	settings := url.Values{
		"settings-general-use_radarr": {"true"},
		"settings-general-use_sonarr": {"true"},
		"settings-radarr-ip":          {"radarr"},
		"settings-radarr-port":        {"7878"},
		"settings-radarr-base_url":    {"/"},
		"settings-radarr-ssl":         {"false"},
		"settings-radarr-apikey":      {radarrKey.Reveal()},
		"settings-sonarr-ip":          {"sonarr"},
		"settings-sonarr-port":        {"8989"},
		"settings-sonarr-base_url":    {"/"},
		"settings-sonarr-ssl":         {"false"},
		"settings-sonarr-apikey":      {sonarrKey.Reveal()},
	}
	endpoint, err := c.api.apiEndpoint("bazarr", bazarrSettingsPath)
	if err != nil {
		return err
	}
	request, err := c.api.request(ctx, http.MethodPost, endpoint, bazarrKey, strings.NewReader(settings.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.api.client.Do(request)
	if err != nil {
		return fmt.Errorf("save Bazarr Arr connections: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("save Bazarr Arr connections: %w", err)
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("save Bazarr Arr connections: unexpected HTTP status %d", response.StatusCode)
	}

	return c.verifyARRConnections(ctx, endpoint, bazarrKey, radarrKey, sonarrKey)
}

func (c *BazarrClient) verifyARRConnections(
	ctx context.Context,
	endpoint string,
	bazarrKey APIKey,
	radarrKey APIKey,
	sonarrKey APIKey,
) error {
	request, err := c.api.request(ctx, http.MethodGet, endpoint, bazarrKey, nil)
	if err != nil {
		return err
	}
	response, err := c.api.client.Do(request)
	if err != nil {
		return fmt.Errorf("verify Bazarr Arr connections: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return fmt.Errorf("verify Bazarr Arr connections: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("verify Bazarr Arr connections: unexpected HTTP status %d", response.StatusCode)
	}

	var current struct {
		General struct {
			UseRadarr bool `json:"use_radarr"`
			UseSonarr bool `json:"use_sonarr"`
		} `json:"general"`
		Radarr bazarrARRSettings `json:"radarr"`
		Sonarr bazarrARRSettings `json:"sonarr"`
	}
	if err := json.Unmarshal(contents, &current); err != nil {
		return fmt.Errorf("decode Bazarr settings: %w", err)
	}
	if !current.General.UseRadarr || !current.General.UseSonarr ||
		!current.Radarr.matches("radarr", 7878, radarrKey) ||
		!current.Sonarr.matches("sonarr", 8989, sonarrKey) {
		return fmt.Errorf("bazarr did not persist the approved Arr connections")
	}
	return nil
}

type bazarrARRSettings struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	BaseURL string `json:"base_url"`
	SSL     bool   `json:"ssl"`
	APIKey  string `json:"apikey"`
}

func (s bazarrARRSettings) matches(host string, port int, apiKey APIKey) bool {
	return s.IP == host && s.Port == port && (s.BaseURL == "" || s.BaseURL == "/") &&
		!s.SSL && s.APIKey == apiKey.Reveal()
}

type BazarrConfigurator interface {
	EnsureARRConnections(
		ctx context.Context,
		bazarrKey APIKey,
		radarrKey APIKey,
		sonarrKey APIKey,
	) error
}

type BazarrAPIKeyReader interface {
	Read(rootPath string) (APIKey, error)
}

type BazarrProvisioner struct {
	bazarrCredentials BazarrAPIKeyReader
	arrCredentials    CredentialReader
	client            BazarrConfigurator
}

func NewBazarrProvisioner(
	bazarrCredentials BazarrAPIKeyReader,
	arrCredentials CredentialReader,
	client BazarrConfigurator,
) *BazarrProvisioner {
	return &BazarrProvisioner{
		bazarrCredentials: bazarrCredentials,
		arrCredentials:    arrCredentials,
		client:            client,
	}
}

func (p *BazarrProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
) error {
	if applicationID != "bazarr" {
		return nil
	}
	bazarrKey, err := p.bazarrCredentials.Read(rootPath)
	if err != nil {
		return fmt.Errorf("read Bazarr API credential: %w", err)
	}
	radarrKey, err := p.arrCredentials.Read(rootPath, "radarr")
	if err != nil {
		return fmt.Errorf("read Radarr API credential for Bazarr: %w", err)
	}
	sonarrKey, err := p.arrCredentials.Read(rootPath, "sonarr")
	if err != nil {
		return fmt.Errorf("read Sonarr API credential for Bazarr: %w", err)
	}
	if err := p.client.EnsureARRConnections(ctx, bazarrKey, radarrKey, sonarrKey); err != nil {
		return fmt.Errorf("ensure Bazarr Arr connections: %w", err)
	}
	return nil
}

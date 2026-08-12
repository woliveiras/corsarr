package provisioning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type prowlarrTarget struct {
	Name           string
	BaseURL        string
	SyncCategories []int
}

var prowlarrTargets = map[string]prowlarrTarget{
	"lazylibrarian": {
		Name: "LazyLibrarian", BaseURL: "http://lazylibrarian:5299",
		SyncCategories: []int{3030, 7000, 7010, 7020, 7030, 7040, 7050, 7060},
	},
	"lidarr": {Name: "Lidarr", BaseURL: "http://lidarr:8686"},
	"radarr": {Name: "Radarr", BaseURL: "http://radarr:7878"},
	"sonarr": {Name: "Sonarr", BaseURL: "http://sonarr:8989"},
}

type ProwlarrClient struct {
	api *ARRClient
}

func NewProwlarrClient(resolver ApplicationEndpointResolver) *ProwlarrClient {
	return &ProwlarrClient{api: NewARRClient(resolver)}
}

func (c *ProwlarrClient) EnsureApplication(
	ctx context.Context,
	applicationID string,
	prowlarrKey APIKey,
	targetKey APIKey,
) error {
	return c.ensureApplication(ctx, applicationID, prowlarrKey, targetKey, "", credentials.Secret{})
}

func (c *ProwlarrClient) EnsureApplicationWithAuthentication(
	ctx context.Context,
	applicationID string,
	prowlarrKey APIKey,
	targetKey APIKey,
	username string,
	password credentials.Secret,
) error {
	return c.ensureApplication(ctx, applicationID, prowlarrKey, targetKey, username, password)
}

func (c *ProwlarrClient) ensureApplication(
	ctx context.Context,
	applicationID string,
	prowlarrKey APIKey,
	targetKey APIKey,
	username string,
	password credentials.Secret,
) error {
	target, supported := prowlarrTargets[applicationID]
	if !supported {
		return fmt.Errorf("prowlarr application is not supported: %s", applicationID)
	}
	const basePath = "/api/v1/applications"
	providers, err := c.api.getProviders(ctx, "prowlarr", prowlarrKey, basePath)
	if err != nil {
		return err
	}
	reservedName := target.Name + " (Corsarr)"
	var provider map[string]any
	for _, candidate := range providers {
		if candidate["name"] == reservedName {
			if !strings.EqualFold(stringValue(candidate["implementation"]), target.Name) {
				return fmt.Errorf("reserved Corsarr Prowlarr application name is already in use")
			}
			provider = candidate
			break
		}
	}
	creating := provider == nil
	if creating {
		schemas, err := c.api.getProviders(ctx, "prowlarr", prowlarrKey, basePath+"/schema")
		if err != nil {
			return err
		}
		for _, schema := range schemas {
			if strings.EqualFold(stringValue(schema["implementation"]), target.Name) {
				provider = schema
				break
			}
		}
		if provider == nil {
			return fmt.Errorf("prowlarr %s schema is unavailable", target.Name)
		}
	}

	provider["name"] = reservedName
	provider["enable"] = true
	provider["syncLevel"] = "fullSync"
	fields, ok := provider["fields"].([]any)
	if !ok {
		return fmt.Errorf("prowlarr application schema has invalid fields")
	}
	values := map[string]any{
		"prowlarrUrl":  "http://prowlarr:9696",
		"baseUrl":      target.BaseURL,
		"apiKey":       targetKey.Reveal(),
		"authUsername": username,
		"authPassword": password.Reveal(),
	}
	if len(target.SyncCategories) > 0 {
		values["syncCategories"] = target.SyncCategories
	}
	for name, value := range values {
		if !setProviderField(fields, name, value) {
			return fmt.Errorf("prowlarr application schema is missing required field: %s", name)
		}
	}

	method := http.MethodPost
	endpointPath := basePath
	if !creating {
		id, ok := numericID(provider["id"])
		if !ok || id < 1 {
			return fmt.Errorf("corsarr Prowlarr application has invalid ID")
		}
		method = http.MethodPut
		endpointPath += "/" + strconv.Itoa(id)
	}
	contents, err := json.Marshal(provider)
	if err != nil {
		return fmt.Errorf("encode Prowlarr application: %w", err)
	}
	endpoint, err := c.api.apiEndpoint("prowlarr", endpointPath)
	if err != nil {
		return err
	}
	request, err := c.api.request(ctx, method, endpoint, prowlarrKey, bytes.NewReader(contents))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.api.client.Do(request)
	if err != nil {
		return fmt.Errorf("save Prowlarr application: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("save Prowlarr application: %w", err)
	}
	wantedStatus := http.StatusCreated
	if !creating {
		wantedStatus = http.StatusAccepted
	}
	if response.StatusCode != wantedStatus && response.StatusCode != http.StatusOK {
		return fmt.Errorf("save Prowlarr application: unexpected HTTP status %d", response.StatusCode)
	}
	return nil
}

type ProwlarrConfigurator interface {
	EnsureApplication(
		ctx context.Context,
		applicationID string,
		prowlarrKey APIKey,
		targetKey APIKey,
	) error
	EnsureApplicationWithAuthentication(
		ctx context.Context,
		applicationID string,
		prowlarrKey APIKey,
		targetKey APIKey,
		username string,
		password credentials.Secret,
	) error
}

type ProwlarrProvisioner struct {
	credentialReader CredentialReader
	secrets          credentials.Store
	client           ProwlarrConfigurator
}

func NewProwlarrProvisioner(
	credentialReader CredentialReader,
	client ProwlarrConfigurator,
	stores ...credentials.Store,
) *ProwlarrProvisioner {
	var secrets credentials.Store
	if len(stores) > 0 {
		secrets = stores[0]
	}
	return &ProwlarrProvisioner{
		credentialReader: credentialReader,
		secrets:          secrets,
		client:           client,
	}
}

func (p *ProwlarrProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
	selected []string,
) error {
	if _, supported := prowlarrTargets[applicationID]; !supported {
		return nil
	}
	if !selectedApplication(selected, "prowlarr") {
		return nil
	}
	prowlarrKey, err := p.credentialReader.Read(rootPath, "prowlarr")
	if err != nil {
		return fmt.Errorf("read Prowlarr API credential: %w", err)
	}
	if applicationID == lazyLibrarianApplicationID {
		if p.secrets == nil {
			return fmt.Errorf("LazyLibrarian credential store is unavailable")
		}
		apiKeySecret, err := p.secrets.Load(ctx, credentials.KeyLazyLibrarianAPIKey)
		if err != nil {
			return fmt.Errorf("load LazyLibrarian API credential: %w", err)
		}
		targetKey, err := newAPIKey(apiKeySecret.Reveal())
		if err != nil {
			return fmt.Errorf("validate LazyLibrarian API credential: %w", err)
		}
		password, err := p.secrets.Load(ctx, credentials.KeyLazyLibrarianPassword)
		if err != nil {
			return fmt.Errorf("load LazyLibrarian authentication credential: %w", err)
		}
		if err := p.client.EnsureApplicationWithAuthentication(
			ctx,
			applicationID,
			prowlarrKey,
			targetKey,
			lazyLibrarianUsername,
			password,
		); err != nil {
			return fmt.Errorf("ensure Prowlarr application: %w", err)
		}
		return nil
	}
	targetKey, err := p.credentialReader.Read(rootPath, applicationID)
	if err != nil {
		return fmt.Errorf("read target Arr API credential: %w", err)
	}
	if err := p.client.EnsureApplication(ctx, applicationID, prowlarrKey, targetKey); err != nil {
		return fmt.Errorf("ensure Prowlarr application: %w", err)
	}
	return nil
}

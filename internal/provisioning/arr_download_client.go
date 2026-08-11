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

const (
	corsarrQBittorrentProviderName = "qBittorrent (Corsarr)"
	qbittorrentInternalPort        = 8081
)

var arrCategoryFields = map[string]string{
	"lidarr": "musicCategory",
	"radarr": "movieCategory",
	"sonarr": "tvCategory",
}

func (c *ARRClient) EnsureQBittorrentDownloadClient(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
	username string,
	password credentials.Secret,
) error {
	apiVersion, supported := arrAPIVersions[applicationID]
	categoryField := arrCategoryFields[applicationID]
	if !supported || categoryField == "" {
		return fmt.Errorf("qBittorrent download client is not supported for application: %s", applicationID)
	}
	basePath := "/api/" + apiVersion + "/downloadclient"

	providers, err := c.getProviders(ctx, applicationID, apiKey, basePath)
	if err != nil {
		return err
	}
	var provider map[string]any
	for _, candidate := range providers {
		if candidate["name"] == corsarrQBittorrentProviderName {
			if !strings.EqualFold(stringValue(candidate["implementation"]), "QBittorrent") {
				return fmt.Errorf("reserved Corsarr download client name is already in use")
			}
			provider = candidate
			break
		}
	}
	creating := provider == nil
	if creating {
		schemas, err := c.getProviders(ctx, applicationID, apiKey, basePath+"/schema")
		if err != nil {
			return err
		}
		for _, schema := range schemas {
			if strings.EqualFold(stringValue(schema["implementation"]), "QBittorrent") {
				provider = schema
				break
			}
		}
		if provider == nil {
			return fmt.Errorf("qBittorrent download client schema is unavailable")
		}
	}

	provider["name"] = corsarrQBittorrentProviderName
	provider["enable"] = true
	provider["priority"] = 1
	provider["removeCompletedDownloads"] = true
	provider["removeFailedDownloads"] = true
	fields, ok := provider["fields"].([]any)
	if !ok {
		return fmt.Errorf("qBittorrent download client schema has invalid fields")
	}
	values := map[string]any{
		"host":        "qbittorrent",
		"port":        qbittorrentInternalPort,
		"useSsl":      false,
		"urlBase":     "",
		"apiKey":      "",
		"username":    username,
		"password":    password.Reveal(),
		categoryField: applicationID,
	}
	for name, value := range values {
		if !setProviderField(fields, name, value) {
			return fmt.Errorf("qBittorrent schema is missing required field: %s", name)
		}
	}

	method := http.MethodPost
	endpointPath := basePath
	if !creating {
		id, ok := numericID(provider["id"])
		if !ok || id < 1 {
			return fmt.Errorf("corsarr qBittorrent download client has invalid ID")
		}
		method = http.MethodPut
		endpointPath += "/" + strconv.Itoa(id)
	}
	contents, err := json.Marshal(provider)
	if err != nil {
		return fmt.Errorf("encode qBittorrent download client: %w", err)
	}
	endpoint, err := c.apiEndpoint(applicationID, endpointPath)
	if err != nil {
		return err
	}
	request, err := c.request(ctx, method, endpoint, apiKey, bytes.NewReader(contents))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("save qBittorrent download client: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("save qBittorrent download client: %w", err)
	}
	wantedStatus := http.StatusCreated
	if !creating {
		wantedStatus = http.StatusAccepted
	}
	if response.StatusCode != wantedStatus && response.StatusCode != http.StatusOK {
		return fmt.Errorf("save qBittorrent download client: unexpected HTTP status %d", response.StatusCode)
	}
	return nil
}

func (c *ARRClient) getProviders(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
	apiPath string,
) ([]map[string]any, error) {
	endpoint, err := c.apiEndpoint(applicationID, apiPath)
	if err != nil {
		return nil, err
	}
	request, err := c.request(ctx, http.MethodGet, endpoint, apiKey, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("load application provider: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return nil, fmt.Errorf("load application provider: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("load application provider: unexpected HTTP status %d", response.StatusCode)
	}
	var providers []map[string]any
	if err := json.Unmarshal(contents, &providers); err != nil {
		return nil, fmt.Errorf("decode application provider: %w", err)
	}
	return providers, nil
}

func setProviderField(fields []any, name string, value any) bool {
	for _, rawField := range fields {
		field, ok := rawField.(map[string]any)
		if !ok || !strings.EqualFold(stringValue(field["name"]), name) {
			continue
		}
		field["value"] = value
		return true
	}
	return false
}

func stringValue(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func numericID(value any) (int, bool) {
	switch number := value.(type) {
	case float64:
		return int(number), number == float64(int(number))
	case int:
		return number, true
	case json.Number:
		parsed, err := strconv.Atoi(number.String())
		return parsed, err == nil
	default:
		return 0, false
	}
}

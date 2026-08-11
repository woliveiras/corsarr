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
	arrAuthenticationMethodForms            = "forms"
	arrAuthenticationMethodNone             = "none"
	arrAuthenticationRequiredLocalAddresses = "disabledForLocalAddresses"
	arrManagedUsername                      = "corsarr"
)

var arrAuthenticationAPIVersions = map[string]string{
	"lidarr":   "v1",
	"prowlarr": "v1",
	"radarr":   "v3",
	"sonarr":   "v3",
}

type ARRAuthenticationInspection struct {
	BootstrapRequired bool
	CorsarrManaged    bool
}

type ARRAuthenticationActivation struct {
	CredentialAccepted bool
}

func (c *ARRClient) InspectAuthentication(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
) (ARRAuthenticationInspection, error) {
	config, err := c.getHostConfig(ctx, applicationID, apiKey)
	if err != nil {
		return ARRAuthenticationInspection{}, err
	}
	return inspectARRAuthentication(config), nil
}

func (c *ARRClient) ActivateAuthentication(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
	username string,
	password credentials.Secret,
) (ARRAuthenticationActivation, error) {
	result := ARRAuthenticationActivation{}
	if username != arrManagedUsername {
		return result, fmt.Errorf("Arr authentication username is not approved")
	}
	if len(password.Reveal()) < 16 || len(password.Reveal()) > 256 ||
		strings.ContainsRune(password.Reveal(), '\x00') {
		return result, fmt.Errorf("Arr authentication credential is invalid")
	}
	config, err := c.getHostConfig(ctx, applicationID, apiKey)
	if err != nil {
		return result, err
	}
	if !inspectARRAuthentication(config).BootstrapRequired {
		return result, fmt.Errorf("Arr authentication bootstrap is no longer available")
	}
	id, ok := numericID(config["id"])
	if !ok || id < 1 {
		return result, fmt.Errorf("Arr host configuration has invalid ID")
	}
	config["authenticationMethod"] = arrAuthenticationMethodForms
	config["authenticationRequired"] = arrAuthenticationRequiredLocalAddresses
	config["username"] = username
	config["password"] = password.Reveal()
	config["passwordConfirmation"] = password.Reveal()

	contents, err := json.Marshal(config)
	if err != nil {
		return result, fmt.Errorf("encode Arr host configuration: %w", err)
	}
	endpoint, err := c.authenticationEndpoint(applicationID, "/"+strconv.Itoa(id))
	if err != nil {
		return result, err
	}
	request, err := c.request(ctx, http.MethodPut, endpoint, apiKey, bytes.NewReader(contents))
	if err != nil {
		return result, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return result, fmt.Errorf("activate Arr authentication: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return result, fmt.Errorf("activate Arr authentication: %w", err)
	}
	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusOK {
		return result, fmt.Errorf(
			"activate Arr authentication: unexpected HTTP status %d",
			response.StatusCode,
		)
	}
	result.CredentialAccepted = true

	verified, err := c.getHostConfig(ctx, applicationID, apiKey)
	if err != nil {
		return result, fmt.Errorf("verify Arr authentication: %w", err)
	}
	inspection := inspectARRAuthentication(verified)
	if !inspection.CorsarrManaged ||
		stringValue(verified["authenticationRequired"]) != arrAuthenticationRequiredLocalAddresses {
		return result, fmt.Errorf("Arr authentication did not persist the approved configuration")
	}
	return result, nil
}

func (c *ARRClient) getHostConfig(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
) (map[string]any, error) {
	endpoint, err := c.authenticationEndpoint(applicationID, "")
	if err != nil {
		return nil, err
	}
	request, err := c.request(ctx, http.MethodGet, endpoint, apiKey, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("read Arr host configuration: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return nil, fmt.Errorf("read Arr host configuration: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"read Arr host configuration: unexpected HTTP status %d",
			response.StatusCode,
		)
	}
	var config map[string]any
	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("decode Arr host configuration: %w", err)
	}
	return config, nil
}

func (c *ARRClient) authenticationEndpoint(applicationID, suffix string) (string, error) {
	version, supported := arrAuthenticationAPIVersions[applicationID]
	if !supported {
		return "", fmt.Errorf("Arr authentication is not supported: %s", applicationID)
	}
	return c.apiEndpoint(applicationID, "/api/"+version+"/config/host"+suffix)
}

func inspectARRAuthentication(config map[string]any) ARRAuthenticationInspection {
	method := stringValue(config["authenticationMethod"])
	username := stringValue(config["username"])
	return ARRAuthenticationInspection{
		BootstrapRequired: method == arrAuthenticationMethodNone && username == "",
		CorsarrManaged:    method == arrAuthenticationMethodForms && username == arrManagedUsername,
	}
}

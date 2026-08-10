package provisioning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

const maxARRResponseSize = 1024 * 1024

var (
	approvedARRRootFolders = map[string]string{
		"lidarr": "/data/library/music",
		"radarr": "/data/library/movies",
		"sonarr": "/data/library/tv",
	}
	arrAPIVersions = map[string]string{
		"lidarr": "v1",
		"radarr": "v3",
		"sonarr": "v3",
	}
)

type ARRClient struct {
	resolver ApplicationEndpointResolver
	client   *http.Client
}

func NewARRClient(resolver ApplicationEndpointResolver) *ARRClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	return &ARRClient{
		resolver: resolver,
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *ARRClient) EnsureRootFolder(
	ctx context.Context,
	applicationID string,
	apiKey APIKey,
	rootPath string,
) error {
	approvedPath, supported := approvedARRRootFolders[applicationID]
	apiVersion := arrAPIVersions[applicationID]
	if !supported || path.Clean(rootPath) != approvedPath {
		return fmt.Errorf("root folder is not approved for application: %s", applicationID)
	}
	endpoint, err := c.apiEndpoint(applicationID, "/api/"+apiVersion+"/rootfolder")
	if err != nil {
		return err
	}

	request, err := c.request(ctx, http.MethodGet, endpoint, apiKey, nil)
	if err != nil {
		return err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("list application root folders: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return fmt.Errorf("list application root folders: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("list application root folders: unexpected HTTP status %d", response.StatusCode)
	}
	var folders []struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(contents, &folders); err != nil {
		return fmt.Errorf("decode application root folders: %w", err)
	}
	for _, folder := range folders {
		if path.Clean(folder.Path) == approvedPath {
			return nil
		}
	}

	body, err := json.Marshal(struct {
		Path string `json:"path"`
	}{Path: approvedPath})
	if err != nil {
		return fmt.Errorf("encode application root folder: %w", err)
	}
	request, err = c.request(ctx, http.MethodPost, endpoint, apiKey, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err = c.client.Do(request)
	if err != nil {
		return fmt.Errorf("create application root folder: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("create application root folder: %w", err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return fmt.Errorf("create application root folder: unexpected HTTP status %d", response.StatusCode)
	}
	return nil
}

func (c *ARRClient) apiEndpoint(applicationID string, apiPath string) (string, error) {
	baseEndpoint, err := c.resolver.ResolveApplicationURL(applicationID)
	if err != nil {
		return "", fmt.Errorf("resolve application API endpoint: %w", err)
	}
	parsed, err := url.Parse(baseEndpoint)
	if err != nil {
		return "", fmt.Errorf("parse application API endpoint: %w", err)
	}
	hostIP := net.ParseIP(parsed.Hostname())
	if parsed.Scheme != "http" || hostIP == nil || !hostIP.IsLoopback() || parsed.User != nil {
		return "", fmt.Errorf("application API endpoint is not an approved loopback HTTP URL")
	}
	parsed.Path = apiPath
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (c *ARRClient) request(
	ctx context.Context,
	method string,
	endpoint string,
	apiKey APIKey,
	body io.Reader,
) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create application API request: %w", err)
	}
	request.Header.Set("X-Api-Key", apiKey.Reveal())
	request.Header.Set("Accept", "application/json")
	return request, nil
}

func boundedResponse(response *http.Response) ([]byte, error) {
	contents, readErr := io.ReadAll(io.LimitReader(response.Body, maxARRResponseSize+1))
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close application API response: %w", closeErr)
	}
	if len(contents) > maxARRResponseSize {
		return nil, fmt.Errorf("response exceeds size limit")
	}
	return contents, nil
}

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
	"regexp"
	"strings"
	"time"

	"github.com/woliveiras/corsarr/internal/credentials"
)

const jellyfinUsername = "corsarr"

const (
	jellyfinSetupAPITimeout      = 2 * time.Minute
	jellyfinSetupAPIPollInterval = 500 * time.Millisecond
)

var jellyfinTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,512}$`)

var jellyfinLibraries = []jellyfinLibrary{
	{Name: "Movies (Corsarr)", CollectionType: "movies", Path: "/data/library/movies"},
	{Name: "Music (Corsarr)", CollectionType: "music", Path: "/data/library/music"},
	{Name: "TV Shows (Corsarr)", CollectionType: "tvshows", Path: "/data/library/tv"},
}

type jellyfinLibrary struct {
	Name           string
	CollectionType string
	Path           string
}

type JellyfinSetupResult struct {
	CredentialAccepted bool
}

type JellyfinClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewJellyfinClient(resolver ApplicationEndpointResolver) *JellyfinClient {
	client := &JellyfinClient{}
	endpoint, err := resolver.ResolveApplicationURL("jellyfin")
	if err == nil {
		parsed, parseErr := url.Parse(endpoint)
		if parseErr == nil {
			hostIP := net.ParseIP(parsed.Hostname())
			if parsed.Scheme == "http" && hostIP != nil && hostIP.IsLoopback() && parsed.User == nil {
				parsed.Path = "/"
				parsed.RawPath = ""
				parsed.RawQuery = ""
				parsed.Fragment = ""
				client.baseURL = parsed
			}
		}
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	client.client = &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client
}

func (c *JellyfinClient) EnsureSetup(
	ctx context.Context,
	password credentials.Secret,
) (JellyfinSetupResult, error) {
	completed, err := c.startupCompleted(ctx)
	if err != nil {
		return JellyfinSetupResult{}, err
	}
	if !completed {
		if err := c.waitForSetupAPI(ctx); err != nil {
			return JellyfinSetupResult{}, err
		}
	}
	token, authenticated, err := c.authenticate(ctx, password)
	if err != nil {
		return JellyfinSetupResult{}, err
	}
	result := JellyfinSetupResult{CredentialAccepted: authenticated}
	if completed && !authenticated {
		return result, fmt.Errorf("authenticate existing Jellyfin administrator: login rejected")
	}
	if !authenticated {
		if err := c.updateStartupUser(ctx, password); err != nil {
			return result, err
		}
		result.CredentialAccepted = true
		token, authenticated, err = c.authenticate(ctx, password)
		if err != nil {
			return result, err
		}
		if !authenticated {
			return result, fmt.Errorf("verify Jellyfin administrator: login rejected")
		}
	}

	if err := c.ensureLibraries(ctx, token); err != nil {
		return result, err
	}
	if completed {
		return result, nil
	}
	if err := c.postJSON(ctx, "/Startup/RemoteAccess", token.Reveal(), map[string]bool{
		"EnableRemoteAccess": false,
	}); err != nil {
		return result, fmt.Errorf("disable Jellyfin remote access default: %w", err)
	}
	if err := c.postJSON(ctx, "/Startup/Complete", token.Reveal(), struct{}{}); err != nil {
		return result, fmt.Errorf("complete Jellyfin startup: %w", err)
	}
	completed, err = c.startupCompleted(ctx)
	if err != nil {
		return result, err
	}
	if !completed {
		return result, fmt.Errorf("jellyfin startup did not complete")
	}
	return result, nil
}

func (c *JellyfinClient) waitForSetupAPI(ctx context.Context) error {
	waitContext, cancel := context.WithTimeout(ctx, jellyfinSetupAPITimeout)
	defer cancel()

	for {
		response, _, err := c.do(waitContext, http.MethodGet, "/Startup/User", "", nil)
		if err != nil {
			return fmt.Errorf("check Jellyfin setup API readiness: %w", err)
		}
		if response.StatusCode == http.StatusOK {
			return nil
		}
		if response.StatusCode != http.StatusServiceUnavailable {
			return fmt.Errorf(
				"check Jellyfin setup API readiness: unexpected HTTP status %d",
				response.StatusCode,
			)
		}

		timer := time.NewTimer(jellyfinSetupAPIPollInterval)
		select {
		case <-waitContext.Done():
			timer.Stop()
			return fmt.Errorf("wait for Jellyfin setup API: %w", waitContext.Err())
		case <-timer.C:
		}
	}
}

func (c *JellyfinClient) startupCompleted(ctx context.Context) (bool, error) {
	waitContext, cancel := context.WithTimeout(ctx, jellyfinSetupAPITimeout)
	defer cancel()

	for {
		response, contents, err := c.do(
			waitContext,
			http.MethodGet,
			"/System/Info/Public",
			"",
			nil,
		)
		if err != nil {
			return false, fmt.Errorf("get Jellyfin public status: %w", err)
		}
		if response.StatusCode == http.StatusOK {
			var status struct {
				StartupWizardCompleted bool `json:"StartupWizardCompleted"`
			}
			if err := json.Unmarshal(contents, &status); err != nil {
				return false, fmt.Errorf("decode Jellyfin public status: %w", err)
			}
			return status.StartupWizardCompleted, nil
		}
		if response.StatusCode != http.StatusServiceUnavailable {
			return false, fmt.Errorf(
				"get Jellyfin public status: unexpected HTTP status %d",
				response.StatusCode,
			)
		}

		timer := time.NewTimer(jellyfinSetupAPIPollInterval)
		select {
		case <-waitContext.Done():
			timer.Stop()
			return false, fmt.Errorf("wait for Jellyfin public status: %w", waitContext.Err())
		case <-timer.C:
		}
	}
}

func (c *JellyfinClient) authenticate(
	ctx context.Context,
	password credentials.Secret,
) (credentials.Secret, bool, error) {
	body, err := json.Marshal(map[string]string{"Username": jellyfinUsername, "Pw": password.Reveal()})
	if err != nil {
		return credentials.Secret{}, false, fmt.Errorf("encode Jellyfin authentication: %w", err)
	}
	response, contents, err := c.do(ctx, http.MethodPost, "/Users/AuthenticateByName", "", bytes.NewReader(body))
	if err != nil {
		return credentials.Secret{}, false, fmt.Errorf("authenticate Jellyfin: %w", err)
	}
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return credentials.Secret{}, false, nil
	}
	if response.StatusCode != http.StatusOK {
		return credentials.Secret{}, false, fmt.Errorf("authenticate Jellyfin: unexpected HTTP status %d", response.StatusCode)
	}
	var result struct {
		AccessToken string `json:"AccessToken"`
	}
	if err := json.Unmarshal(contents, &result); err != nil {
		return credentials.Secret{}, false, fmt.Errorf("decode Jellyfin authentication: %w", err)
	}
	if !jellyfinTokenPattern.MatchString(result.AccessToken) {
		return credentials.Secret{}, false, fmt.Errorf("jellyfin returned an invalid access token")
	}
	return credentials.NewSecret(result.AccessToken), true, nil
}

func (c *JellyfinClient) updateStartupUser(ctx context.Context, password credentials.Secret) error {
	err := c.postJSON(ctx, "/Startup/User", "", map[string]string{
		"Name": jellyfinUsername, "Password": password.Reveal(),
	})
	if err != nil {
		return fmt.Errorf("create Jellyfin administrator: %w", err)
	}
	return nil
}

func (c *JellyfinClient) ensureLibraries(ctx context.Context, token credentials.Secret) error {
	response, contents, err := c.do(ctx, http.MethodGet, "/Library/VirtualFolders", token.Reveal(), nil)
	if err != nil {
		return fmt.Errorf("list Jellyfin libraries: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("list Jellyfin libraries: unexpected HTTP status %d", response.StatusCode)
	}
	var existing []struct {
		Name           string   `json:"Name"`
		Locations      []string `json:"Locations"`
		CollectionType string   `json:"CollectionType"`
	}
	if err := json.Unmarshal(contents, &existing); err != nil {
		return fmt.Errorf("decode Jellyfin libraries: %w", err)
	}
	for _, desired := range jellyfinLibraries {
		found := false
		for _, candidate := range existing {
			if candidate.Name != desired.Name {
				continue
			}
			if candidate.CollectionType != desired.CollectionType ||
				len(candidate.Locations) != 1 || path.Clean(candidate.Locations[0]) != desired.Path {
				return fmt.Errorf("reserved Corsarr Jellyfin library name is already in use: %s", desired.Name)
			}
			found = true
			break
		}
		if found {
			continue
		}
		query := url.Values{
			"name":           {desired.Name},
			"collectionType": {desired.CollectionType},
			"paths":          {desired.Path},
			"refreshLibrary": {"false"},
		}
		apiPath := "/Library/VirtualFolders?" + query.Encode()
		if err := c.postJSON(ctx, apiPath, token.Reveal(), struct{}{}); err != nil {
			return fmt.Errorf("create Jellyfin library %s: %w", desired.Name, err)
		}
	}
	return nil
}

func (c *JellyfinClient) postJSON(ctx context.Context, apiPath string, token string, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode Jellyfin request: %w", err)
	}
	response, _, err := c.do(ctx, http.MethodPost, apiPath, token, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected HTTP status %d", response.StatusCode)
	}
	return nil
}

func (c *JellyfinClient) do(
	ctx context.Context,
	method string,
	apiPath string,
	token string,
	body io.Reader,
) (*http.Response, []byte, error) {
	if c.baseURL == nil {
		return nil, nil, fmt.Errorf("jellyfin endpoint is not an approved loopback HTTP URL")
	}
	endpoint := *c.baseURL
	parsedPath, err := url.Parse(apiPath)
	if err != nil || !strings.HasPrefix(parsedPath.Path, "/") || parsedPath.Host != "" {
		return nil, nil, fmt.Errorf("jellyfin API path is invalid")
	}
	endpoint.Path = parsedPath.Path
	endpoint.RawQuery = parsedPath.RawQuery
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, nil, fmt.Errorf("create Jellyfin request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	authorization := `MediaBrowser Client="Corsarr", Device="Corsarr Desktop", DeviceId="corsarr-desktop", Version="0.1.0"`
	if token != "" {
		authorization += `, Token="` + token + `"`
	}
	request.Header.Set("Authorization", authorization)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return response, nil, err
	}
	return response, contents, nil
}

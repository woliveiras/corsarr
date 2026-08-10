package provisioning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type SeerrClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewSeerrClient(resolver ApplicationEndpointResolver) *SeerrClient {
	client := &SeerrClient{}
	endpoint, err := resolver.ResolveApplicationURL("jellyseerr")
	if err == nil {
		parsed, parseErr := url.Parse(endpoint)
		if parseErr == nil {
			hostIP := net.ParseIP(parsed.Hostname())
			if parsed.Scheme == "http" && hostIP != nil && hostIP.IsLoopback() && parsed.User == nil {
				parsed.Path, parsed.RawPath, parsed.RawQuery, parsed.Fragment = "/", "", "", ""
				client.baseURL = parsed
			}
		}
	}
	jar, _ := cookiejar.New(nil)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	client.client = &http.Client{
		Transport: transport, Jar: jar, Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return client
}

func (c *SeerrClient) EnsureSetup(
	ctx context.Context,
	jellyfinPassword credentials.Secret,
	radarrKey APIKey,
	sonarrKey APIKey,
) error {
	initialized, err := c.initialized(ctx)
	if err != nil {
		return err
	}
	login := map[string]any{"username": jellyfinUsername, "password": jellyfinPassword.Reveal()}
	if !initialized {
		login["hostname"], login["port"], login["useSsl"], login["serverType"] = "jellyfin", 8096, false, 2
	}
	if _, err := c.jsonRequest(ctx, http.MethodPost, "/api/v1/auth/jellyfin", login, http.StatusOK); err != nil {
		return fmt.Errorf("authenticate Seerr through Jellyfin: %w", err)
	}
	if err := c.enableJellyfinLibraries(ctx); err != nil {
		return err
	}
	if err := c.ensureARR(ctx, "radarr", radarrKey); err != nil {
		return err
	}
	if err := c.ensureARR(ctx, "sonarr", sonarrKey); err != nil {
		return err
	}
	if !initialized {
		contents, err := c.jsonRequest(ctx, http.MethodPost, "/api/v1/settings/initialize", struct{}{}, http.StatusOK)
		if err != nil {
			return fmt.Errorf("initialize Seerr: %w", err)
		}
		var status struct {
			Initialized bool `json:"initialized"`
		}
		if json.Unmarshal(contents, &status) != nil || !status.Initialized {
			return fmt.Errorf("seerr did not persist initialization")
		}
	}
	return nil
}

func (c *SeerrClient) initialized(ctx context.Context) (bool, error) {
	contents, err := c.jsonRequest(ctx, http.MethodGet, "/api/v1/settings/public", nil, http.StatusOK)
	if err != nil {
		return false, fmt.Errorf("get Seerr public settings: %w", err)
	}
	var status struct {
		Initialized bool `json:"initialized"`
	}
	if err := json.Unmarshal(contents, &status); err != nil {
		return false, fmt.Errorf("decode Seerr public settings: %w", err)
	}
	return status.Initialized, nil
}

func (c *SeerrClient) enableJellyfinLibraries(ctx context.Context) error {
	contents, err := c.jsonRequest(ctx, http.MethodGet, "/api/v1/settings/jellyfin/library?sync=true", nil, http.StatusOK)
	if err != nil {
		return fmt.Errorf("sync Seerr Jellyfin libraries: %w", err)
	}
	var libraries []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(contents, &libraries); err != nil {
		return fmt.Errorf("decode Seerr Jellyfin libraries: %w", err)
	}
	ids := make([]string, 0, len(libraries))
	for _, library := range libraries {
		if library.ID == "" || strings.ContainsAny(library.ID, ",&\r\n") {
			return fmt.Errorf("seerr returned an invalid Jellyfin library ID")
		}
		ids = append(ids, library.ID)
	}
	if len(ids) == 0 {
		return fmt.Errorf("seerr did not discover Jellyfin libraries")
	}
	sort.Strings(ids)
	query := url.Values{"enable": {strings.Join(ids, ",")}}
	if _, err := c.jsonRequest(ctx, http.MethodGet, "/api/v1/settings/jellyfin/library?"+query.Encode(), nil, http.StatusOK); err != nil {
		return fmt.Errorf("enable Seerr Jellyfin libraries: %w", err)
	}
	return nil
}

func (c *SeerrClient) ensureARR(ctx context.Context, app string, apiKey APIKey) error {
	port, root := 7878, "/data/library/movies"
	reservedName := "Radarr (Corsarr)"
	if app == "sonarr" {
		port, root, reservedName = 8989, "/data/library/tv", "Sonarr (Corsarr)"
	}
	connection := map[string]any{
		"hostname": app, "port": port, "apiKey": apiKey.Reveal(), "useSsl": false, "baseUrl": "",
	}
	testResult, err := c.jsonRequest(ctx, http.MethodPost, "/api/v1/settings/"+app+"/test", connection, http.StatusOK)
	if err != nil {
		return fmt.Errorf("test Seerr %s connection: %w", app, err)
	}
	var options struct {
		Profiles []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"profiles"`
		RootFolders []struct {
			Path string `json:"path"`
		} `json:"rootFolders"`
	}
	if err := json.Unmarshal(testResult, &options); err != nil {
		return fmt.Errorf("decode Seerr %s connection options: %w", app, err)
	}
	profileID, profileName, ok := selectSeerrProfile(options.Profiles)
	if !ok || !hasSeerrRoot(options.RootFolders, root) {
		return fmt.Errorf("seerr %s connection lacks an approved profile or root folder", app)
	}
	for key, value := range map[string]any{
		"name": reservedName, "activeProfileId": profileID, "activeProfileName": profileName,
		"activeDirectory": root, "tags": []int{}, "is4k": false, "isDefault": true,
		"externalUrl": "", "syncEnabled": true, "preventSearch": false, "tagRequests": false,
	} {
		connection[key] = value
	}
	if app == "radarr" {
		connection["minimumAvailability"] = "released"
	} else {
		connection["seriesType"], connection["animeSeriesType"] = "standard", "anime"
		connection["animeTags"], connection["enableSeasonFolders"], connection["monitorNewItems"] = []int{}, true, "all"
	}

	existingContents, err := c.jsonRequest(ctx, http.MethodGet, "/api/v1/settings/"+app, nil, http.StatusOK)
	if err != nil {
		return fmt.Errorf("list Seerr %s connections: %w", app, err)
	}
	var existing []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(existingContents, &existing); err != nil {
		return fmt.Errorf("decode Seerr %s connections: %w", app, err)
	}
	method, apiPath, wantedStatus := http.MethodPost, "/api/v1/settings/"+app, http.StatusCreated
	for _, candidate := range existing {
		if candidate.Name == reservedName {
			method, apiPath, wantedStatus = http.MethodPut, fmt.Sprintf("/api/v1/settings/%s/%d", app, candidate.ID), http.StatusOK
			break
		}
	}
	if _, err := c.jsonRequest(ctx, method, apiPath, connection, wantedStatus); err != nil {
		return fmt.Errorf("save Seerr %s connection: %w", app, err)
	}
	return nil
}

func selectSeerrProfile(profiles []struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}) (int, string, bool) {
	if len(profiles) == 0 {
		return 0, "", false
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	for _, profile := range profiles {
		if strings.EqualFold(profile.Name, "Any") {
			return profile.ID, profile.Name, true
		}
	}
	return profiles[0].ID, profiles[0].Name, profiles[0].ID >= 0 && profiles[0].Name != ""
}

func hasSeerrRoot(roots []struct {
	Path string `json:"path"`
}, wanted string) bool {
	for _, root := range roots {
		if root.Path == wanted {
			return true
		}
	}
	return false
}

func (c *SeerrClient) jsonRequest(ctx context.Context, method, apiPath string, value any, wantedStatus int) ([]byte, error) {
	var body io.Reader
	if value != nil {
		contents, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(contents)
	}
	if c.baseURL == nil {
		return nil, fmt.Errorf("seerr endpoint is not an approved loopback HTTP URL")
	}
	endpoint := *c.baseURL
	parsed, err := url.Parse(apiPath)
	if err != nil || !strings.HasPrefix(parsed.Path, "/api/v1/") || parsed.Host != "" {
		return nil, fmt.Errorf("seerr API path is invalid")
	}
	endpoint.Path, endpoint.RawQuery = parsed.Path, parsed.RawQuery
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if value != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != wantedStatus {
		return nil, fmt.Errorf("unexpected HTTP status %d", response.StatusCode)
	}
	return contents, nil
}

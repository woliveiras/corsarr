package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/woliveiras/corsarr/internal/credentials"
)

var qbittorrentCategories = map[string]string{
	"lidarr": "/data/downloads/complete/lidarr",
	"radarr": "/data/downloads/complete/radarr",
	"sonarr": "/data/downloads/complete/sonarr",
}

type QBittorrentClient struct {
	resolver ApplicationEndpointResolver
}

func NewQBittorrentClient(resolver ApplicationEndpointResolver) *QBittorrentClient {
	return &QBittorrentClient{resolver: resolver}
}

func (c *QBittorrentClient) Login(
	ctx context.Context,
	username string,
	password credentials.Secret,
) (*QBittorrentSession, error) {
	baseURL, err := c.baseURL()
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create qBittorrent session: %w", err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	session := &QBittorrentSession{client: httpClient, baseURL: baseURL}
	form := url.Values{"username": {username}, "password": {password.Reveal()}}
	response, err := c.postForm(ctx, session, "/api/v2/auth/login", form)
	if err != nil {
		return nil, fmt.Errorf("authenticate qBittorrent: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return nil, fmt.Errorf("authenticate qBittorrent: %w", err)
	}
	if response.StatusCode != http.StatusOK || strings.TrimSpace(string(contents)) != "Ok." {
		return nil, fmt.Errorf("authenticate qBittorrent: login rejected")
	}
	return session, nil
}

func (c *QBittorrentClient) SetCredentials(
	ctx context.Context,
	session *QBittorrentSession,
	username string,
	password credentials.Secret,
) error {
	return c.setPreferences(ctx, session, map[string]any{
		"web_ui_username": username,
		"web_ui_password": password.Reveal(),
	})

}

func (c *QBittorrentClient) EnsureDownloadPaths(
	ctx context.Context,
	session *QBittorrentSession,
) error {
	return c.setPreferences(ctx, session, map[string]any{
		"save_path":         "/data/downloads/complete",
		"temp_path":         "/data/downloads/incomplete",
		"temp_path_enabled": true,
	})
}

func (c *QBittorrentClient) setPreferences(
	ctx context.Context,
	session *QBittorrentSession,
	values map[string]any,
) error {
	preferences, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("encode qBittorrent preferences: %w", err)
	}
	response, err := c.postForm(
		ctx,
		session,
		"/api/v2/app/setPreferences",
		url.Values{"json": {string(preferences)}},
	)
	if err != nil {
		return fmt.Errorf("set qBittorrent preferences: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("set qBittorrent preferences: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("set qBittorrent preferences: unexpected HTTP status %d", response.StatusCode)
	}
	return nil
}

func (c *QBittorrentClient) EnsureCategories(
	ctx context.Context,
	session *QBittorrentSession,
) error {
	response, err := c.do(ctx, session, http.MethodGet, "/api/v2/torrents/categories", nil)
	if err != nil {
		return fmt.Errorf("list qBittorrent categories: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return fmt.Errorf("list qBittorrent categories: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("list qBittorrent categories: unexpected HTTP status %d", response.StatusCode)
	}
	var existing map[string]struct {
		Name     string `json:"name"`
		SavePath string `json:"savePath"`
	}
	if err := json.Unmarshal(contents, &existing); err != nil {
		return fmt.Errorf("decode qBittorrent categories: %w", err)
	}

	categoryNames := make([]string, 0, len(qbittorrentCategories))
	for name := range qbittorrentCategories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)
	for _, name := range categoryNames {
		desiredPath := qbittorrentCategories[name]
		category, found := existing[name]
		if found && path.Clean(category.SavePath) == desiredPath {
			continue
		}
		operation := "/api/v2/torrents/createCategory"
		if found {
			operation = "/api/v2/torrents/editCategory"
		}
		response, err := c.postForm(ctx, session, operation, url.Values{
			"category": {name},
			"savePath": {desiredPath},
		})
		if err != nil {
			return fmt.Errorf("reconcile qBittorrent category %s: %w", name, err)
		}
		if _, err := boundedResponse(response); err != nil {
			return fmt.Errorf("reconcile qBittorrent category %s: %w", name, err)
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf(
				"reconcile qBittorrent category %s: unexpected HTTP status %d",
				name,
				response.StatusCode,
			)
		}
	}
	return nil
}

func (c *QBittorrentClient) baseURL() (*url.URL, error) {
	endpoint, err := c.resolver.ResolveApplicationURL(qbittorrentApplicationID)
	if err != nil {
		return nil, fmt.Errorf("resolve qBittorrent endpoint: %w", err)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse qBittorrent endpoint: %w", err)
	}
	hostIP := net.ParseIP(parsed.Hostname())
	if parsed.Scheme != "http" || hostIP == nil || !hostIP.IsLoopback() || parsed.User != nil {
		return nil, fmt.Errorf("qBittorrent endpoint is not an approved loopback HTTP URL")
	}
	parsed.Path = "/"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func (c *QBittorrentClient) postForm(
	ctx context.Context,
	session *QBittorrentSession,
	apiPath string,
	form url.Values,
) (*http.Response, error) {
	return c.do(ctx, session, http.MethodPost, apiPath, strings.NewReader(form.Encode()))
}

func (c *QBittorrentClient) do(
	ctx context.Context,
	session *QBittorrentSession,
	method string,
	apiPath string,
	body *strings.Reader,
) (*http.Response, error) {
	if session == nil || session.client == nil || session.baseURL == nil {
		return nil, fmt.Errorf("qBittorrent session is invalid")
	}
	endpoint := *session.baseURL
	endpoint.Path = apiPath
	var requestBody *strings.Reader
	if body == nil {
		requestBody = strings.NewReader("")
	} else {
		requestBody = body
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Referer", session.baseURL.String())
	request.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return session.client.Do(request)
}

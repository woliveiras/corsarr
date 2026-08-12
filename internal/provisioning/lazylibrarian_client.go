package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	lazyLibrarianDownloadPath = "/data/downloads/complete/lazylibrarian"
	lazyLibrarianLibraryPath  = "/data/library/books"
)

type LazyLibrarianClient struct {
	resolver     ApplicationEndpointResolver
	client       *http.Client
	pollInterval time.Duration
	restartGrace time.Duration
	restartWait  time.Duration
}

func NewLazyLibrarianClient(resolver ApplicationEndpointResolver) *LazyLibrarianClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	return &LazyLibrarianClient{
		resolver: resolver,
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		pollInterval: 500 * time.Millisecond,
		restartGrace: time.Second,
		restartWait:  2 * time.Minute,
	}
}

func (c *LazyLibrarianClient) EnsureSetup(
	ctx context.Context,
	setup LazyLibrarianSetupRequest,
) error {
	if setup.Username != lazyLibrarianUsername ||
		len(setup.Password.Reveal()) < 16 ||
		!arrAPIKeyPattern.MatchString(setup.APIKey.Reveal()) {
		return fmt.Errorf("LazyLibrarian credentials are invalid")
	}
	if setup.ConfigureQBittorrent &&
		(setup.QBittorrentUsername != qbittorrentUsername ||
			setup.QBittorrentPassword.Reveal() == "") {
		return fmt.Errorf("qBittorrent credentials are invalid")
	}

	baseURL, err := c.baseURL()
	if err != nil {
		return err
	}
	form := url.Values{
		"api_enabled":   {"1"},
		"api_key":       {setup.APIKey.Reveal()},
		"auth_type":     {"BASIC"},
		"download_dir":  {lazyLibrarianDownloadPath},
		"ebook_dir":     {lazyLibrarianLibraryPath},
		"hostredact":    {"1"},
		"http_pass":     {setup.Password.Reveal()},
		"http_user":     {setup.Username},
		"logfileredact": {"1"},
		"logredact":     {"1"},
	}
	if setup.ConfigureQBittorrent {
		form.Set("qbittorrent_dir", lazyLibrarianDownloadPath)
		form.Set("qbittorrent_host", "http://qbittorrent")
		form.Set("qbittorrent_label", lazyLibrarianApplicationID)
		form.Set("qbittorrent_pass", setup.QBittorrentPassword.Reveal())
		form.Set("qbittorrent_port", "8081")
		form.Set("qbittorrent_user", setup.QBittorrentUsername)
		form.Set("tor_downloader_qbittorrent", "1")
	}
	if err := c.postConfiguration(ctx, baseURL, setup, form); err != nil {
		return err
	}
	if err := c.restart(ctx, baseURL, setup); err != nil {
		return err
	}
	if err := c.waitForAPI(ctx, baseURL, setup); err != nil {
		return err
	}
	if setup.ConfigureQBittorrent {
		if err := c.verifyQBittorrent(ctx, baseURL, setup); err != nil {
			return err
		}
	}
	return nil
}

func (c *LazyLibrarianClient) postConfiguration(
	ctx context.Context,
	baseURL *url.URL,
	setup LazyLibrarianSetupRequest,
	form url.Values,
) error {
	endpoint := *baseURL
	endpoint.Path = "/config_update"
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.String(),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("create LazyLibrarian configuration request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(setup.Username, setup.Password.Reveal())
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("save LazyLibrarian configuration: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("save LazyLibrarian configuration: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf(
			"save LazyLibrarian configuration: unexpected HTTP status %d",
			response.StatusCode,
		)
	}
	return nil
}

func (c *LazyLibrarianClient) restart(
	ctx context.Context,
	baseURL *url.URL,
	setup LazyLibrarianSetupRequest,
) error {
	response, err := c.apiRequest(ctx, baseURL, setup, "restart")
	if err != nil {
		return fmt.Errorf("restart LazyLibrarian after configuration: %w", err)
	}
	if _, err := boundedResponse(response); err != nil {
		return fmt.Errorf("restart LazyLibrarian after configuration: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"restart LazyLibrarian after configuration: unexpected HTTP status %d",
			response.StatusCode,
		)
	}
	return nil
}

func (c *LazyLibrarianClient) waitForAPI(
	ctx context.Context,
	baseURL *url.URL,
	setup LazyLibrarianSetupRequest,
) error {
	waitContext, cancel := context.WithTimeout(ctx, c.restartWait)
	defer cancel()
	if err := waitForLazyLibrarianPoll(waitContext, c.restartGrace); err != nil {
		return fmt.Errorf("LazyLibrarian did not return after configuration: %w", err)
	}
	consecutiveSuccesses := 0
	for {
		response, err := c.apiRequest(waitContext, baseURL, setup, "getVersion")
		if err == nil {
			contents, readErr := boundedResponse(response)
			if readErr == nil && response.StatusCode == http.StatusOK {
				var status struct {
					Success bool `json:"Success"`
				}
				if json.Unmarshal(contents, &status) == nil && status.Success {
					consecutiveSuccesses++
					if consecutiveSuccesses == 2 {
						return nil
					}
				} else {
					consecutiveSuccesses = 0
				}
			} else {
				consecutiveSuccesses = 0
			}
		} else {
			consecutiveSuccesses = 0
		}
		if err := waitForLazyLibrarianPoll(waitContext, c.pollInterval); err != nil {
			return fmt.Errorf("LazyLibrarian did not return after configuration: %w", waitContext.Err())
		}
	}
}

func waitForLazyLibrarianPoll(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *LazyLibrarianClient) verifyQBittorrent(
	ctx context.Context,
	baseURL *url.URL,
	setup LazyLibrarianSetupRequest,
) error {
	endpoint := *baseURL
	endpoint.Path = "/test_qbittorrent"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create LazyLibrarian qBittorrent verification request: %w", err)
	}
	request.SetBasicAuth(setup.Username, setup.Password.Reveal())
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("verify LazyLibrarian qBittorrent connection: %w", err)
	}
	contents, err := boundedResponse(response)
	if err != nil {
		return fmt.Errorf("verify LazyLibrarian qBittorrent connection: %w", err)
	}
	if response.StatusCode != http.StatusOK ||
		!strings.Contains(strings.ToLower(string(contents)), "login successful") {
		return fmt.Errorf("LazyLibrarian could not connect to qBittorrent")
	}
	return nil
}

func (c *LazyLibrarianClient) apiRequest(
	ctx context.Context,
	baseURL *url.URL,
	setup LazyLibrarianSetupRequest,
	command string,
) (*http.Response, error) {
	endpoint := *baseURL
	endpoint.Path = "/api"
	query := endpoint.Query()
	query.Set("apikey", setup.APIKey.Reveal())
	query.Set("cmd", command)
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(setup.Username, setup.Password.Reveal())
	response, err := c.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// A url.Error includes the complete request URL. The LazyLibrarian API
		// requires its key in the query, so never propagate that error verbatim.
		return nil, fmt.Errorf("perform LazyLibrarian API command %s: request failed", command)
	}
	return response, nil
}

func (c *LazyLibrarianClient) baseURL() (*url.URL, error) {
	endpoint, err := c.resolver.ResolveApplicationURL(lazyLibrarianApplicationID)
	if err != nil {
		return nil, fmt.Errorf("resolve LazyLibrarian endpoint: %w", err)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse LazyLibrarian endpoint: %w", err)
	}
	hostIP := net.ParseIP(parsed.Hostname())
	if parsed.Scheme != "http" || hostIP == nil || !hostIP.IsLoopback() ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("LazyLibrarian endpoint is not an approved loopback HTTP URL")
	}
	parsed.Path = ""
	parsed.RawPath = ""
	return parsed, nil
}

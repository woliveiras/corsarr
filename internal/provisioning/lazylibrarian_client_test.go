package provisioning

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestLazyLibrarianClientPersistsAndVerifiesManagedSetup(t *testing.T) {
	var configured bool
	var restarted bool
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/config_update":
			if request.Method != http.MethodPost {
				t.Errorf("expected configuration POST, got %s", request.Method)
			}
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse LazyLibrarian configuration: %v", err)
			}
			assertLazyLibrarianForm(t, request)
			configured = true
			response.WriteHeader(http.StatusSeeOther)
		case "/api":
			assertLazyLibrarianAuthorization(t, request)
			if request.URL.Query().Get("apikey") != "0123456789abcdef0123456789abcdef" {
				t.Error("expected generated LazyLibrarian API key")
			}
			switch request.URL.Query().Get("cmd") {
			case "restart":
				restarted = true
			case "getVersion":
				if !restarted {
					t.Error("verified API before restart")
				}
			default:
				t.Errorf("unexpected API command %q", request.URL.Query().Get("cmd"))
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"Success": true})
		case "/test_qbittorrent":
			assertLazyLibrarianAuthorization(t, request)
			_, _ = response.Write([]byte("qBittorrent login successful, api: 2.11"))
		default:
			t.Errorf("unexpected LazyLibrarian endpoint %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewLazyLibrarianClient(readinessResolver{url: server.URL})
	client.pollInterval = time.Millisecond
	client.restartGrace = time.Millisecond
	client.restartWait = time.Second
	err := client.EnsureSetup(context.Background(), LazyLibrarianSetupRequest{
		Username:             "corsarr",
		Password:             credentials.NewSecret("lazylibrarian-private-password"),
		APIKey:               APIKey{value: "0123456789abcdef0123456789abcdef"},
		ConfigureQBittorrent: true,
		QBittorrentUsername:  "corsarr",
		QBittorrentPassword:  credentials.NewSecret("qbittorrent-private-password"),
	})
	if err != nil {
		t.Fatalf("ensure LazyLibrarian setup: %v", err)
	}
	if !configured || !restarted {
		t.Fatalf("expected persisted and restarted setup: configured=%t restarted=%t", configured, restarted)
	}
}

func TestLazyLibrarianClientDoesNotExposeAPIKeyInTransportErrors(t *testing.T) {
	const apiKey = "0123456789abcdef0123456789abcdef"
	client := NewLazyLibrarianClient(readinessResolver{})
	client.client.Transport = lazyLibrarianRoundTripper(func(request *http.Request) (*http.Response, error) {
		return nil, errors.New(request.URL.String())
	})
	baseURL, err := url.Parse("http://127.0.0.1:5299")
	if err != nil {
		t.Fatalf("parse test URL: %v", err)
	}

	_, err = client.apiRequest(context.Background(), baseURL, LazyLibrarianSetupRequest{
		Username: "corsarr",
		Password: credentials.NewSecret("lazylibrarian-private-password"),
		APIKey:   APIKey{value: apiKey},
	}, "getVersion")
	if err == nil {
		t.Fatal("expected transport failure")
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("transport error exposed LazyLibrarian API key: %v", err)
	}
}

func assertLazyLibrarianForm(t *testing.T, request *http.Request) {
	t.Helper()
	assertLazyLibrarianAuthorization(t, request)
	want := map[string]string{
		"api_enabled":                "1",
		"api_key":                    "0123456789abcdef0123456789abcdef",
		"auth_type":                  "BASIC",
		"download_dir":               "/data/downloads/complete/lazylibrarian",
		"ebook_dir":                  "/data/library/books",
		"http_user":                  "corsarr",
		"http_pass":                  "lazylibrarian-private-password",
		"qbittorrent_dir":            "/data/downloads/complete/lazylibrarian",
		"qbittorrent_host":           "http://qbittorrent",
		"qbittorrent_label":          "lazylibrarian",
		"qbittorrent_pass":           "qbittorrent-private-password",
		"qbittorrent_port":           "8081",
		"qbittorrent_user":           "corsarr",
		"tor_downloader_qbittorrent": "1",
	}
	for key, value := range want {
		if request.Form.Get(key) != value {
			t.Errorf("expected %s=%q, got %q", key, value, request.Form.Get(key))
		}
	}
}

func assertLazyLibrarianAuthorization(t *testing.T, request *http.Request) {
	t.Helper()
	want := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte("corsarr:lazylibrarian-private-password"),
	)
	if request.Header.Get("Authorization") != want {
		t.Error("expected managed LazyLibrarian basic authentication")
	}
}

type lazyLibrarianRoundTripper func(*http.Request) (*http.Response, error)

func (f lazyLibrarianRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

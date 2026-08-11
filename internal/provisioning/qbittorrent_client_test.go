package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestQBittorrentClientAuthenticatesAndSetsCredentials(t *testing.T) {
	var endpoint string
	operations := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Referer() != endpoint+"/" {
			t.Errorf("expected exact loopback referer %q, got %q", endpoint+"/", request.Referer())
		}
		switch request.URL.Path {
		case "/api/v2/auth/login":
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse login form: %v", err)
			}
			operations = append(operations, "login:"+request.Form.Get("username")+":"+request.Form.Get("password"))
			http.SetCookie(response, &http.Cookie{Name: "SID", Value: "session", Path: "/"})
			_, _ = response.Write([]byte("Ok."))
		case "/api/v2/app/setPreferences":
			if _, err := request.Cookie("SID"); err != nil {
				t.Errorf("expected authenticated session cookie: %v", err)
			}
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse preferences form: %v", err)
			}
			var preferences map[string]string
			if err := json.Unmarshal([]byte(request.Form.Get("json")), &preferences); err != nil {
				t.Errorf("decode preferences: %v", err)
			}
			operations = append(operations, "set:"+preferences["web_ui_username"]+":"+preferences["web_ui_password"])
		default:
			t.Errorf("unexpected endpoint %s", request.URL.Path)
		}
	}))
	defer server.Close()
	endpoint = server.URL

	client := NewQBittorrentClient(readinessResolver{url: server.URL})
	session, err := client.Login(context.Background(), "admin", credentials.NewSecret("temporary"))
	if err != nil {
		t.Fatalf("login qBittorrent: %v", err)
	}
	if err := client.SetCredentials(
		context.Background(),
		session,
		"corsarr",
		credentials.NewSecret("permanent"),
	); err != nil {
		t.Fatalf("set qBittorrent credentials: %v", err)
	}
	want := []string{"login:admin:temporary", "set:corsarr:permanent"}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("unexpected operations\nwant: %v\n got: %v", want, operations)
	}
}

func TestQBittorrentClientAcceptsNoContentLoginResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.SetCookie(response, &http.Cookie{Name: "SID", Value: "session", Path: "/"})
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewQBittorrentClient(readinessResolver{url: server.URL})
	if _, err := client.Login(
		context.Background(),
		"admin",
		credentials.NewSecret("temporary"),
	); err != nil {
		t.Fatalf("accept successful qBittorrent login without a response body: %v", err)
	}
}

func TestQBittorrentClientRejectsNoContentLoginWithoutSessionCookie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewQBittorrentClient(readinessResolver{url: server.URL})
	if _, err := client.Login(
		context.Background(),
		"admin",
		credentials.NewSecret("temporary"),
	); err == nil {
		t.Fatal("expected a qBittorrent login without a session cookie to fail")
	}
}

func TestQBittorrentClientRealLoginContract(t *testing.T) {
	endpoint := os.Getenv("CORSARR_TEST_QBITTORRENT_URL")
	password := os.Getenv("CORSARR_TEST_QBITTORRENT_PASSWORD")
	if endpoint == "" || password == "" {
		t.Skip("set the bounded qBittorrent test endpoint and temporary credential")
	}

	client := NewQBittorrentClient(readinessResolver{url: endpoint})
	if _, err := client.Login(
		context.Background(),
		"admin",
		credentials.NewSecret(password),
	); err != nil {
		t.Fatalf("authenticate against the running qBittorrent contract: %v", err)
	}
}

func TestQBittorrentClientReconcilesApprovedCategories(t *testing.T) {
	operations := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(response, &http.Cookie{Name: "SID", Value: "session", Path: "/"})
			_, _ = response.Write([]byte("Ok."))
		case "/api/v2/torrents/categories":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"radarr": map[string]string{"name": "radarr", "savePath": "/wrong"},
				"sonarr": map[string]string{"name": "sonarr", "savePath": "/data/downloads/complete/sonarr"},
			})
		case "/api/v2/torrents/createCategory", "/api/v2/torrents/editCategory":
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse category form: %v", err)
			}
			operations = append(operations, request.URL.Path+":"+request.Form.Get("category")+":"+request.Form.Get("savePath"))
		default:
			t.Errorf("unexpected endpoint %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewQBittorrentClient(readinessResolver{url: server.URL})
	session, err := client.Login(context.Background(), "corsarr", credentials.NewSecret("password"))
	if err != nil {
		t.Fatalf("login qBittorrent: %v", err)
	}
	if err := client.EnsureCategories(context.Background(), session); err != nil {
		t.Fatalf("reconcile categories: %v", err)
	}
	want := []string{
		"/api/v2/torrents/createCategory:lidarr:/data/downloads/complete/lidarr",
		"/api/v2/torrents/editCategory:radarr:/data/downloads/complete/radarr",
	}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("unexpected category operations\nwant: %v\n got: %v", want, operations)
	}
}

func TestQBittorrentClientSetsApprovedDownloadPaths(t *testing.T) {
	var preferences map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(response, &http.Cookie{Name: "SID", Value: "session", Path: "/"})
			_, _ = response.Write([]byte("Ok."))
		case "/api/v2/app/setPreferences":
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse preferences form: %v", err)
			}
			if err := json.Unmarshal([]byte(request.Form.Get("json")), &preferences); err != nil {
				t.Errorf("decode preferences: %v", err)
			}
		default:
			t.Errorf("unexpected endpoint %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewQBittorrentClient(readinessResolver{url: server.URL})
	session, err := client.Login(context.Background(), "corsarr", credentials.NewSecret("password"))
	if err != nil {
		t.Fatalf("login qBittorrent: %v", err)
	}
	if err := client.EnsureDownloadPaths(context.Background(), session); err != nil {
		t.Fatalf("set download paths: %v", err)
	}
	if preferences["save_path"] != "/data/downloads/complete" ||
		preferences["temp_path"] != "/data/downloads/incomplete" ||
		preferences["temp_path_enabled"] != true {
		t.Fatalf("unexpected download preferences %#v", preferences)
	}
}

func TestQBittorrentClientDoesNotRedirectCredential(t *testing.T) {
	targetReceivedCredential := false
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err == nil && request.Form.Get("password") != "" {
			targetReceivedCredential = true
		}
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL, http.StatusFound)
	}))
	defer origin.Close()

	client := NewQBittorrentClient(readinessResolver{url: origin.URL})
	if _, err := client.Login(context.Background(), "admin", credentials.NewSecret("temporary")); err == nil {
		t.Fatal("expected redirected login to fail")
	}
	if targetReceivedCredential {
		t.Fatal("expected redirected endpoint not to receive credential")
	}
}

package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestSeerrClientInitializesFromJellyfinAndCreatesArrConnections(t *testing.T) {
	initialized := false
	loginAttempts := 0
	created := make(map[string]map[string]any)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case "GET /api/v1/settings/public":
			_ = json.NewEncoder(response).Encode(map[string]bool{"initialized": initialized})
		case "POST /api/v1/auth/jellyfin":
			loginAttempts++
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if loginAttempts == 1 {
				if _, hasHostname := body["hostname"]; hasHostname {
					t.Errorf("expected existing Jellyfin login before configuration %#v", body)
				}
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			if body["username"] != "corsarr" || body["password"] != "private-password" ||
				body["hostname"] != "jellyfin" || body["serverType"] != float64(2) ||
				body["urlBase"] != "" {
				t.Errorf("unexpected Seerr login %#v", body)
			}
			http.SetCookie(response, &http.Cookie{Name: "connect.sid", Value: "local-session", Path: "/api"})
			_, _ = response.Write([]byte(`{"id":1}`))
		case "GET /api/v1/settings/jellyfin/library":
			assertSeerrSession(t, request)
			if request.URL.Query().Get("sync") == "true" {
				_ = json.NewEncoder(response).Encode([]map[string]any{
					{"id": "movies", "name": "Movies (Corsarr)", "enabled": false},
					{"id": "shows", "name": "TV Shows (Corsarr)", "enabled": false},
				})
				return
			}
			if request.URL.Query().Get("enable") != "movies,shows" {
				t.Errorf("unexpected enabled libraries %q", request.URL.Query().Get("enable"))
			}
			_, _ = response.Write([]byte("[]"))
		case "GET /api/v1/settings/radarr", "GET /api/v1/settings/sonarr":
			assertSeerrSession(t, request)
			_, _ = response.Write([]byte("[]"))
		case "POST /api/v1/settings/radarr/test":
			writeSeerrTestResponse(response, "/data/library/movies")
		case "POST /api/v1/settings/sonarr/test":
			writeSeerrTestResponse(response, "/data/library/tv")
		case "POST /api/v1/settings/radarr":
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			created["radarr"] = body
			response.WriteHeader(http.StatusCreated)
		case "POST /api/v1/settings/sonarr":
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			created["sonarr"] = body
			response.WriteHeader(http.StatusCreated)
		case "POST /api/v1/settings/initialize":
			initialized = true
			_, _ = response.Write([]byte(`{"initialized":true}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewSeerrClient(readinessResolver{url: server.URL})
	err := client.EnsureSetup(
		context.Background(),
		credentials.NewSecret("private-password"),
		APIKey{value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		APIKey{value: "cccccccccccccccccccccccccccccccc"},
	)
	if err != nil {
		t.Fatalf("configure Seerr: %v", err)
	}
	if !initialized {
		t.Fatal("expected Seerr initialization")
	}
	if loginAttempts != 2 {
		t.Fatalf("expected existing login followed by first-run configuration, got %d attempts", loginAttempts)
	}
	for _, app := range []string{"radarr", "sonarr"} {
		settings := created[app]
		if settings["name"] != map[string]string{"radarr": "Radarr (Corsarr)", "sonarr": "Sonarr (Corsarr)"}[app] ||
			settings["activeProfileName"] != "Any" || settings["isDefault"] != true || settings["syncEnabled"] != true {
			t.Fatalf("unexpected %s settings %#v", app, settings)
		}
	}
}

func assertSeerrSession(t *testing.T, request *http.Request) {
	t.Helper()
	cookie, err := request.Cookie("connect.sid")
	if err != nil || cookie.Value != "local-session" {
		t.Fatalf("expected Seerr session cookie, got %#v (%v)", cookie, err)
	}
}

func writeSeerrTestResponse(response http.ResponseWriter, root string) {
	_ = json.NewEncoder(response).Encode(map[string]any{
		"profiles":    []map[string]any{{"id": 5, "name": "HD"}, {"id": 1, "name": "Any"}},
		"rootFolders": []map[string]any{{"id": 1, "path": root}},
		"tags":        []any{},
		"urlBase":     "",
	})
}

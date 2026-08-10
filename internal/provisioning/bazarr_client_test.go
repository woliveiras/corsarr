package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestBazarrClientConnectsRadarrAndSonarrThroughOfficialSettingsAPI(t *testing.T) {
	var saved url.Values
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-API-KEY") != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
			t.Errorf("expected Bazarr API key")
		}
		switch request.Method + " " + request.URL.Path {
		case "POST /api/system/settings":
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse Bazarr settings: %v", err)
			}
			saved = request.PostForm
			response.WriteHeader(http.StatusNoContent)
		case "GET /api/system/settings":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"general": map[string]any{"use_radarr": true, "use_sonarr": true},
				"radarr": map[string]any{
					"ip": "radarr", "port": 7878, "base_url": "/", "ssl": false,
					"apikey": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				},
				"sonarr": map[string]any{
					"ip": "sonarr", "port": 8989, "base_url": "/", "ssl": false,
					"apikey": "cccccccccccccccccccccccccccccccc",
				},
			})
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewBazarrClient(readinessResolver{url: server.URL})
	err := client.EnsureARRConnections(
		context.Background(),
		APIKey{value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		APIKey{value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		APIKey{value: "cccccccccccccccccccccccccccccccc"},
	)
	if err != nil {
		t.Fatalf("configure Bazarr: %v", err)
	}
	assertFormValue(t, saved, "settings-general-use_radarr", "true")
	assertFormValue(t, saved, "settings-general-use_sonarr", "true")
	assertFormValue(t, saved, "settings-radarr-ip", "radarr")
	assertFormValue(t, saved, "settings-radarr-apikey", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	assertFormValue(t, saved, "settings-sonarr-ip", "sonarr")
	assertFormValue(t, saved, "settings-sonarr-apikey", "cccccccccccccccccccccccccccccccc")
}

func TestBazarrProvisionerReadsOnlyFixedApplicationCredentials(t *testing.T) {
	reader := &multiCredentialReader{keys: map[string]APIKey{
		"radarr": {value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		"sonarr": {value: "cccccccccccccccccccccccccccccccc"},
	}}
	bazarrReader := &recordingBazarrCredentialReader{key: APIKey{value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	client := &recordingBazarrConfigurator{}
	provisioner := NewBazarrProvisioner(bazarrReader, reader, client)

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "bazarr"); err != nil {
		t.Fatalf("provision Bazarr: %v", err)
	}
	if bazarrReader.rootPath != "/host/Corsarr" {
		t.Fatalf("unexpected Bazarr credential root %q", bazarrReader.rootPath)
	}
	if len(reader.applications) != 2 || reader.applications[0] != "radarr" || reader.applications[1] != "sonarr" {
		t.Fatalf("unexpected Arr credential reads %v", reader.applications)
	}
	if client.bazarrKey.Reveal() == "" || client.radarrKey.Reveal() == "" || client.sonarrKey.Reveal() == "" {
		t.Fatal("expected all backend credentials")
	}
}

func assertFormValue(t *testing.T, values url.Values, name string, want string) {
	t.Helper()
	if got := values.Get(name); got != want {
		t.Fatalf("expected %s=%q, got %q", name, want, got)
	}
}

type recordingBazarrCredentialReader struct {
	rootPath string
	key      APIKey
}

func (r *recordingBazarrCredentialReader) Read(rootPath string) (APIKey, error) {
	r.rootPath = rootPath
	return r.key, nil
}

type recordingBazarrConfigurator struct {
	bazarrKey APIKey
	radarrKey APIKey
	sonarrKey APIKey
}

func (c *recordingBazarrConfigurator) EnsureARRConnections(
	_ context.Context,
	bazarrKey APIKey,
	radarrKey APIKey,
	sonarrKey APIKey,
) error {
	c.bazarrKey = bazarrKey
	c.radarrKey = radarrKey
	c.sonarrKey = sonarrKey
	return nil
}

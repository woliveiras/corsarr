package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestARRClientCreatesQBittorrentFromOfficialSchema(t *testing.T) {
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Api-Key") != "0123456789abcdef0123456789abcdef" {
			t.Errorf("expected Arr API key")
		}
		switch request.Method + " " + request.URL.Path {
		case "GET /api/v3/downloadclient":
			_, _ = response.Write([]byte("[]"))
		case "GET /api/v3/downloadclient/schema":
			_ = json.NewEncoder(response).Encode([]map[string]any{{
				"name":               "qBittorrent",
				"implementation":     "QBittorrent",
				"implementationName": "qBittorrent",
				"configContract":     "QBittorrentSettings",
				"fields": []map[string]any{
					{"name": "host", "value": "localhost"},
					{"name": "port", "value": 8080},
					{"name": "useSsl", "value": false},
					{"name": "urlBase", "value": ""},
					{"name": "apiKey", "value": ""},
					{"name": "username", "value": ""},
					{"name": "password", "value": ""},
					{"name": "movieCategory", "value": "radarr"},
				},
			}})
		case "POST /api/v3/downloadclient":
			if err := json.NewDecoder(request.Body).Decode(&created); err != nil {
				t.Errorf("decode download client: %v", err)
			}
			response.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	err := client.EnsureQBittorrentDownloadClient(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"corsarr",
		credentials.NewSecret("qbit-password"),
	)
	if err != nil {
		t.Fatalf("create qBittorrent download client: %v", err)
	}
	if created["name"] != "qBittorrent (Corsarr)" || created["enable"] != true {
		t.Fatalf("unexpected provider metadata %#v", created)
	}
	fields := providerFieldsByName(t, created)
	assertProviderField(t, fields, "host", "qbittorrent")
	assertProviderField(t, fields, "port", float64(8080))
	assertProviderField(t, fields, "username", "corsarr")
	assertProviderField(t, fields, "password", "qbit-password")
	assertProviderField(t, fields, "movieCategory", "radarr")
}

func TestARRClientUpdatesOnlyCorsarrOwnedDownloadClient(t *testing.T) {
	updatedID := ""
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case "GET /api/v1/downloadclient":
			_ = json.NewEncoder(response).Encode([]map[string]any{
				{"id": 7, "name": "My qBittorrent", "implementation": "QBittorrent"},
				{
					"id": 9, "name": "qBittorrent (Corsarr)", "implementation": "QBittorrent",
					"configContract": "QBittorrentSettings",
					"fields": []map[string]any{
						{"name": "host", "value": "old"},
						{"name": "port", "value": 1},
						{"name": "useSsl", "value": true},
						{"name": "urlBase", "value": "/old"},
						{"name": "apiKey", "value": "old"},
						{"name": "username", "value": "old"},
						{"name": "password", "value": "********"},
						{"name": "musicCategory", "value": "old"},
					},
				},
			})
		case "PUT /api/v1/downloadclient/9":
			updatedID = "9"
			response.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	err := client.EnsureQBittorrentDownloadClient(
		context.Background(),
		"lidarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"corsarr",
		credentials.NewSecret("qbit-password"),
	)
	if err != nil {
		t.Fatalf("update qBittorrent download client: %v", err)
	}
	if updatedID != "9" {
		t.Fatalf("expected only Corsarr provider updated, got %q", updatedID)
	}
}

func providerFieldsByName(t *testing.T, provider map[string]any) map[string]any {
	t.Helper()
	fields := map[string]any{}
	for _, rawField := range provider["fields"].([]any) {
		field := rawField.(map[string]any)
		fields[field["name"].(string)] = field["value"]
	}
	return fields
}

func assertProviderField(t *testing.T, fields map[string]any, name string, want any) {
	t.Helper()
	if fields[name] != want {
		t.Fatalf("unexpected field %s: want %#v, got %#v", name, want, fields[name])
	}
}

package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProwlarrClientCreatesCorsarrApplicationFromSchema(t *testing.T) {
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Api-Key") != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
			t.Errorf("expected Prowlarr API key")
		}
		switch request.Method + " " + request.URL.Path {
		case "GET /api/v1/applications":
			_, _ = response.Write([]byte("[]"))
		case "GET /api/v1/applications/schema":
			_ = json.NewEncoder(response).Encode([]map[string]any{{
				"name": "Radarr", "implementation": "Radarr", "configContract": "RadarrSettings",
				"fields": []map[string]any{
					{"name": "prowlarrUrl", "value": "http://localhost:9696"},
					{"name": "baseUrl", "value": "http://localhost:7878"},
					{"name": "apiKey", "value": ""},
					{"name": "authUsername", "value": ""},
					{"name": "authPassword", "value": ""},
					{"name": "syncCategories", "value": []any{2000.0}},
				},
			}})
		case "POST /api/v1/applications":
			if err := json.NewDecoder(request.Body).Decode(&created); err != nil {
				t.Errorf("decode Prowlarr application: %v", err)
			}
			response.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProwlarrClient(readinessResolver{url: server.URL})
	err := client.EnsureApplication(
		context.Background(),
		"radarr",
		APIKey{value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		APIKey{value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	)
	if err != nil {
		t.Fatalf("create Prowlarr application: %v", err)
	}
	if created["name"] != "Radarr (Corsarr)" || created["syncLevel"] != "fullSync" {
		t.Fatalf("unexpected Prowlarr metadata %#v", created)
	}
	fields := providerFieldsByName(t, created)
	assertProviderField(t, fields, "prowlarrUrl", "http://prowlarr:9696")
	assertProviderField(t, fields, "baseUrl", "http://radarr:7878")
	assertProviderField(t, fields, "apiKey", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
}

func TestProwlarrProvisionerReadsBothFixedCredentials(t *testing.T) {
	reader := &multiCredentialReader{keys: map[string]APIKey{
		"prowlarr": {value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		"sonarr":   {value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	}}
	client := &recordingProwlarrConfigurator{}
	provisioner := NewProwlarrProvisioner(reader, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"sonarr",
		[]string{"prowlarr", "sonarr"},
	); err != nil {
		t.Fatalf("provision Prowlarr application: %v", err)
	}
	if client.applicationID != "sonarr" || client.prowlarrKey.Reveal() == "" || client.targetKey.Reveal() == "" {
		t.Fatalf("unexpected Prowlarr configuration %#v", client)
	}
	if len(reader.applications) != 2 || reader.applications[0] != "prowlarr" || reader.applications[1] != "sonarr" {
		t.Fatalf("unexpected credential reads %v", reader.applications)
	}
}

func TestProwlarrProvisionerSkipsUnselectedProwlarr(t *testing.T) {
	reader := &multiCredentialReader{}
	client := &recordingProwlarrConfigurator{}
	provisioner := NewProwlarrProvisioner(reader, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"sonarr",
		[]string{"sonarr"},
	); err != nil {
		t.Fatalf("skip unselected Prowlarr: %v", err)
	}
	if len(reader.applications) != 0 || client.applicationID != "" {
		t.Fatalf("expected external indexer manager to remain untouched")
	}
}

type multiCredentialReader struct {
	keys         map[string]APIKey
	applications []string
}

func (r *multiCredentialReader) Read(_ string, applicationID string) (APIKey, error) {
	r.applications = append(r.applications, applicationID)
	return r.keys[applicationID], nil
}

type recordingProwlarrConfigurator struct {
	applicationID string
	prowlarrKey   APIKey
	targetKey     APIKey
}

func (c *recordingProwlarrConfigurator) EnsureApplication(
	_ context.Context,
	applicationID string,
	prowlarrKey APIKey,
	targetKey APIKey,
) error {
	c.applicationID = applicationID
	c.prowlarrKey = prowlarrKey
	c.targetKey = targetKey
	return nil
}

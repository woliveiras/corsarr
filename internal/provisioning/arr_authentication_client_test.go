package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestARRClientActivatesFormsAuthenticationAndPreservesHostConfig(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		if request.Header.Get("X-Api-Key") != "0123456789abcdef0123456789abcdef" {
			t.Fatal("expected API key authentication")
		}
		switch requests {
		case 1, 2:
			if request.Method != http.MethodGet || request.URL.Path != "/api/v3/config/host" {
				t.Fatalf("unexpected inspection %s %s", request.Method, request.URL.Path)
			}
			_ = json.NewEncoder(response).Encode(map[string]any{
				"id": 1, "bindAddress": "*", "port": 7878,
				"authenticationMethod": "none", "authenticationRequired": "enabled",
				"username": "", "password": "", "passwordConfirmation": "",
			})
		case 3:
			if request.Method != http.MethodPut || request.URL.Path != "/api/v3/config/host/1" {
				t.Fatalf("unexpected activation %s %s", request.Method, request.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode host config: %v", err)
			}
			if body["bindAddress"] != "*" || body["port"] != float64(7878) {
				t.Fatalf("existing host configuration was not preserved: %#v", body)
			}
			if body["authenticationMethod"] != "forms" ||
				body["authenticationRequired"] != "disabledForLocalAddresses" ||
				body["username"] != "corsarr" || body["password"] != "private-password" ||
				body["passwordConfirmation"] != "private-password" {
				t.Fatalf("unexpected authentication configuration %#v", body)
			}
			response.WriteHeader(http.StatusAccepted)
		case 4:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"id": 1, "authenticationMethod": "forms",
				"authenticationRequired": "disabledForLocalAddresses", "username": "corsarr",
			})
		default:
			t.Fatalf("unexpected request %d", requests)
		}
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	inspection, err := client.InspectAuthentication(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
	)
	if err != nil || !inspection.BootstrapRequired {
		t.Fatalf("inspect pristine authentication: %#v, %v", inspection, err)
	}
	result, err := client.ActivateAuthentication(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"corsarr",
		credentials.NewSecret("private-password"),
	)
	if err != nil || !result.CredentialAccepted {
		t.Fatalf("activate authentication: %#v, %v", result, err)
	}
	if requests != 4 {
		t.Fatalf("expected inspection, guarded activation and verification, got %d requests", requests)
	}
}

func TestARRClientIdentifiesManualAuthenticationWithoutMutation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("manual authentication must not be mutated with %s", request.Method)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"id": 1, "authenticationMethod": "forms",
			"authenticationRequired": "enabled", "username": "william",
		})
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	inspection, err := client.InspectAuthentication(
		context.Background(),
		"prowlarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
	)
	if err != nil {
		t.Fatalf("inspect manual authentication: %v", err)
	}
	if inspection.BootstrapRequired || inspection.CorsarrManaged {
		t.Fatalf("unexpected manual authentication classification %#v", inspection)
	}
}

func TestARRClientRejectsAuthenticationConfigWithoutNumericID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{
			"authenticationMethod": "none", "authenticationRequired": "enabled", "username": "",
		})
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	_, err := client.ActivateAuthentication(
		context.Background(),
		"sonarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"corsarr",
		credentials.NewSecret("private-password"),
	)
	if err == nil {
		t.Fatal("expected missing host config ID to be rejected")
	}
}

func TestARRClientDoesNotOverwriteAuthenticationChangedBeforeActivation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("authentication changed by the user must not be mutated with %s", request.Method)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"id": 1, "authenticationMethod": "forms",
			"authenticationRequired": "enabled", "username": "william",
		})
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	result, err := client.ActivateAuthentication(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"corsarr",
		credentials.NewSecret("private-password"),
	)
	if err == nil {
		t.Fatal("expected changed authentication to abort activation")
	}
	if result.CredentialAccepted {
		t.Fatal("credential must not be accepted before the guarded update")
	}
}

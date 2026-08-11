package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestARRClientKeepsExistingRootFolder(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		if request.Header.Get("X-Api-Key") != "0123456789abcdef0123456789abcdef" {
			t.Fatal("expected API key header")
		}
		if request.Method != http.MethodGet || request.URL.Path != "/api/v3/rootfolder" {
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		_ = json.NewEncoder(response).Encode([]map[string]any{{"id": 1, "path": "/data/library/movies/"}})
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	err := client.EnsureRootFolder(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"/data/library/movies",
	)
	if err != nil {
		t.Fatalf("ensure existing root folder: %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected GET only, got %d requests", requests)
	}
}

func TestARRClientCreatesMissingRootFolder(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		switch requests {
		case 1:
			_ = json.NewEncoder(response).Encode([]any{})
		case 2:
			if request.Method != http.MethodPost || request.URL.Path != "/api/v1/rootfolder" {
				t.Fatalf("unexpected create request %s %s", request.Method, request.URL.Path)
			}
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode create request: %v", err)
			}
			if body["path"] != "/data/library/music" {
				t.Fatalf("unexpected root path %q", body["path"])
			}
			response.WriteHeader(http.StatusCreated)
		default:
			t.Fatalf("unexpected request %d", requests)
		}
	}))
	defer server.Close()

	client := NewARRClient(readinessResolver{url: server.URL})
	err := client.EnsureRootFolder(
		context.Background(),
		"lidarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"/data/library/music",
	)
	if err != nil {
		t.Fatalf("create root folder: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected GET and POST, got %d requests", requests)
	}
}

func TestARRClientDoesNotRedirectAPIKey(t *testing.T) {
	targetCalls := 0
	targetReceivedKey := false
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		targetCalls++
		targetReceivedKey = request.Header.Get("X-Api-Key") != ""
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL, http.StatusFound)
	}))
	defer origin.Close()

	client := NewARRClient(readinessResolver{url: origin.URL})
	err := client.EnsureRootFolder(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"/data/library/movies",
	)
	if err == nil {
		t.Fatal("expected redirect response to be rejected")
	}
	if targetCalls != 0 || targetReceivedKey {
		t.Fatalf("expected redirect target not to receive credential, calls=%d key=%v", targetCalls, targetReceivedKey)
	}
}

func TestARRClientRejectsUnapprovedRootFolder(t *testing.T) {
	client := NewARRClient(readinessResolver{url: "http://127.0.0.1:7878"})
	err := client.EnsureRootFolder(
		context.Background(),
		"radarr",
		APIKey{value: "0123456789abcdef0123456789abcdef"},
		"/etc",
	)
	if err == nil {
		t.Fatal("expected arbitrary root folder to be rejected")
	}
}

func TestARRProvisionerUsesOnlyApprovedRootFolder(t *testing.T) {
	reader := &recordingCredentialReader{credential: APIKey{value: "0123456789abcdef0123456789abcdef"}}
	client := &recordingRootFolderClient{}
	provisioner := NewARRProvisioner(reader, client)

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "sonarr", nil); err != nil {
		t.Fatalf("provision Sonarr: %v", err)
	}
	if reader.rootPath != "/host/Corsarr" || reader.applicationID != "sonarr" {
		t.Fatalf("unexpected credential lookup: %#v", reader)
	}
	if client.rootPath != "/data/library/tv" || client.applicationID != "sonarr" {
		t.Fatalf("unexpected root folder call: %#v", client)
	}

	reader.calls = 0
	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "jellyfin", nil); err != nil {
		t.Fatalf("skip unsupported app: %v", err)
	}
	if reader.calls != 0 {
		t.Fatalf("expected unsupported app not to read credentials, got %d calls", reader.calls)
	}
}

type recordingCredentialReader struct {
	credential    APIKey
	rootPath      string
	applicationID string
	calls         int
}

func (r *recordingCredentialReader) Read(rootPath string, applicationID string) (APIKey, error) {
	r.calls++
	r.rootPath = rootPath
	r.applicationID = applicationID
	return r.credential, nil
}

type recordingRootFolderClient struct {
	applicationID string
	rootPath      string
}

func (c *recordingRootFolderClient) EnsureRootFolder(
	_ context.Context,
	applicationID string,
	_ APIKey,
	rootPath string,
) error {
	c.applicationID = applicationID
	c.rootPath = rootPath
	return nil
}

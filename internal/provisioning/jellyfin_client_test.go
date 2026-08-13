package provisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestJellyfinClientCompletesStartupAndCreatesApprovedLibraries(t *testing.T) {
	password := credentials.NewSecret("private-password")
	userConfigured := false
	wizardComplete := false
	setupAPIReady := false
	publicStatusChecks := 0
	setupReadinessChecks := 0
	createdLibraries := make([]string, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case "GET /System/Info/Public":
			publicStatusChecks++
			if publicStatusChecks == 1 {
				response.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"StartupWizardCompleted": wizardComplete})
		case "GET /Startup/User":
			setupReadinessChecks++
			if setupReadinessChecks == 1 {
				response.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			setupAPIReady = true
			_ = json.NewEncoder(response).Encode(map[string]string{"Name": "abc"})
		case "POST /Users/AuthenticateByName":
			if !setupAPIReady {
				response.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			var body struct {
				Username string `json:"Username"`
				Password string `json:"Pw"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body.Username != "corsarr" || body.Password != password.Reveal() || !userConfigured {
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(response).Encode(map[string]string{"AccessToken": "0123456789abcdef0123456789abcdef"})
		case "POST /Startup/User":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["Name"] != "corsarr" || body["Password"] != password.Reveal() {
				t.Errorf("unexpected startup user %#v", body)
			}
			userConfigured = true
			response.WriteHeader(http.StatusNoContent)
		case "GET /Library/VirtualFolders":
			assertJellyfinToken(t, request, "0123456789abcdef0123456789abcdef")
			_, _ = response.Write([]byte("[]"))
		case "POST /Library/VirtualFolders":
			assertJellyfinToken(t, request, "0123456789abcdef0123456789abcdef")
			createdLibraries = append(createdLibraries, request.URL.Query().Get("name")+":"+
				request.URL.Query().Get("collectionType")+":"+request.URL.Query().Get("paths"))
			response.WriteHeader(http.StatusNoContent)
		case "POST /Startup/RemoteAccess":
			assertJellyfinToken(t, request, "0123456789abcdef0123456789abcdef")
			var body map[string]bool
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["EnableRemoteAccess"] {
				t.Error("expected remote access disabled by default")
			}
			response.WriteHeader(http.StatusNoContent)
		case "POST /Startup/Complete":
			assertJellyfinToken(t, request, "0123456789abcdef0123456789abcdef")
			wizardComplete = true
			response.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	result, err := NewJellyfinClient(readinessResolver{url: server.URL}).EnsureSetup(context.Background(), password)
	if err != nil {
		t.Fatalf("configure Jellyfin: %v", err)
	}
	if !result.CredentialAccepted {
		t.Fatal("expected credential to be accepted")
	}
	if setupReadinessChecks != 2 {
		t.Fatalf("expected Jellyfin setup API retry, got %d checks", setupReadinessChecks)
	}
	if publicStatusChecks != 3 {
		t.Fatalf("expected Jellyfin public status retry and completion check, got %d checks", publicStatusChecks)
	}
	sort.Strings(createdLibraries)
	want := []string{
		"Movies (Corsarr):movies:/data/library/movies",
		"Music (Corsarr):music:/data/library/music",
		"TV Shows (Corsarr):tvshows:/data/library/tv",
	}
	if len(createdLibraries) != len(want) {
		t.Fatalf("unexpected libraries %v", createdLibraries)
	}
	for index := range want {
		if createdLibraries[index] != want[index] {
			t.Fatalf("unexpected libraries\nwant: %v\n got: %v", want, createdLibraries)
		}
	}
}

func TestJellyfinClientPreservesExistingLibraries(t *testing.T) {
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case "GET /System/Info/Public":
			_, _ = response.Write([]byte(`{"StartupWizardCompleted":true}`))
		case "POST /Users/AuthenticateByName":
			_, _ = response.Write([]byte(`{"AccessToken":"0123456789abcdef0123456789abcdef"}`))
		case "GET /Library/VirtualFolders":
			_ = json.NewEncoder(response).Encode([]map[string]any{
				{"Name": "Movies (Corsarr)", "CollectionType": "movies", "Locations": []string{"/data/library/movies"}},
				{"Name": "User library", "CollectionType": "mixed", "Locations": []string{"/other"}},
			})
		case "POST /Library/VirtualFolders":
			posts++
			response.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	_, err := NewJellyfinClient(readinessResolver{url: server.URL}).EnsureSetup(
		context.Background(),
		credentials.NewSecret("private-password"),
	)
	if err != nil {
		t.Fatalf("reconcile Jellyfin: %v", err)
	}
	if posts != 2 {
		t.Fatalf("expected only two missing Corsarr libraries, got %d", posts)
	}
}

func TestJellyfinClientDoesNotRedirectAccessToken(t *testing.T) {
	targetReceivedToken := false
	target := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Header.Get("Authorization"), "Token=") {
			targetReceivedToken = true
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case "GET /System/Info/Public":
			_, _ = response.Write([]byte(`{"StartupWizardCompleted":true}`))
		case "POST /Users/AuthenticateByName":
			_, _ = response.Write([]byte(`{"AccessToken":"0123456789abcdef0123456789abcdef"}`))
		case "GET /Library/VirtualFolders":
			http.Redirect(response, request, target.URL, http.StatusTemporaryRedirect)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer origin.Close()

	_, err := NewJellyfinClient(readinessResolver{url: origin.URL}).EnsureSetup(
		context.Background(),
		credentials.NewSecret("private-password"),
	)
	if err == nil {
		t.Fatal("expected redirected library request to fail")
	}
	if targetReceivedToken {
		t.Fatal("expected Jellyfin access token not to be redirected")
	}
}

func assertJellyfinToken(t *testing.T, request *http.Request, token string) {
	t.Helper()
	header := request.Header.Get("Authorization")
	if header == "" || !containsAll(header, `MediaBrowser `, `Token="`+token+`"`) {
		t.Fatalf("expected Jellyfin authorization token, got %q", header)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}

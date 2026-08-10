package provisioning

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPReadinessWaitsUntilLocalApplicationResponds(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	waiter := NewHTTPReadiness(
		readinessResolver{url: server.URL},
		250*time.Millisecond,
		time.Millisecond,
	)
	if err := waiter.Wait(context.Background(), "radarr"); err != nil {
		t.Fatalf("wait for application readiness: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPReadinessDoesNotFollowRedirects(t *testing.T) {
	redirectTargetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectTargetCalls++
	}))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL, http.StatusFound)
	}))
	defer server.Close()

	waiter := NewHTTPReadiness(
		readinessResolver{url: server.URL},
		250*time.Millisecond,
		time.Millisecond,
	)
	if err := waiter.Wait(context.Background(), "jellyfin"); err != nil {
		t.Fatalf("wait for redirecting application: %v", err)
	}
	if redirectTargetCalls != 0 {
		t.Fatalf("expected redirect target not to be contacted, got %d calls", redirectTargetCalls)
	}
}

type readinessResolver struct {
	url string
	err error
}

func (r readinessResolver) ResolveApplicationURL(string) (string, error) {
	return r.url, r.err
}

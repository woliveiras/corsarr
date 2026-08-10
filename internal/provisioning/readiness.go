package provisioning

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type ApplicationEndpointResolver interface {
	ResolveApplicationURL(id string) (string, error)
}

// HTTPReadiness waits for the allowlisted loopback web endpoint to accept HTTP.
// Redirects are deliberately not followed and no credentials are sent.
type HTTPReadiness struct {
	resolver     ApplicationEndpointResolver
	client       *http.Client
	timeout      time.Duration
	pollInterval time.Duration
}

func NewHTTPReadiness(
	resolver ApplicationEndpointResolver,
	timeout time.Duration,
	pollInterval time.Duration,
) *HTTPReadiness {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	return &HTTPReadiness{
		resolver: resolver,
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		timeout:      timeout,
		pollInterval: pollInterval,
	}
}

func (w *HTTPReadiness) Wait(ctx context.Context, applicationID string) error {
	endpoint, err := w.resolver.ResolveApplicationURL(applicationID)
	if err != nil {
		return fmt.Errorf("resolve application readiness endpoint: %w", err)
	}
	waitContext, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	var lastResult string
	for {
		request, err := http.NewRequestWithContext(waitContext, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("create application readiness request: %w", err)
		}
		response, requestErr := w.client.Do(request)
		if requestErr == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4*1024))
			_ = response.Body.Close()
			if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusInternalServerError {
				return nil
			}
			lastResult = response.Status
		} else {
			lastResult = requestErr.Error()
		}

		timer := time.NewTimer(w.pollInterval)
		select {
		case <-waitContext.Done():
			timer.Stop()
			return fmt.Errorf("application did not become ready (%s): %w", lastResult, waitContext.Err())
		case <-timer.C:
		}
	}
}

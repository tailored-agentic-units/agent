package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/tailored-agentic-units/agent/client"
	"github.com/tailored-agentic-units/agent/mock"
	"github.com/tailored-agentic-units/agent/request"
	"github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/protocol/model"
)

// fastRetryConfig returns a ClientConfig with tiny, deterministic backoff so
// retry tests run fast. MaxRetries=2 means the retry loop makes up to 3
// attempts (attempt 0..2).
func fastRetryConfig() *config.ClientConfig {
	return &config.ClientConfig{
		Timeout:            "30s",
		ConnectionTimeout:  "10s",
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries:     2,
			InitialBackoff: "1ms",
			MaxBackoff:     "5ms",
			Jitter:         false,
		},
	}
}

// newChatRequestTo builds a chat request whose provider targets the given URL
// and whose format parses any 200 response into a non-nil sentinel.
func newChatRequestTo(url string) request.Request {
	prov := mock.NewMockProvider(
		mock.WithBaseURL(url),
		mock.WithEndpoint(""),
	)
	fmt := mock.NewMockFormat(
		mock.WithParseResult("ok", nil),
	)
	mdl := &model.Model{Name: "test-model"}
	return request.NewChat(prov, fmt, mdl, nil, nil)
}

// TestExecute_RetriesThenSucceeds verifies that transient 500s are retried and
// that Execute returns the eventual success. The server fails the first N calls
// (N <= MaxRetries) with 500, then returns 200.
func TestExecute_RetriesThenSucceeds(t *testing.T) {
	const failFirst = 2 // == MaxRetries, so the 3rd attempt (index 2) succeeds

	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if int(n) <= failFirst {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := client.New(fastRetryConfig())
	result, err := c.Execute(context.Background(), newChatRequestTo(server.URL))

	if err != nil {
		t.Fatalf("Execute returned error after retries: %v", err)
	}
	if result == nil {
		t.Fatal("Execute returned nil result on success")
	}
	if got := atomic.LoadInt32(&count); int(got) != failFirst+1 {
		t.Errorf("got %d requests, want %d (failFirst+1)", got, failFirst+1)
	}
}

// TestExecute_RetryableCodes verifies each retryable status triggers more than
// one attempt. With MaxRetries=2 an always-failing server is hit 3 times.
func TestExecute_RetryableCodes(t *testing.T) {
	codes := []int{
		http.StatusRequestTimeout,      // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
	}

	for _, code := range codes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			var count int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&count, 1)
				w.WriteHeader(code)
			}))
			defer server.Close()

			c := client.New(fastRetryConfig())
			_, err := c.Execute(context.Background(), newChatRequestTo(server.URL))

			if err == nil {
				t.Fatalf("expected error for status %d, got nil", code)
			}
			if got := atomic.LoadInt32(&count); got <= 1 {
				t.Errorf("status %d: got %d requests, want > 1 (should retry)", code, got)
			}
		})
	}
}

// TestExecute_NonRetryableCodes verifies deterministic 4xx errors are not
// retried (exactly one attempt). This guards the content-filter 400 fast-fail.
func TestExecute_NonRetryableCodes(t *testing.T) {
	codes := []int{
		http.StatusBadRequest,   // 400 (e.g. content-filter rejection)
		http.StatusUnauthorized, // 401
		http.StatusNotFound,     // 404
	}

	for _, code := range codes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			var count int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&count, 1)
				w.WriteHeader(code)
			}))
			defer server.Close()

			c := client.New(fastRetryConfig())
			_, err := c.Execute(context.Background(), newChatRequestTo(server.URL))

			if err == nil {
				t.Fatalf("expected error for status %d, got nil", code)
			}
			if got := atomic.LoadInt32(&count); got != 1 {
				t.Errorf("status %d: got %d requests, want exactly 1 (no retry)", code, got)
			}
		})
	}
}

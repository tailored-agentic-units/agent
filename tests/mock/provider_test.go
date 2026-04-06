package mock_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/tailored-agentic-units/agent/mock"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/provider"
)

func TestNewMockProvider(t *testing.T) {
	p := mock.NewMockProvider()

	if p == nil {
		t.Fatal("NewMockProvider returned nil")
	}
}

func TestMockProvider_Name(t *testing.T) {
	p := mock.NewMockProvider()

	if p.Name() != "mock-provider" {
		t.Errorf("got name %q, want %q", p.Name(), "mock-provider")
	}
}

func TestMockProvider_Endpoint(t *testing.T) {
	customEndpoints := map[protocol.Protocol]string{
		protocol.Chat:   "/chat",
		protocol.Vision: "/vision",
	}

	p := mock.NewMockProvider(
		mock.WithBaseURL("https://custom.api"),
		mock.WithEndpointMapping(customEndpoints),
	)

	endpoint, err := p.Endpoint(protocol.Chat)

	if err != nil {
		t.Fatalf("Endpoint failed: %v", err)
	}

	if endpoint != "https://custom.api/chat" {
		t.Errorf("got endpoint %q, want %q", endpoint, "https://custom.api/chat")
	}
}

func TestMockProvider_PrepareRequest(t *testing.T) {
	expectedRequest := &provider.Request{
		URL:     "https://test.api/chat",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(`{"test":"data"}`),
	}

	p := mock.NewMockProvider(
		mock.WithPrepareResponse(expectedRequest, nil),
	)

	request, err := p.PrepareRequest(context.Background(), protocol.Chat, []byte(`{}`), nil)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request != expectedRequest {
		t.Error("returned different request than configured")
	}
}

func TestMockProvider_BaseURL(t *testing.T) {
	p := mock.NewMockProvider(
		mock.WithBaseURL("https://custom.api"),
	)

	if p.BaseURL() != "https://custom.api" {
		t.Errorf("got baseURL %q, want %q", p.BaseURL(), "https://custom.api")
	}
}

func TestMockProvider_SetHeaders(t *testing.T) {
	headers := map[string]string{
		"X-Custom": "test-value",
	}

	p := mock.NewMockProvider(
		mock.WithProviderHeaders(headers),
	)

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	err := p.SetHeaders(context.Background(), req)

	if err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	if req.Header.Get("X-Custom") != "test-value" {
		t.Errorf("got header %q, want %q", req.Header.Get("X-Custom"), "test-value")
	}
}

func TestMockProvider_SetHeaders_Error(t *testing.T) {
	expectedErr := errors.New("auth failed")

	p := mock.NewMockProvider(
		mock.WithSetHeadersError(expectedErr),
	)

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	err := p.SetHeaders(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from SetHeaders, got nil")
	}

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMockProvider_Stream(t *testing.T) {
	p := mock.NewMockProvider()

	sr := p.Stream()
	if sr != nil {
		t.Error("expected nil StreamReader by default")
	}
}

func TestMockProvider_PrepareStreamRequest(t *testing.T) {
	p := mock.NewMockProvider(
		mock.WithBaseURL("https://test.api"),
		mock.WithEndpoint("/chat"),
	)

	req, err := p.PrepareStreamRequest(context.Background(), protocol.Chat, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	if req.Headers["Accept"] != "text/event-stream" {
		t.Errorf("got Accept header %q, want %q", req.Headers["Accept"], "text/event-stream")
	}

	if req.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", req.Headers["Cache-Control"], "no-cache")
	}
}

package mock

import (
	"context"
	"net/http"

	"github.com/tailored-agentic-units/protocol"
	protostreaming "github.com/tailored-agentic-units/protocol/streaming"
	"github.com/tailored-agentic-units/provider"
)

// MockProvider implements provider.Provider interface for testing.
type MockProvider struct {
	name     string
	baseURL  string
	headers  map[string]string
	endpoint string

	marshalResponse       []byte
	marshalError          error
	prepareResponse       *provider.Request
	prepareError          error
	setHeadersError       error
	endpointError         error
	customEndpointMapping map[protocol.Protocol]string
	stream                protostreaming.StreamReader
}

// NewMockProvider creates a new MockProvider with default configuration.
func NewMockProvider(opts ...MockProviderOption) *MockProvider {
	m := &MockProvider{
		name:                  "mock-provider",
		baseURL:               "http://mock-provider.local",
		headers:               make(map[string]string),
		endpoint:              "/mock/endpoint",
		customEndpointMapping: make(map[protocol.Protocol]string),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockProviderOption configures a MockProvider.
type MockProviderOption func(*MockProvider)

// WithProviderName sets the provider name.
func WithProviderName(name string) MockProviderOption {
	return func(m *MockProvider) {
		m.name = name
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(url string) MockProviderOption {
	return func(m *MockProvider) {
		m.baseURL = url
	}
}

// WithProviderHeaders sets custom headers.
func WithProviderHeaders(headers map[string]string) MockProviderOption {
	return func(m *MockProvider) {
		m.headers = headers
	}
}

// WithEndpoint sets the default endpoint.
func WithEndpoint(endpoint string) MockProviderOption {
	return func(m *MockProvider) {
		m.endpoint = endpoint
	}
}

// WithEndpointMapping sets custom endpoint mapping for protocols.
func WithEndpointMapping(mapping map[protocol.Protocol]string) MockProviderOption {
	return func(m *MockProvider) {
		m.customEndpointMapping = mapping
	}
}

// WithMarshalResponse sets the response for Marshal.
func WithMarshalResponse(body []byte, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.marshalResponse = body
		m.marshalError = err
	}
}

// WithPrepareResponse sets the response for PrepareRequest.
func WithPrepareResponse(response *provider.Request, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.prepareResponse = response
		m.prepareError = err
	}
}

// WithSetHeadersError sets an error for SetHeaders.
func WithSetHeadersError(err error) MockProviderOption {
	return func(m *MockProvider) {
		m.setHeadersError = err
	}
}

// WithEndpointError sets an error for Endpoint.
func WithEndpointError(err error) MockProviderOption {
	return func(m *MockProvider) {
		m.endpointError = err
	}
}

// WithStreamReader sets the stream reader for streaming responses.
func WithStreamReader(sr protostreaming.StreamReader) MockProviderOption {
	return func(m *MockProvider) {
		m.stream = sr
	}
}

// Name returns the provider name.
func (m *MockProvider) Name() string {
	return m.name
}

// BaseURL returns the base URL.
func (m *MockProvider) BaseURL() string {
	return m.baseURL
}

// Endpoint returns the configured endpoint for a protocol.
func (m *MockProvider) Endpoint(proto protocol.Protocol) (string, error) {
	if m.endpointError != nil {
		return "", m.endpointError
	}

	if endpoint, ok := m.customEndpointMapping[proto]; ok {
		return m.baseURL + endpoint, nil
	}

	return m.baseURL + m.endpoint, nil
}

// Stream returns the stream reader for streaming responses.
func (m *MockProvider) Stream() protostreaming.StreamReader {
	return m.stream
}

// SetHeaders sets the configured headers on the request.
func (m *MockProvider) SetHeaders(_ context.Context, req *http.Request) error {
	if m.setHeadersError != nil {
		return m.setHeadersError
	}

	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
	return nil
}

// PrepareRequest returns the predetermined request.
func (m *MockProvider) PrepareRequest(_ context.Context, proto protocol.Protocol, body []byte, _ map[string]string) (*provider.Request, error) {
	if m.prepareError != nil {
		return nil, m.prepareError
	}

	if m.prepareResponse != nil {
		return m.prepareResponse, nil
	}

	endpoint, _ := m.Endpoint(proto)
	return &provider.Request{
		URL:     endpoint,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    body,
	}, nil
}

// PrepareStreamRequest returns a prepared request with streaming headers.
func (m *MockProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*provider.Request, error) {
	req, err := m.PrepareRequest(ctx, proto, body, headers)
	if err != nil {
		return nil, err
	}

	req.Headers["Accept"] = "text/event-stream"
	req.Headers["Cache-Control"] = "no-cache"

	return req, nil
}

// Verify MockProvider implements provider.Provider interface.
var _ provider.Provider = (*MockProvider)(nil)

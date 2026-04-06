package mock

import (
	"context"
	"net/http"
	"time"

	"github.com/tailored-agentic-units/agent/client"
	"github.com/tailored-agentic-units/agent/request"
	"github.com/tailored-agentic-units/protocol/response"
)

// MockClient implements client.Client interface for testing.
type MockClient struct {
	healthy bool

	executeResponse any
	executeError    error
	streamResponses []*response.StreamingResponse
	streamError     error
	httpClient      *http.Client
}

// NewMockClient creates a new MockClient with default configuration.
func NewMockClient(opts ...MockClientOption) *MockClient {
	m := &MockClient{
		healthy: true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockClientOption configures a MockClient.
type MockClientOption func(*MockClient)

// WithExecuteResponse sets the response for Execute.
func WithExecuteResponse(resp any, err error) MockClientOption {
	return func(m *MockClient) {
		m.executeResponse = resp
		m.executeError = err
	}
}

// WithStreamResponse sets the responses for ExecuteStream.
func WithStreamResponse(responses []*response.StreamingResponse, err error) MockClientOption {
	return func(m *MockClient) {
		m.streamResponses = responses
		m.streamError = err
	}
}

// WithHealthy sets the health status.
func WithHealthy(healthy bool) MockClientOption {
	return func(m *MockClient) {
		m.healthy = healthy
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) MockClientOption {
	return func(m *MockClient) {
		m.httpClient = c
	}
}

// HTTPClient returns the configured HTTP client.
func (m *MockClient) HTTPClient() *http.Client {
	return m.httpClient
}

// Execute returns the predetermined response.
func (m *MockClient) Execute(_ context.Context, _ request.Request) (any, error) {
	return m.executeResponse, m.executeError
}

// ExecuteStream returns a channel with predetermined streaming responses.
func (m *MockClient) ExecuteStream(_ context.Context, _ request.Request) (<-chan *response.StreamingResponse, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingResponse, len(m.streamResponses))
	for _, sr := range m.streamResponses {
		ch <- sr
	}
	close(ch)

	return ch, nil
}

// IsHealthy returns the mock health status.
func (m *MockClient) IsHealthy() bool {
	return m.healthy
}

// Verify MockClient implements client.Client interface.
var _ client.Client = (*MockClient)(nil)

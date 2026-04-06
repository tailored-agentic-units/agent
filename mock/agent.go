package mock

import (
	"context"

	"github.com/tailored-agentic-units/agent"
	"github.com/tailored-agentic-units/agent/client"
	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/model"
	"github.com/tailored-agentic-units/protocol/response"
	"github.com/tailored-agentic-units/provider"
)

// MockAgent implements agent.Agent interface for testing.
// All methods return predetermined responses configured during construction.
type MockAgent struct {
	id string

	// Protocol responses
	chatResponse       *response.Response
	chatError          error
	visionResponse     *response.Response
	visionError        error
	toolsResponse      *response.Response
	toolsError         error
	embeddingsResponse *response.EmbeddingsResponse
	embeddingsError    error

	// Streaming responses
	streamResponses []response.StreamingResponse
	streamError     error

	// Dependencies
	mockClient   client.Client
	mockProvider provider.Provider
	mockFormat   format.Format
	mockModel    *model.Model
}

// NewMockAgent creates a new MockAgent with default configuration.
// Use option functions to configure specific behaviors.
func NewMockAgent(opts ...MockAgentOption) *MockAgent {
	m := &MockAgent{
		id:              "mock-agent-id",
		mockClient:      NewMockClient(),
		mockProvider:    NewMockProvider(),
		mockFormat:      NewMockFormat(),
		mockModel:       &model.Model{
			Name:    "mock-model",
			Options: make(map[protocol.Protocol]map[string]any),
		},
		streamResponses: []response.StreamingResponse{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockAgentOption configures a MockAgent.
type MockAgentOption func(*MockAgent)

// WithAgentID sets the agent ID.
func WithAgentID(id string) MockAgentOption {
	return func(m *MockAgent) {
		m.id = id
	}
}

// WithChatResponse sets the chat response and error.
func WithChatResponse(resp *response.Response, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.chatResponse = resp
		m.chatError = err
	}
}

// WithVisionResponse sets the vision response and error.
func WithVisionResponse(resp *response.Response, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.visionResponse = resp
		m.visionError = err
	}
}

// WithToolsResponse sets the tools response and error.
func WithToolsResponse(resp *response.Response, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.toolsResponse = resp
		m.toolsError = err
	}
}

// WithEmbeddingsResponse sets the embeddings response and error.
func WithEmbeddingsResponse(resp *response.EmbeddingsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.embeddingsResponse = resp
		m.embeddingsError = err
	}
}

// WithStreamResponses sets the streaming responses for stream methods.
func WithStreamResponses(responses []response.StreamingResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.streamResponses = responses
		m.streamError = err
	}
}

// WithClient sets a custom client.
func WithClient(c client.Client) MockAgentOption {
	return func(m *MockAgent) {
		m.mockClient = c
	}
}

// WithProvider sets a custom provider.
func WithProvider(p provider.Provider) MockAgentOption {
	return func(m *MockAgent) {
		m.mockProvider = p
	}
}

// WithFormat sets a custom format.
func WithFormat(f format.Format) MockAgentOption {
	return func(m *MockAgent) {
		m.mockFormat = f
	}
}

// WithModel sets a custom model.
func WithModel(mdl *model.Model) MockAgentOption {
	return func(m *MockAgent) {
		m.mockModel = mdl
	}
}

// ID returns the mock agent's unique identifier.
func (m *MockAgent) ID() string {
	return m.id
}

// Client returns the mock client.
func (m *MockAgent) Client() client.Client {
	return m.mockClient
}

// Provider returns the mock provider.
func (m *MockAgent) Provider() provider.Provider {
	return m.mockProvider
}

// Format returns the mock format.
func (m *MockAgent) Format() format.Format {
	return m.mockFormat
}

// Model returns the mock model.
func (m *MockAgent) Model() *model.Model {
	return m.mockModel
}

// Chat returns the predetermined chat response.
func (m *MockAgent) Chat(_ context.Context, _ []protocol.Message, _ ...map[string]any) (*response.Response, error) {
	return m.chatResponse, m.chatError
}

// ChatStream returns a channel with predetermined streaming responses.
func (m *MockAgent) ChatStream(_ context.Context, _ []protocol.Message, _ ...map[string]any) (<-chan *response.StreamingResponse, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingResponse, len(m.streamResponses))
	for i := range m.streamResponses {
		ch <- &m.streamResponses[i]
	}
	close(ch)

	return ch, nil
}

// Vision returns the predetermined vision response.
func (m *MockAgent) Vision(_ context.Context, _ []protocol.Message, _ []format.Image, _ ...map[string]any) (*response.Response, error) {
	return m.visionResponse, m.visionError
}

// VisionStream returns a channel with predetermined streaming responses.
func (m *MockAgent) VisionStream(_ context.Context, _ []protocol.Message, _ []format.Image, _ ...map[string]any) (<-chan *response.StreamingResponse, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingResponse, len(m.streamResponses))
	for i := range m.streamResponses {
		ch <- &m.streamResponses[i]
	}
	close(ch)

	return ch, nil
}

// Tools returns the predetermined tools response.
func (m *MockAgent) Tools(_ context.Context, _ []protocol.Message, _ []format.ToolDefinition, _ ...map[string]any) (*response.Response, error) {
	return m.toolsResponse, m.toolsError
}

// Embed returns the predetermined embeddings response.
func (m *MockAgent) Embed(_ context.Context, _ string, _ ...map[string]any) (*response.EmbeddingsResponse, error) {
	return m.embeddingsResponse, m.embeddingsError
}

// Verify MockAgent implements agent.Agent interface.
var _ agent.Agent = (*MockAgent)(nil)

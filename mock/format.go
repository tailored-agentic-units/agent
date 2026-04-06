package mock

import (
	"encoding/json"

	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/response"
)

// MockFormat implements format.Format interface for testing.
type MockFormat struct {
	name string

	marshalResponse []byte
	marshalError    error
	parseResponse   any
	parseError      error
	chunkResponse   *response.StreamingResponse
	chunkError      error
}

// NewMockFormat creates a new MockFormat with default configuration.
func NewMockFormat(opts ...MockFormatOption) *MockFormat {
	m := &MockFormat{
		name: "mock-format",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockFormatOption configures a MockFormat.
type MockFormatOption func(*MockFormat)

// WithFormatName sets the format name.
func WithFormatName(name string) MockFormatOption {
	return func(m *MockFormat) {
		m.name = name
	}
}

// WithMarshalResult sets the response for Marshal.
func WithMarshalResult(body []byte, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.marshalResponse = body
		m.marshalError = err
	}
}

// WithParseResult sets the response for Parse.
func WithParseResult(resp any, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.parseResponse = resp
		m.parseError = err
	}
}

// WithChunkResult sets the response for ParseStreamChunk.
func WithChunkResult(resp *response.StreamingResponse, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.chunkResponse = resp
		m.chunkError = err
	}
}

// Name returns the format name.
func (m *MockFormat) Name() string {
	return m.name
}

// Marshal returns the predetermined marshaled body.
func (m *MockFormat) Marshal(_ protocol.Protocol, data any) ([]byte, error) {
	if m.marshalError != nil {
		return nil, m.marshalError
	}

	if m.marshalResponse != nil {
		return m.marshalResponse, nil
	}

	return json.Marshal(data)
}

// Parse returns the predetermined response.
func (m *MockFormat) Parse(_ protocol.Protocol, _ []byte) (any, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}

	if m.parseResponse != nil {
		return m.parseResponse, nil
	}

	return &response.Response{}, nil
}

// ParseStreamChunk returns the predetermined chunk response.
func (m *MockFormat) ParseStreamChunk(_ protocol.Protocol, _ []byte) (*response.StreamingResponse, error) {
	if m.chunkError != nil {
		return nil, m.chunkError
	}

	if m.chunkResponse != nil {
		return m.chunkResponse, nil
	}

	return &response.StreamingResponse{}, nil
}

// Verify MockFormat implements format.Format interface.
var _ format.Format = (*MockFormat)(nil)

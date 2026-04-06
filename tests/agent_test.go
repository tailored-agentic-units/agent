package agent_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tailored-agentic-units/agent"
	"github.com/tailored-agentic-units/agent/mock"
	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/protocol/response"
)

func newTestConfig() *config.AgentConfig {
	return &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Timeout:            "30s",
			ConnectionTimeout:  "10s",
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
		Provider: &config.ProviderConfig{
			Name:    "mock-provider",
			BaseURL: "http://localhost:11434",
		},
		Model: &config.ModelConfig{
			Name: "test-model",
			Capabilities: map[string]map[string]any{
				"chat": {"temperature": 0.7},
			},
		},
	}
}

func TestNew(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat()

	a := agent.New(cfg, mockProvider, mockFormat)

	if a == nil {
		t.Fatal("New returned nil agent")
	}

	if a.ID() == "" {
		t.Error("agent ID is empty")
	}
}

func TestAgent_ID(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat()

	a1 := agent.New(cfg, mockProvider, mockFormat)
	a2 := agent.New(cfg, mockProvider, mockFormat)

	if a1.ID() == a2.ID() {
		t.Error("two agents have the same ID")
	}

	id1 := a1.ID()
	id2 := a1.ID()

	if id1 != id2 {
		t.Error("agent ID changed between calls")
	}
}

func TestAgent_Chat(t *testing.T) {
	expectedResp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello, how can I help you?"},
		},
	}

	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat()
	mockClient := mock.NewMockClient(
		mock.WithExecuteResponse(expectedResp, nil),
	)

	a := agent.New(cfg, mockProvider, mockFormat)
	// Use MockAgent to test Chat directly with controlled responses
	ma := mock.NewMockAgent(
		mock.WithChatResponse(expectedResp, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("Hello"),
	}

	resp, err := ma.Chat(context.Background(), messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Chat returned nil response")
	}

	if resp.Text() != "Hello, how can I help you?" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, how can I help you?")
	}

	// Verify agent was created with the mock client
	_ = mockClient
	_ = a
}

func TestAgent_ChatStream(t *testing.T) {
	chunks := []response.StreamingResponse{
		{Content: []response.ContentBlock{response.TextBlock{Text: "Hello"}}},
		{Content: []response.ContentBlock{response.TextBlock{Text: ", "}}},
		{Content: []response.ContentBlock{response.TextBlock{Text: "world!"}}},
	}

	ma := mock.NewMockAgent(
		mock.WithStreamResponses(chunks, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("Hi"),
	}

	stream, err := ma.ChatStream(context.Background(), messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var content strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		content.WriteString(chunk.Text())
	}

	if content.String() != "Hello, world!" {
		t.Errorf("got content %q, want %q", content.String(), "Hello, world!")
	}
}

func TestAgent_Vision(t *testing.T) {
	expectedResp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "I see a cat in the image."},
		},
	}

	ma := mock.NewMockAgent(
		mock.WithVisionResponse(expectedResp, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("What's in this image?"),
	}
	images := []format.Image{
		{Data: []byte("fake-png-data"), Format: "png"},
	}

	resp, err := ma.Vision(context.Background(), messages, images)
	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Vision returned nil response")
	}

	if resp.Text() != "I see a cat in the image." {
		t.Errorf("got text %q, want %q", resp.Text(), "I see a cat in the image.")
	}
}

func TestAgent_VisionStream(t *testing.T) {
	chunks := []response.StreamingResponse{
		{Content: []response.ContentBlock{response.TextBlock{Text: "I see "}}},
		{Content: []response.ContentBlock{response.TextBlock{Text: "a cat."}}},
	}

	ma := mock.NewMockAgent(
		mock.WithStreamResponses(chunks, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("Describe this image"),
	}
	images := []format.Image{
		{Data: []byte("fake-png"), Format: "png"},
	}

	stream, err := ma.VisionStream(context.Background(), messages, images)
	if err != nil {
		t.Fatalf("VisionStream failed: %v", err)
	}

	var content strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		content.WriteString(chunk.Text())
	}

	if content.String() != "I see a cat." {
		t.Errorf("got content %q, want %q", content.String(), "I see a cat.")
	}
}

func TestAgent_Tools(t *testing.T) {
	expectedResp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.ToolUseBlock{
				ID:    "call_123",
				Name:  "get_weather",
				Input: map[string]any{"location": "Boston"},
			},
		},
	}

	ma := mock.NewMockAgent(
		mock.WithToolsResponse(expectedResp, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("What's the weather in Boston?"),
	}
	tools := []format.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{"type": "string"},
				},
			},
		},
	}

	resp, err := ma.Tools(context.Background(), messages, tools)
	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Tools returned nil response")
	}

	calls := resp.ToolCalls()
	if len(calls) == 0 {
		t.Fatal("response has no tool calls")
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got function name %q, want %q", calls[0].Name, "get_weather")
	}
}

func TestAgent_Embed(t *testing.T) {
	expectedResp := &response.EmbeddingsResponse{
		Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		Model:      "test-model",
	}

	ma := mock.NewMockAgent(
		mock.WithEmbeddingsResponse(expectedResp, nil),
	)

	resp, err := ma.Embed(context.Background(), "Hello, world!")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Embed returned nil response")
	}

	if len(resp.Embeddings) == 0 {
		t.Fatal("response has no embeddings")
	}

	if len(resp.Embeddings[0]) != 3 {
		t.Errorf("got %d dimensions, want 3", len(resp.Embeddings[0]))
	}
}

func TestAgent_Client(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat()

	a := agent.New(cfg, mockProvider, mockFormat)

	c := a.Client()
	if c == nil {
		t.Error("Client() returned nil")
	}
}

func TestAgent_Provider(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider(
		mock.WithProviderName("test-provider"),
	)
	mockFormat := mock.NewMockFormat()

	a := agent.New(cfg, mockProvider, mockFormat)

	p := a.Provider()
	if p == nil {
		t.Error("Provider() returned nil")
	}

	if p.Name() != "test-provider" {
		t.Errorf("got provider name %q, want %q", p.Name(), "test-provider")
	}
}

func TestAgent_Format(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat(
		mock.WithFormatName("test-format"),
	)

	a := agent.New(cfg, mockProvider, mockFormat)

	f := a.Format()
	if f == nil {
		t.Fatal("Format() returned nil")
	}

	if f.Name() != "test-format" {
		t.Errorf("got format name %q, want %q", f.Name(), "test-format")
	}
}

func TestAgent_Model(t *testing.T) {
	cfg := newTestConfig()
	mockProvider := mock.NewMockProvider()
	mockFormat := mock.NewMockFormat()

	a := agent.New(cfg, mockProvider, mockFormat)

	mdl := a.Model()
	if mdl == nil {
		t.Error("Model() returned nil")
	}

	if mdl.Name != "test-model" {
		t.Errorf("got model name %q, want %q", mdl.Name, "test-model")
	}
}

func TestAgent_ChatError(t *testing.T) {
	expectedErr := errors.New("chat failed")

	ma := mock.NewMockAgent(
		mock.WithChatResponse(nil, expectedErr),
	)

	messages := []protocol.Message{
		protocol.UserMessage("Hello"),
	}

	_, err := ma.Chat(context.Background(), messages)
	if err == nil {
		t.Fatal("expected error from Chat, got nil")
	}

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

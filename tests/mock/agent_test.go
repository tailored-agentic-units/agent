package mock_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tailored-agentic-units/agent/mock"
	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/response"
)

func TestNewMockAgent(t *testing.T) {
	agent := mock.NewMockAgent(
		mock.WithAgentID("test-id"),
	)

	if agent == nil {
		t.Fatal("NewMockAgent returned nil")
	}

	if agent.ID() != "test-id" {
		t.Errorf("got ID %q, want %q", agent.ID(), "test-id")
	}
}

func TestMockAgent_Chat(t *testing.T) {
	expectedResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello"},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithAgentID("test-id"),
		mock.WithChatResponse(expectedResponse, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("test"),
	}

	resp, err := agent.Chat(context.Background(), messages)

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Vision(t *testing.T) {
	expectedResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "I see an image"},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithAgentID("test-id"),
		mock.WithVisionResponse(expectedResponse, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("test"),
	}

	resp, err := agent.Vision(context.Background(), messages, []format.Image{{Data: []byte("fake"), Format: "png"}})

	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Tools(t *testing.T) {
	expectedResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.ToolUseBlock{
				ID:    "call_123",
				Name:  "test_func",
				Input: map[string]any{},
			},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithAgentID("test-id"),
		mock.WithToolsResponse(expectedResponse, nil),
	)

	messages := []protocol.Message{
		protocol.UserMessage("test"),
	}

	resp, err := agent.Tools(context.Background(), messages, nil)

	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Embed(t *testing.T) {
	expectedResponse := &response.EmbeddingsResponse{
		Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		Model:      "test-model",
	}

	agent := mock.NewMockAgent(
		mock.WithAgentID("test-id"),
		mock.WithEmbeddingsResponse(expectedResponse, nil),
	)

	resp, err := agent.Embed(context.Background(), "test")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestNewSimpleChatAgent(t *testing.T) {
	agent := mock.NewSimpleChatAgent("test-id", "Hello, world!")

	messages := []protocol.Message{
		protocol.UserMessage("test"),
	}

	resp, err := agent.Chat(context.Background(), messages)

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Text() != "Hello, world!" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, world!")
	}
}

func TestNewStreamingChatAgent(t *testing.T) {
	agent := mock.NewStreamingChatAgent("test-id", []string{"Hello", ", ", "world!"})

	messages := []protocol.Message{
		protocol.UserMessage("test"),
	}

	stream, err := agent.ChatStream(context.Background(), messages)

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

func TestMockAgent_Client(t *testing.T) {
	agent := mock.NewMockAgent()

	c := agent.Client()
	if c == nil {
		t.Error("Client() returned nil")
	}
}

func TestMockAgent_Provider(t *testing.T) {
	agent := mock.NewMockAgent()

	p := agent.Provider()
	if p == nil {
		t.Error("Provider() returned nil")
	}
}

func TestMockAgent_Format(t *testing.T) {
	agent := mock.NewMockAgent()

	f := agent.Format()
	if f == nil {
		t.Error("Format() returned nil")
	}
}

func TestMockAgent_Model(t *testing.T) {
	agent := mock.NewMockAgent()

	m := agent.Model()
	if m == nil {
		t.Error("Model() returned nil")
	}

	if m.Name != "mock-model" {
		t.Errorf("got model name %q, want %q", m.Name, "mock-model")
	}
}

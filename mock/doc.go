// Package mock provides mock implementations of agent interfaces for testing.
//
// This package enables testing of code that depends on agent without requiring
// real LLM providers or network connections. Each mock is configurable with
// predetermined responses and behavior.
//
// # Mock Implementations
//
// MockAgent: Implements agent.Agent interface with configurable protocol responses
//
// MockClient: Implements client.Client interface for client layer testing
//
// MockProvider: Implements provider.Provider interface with endpoint mapping
//
// MockFormat: Implements format.Format interface for wire format testing
//
// # Usage Example
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithChatResponse(&response.Response{
//	        Role: "assistant",
//	        Content: []response.ContentBlock{response.TextBlock{Text: "Hello"}},
//	    }, nil),
//	)
//
//	resp, err := mockAgent.Chat(context.Background(), messages)
//
// # Streaming Support
//
// Streaming methods return pre-populated channels:
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithStreamResponses([]response.StreamingResponse{
//	        {Content: []response.ContentBlock{response.TextBlock{Text: "chunk1"}}},
//	    }, nil),
//	)
//
//	chunks, _ := mockAgent.ChatStream(context.Background(), messages)
//	for chunk := range chunks {
//	    // Process test chunks
//	}
package mock

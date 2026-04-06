package client_test

import (
	"testing"
	"time"

	"github.com/tailored-agentic-units/agent/client"
	"github.com/tailored-agentic-units/protocol/config"
)

func TestNew(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            "30s",
		ConnectionTimeout:  "10s",
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if c == nil {
		t.Fatal("New returned nil client")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            "30s",
		ConnectionTimeout:  "10s",
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if !c.IsHealthy() {
		t.Error("expected client to be healthy initially")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            "5s",
		ConnectionTimeout:  "2s",
		ConnectionPoolSize: 20,
	}

	c := client.New(cfg)

	httpClient := c.HTTPClient()

	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 5*time.Second {
		t.Errorf("got timeout %v, want %v", httpClient.Timeout, 5*time.Second)
	}
}

func TestClient_HTTPClient_ZeroTimeout(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            "",
		ConnectionTimeout:  "",
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	httpClient := c.HTTPClient()
	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 0 {
		t.Errorf("got timeout %v, want 0 for empty string duration", httpClient.Timeout)
	}
}

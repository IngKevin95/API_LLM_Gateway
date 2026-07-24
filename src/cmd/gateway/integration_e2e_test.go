package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/api/openai"
)

// mockProcessor for e2e test
type testProcessor struct{}

func (tp *testProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	return &adapter.Response{
		ProviderID: "mock",
		Content:    "Hello from E2E integration test!",
	}, nil
}

func (tp *testProcessor) ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error) {
	return nil, nil // Stub for now
}

func (tp *testProcessor) ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error) {
	return &adapter.Embedding{
		ProviderID: "mock",
		Vectors:    [][]float64{{0.1, 0.2, 0.3}},
	}, nil
}

func TestE2E_HTTP_Wiring(t *testing.T) {
	// Setup the mux and handler
	mux := http.NewServeMux()
	
	processor := &testProcessor{}
	
	openaiHandler := openai.NewHandler(processor)
	mux.HandleFunc("POST /v1/chat/completions", openaiHandler.HandleChatCompletions)
	
	// Create a test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Test /v1/chat/completions
	reqBody := []byte(`{"model": "gpt-4", "messages": [{"role": "user", "content": "hola"}]}`)
	resp, err := http.Post(server.URL+"/v1/chat/completions", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	
	choices, ok := responseBody["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("Invalid or missing choices in response: %v", responseBody)
	}
	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	if message["content"] != "Hello from E2E integration test!" {
		t.Errorf("Unexpected content: %v", message["content"])
	}
}

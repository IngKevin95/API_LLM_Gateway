package request

import (
	"testing"
)

func TestNormalizedRequest_StoresFormat(t *testing.T) {
	nr := &NormalizedRequest{
		Format: "openai",
		Model:  "gpt-4",
	}

	if nr.Format != "openai" {
		t.Errorf("expected Format='openai', got '%s'", nr.Format)
	}
	if nr.Model != "gpt-4" {
		t.Errorf("expected Model='gpt-4', got '%s'", nr.Model)
	}
}

func TestNormalizedRequest_StoresMessages(t *testing.T) {
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "hello",
		},
	}

	nr := &NormalizedRequest{
		Messages: messages,
	}

	if len(nr.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(nr.Messages))
	}
	if nr.Messages[0]["role"] != "user" {
		t.Errorf("expected role='user', got '%v'", nr.Messages[0]["role"])
	}
}

func TestNormalizedRequest_StoresParameters(t *testing.T) {
	params := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
	}

	nr := &NormalizedRequest{
		Parameters: params,
	}

	if nr.Parameters["temperature"] != 0.7 {
		t.Errorf("expected temperature=0.7, got %v", nr.Parameters["temperature"])
	}
}

package cost

import (
	"context"
	"errors"
	"testing"

	"api-llm-gateway/internal/adapter"
)

type mockAdapter struct {
	response adapter.Response
	err      error
}

func (m *mockAdapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	return m.response, m.err
}

func (m *mockAdapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAdapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, errors.New("not implemented")
}

func TestMiddleware_Chat_Success(t *testing.T) {
	fakeReg := &mockModelFinder{
		costPer1M: 10,
		exists:    true,
	}

	var recordedRecord CostRecord
	sink := func(ctx context.Context, r CostRecord) error {
		recordedRecord = r
		return nil
	}
	tracker := NewTracker(fakeReg, sink)

	mockAd := &mockAdapter{
		response: adapter.Response{Content: "Hello world!"}, // 12 chars
	}

	mw := NewMiddleware(tracker, "openai", mockAd)

	req := adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Content: "Hi"}}, // 2 chars -> Prompt=2, Completion=12, Total=14
	}

	ctx := context.WithValue(context.Background(), "agentID", "agent-x")
	resp, err := mw.Chat(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello world!" {
		t.Errorf("expected Hello world!, got %v", resp.Content)
	}

	// Tracker Sink is synchronous in this test
	if recordedRecord.AgentID != "agent-x" {
		t.Errorf("expected agent-x, got %s", recordedRecord.AgentID)
	}
	if recordedRecord.ProviderID != "openai" {
		t.Errorf("expected openai, got %s", recordedRecord.ProviderID)
	}
	// 14 tokens * 10 / 1000000 = 0.00014
	if recordedRecord.Cost != 0.00014 {
		t.Errorf("expected 0.00014 cost, got %f", recordedRecord.Cost)
	}
}

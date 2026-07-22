package semanticcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/semanticcache"
)

type mockVectorSearch struct {
	sim   float64
	resp  string
	delay time.Duration
}

func (m *mockVectorSearch) FindSimilar(ctx context.Context, prompt string) (float64, string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return 0, "", ctx.Err()
		}
	}
	return m.sim, m.resp, nil
}

func TestCache_Hit(t *testing.T) {
	vs := &mockVectorSearch{sim: 0.99, resp: "Cached response"}
	cache := semanticcache.New(vs, 0.98, 10)

	resp, hit := cache.Lookup(context.Background(), "Long prompt for testing semantic hit")
	if !hit {
		t.Errorf("Expected cache hit")
	}
	if resp != "Cached response" {
		t.Errorf("Expected 'Cached response', got %s", resp)
	}
}

func TestCache_Miss(t *testing.T) {
	vs := &mockVectorSearch{sim: 0.80, resp: "Other response"}
	cache := semanticcache.New(vs, 0.98, 10)

	_, hit := cache.Lookup(context.Background(), "Long prompt for testing semantic miss")
	if hit {
		t.Errorf("Expected cache miss")
	}
}

func TestCache_BypassShort(t *testing.T) {
	vs := &mockVectorSearch{sim: 0.99, resp: "Cached"}
	cache := semanticcache.New(vs, 0.98, 10)

	_, hit := cache.Lookup(context.Background(), "Short") // < 10 chars
	if hit {
		t.Errorf("Expected bypass for short prompt")
	}
}

func TestCache_Timeout(t *testing.T) {
	vs := &mockVectorSearch{sim: 0.99, delay: 100 * time.Millisecond}
	cache := semanticcache.New(vs, 0.98, 10)

	// This should timeout internally if we set a timeout limit
	// Wait, the cache might enforce the 50ms limit internally.
	start := time.Now()
	_, hit := cache.Lookup(context.Background(), "Long prompt but slow backend")
	if time.Since(start) > 60*time.Millisecond {
		t.Errorf("Lookup blocked too long!")
	}
	if hit {
		t.Errorf("Expected miss due to timeout")
	}
}

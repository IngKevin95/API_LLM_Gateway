package quota

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryManager_ReserveAndCommit(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("openai", Consumption{Tokens: 1000, Requests: 10})

	// Reserve valid
	ok := m.Reserve("openai", Consumption{Tokens: 100, Requests: 1})
	if !ok {
		t.Errorf("Expected Reserve to succeed")
	}

	// Exceed limit
	ok = m.Reserve("openai", Consumption{Tokens: 1000, Requests: 1})
	if ok {
		t.Errorf("Expected Reserve to fail due to token limit")
	}

	// Commit actual usage (we estimated 100, actual 50)
	err := m.Commit("openai", Consumption{Tokens: 100, Requests: 1}, Consumption{Tokens: 50, Requests: 1})
	if err != nil {
		t.Errorf("Unexpected error in Commit: %v", err)
	}

	// We can now reserve 950 more tokens (1000 - 50)
	ok = m.Reserve("openai", Consumption{Tokens: 950, Requests: 1})
	if !ok {
		t.Errorf("Expected Reserve to succeed after adjusting estimate via Commit")
	}
}

func TestInMemoryManager_WindowReset(t *testing.T) {
	now := time.Date(2026, 7, 22, 23, 59, 59, 0, time.UTC)
	m := NewInMemoryManagerWithClock(func() time.Time { return now })
	m.SetLimit("anthropic", Consumption{Tokens: 100, Requests: 5})

	ok := m.Reserve("anthropic", Consumption{Tokens: 100, Requests: 5})
	if !ok {
		t.Errorf("Expected Reserve to succeed")
	}

	// Advance time to next day
	now = now.Add(2 * time.Second)

	// Limit should be reset, so we can reserve again
	ok = m.Reserve("anthropic", Consumption{Tokens: 100, Requests: 5})
	if !ok {
		t.Errorf("Expected Reserve to succeed after window reset")
	}
}

func TestInMemoryManager_Concurrency(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("local", Consumption{Tokens: 10000, Requests: 1000})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Reserve("local", Consumption{Tokens: 10, Requests: 1})
			m.Commit("local", Consumption{Tokens: 10, Requests: 1}, Consumption{Tokens: 10, Requests: 1})
		}()
	}
	wg.Wait()

	// 1000 tokens and 100 requests consumed.
	ok := m.Reserve("local", Consumption{Tokens: 9000, Requests: 900})
	if !ok {
		t.Errorf("Expected remaining quota to be valid")
	}
	ok = m.Reserve("local", Consumption{Tokens: 1, Requests: 1})
	if ok {
		t.Errorf("Expected limit to be exactly reached")
	}
}

func TestInMemoryManager_Overshoot(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("overshoot", Consumption{Tokens: 100, Requests: 5})

	// Reserve valid estimate (50)
	ok := m.Reserve("overshoot", Consumption{Tokens: 50, Requests: 1})
	if !ok {
		t.Errorf("Expected Reserve to succeed")
	}

	// Commit actual usage (150) -> overshoot by 100
	m.Commit("overshoot", Consumption{Tokens: 50, Requests: 1}, Consumption{Tokens: 150, Requests: 1})

	// Next reservation should fail because we are in deficit (used: 150, limit: 100)
	ok = m.Reserve("overshoot", Consumption{Tokens: 1, Requests: 1})
	if ok {
		t.Errorf("Expected Reserve to fail due to previous overshoot deficit")
	}
}

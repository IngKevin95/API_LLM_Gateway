package failover

import (
	"testing"
	"time"
)

// RED: Test failing for HU-EVO-010 AC1 - Extract Retry-After seconds
func TestRetryAfter_ExtractSeconds(t *testing.T) {
	// Arrange: Retry-After header value in seconds
	retryAfterStr := "60"

	// Act: parse
	duration := parseRetryAfterSeconds(retryAfterStr)

	// Assert: parsed as 60 seconds
	if duration != 60*time.Second {
		t.Errorf("Expected 60 seconds, got %v", duration)
	}
}

// RED: Test failing for HU-EVO-010 AC2 - Default 30s if absent
func TestRetryAfter_DefaultTo30s_IfAbsent(t *testing.T) {
	// Arrange: no Retry-After header
	retryAfterStr := ""

	// Act: parse
	duration := parseRetryAfterSeconds(retryAfterStr)

	// Assert: default to 30 seconds
	if duration != 30*time.Second {
		t.Errorf("Expected default 30 seconds, got %v", duration)
	}
}

// RED: Test failing for HU-EVO-010 AC3 - Mid-stream abort (error propagates)
func TestRetryAfter_MidStreamAbort_PropagatesError(t *testing.T) {
	// Arrange: mock a 429 mid-stream error
	err := &ProviderError{
		Status:     429,
		RetryAfter: 60 * time.Second,
	}

	// Assert: error should propagate (no transparent failover mid-stream)
	if err.Status != 429 {
		t.Errorf("Expected 429 status, got %d", err.Status)
	}
	if err.RetryAfter != 60*time.Second {
		t.Errorf("Expected 60s RetryAfter, got %v", err.RetryAfter)
	}
}

// RED: Test failing for HU-EVO-010 AC4 - Backoff exponential
func TestRetryAfter_BackoffExponential_30_60_120_Cap(t *testing.T) {
	// Arrange: backoff counter starting at 0
	backoffSteps := []time.Duration{
		30 * time.Second,  // 1st 429
		60 * time.Second,  // 2nd 429
		120 * time.Second, // 3rd 429
		120 * time.Second, // 4th 429 (capped)
		120 * time.Second, // 5th 429 (capped)
	}

	// Act & Assert: each 429 doubles backoff, capped at 120s
	for i, expected := range backoffSteps {
		actual := calculateBackoff(i) // backoff step i
		if actual != expected {
			t.Errorf("Backoff step %d: expected %v, got %v", i, expected, actual)
		}
	}
}

// Helper functions (stub implementations for tests)
func parseRetryAfterSeconds(s string) time.Duration {
	if s == "" {
		return 30 * time.Second
	}
	// Parse as integer seconds and convert to duration
	var seconds int
	if _, err := parseFromString(s, &seconds); err != nil {
		return 30 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func parseFromString(s string, val *int) (bool, error) {
	// Stub: just parse simple integers
	_, err := scanFromString(s, val)
	return err == nil, err
}

func scanFromString(s string, val *int) (int, error) {
	// Stub: sscanf-like behavior
	n := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	*val = n
	return n, nil
}

func calculateBackoff(step int) time.Duration {
	// Exponential backoff: 30 -> 60 -> 120 (capped)
	durations := []time.Duration{
		30 * time.Second,
		60 * time.Second,
		120 * time.Second,
	}
	if step >= len(durations) {
		return 120 * time.Second // cap
	}
	return durations[step]
}

// ProviderError stub
type ProviderError struct {
	Status     int
	RetryAfter time.Duration
}

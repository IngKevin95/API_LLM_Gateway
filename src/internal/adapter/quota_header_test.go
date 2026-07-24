package adapter

import (
	"net/http"
	"testing"
	"time"
)

// RED: Test failing for HU-EVO-006 AC1 - OpenAI headers parsing
func TestExtractQuota_OpenAI_ValidHeaders(t *testing.T) {
	// Arrange: mock response headers like OpenAI (standard names)
	headers := http.Header{}
	headers.Set("x-ratelimit-limit-requests", "10000")
	headers.Set("x-ratelimit-remaining-requests", "9950")
	headers.Set("x-ratelimit-reset-requests", "1721760000")

	// Act: extract quota info
	quota := ExtractQuota(headers)

	// Assert: verify normalized QuotaInfo
	if quota.Limit != 10000 {
		t.Errorf("Expected Limit=10000, got %d", quota.Limit)
	}
	if quota.Remaining != 9950 {
		t.Errorf("Expected Remaining=9950, got %d", quota.Remaining)
	}
	if quota.ResetAt == nil {
		t.Errorf("Expected ResetAt to be set, got nil")
	}
	// Unix timestamp 1721760000 should parse correctly
	expected := time.Unix(1721760000, 0)
	if quota.ResetAt != nil && !quota.ResetAt.Equal(expected) {
		t.Errorf("Expected ResetAt=%v, got %v", expected, *quota.ResetAt)
	}
}

// RED: Test failing for HU-EVO-006 AC2 - Anthropic headers normalization
func TestExtractQuota_Anthropic_HeaderNormalization(t *testing.T) {
	headers := http.Header{}
	headers.Set("Anthropic-Ratelimit-Requests-Limit", "60000")
	headers.Set("Anthropic-Ratelimit-Requests-Remaining", "59999")
	headers.Set("Anthropic-Ratelimit-Requests-Reset", "1721760000")

	quota := ExtractQuota(headers)

	if quota.Limit != 60000 {
		t.Errorf("Expected Limit=60000, got %d", quota.Limit)
	}
	if quota.Remaining != 59999 {
		t.Errorf("Expected Remaining=59999, got %d", quota.Remaining)
	}
}

// RED: Test failing for HU-EVO-006 AC3 - Case-insensitive parsing (Groq)
func TestExtractQuota_CaseInsensitive_LowercaseHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-ratelimit-limit-requests", "30")
	headers.Set("x-ratelimit-remaining-requests", "29")
	headers.Set("x-ratelimit-reset-requests", "1721760000")

	quota := ExtractQuota(headers)

	if quota.Limit != 30 {
		t.Errorf("Expected Limit=30, got %d", quota.Limit)
	}
	if quota.Remaining != 29 {
		t.Errorf("Expected Remaining=29, got %d", quota.Remaining)
	}
}

// RED: Test failing for HU-EVO-006 AC4 - Graceful fallback when headers missing
func TestExtractQuota_MissingHeaders_GracefulFallback(t *testing.T) {
	headers := http.Header{} // Empty, no rate-limit headers

	quota := ExtractQuota(headers)

	if quota.Limit != 0 {
		t.Errorf("Expected Limit=0 for missing headers, got %d", quota.Limit)
	}
	if quota.Remaining != 0 {
		t.Errorf("Expected Remaining=0 for missing headers, got %d", quota.Remaining)
	}
	if quota.ResetAt != nil {
		t.Errorf("Expected ResetAt=nil for missing headers, got %v", quota.ResetAt)
	}
}

// RED: Test failing for HU-EVO-006 AC5 - Multiple reset header formats
func TestExtractQuota_RetryAfterSeconds(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-ratelimit-limit-requests", "1000")
	headers.Set("x-ratelimit-remaining-requests", "100")
	headers.Set("Retry-After", "60") // Seconds format

	quota := ExtractQuota(headers)

	if quota.Limit != 1000 {
		t.Errorf("Expected Limit=1000, got %d", quota.Limit)
	}
	// ResetAt should be calculated as now + 60 seconds
	// Allow 2 second tolerance for test execution time
	if quota.ResetAt == nil {
		t.Errorf("Expected ResetAt to be set from Retry-After, got nil")
	}
	if quota.ResetAt != nil {
		delta := time.Until(*quota.ResetAt)
		if delta < 55*time.Second || delta > 65*time.Second {
			t.Errorf("Expected ResetAt ~60 seconds from now, got %v from now", delta)
		}
	}
}

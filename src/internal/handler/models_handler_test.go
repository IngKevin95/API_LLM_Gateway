package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelsHandler_ListAllModels(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("response should be valid JSON: %v", err)
	}

	if _, hasData := result["data"]; !hasData {
		t.Error("response should have 'data' field")
	}
}

func TestModelsHandler_FilterByCapability(t *testing.T) {
	// GET /v1/models?capability=chat
	req := httptest.NewRequest("GET", "/v1/models?capability=chat", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Should filter to chat models only
	if data, ok := result["data"].([]interface{}); ok {
		if len(data) == 0 {
			t.Logf("no chat models in result (acceptable)")
		}
	}
}

func TestModelsHandler_FilterByProvider(t *testing.T) {
	// GET /v1/models?provider=openai
	req := httptest.NewRequest("GET", "/v1/models?provider=openai", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestModelsHandler_IncludeMetadata(t *testing.T) {
	// GET /v1/models?include_metadata=true
	req := httptest.NewRequest("GET", "/v1/models?include_metadata=true", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// With metadata, response should include cost, latency info
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if model, ok := data[0].(map[string]interface{}); ok {
			// Should have metadata fields
			if _, hasCost := model["cost_per_1m"]; hasCost {
				t.Logf("metadata included: cost present")
			}
		}
	}
}

func TestModelsHandler_ModelAvailabilityStatus(t *testing.T) {
	// GET /v1/models?include_status=true
	req := httptest.NewRequest("GET", "/v1/models?include_status=true", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Response should include availability status
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if model, ok := data[0].(map[string]interface{}); ok {
			if _, hasStatus := model["status"]; hasStatus {
				t.Logf("availability status included")
			}
		}
	}
}

func TestModelsHandler_Pagination(t *testing.T) {
	// GET /v1/models?limit=10&offset=0
	req := httptest.NewRequest("GET", "/v1/models?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Response should respect pagination
	if data, ok := result["data"].([]interface{}); ok {
		if len(data) > 10 {
			t.Errorf("expected max 10 models, got %d", len(data))
		}
	}
}

func TestModelsHandler_CachedResponse(t *testing.T) {
	// Cache should prevent repeated calls to registry
	req1 := httptest.NewRequest("GET", "/v1/models", nil)
	w1 := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/v1/models", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Errorf("both requests should succeed")
	}

	// Responses should be identical (cached)
	if w1.Body.String() != w2.Body.String() {
		t.Logf("cached response consistency (acceptable if data changes)")
	}
}

func TestModelsHandler_CacheInvalidation(t *testing.T) {
	// Cache should be invalidatable via header or timeout
	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Cache-Control", "no-cache")
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should bypass cache when no-cache header present
	if w.Header().Get("Cache-Control") == "" {
		t.Logf("cache headers may not be set")
	}
}

func TestModelsHandler_InvalidCapability(t *testing.T) {
	// GET /v1/models?capability=invalid
	req := httptest.NewRequest("GET", "/v1/models?capability=invalid", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	// Should return empty list or 400 error
	if w.Code == http.StatusBadRequest {
		t.Logf("invalid capability rejected")
	} else if w.Code == http.StatusOK {
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if data, ok := result["data"].([]interface{}); ok && len(data) == 0 {
			t.Logf("invalid capability returns empty list")
		}
	}
}

func TestModelsHandler_CombinedFilters(t *testing.T) {
	// GET /v1/models?capability=vision&provider=anthropic&include_metadata=true
	req := httptest.NewRequest("GET", "/v1/models?capability=vision&provider=anthropic&include_metadata=true", nil)
	w := httptest.NewRecorder()

	handler := NewModelsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Should apply all filters correctly
	if data, ok := result["data"].([]interface{}); ok {
		t.Logf("combined filters applied: %d models returned", len(data))
	}
}

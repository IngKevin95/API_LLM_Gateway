package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"api-llm-gateway/internal/registry"
)

// ModelsHandler implements enhanced /v1/models endpoint
type ModelsHandler struct {
	registry *registry.Registry
	cache    *modelsCache
}

// modelsCache caches /v1/models response
type modelsCache struct {
	data      []byte
	timestamp time.Time
	ttl       time.Duration
	mu        sync.RWMutex
}

// NewModelsHandler creates a new /v1/models endpoint handler
// In real usage, registry would be injected from main; for testing use nil
func NewModelsHandler() *ModelsHandler {
	return &ModelsHandler{
		registry: nil, // Will be injected in production
		cache: &modelsCache{
			ttl: 5 * time.Minute,
		},
	}
}

// ServeHTTP implements http.Handler for the /v1/models endpoint
func (h *ModelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	capability := r.URL.Query().Get("capability")
	provider := r.URL.Query().Get("provider")
	includeMetadata := r.URL.Query().Get("include_metadata") == "true"
	includeStatus := r.URL.Query().Get("include_status") == "true"
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse pagination
	limit := 100 // default
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Check cache (unless no-cache header present)
	noCache := r.Header.Get("Cache-Control") == "no-cache"
	if !noCache && h.cache.isValid() {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.WriteHeader(http.StatusOK)
		w.Write(h.cache.data)
		return
	}

	// Get models from registry
	var models []map[string]interface{}

	// If registry is nil (testing), return empty list
	if h.registry == nil {
		models = []map[string]interface{}{}
	} else {
		// Get model names and build responses
		for _, modelName := range h.registry.ModelNames() {
			model, ok := h.registry.FindModel(modelName)
			if !ok {
				continue
			}

			// Filter by capability
			if capability != "" && !hasCapability(model, capability) {
				continue
			}

			// Filter by provider
			if provider != "" && model.ProviderID != provider {
				continue
			}

			// Build model response
			modelResp := map[string]interface{}{
				"id":       model.Name,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": model.ProviderID,
			}

			// Add optional metadata
			if includeMetadata {
				modelResp["cost_per_1m"] = model.CostPer1M
				modelResp["capabilities"] = model.Capabilities
			}

			// Add optional status
			if includeStatus {
				// ponytail: hardcoded "available"; in production would check health system
				modelResp["status"] = "available"
			}

			models = append(models, modelResp)
		}
	}

	// Apply pagination
	totalCount := len(models)
	if offset >= len(models) {
		models = []map[string]interface{}{}
	} else {
		endIdx := offset + limit
		if endIdx > len(models) {
			endIdx = len(models)
		}
		models = models[offset:endIdx]
	}

	// Build response
	response := map[string]interface{}{
		"object": "list",
		"data":   models,
	}

	// Add pagination info if requested
	if limitStr != "" || offsetStr != "" {
		response["pagination"] = map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  totalCount,
		}
	}

	// Encode response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)

	// Cache the response
	respBytes, _ := json.Marshal(response)
	h.cache.set(respBytes)

	w.Write(respBytes)
}

// hasCapability checks if model supports capability
func hasCapability(model registry.Model, capability string) bool {
	for _, cap := range model.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// Cache methods
func (mc *modelsCache) isValid() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.data != nil && time.Since(mc.timestamp) < mc.ttl
}

func (mc *modelsCache) set(data []byte) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.data = data
	mc.timestamp = time.Now()
}

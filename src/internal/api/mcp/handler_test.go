package mcp_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/api/mcp"
)

func TestHandleDiscovery_Success(t *testing.T) {
	handler := mcp.NewHandler("secret-token")

	req, _ := http.NewRequest("GET", "/mcp/discovery", nil)
	req.Header.Set("Authorization", "Bearer secret-token")

	rr := httptest.NewRecorder()
	handler.HandleDiscovery(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	bodyStr := rr.Body.String()
	if !strings.Contains(bodyStr, `"tools"`) {
		t.Errorf("expected tools in response")
	}
}

func TestHandleDiscovery_Unauthorized(t *testing.T) {
	handler := mcp.NewHandler("secret-token")

	req, _ := http.NewRequest("GET", "/mcp/discovery", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")

	rr := httptest.NewRecorder()
	handler.HandleDiscovery(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

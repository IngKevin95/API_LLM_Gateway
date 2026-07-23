package mcp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/api/mcp"
)

// stubModels implementa ModelSource para tests.
type stubModels struct {
	names []string
}

func (s *stubModels) ModelNames() []string      { return s.names }
func (s *stubModels) HasCapability(string) bool { return true }

func postMCP(h http.Handler, body any, token string) *httptest.ResponseRecorder {
	buf, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// HU-033 AC1 — Descubrimiento: tools/list retorna catálogo con modelos y herramientas.
func TestMCP_ToolsList_Discovery(t *testing.T) {
	models := &stubModels{names: []string{"gpt-4o", "claude-3-haiku"}}
	h := mcp.NewHandler("", models)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}, "")

	if rr.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("esperaba result, obtuve: %+v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Errorf("esperaba tools no vacío, obtuve: %+v", result)
	}
}

// HU-033 AC2 — Error permisos: token inválido → 403 Forbidden.
func TestMCP_Unauthorized_403(t *testing.T) {
	h := mcp.NewHandler("secret-token", nil)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}, "wrong-token")

	if rr.Code != http.StatusForbidden {
		t.Errorf("esperaba 403, obtuve %d", rr.Code)
	}
}

// HU-033 AC2 — Token correcto permite acceso.
func TestMCP_AuthorizedAccess_200(t *testing.T) {
	h := mcp.NewHandler("secret-token", nil)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}, "secret-token")

	if rr.Code != http.StatusOK {
		t.Errorf("esperaba 200, obtuve %d", rr.Code)
	}
}

// HU-033 AC3 — Payload malformado → 400 Bad Request con descripción del error.
func TestMCP_MalformedPayload_400(t *testing.T) {
	h := mcp.NewHandler("", nil)

	req, _ := http.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{not valid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuve %d", rr.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] == nil {
		t.Error("esperaba campo 'error' en respuesta JSON-RPC")
	}
}

// HU-033 AC3 — method vacío → 400.
func TestMCP_MissingMethod_400(t *testing.T) {
	h := mcp.NewHandler("", nil)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		// sin "method"
	}, "")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuve %d", rr.Code)
	}
}

// HU-033 AC4 — Versión de protocolo incompatible → 426 Upgrade Required.
func TestMCP_VersionMismatch_426(t *testing.T) {
	h := mcp.NewHandler("", nil)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2023-01-01", // versión antigua
			"clientInfo":      map[string]any{"name": "old-client"},
		},
	}, "")

	if rr.Code != http.StatusUpgradeRequired {
		t.Errorf("esperaba 426, obtuve %d: %s", rr.Code, rr.Body.String())
	}
}

// HU-033 AC4 — Versión compatible → 200 con serverInfo.
func TestMCP_Initialize_CompatibleVersion_200(t *testing.T) {
	h := mcp.NewHandler("", nil)

	rr := postMCP(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo":      map[string]any{"name": "claude-code", "version": "1.0"},
		},
	}, "")

	if rr.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	result := resp["result"].(map[string]any)
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion esperado '2024-11-05', obtuve %v", result["protocolVersion"])
	}
	if result["serverInfo"] == nil {
		t.Error("esperaba serverInfo en respuesta initialize")
	}
}

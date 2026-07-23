// Package mcp implementa un servidor MCP (Model Context Protocol) sobre HTTP.
// El protocolo usa JSON-RPC 2.0 con transport HTTP streamable (POST /mcp).
//
// Métodos soportados:
//   - initialize      : handshake + negociación de versión del protocolo
//   - tools/list      : devuelve el catálogo de herramientas (capacidades del Gateway)
//   - tools/call      : ejecuta una herramienta (stub — wiring real en Fase 4)
//   - notifications/* : ignorados (servidor stateless)
package mcp

import (
	"encoding/json"
	"net/http"
)

const (
	// supportedProtocolVersion es la versión mínima de MCP que acepta el Gateway.
	supportedProtocolVersion = "2024-11-05"
	serverName               = "api-llm-gateway"
	serverVersion            = "1.0.0"
)

// ModelSource provee el catálogo de modelos/capacidades (lo satisface *registry.Registry).
type ModelSource interface {
	ModelNames() []string
	HasCapability(cap string) bool
}

// Handler implementa el endpoint MCP JSON-RPC 2.0.
type Handler struct {
	authToken string
	models    ModelSource
}

// NewHandler crea un Handler MCP. authToken vacío deshabilita la autenticación.
func NewHandler(authToken string, models ModelSource) *Handler {
	return &Handler{authToken: authToken, models: models}
}

// ServeHTTP despacha POST /mcp al router JSON-RPC.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// AC2 — AuthN: verifica Bearer token si está configurado.
	if h.authToken != "" {
		bearer := r.Header.Get("Authorization")
		if bearer != "Bearer "+h.authToken {
			writeError(w, nil, codeUnauthorized, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, nil, codeParseError, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// AC3 — Payload malformado: valida estructura JSON-RPC base.
	var req jsonRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil || req.Method == "" {
		writeError(w, nil, codeInvalidRequest, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case "initialize":
		h.handleInitialize(w, &req)
	case "tools/list":
		h.handleToolsList(w, &req)
	case "tools/call":
		h.handleToolsCall(w, &req)
	default:
		writeError(w, req.ID, codeMethodNotFound, "Method not found: "+req.Method, http.StatusOK)
	}
}

// handleInitialize responde al handshake MCP negociando la versión del protocolo.
// AC4 — versión incompatible → 426 Upgrade Required.
func (h *Handler) handleInitialize(w http.ResponseWriter, req *jsonRPCRequest) {
	var params initializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeError(w, req.ID, codeInvalidParams, "Invalid initialize params", http.StatusBadRequest)
			return
		}
	}

	// Validar versión del protocolo.
	if params.ProtocolVersion != "" && params.ProtocolVersion < supportedProtocolVersion {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUpgradeRequired)
		_ = json.NewEncoder(w).Encode(jsonRPCError{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &rpcError{
				Code:    codeVersionMismatch,
				Message: "Protocol version " + params.ProtocolVersion + " not supported. Minimum: " + supportedProtocolVersion,
			},
		})
		return
	}

	result := initializeResult{
		ProtocolVersion: supportedProtocolVersion,
		Capabilities: serverCapabilities{
			Tools: &toolsCapability{ListChanged: false},
		},
		ServerInfo: serverInfo{
			Name:    serverName,
			Version: serverVersion,
		},
	}
	writeResult(w, req.ID, result)
}

// handleToolsList devuelve el catálogo de herramientas del Gateway.
// AC1 — descubrimiento: modelos, cuotas y herramientas soportadas.
func (h *Handler) handleToolsList(w http.ResponseWriter, req *jsonRPCRequest) {
	tools := builtinTools()

	// Agrega herramientas dinámicas por modelo si el source está disponible.
	if h.models != nil {
		for _, name := range h.models.ModelNames() {
			tools = append(tools, toolDef{
				Name:        "route_" + name,
				Description: "Envía un chat a través del modelo " + name,
				InputSchema: modelToolSchema(name),
			})
		}
	}

	writeResult(w, req.ID, map[string]any{"tools": tools})
}

// handleToolsCall ejecuta una herramienta MCP.
func (h *Handler) handleToolsCall(w http.ResponseWriter, req *jsonRPCRequest) {
	var params toolCallParams
	if req.Params == nil {
		writeError(w, req.ID, codeInvalidParams, "params required for tools/call", http.StatusOK)
		return
	}
	if err := json.Unmarshal(req.Params, &params); err != nil || params.Name == "" {
		writeError(w, req.ID, codeInvalidParams, "Invalid tools/call params", http.StatusOK)
		return
	}

	// Stub: retorna un resultado vacío con indicación de que el wiring real
	// se completa al integrar cmd/gateway en Fase 4.
	writeResult(w, req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": "tool " + params.Name + " acknowledged (wiring pendiente)"},
		},
		"isError": false,
	})
}

// --- JSON-RPC 2.0 wire types ---

type jsonRPCRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result"`
}

type jsonRPCError struct {
	Jsonrpc string     `json:"jsonrpc"`
	ID      any        `json:"id"`
	Error   *rpcError  `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSON-RPC error codes (MCP usa los estándar JSON-RPC + extensiones).
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeUnauthorized   = -32000
	codeVersionMismatch = -32001
)

// --- Tipos MCP ---

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities,omitempty"`
	ClientInfo      map[string]any `json:"clientInfo,omitempty"`
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      serverInfo         `json:"serverInfo"`
}

type serverCapabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// builtinTools devuelve las herramientas estáticas del Gateway.
func builtinTools() []toolDef {
	return []toolDef{
		{
			Name:        "list_capabilities",
			Description: "Lista las capacidades disponibles en el Gateway (chat, coding, vision, embedding, etc.)",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			Name:        "get_quota",
			Description: "Consulta la cuota restante para un proveedor o clave API",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"provider": map[string]any{"type": "string", "description": "ID del proveedor"},
				},
			},
		},
		{
			Name:        "route_chat",
			Description: "Envía un mensaje de chat al mejor modelo disponible según el score del Router",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"capability": map[string]any{"type": "string"},
					"messages":   map[string]any{"type": "array"},
				},
				"required": []string{"capability", "messages"},
			},
		},
	}
}

// modelToolSchema genera un InputSchema mínimo para un tool de modelo específico.
func modelToolSchema(modelName string) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"messages": map[string]any{"type": "array", "description": "Mensajes para " + modelName},
		},
		"required": []string{"messages"},
	}
}

// --- helpers de respuesta ---

func writeResult(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jsonRPCResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeError(w http.ResponseWriter, id any, code int, msg string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(jsonRPCError{
		Jsonrpc: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	})
}

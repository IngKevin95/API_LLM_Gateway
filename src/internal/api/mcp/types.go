package mcp

// DiscoveryResponse is kept for backward compatibility with the stub tests.
// New protocol types live in handler.go (JSON-RPC 2.0).
type DiscoveryResponse struct {
	Version string `json:"version"`
	Tools   []Tool `json:"tools"`
}

// Tool es un tipo legacy del stub; el protocolo real usa toolDef (handler.go).
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

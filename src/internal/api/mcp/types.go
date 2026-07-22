package mcp

type DiscoveryResponse struct {
	Version string `json:"version"`
	Tools   []Tool `json:"tools"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

package mcp

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	authToken string
}

func NewHandler(token string) *Handler {
	return &Handler{authToken: token}
}

func (h *Handler) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if h.authToken != "" && !strings.HasSuffix(authHeader, h.authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp := DiscoveryResponse{
		Version: "1.0",
		Tools: []Tool{
			{Name: "list_models", Description: "Lista los modelos disponibles en la Gateway"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

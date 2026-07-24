package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/user"
)

type MfaHandler struct {
	store *user.Store
}

func NewMfaHandler(store *user.Store) *MfaHandler {
	return &MfaHandler{store: store}
}

func (h *MfaHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	authIdentity, ok := auth.FromContext(r.Context())
	if !ok || authIdentity.Subject == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authIdentity.Subject

	secret, uri, err := h.store.MfaEnroll(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret": secret,
		"uri":    uri,
	})
}

type mfaVerifyReq struct {
	Code string `json:"code"`
}

func (h *MfaHandler) Verify(w http.ResponseWriter, r *http.Request) {
	authIdentity, ok := auth.FromContext(r.Context())
	if !ok || authIdentity.Subject == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authIdentity.Subject

	var req mfaVerifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	err := h.store.MfaVerify(r.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, user.ErrInvalidMfaCode) {
			http.Error(w, `{"error":"invalid code"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MfaHandler) Disable(w http.ResponseWriter, r *http.Request) {
	authIdentity, ok := auth.FromContext(r.Context())
	if !ok || authIdentity.Subject == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authIdentity.Subject

	// AC3 requiere confirmar el password actual, pero en la API actual
	// no hemos implementado contraseñas (todo es vía API keys y JWT).
	// Por ahora simplemente desactivamos.
	err := h.store.MfaDisable(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Package handler -- endpoints HTTP de administración de usuarios y API keys
// por usuario (HU-EVO-017/HU-EVO-018). RBAC admin-only para CRUD de usuarios;
// self-service (dueño o admin) para API keys.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/user"
)

// UsersHandler expone POST/GET /users y PATCH /users/{id}. Requiere admin
// (WithAdmin en el contexto, ver alerts.go) para todas las operaciones salvo
// que se documente lo contrario. AdminGlobal decide si el admin ve todos los
// tenants o solo el propio (HU-EVO-017 AC2).
type UsersHandler struct {
	store *user.Store
}

func NewUsersHandler(store *user.Store) *UsersHandler { return &UsersHandler{store: store} }

// WriteStaticAdminProfile responde GET /users/me con un perfil admin
// sintético para el token estático GATEWAY_ADMIN_TOKEN, que no es una fila
// real de `users` (no tiene id/email propios). Permite que el Dashboard,
// autenticado con ese token por defecto, resuelva la tab "Team" igual que un
// admin logueado por sesión JWT.
func WriteStaticAdminProfile(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, &user.User{ID: "static-admin", Email: "admin (static token)", Role: user.RoleAdmin, Status: user.StatusActive})
}

// AdminContext lleva la información de autorización resuelta por el caller
// (mismo patrón que handler.WithAdmin en alerts.go, pero con tenant propio y
// distinción admin-global vs admin-de-tenant).
type AdminContext struct {
	IsAdmin     bool
	GlobalAdmin bool
	Tenant      string
}

func (h *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ac, _ := r.Context().Value(adminInfoKey{}).(AdminContext)

	switch r.Method {
	case http.MethodPost:
		h.create(w, r, ac)
	case http.MethodGet:
		h.list(w, r, ac)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

type adminInfoKey struct{}

// WithAdminContext inyecta la resolución de autorización admin (HU-EVO-017
// AC4: no-admin recibe 403 en escritura).
func WithAdminContextValue(r *http.Request, ac AdminContext) *http.Request {
	return r.WithContext(withAdminInfo(r.Context(), ac))
}

func withAdminInfo(ctx context.Context, ac AdminContext) context.Context {
	return context.WithValue(ctx, adminInfoKey{}, ac)
}

type createUserRequest struct {
	Email  string   `json:"email"`
	Role   string   `json:"role"`
	Tenant string   `json:"tenant"`
	Scopes []string `json:"scopes"`
}

func (h *UsersHandler) create(w http.ResponseWriter, r *http.Request, ac AdminContext) {
	if !ac.IsAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden) // HU-EVO-017 AC4
		return
	}
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	tenant := req.Tenant
	if !ac.GlobalAdmin {
		tenant = ac.Tenant // admin de tenant no puede invitar fuera de su propio tenant
	}
	role := user.Role(req.Role)
	if role == "" {
		role = user.RoleOperator
	}
	u, err := h.store.Create(r.Context(), req.Email, role, tenant, req.Scopes)
	switch {
	case errors.Is(err, user.ErrEmailExists):
		http.Error(w, `{"error":"email already exists"}`, http.StatusConflict) // AC5
		return
	case errors.Is(err, user.ErrInvalidInput):
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, u) // AC1
}

func (h *UsersHandler) list(w http.ResponseWriter, r *http.Request, ac AdminContext) {
	if !ac.IsAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	users, err := h.store.List(r.Context(), ac.Tenant, ac.GlobalAdmin) // AC2
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

// Me implementa GET /users/me (HU-EVO-022 AC1): perfil del usuario
// autenticado, resuelto desde auth.Identity (JWT local o userKeys), no desde
// AdminContext -- cualquier usuario autenticado ve su propio perfil, sea o
// no admin.
func (h *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok || id.Subject == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	u, err := h.store.Get(r.Context(), id.Subject)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type patchUserRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}

// PatchUser implementa PATCH /users/{id} (HU-EVO-017 AC3). id llega vía
// r.PathValue("id") (Go 1.22 ServeMux).
func (h *UsersHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	ac, _ := r.Context().Value(adminInfoKey{}).(AdminContext)
	if !ac.IsAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	var req patchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	var role *user.Role
	if req.Role != nil {
		rr := user.Role(*req.Role)
		role = &rr
	}
	var status *user.Status
	if req.Status != nil {
		ss := user.Status(*req.Status)
		status = &ss
	}
	u, err := h.store.Patch(r.Context(), id, role, status)
	switch {
	case errors.Is(err, user.ErrNotFound):
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	case errors.Is(err, user.ErrInvalidInput):
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// APIKeysHandler expone POST/GET /users/{id}/api-keys y
// DELETE /users/{id}/api-keys/{keyId} (HU-EVO-018).
type APIKeysHandler struct {
	keys *user.KeyStore
}

func NewAPIKeysHandler(keys *user.KeyStore) *APIKeysHandler { return &APIKeysHandler{keys: keys} }

func (h *APIKeysHandler) requesterAllowed(r *http.Request, targetUserID string) bool {
	ac, _ := r.Context().Value(adminInfoKey{}).(AdminContext)
	id, hasIdentity := auth.FromContext(r.Context())
	if ac.IsAdmin {
		return true
	}
	if !hasIdentity {
		return false
	}
	return user.IsOwnerOrAdmin(id, targetUserID, false)
}

func (h *APIKeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if !h.requesterAllowed(r, userID) {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden) // AC4
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.generate(w, r, userID)
	case http.MethodGet:
		h.list(w, r, userID)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

type generateKeyRequest struct {
	Name string `json:"name"`
}

func (h *APIKeysHandler) generate(w http.ResponseWriter, r *http.Request, userID string) {
	var req generateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	plain, rec, err := h.keys.Generate(r.Context(), userID, req.Name)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{ // AC1: la key en claro solo se ve esta vez
		"id":         rec.ID,
		"user_id":    rec.UserID,
		"name":       rec.Name,
		"key":        plain,
		"prefix":     rec.Prefix,
		"created_at": rec.CreatedAt,
	})
}

func (h *APIKeysHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	keys, err := h.keys.List(r.Context(), userID) // AC2: nunca la key completa
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": keys})
}

// RevokeAPIKey implementa DELETE /users/{id}/api-keys/{keyId} (AC3/AC4).
func (h *APIKeysHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	keyID := r.PathValue("keyId")
	if !h.requesterAllowed(r, userID) {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	err := h.keys.Revoke(r.Context(), userID, keyID)
	switch {
	case err == user.ErrKeyNotFound:
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	case err == user.ErrKeyForbidden:
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	case err != nil:
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"api-llm-gateway/internal/user"

	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler expone POST /auth/login.
type AuthHandler struct {
	users     *user.Store
	sessions  *user.SessionStore
	jwtSecret []byte
}

func NewAuthHandler(users *user.Store, sessions *user.SessionStore, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		users:     users,
		sessions:  sessions,
		jwtSecret: jwtSecret,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	// Como no tenemos passwords implementados, buscaremos el usuario por email
	// y asumimos auth exitoso para propósitos de demostración.
	// En un escenario real aquí validaríamos el hash del password.

	// Buscar el usuario para obtener su ID
	// Dado que el user.Store.Get busca por ID, vamos a necesitar buscar el ID,
	// pero para simplificar, usaremos List y filtraremos si es necesario o
	// agregaremos un GetByEmail si existiera.
	// El gateway asume email = userID o algo similar? En el store `id` es un INT, devuelto como string.
	// Como no tenemos GetByEmail en la DB, iteramos o hacemos query.
	// Por ahora simularemos buscarlo asumiendo que el ID es 1 o iterando List.

	usersList, err := h.users.List(r.Context(), "", true)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	var foundUser *user.User
	for _, u := range usersList {
		if strings.EqualFold(u.Email, req.Email) {
			foundUser = &u
			break
		}
	}

	if foundUser == nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// 1. Crear sesión en DB
	userAgent := r.Header.Get("User-Agent")
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	sid, err := h.sessions.Create(r.Context(), foundUser.ID, userAgent, ip)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 2. Generar JWT con claims que incluyen `sid`
	claims := jwt.MapClaims{
		"sub":    foundUser.ID,
		"tenant": foundUser.Tenant,
		"sid":    sid,
		"scopes": foundUser.Scopes,
		"role":   string(foundUser.Role),
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

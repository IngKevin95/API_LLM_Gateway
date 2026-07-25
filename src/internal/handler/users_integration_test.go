//go:build integration
// +build integration

// Integración HTTP real de /users y /users/{id}/api-keys contra PostgreSQL
// real (mismo patrón docker que internal/user/user_integration_test.go).
// Correr con: go test ./internal/handler/... -tags=integration -run TestUsersHTTP -v

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"

	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/user"
)

// startTestPostgres (contenedor "alerts-handler-it-pg") y testPGContainer
// están definidos en alerts_integration_test.go, mismo paquete; se reutilizan
// aquí para no chocar puertos/contenedores entre corridas.

func adminReq(method, url string, body any) *http.Request {
	var r *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		r = httptest.NewRequest(method, url, bytes.NewReader(b))
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return WithAdminContextValue(r, AdminContext{IsAdmin: true, GlobalAdmin: true})
}

func nonAdminReq(method, url string, body any) *http.Request {
	var r *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		r = httptest.NewRequest(method, url, bytes.NewReader(b))
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return WithAdminContextValue(r, AdminContext{})
}

// HU-EVO-017 AC1: POST /users admin -> 201.
func TestUsersHTTP_Create_Admin_Returns201(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	h := NewUsersHandler(store)

	req := adminReq(http.MethodPost, "/users", createUserRequest{Email: "new@example.com", Role: "operator", Tenant: "t1"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %s", w.Code, w.Body.String())
	}
}

// HU-EVO-017 AC4: POST /users no-admin -> 403, sin crear registro.
func TestUsersHTTP_Create_NonAdmin_Returns403(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	h := NewUsersHandler(store)

	req := nonAdminReq(http.MethodPost, "/users", createUserRequest{Email: "blocked@example.com", Role: "operator"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403, obtuve %d", w.Code)
	}

	list, err := store.List(context.Background(), "", true)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, u := range list {
		if u.Email == "blocked@example.com" {
			t.Fatalf("no debería haberse creado el usuario tras 403")
		}
	}
}

// HU-EVO-017 AC5: email duplicado -> 409.
func TestUsersHTTP_Create_DuplicateEmail_Returns409(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	h := NewUsersHandler(store)

	body := createUserRequest{Email: "dup@example.com", Role: "operator", Tenant: "t1"}
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, adminReq(http.MethodPost, "/users", body))
	if w1.Code != http.StatusCreated {
		t.Fatalf("1er POST esperaba 201, obtuve %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, adminReq(http.MethodPost, "/users", body))
	if w2.Code != http.StatusConflict {
		t.Fatalf("2do POST esperaba 409, obtuve %d", w2.Code)
	}
}

// HU-EVO-017 AC2: GET /users como admin de tenant (no global) solo ve su tenant.
func TestUsersHTTP_List_TenantAdmin_OnlySeesOwnTenant(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	h := NewUsersHandler(store)

	h.ServeHTTP(httptest.NewRecorder(), adminReq(http.MethodPost, "/users", createUserRequest{Email: "a@t1.com", Role: "operator", Tenant: "t1"}))
	h.ServeHTTP(httptest.NewRecorder(), adminReq(http.MethodPost, "/users", createUserRequest{Email: "b@t2.com", Role: "operator", Tenant: "t2"}))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = WithAdminContextValue(req, AdminContext{IsAdmin: true, GlobalAdmin: false, Tenant: "t1"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []user.User `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Tenant != "t1" {
		t.Fatalf("esperaba 1 usuario de t1, obtuve %+v", resp.Data)
	}
}

// HU-EVO-017 AC3: PATCH /users/{id} suspende -> siguiente auth de sus keys falla.
func TestUsersHTTP_Patch_Suspend_BlocksAPIKeyAuth(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	keys, err := user.NewKeyStore(db, store)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	uh := NewUsersHandler(store)
	kh := NewAPIKeysHandler(keys)

	w := httptest.NewRecorder()
	uh.ServeHTTP(w, adminReq(http.MethodPost, "/users", createUserRequest{Email: "f@t1.com", Role: "operator", Tenant: "t1"}))
	var created user.User
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created user: %v", err)
	}

	active := "active"
	patchReq := httptest.NewRequest(http.MethodPatch, "/users/"+created.ID, bytes.NewReader(mustJSON(patchUserRequest{Status: &active})))
	patchReq.SetPathValue("id", created.ID)
	patchReq = WithAdminContextValue(patchReq, AdminContext{IsAdmin: true, GlobalAdmin: true})
	pw := httptest.NewRecorder()
	uh.PatchUser(pw, patchReq)
	if pw.Code != http.StatusOK {
		t.Fatalf("PATCH activar esperaba 200, obtuve %d: %s", pw.Code, pw.Body.String())
	}

	genReq := httptest.NewRequest(http.MethodPost, "/users/"+created.ID+"/api-keys", bytes.NewReader(mustJSON(generateKeyRequest{Name: "k1"})))
	genReq.SetPathValue("id", created.ID)
	genReq = WithAdminContextValue(genReq, AdminContext{IsAdmin: true, GlobalAdmin: true})
	gw := httptest.NewRecorder()
	kh.ServeHTTP(gw, genReq)
	if gw.Code != http.StatusCreated {
		t.Fatalf("generar key esperaba 201, obtuve %d: %s", gw.Code, gw.Body.String())
	}
	var genResp struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(gw.Body.Bytes(), &genResp); err != nil {
		t.Fatalf("unmarshal genResp: %v", err)
	}

	if _, ok := keys.Authenticate(context.Background(), genResp.Key); !ok {
		t.Fatalf("esperaba autenticar mientras el usuario está activo")
	}

	suspended := "suspended"
	patchReq2 := httptest.NewRequest(http.MethodPatch, "/users/"+created.ID, bytes.NewReader(mustJSON(patchUserRequest{Status: &suspended})))
	patchReq2.SetPathValue("id", created.ID)
	patchReq2 = WithAdminContextValue(patchReq2, AdminContext{IsAdmin: true, GlobalAdmin: true})
	pw2 := httptest.NewRecorder()
	uh.PatchUser(pw2, patchReq2)
	if pw2.Code != http.StatusOK {
		t.Fatalf("PATCH suspender esperaba 200, obtuve %d", pw2.Code)
	}

	if _, ok := keys.Authenticate(context.Background(), genResp.Key); ok {
		t.Fatalf("usuario suspendido: la key NO debería seguir autenticando")
	}
}

// HU-EVO-018 AC4: no-admin intentando revocar key de otro usuario -> 403.
func TestAPIKeysHTTP_Revoke_OtherUser_Returns403(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	keys, err := user.NewKeyStore(db, store)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	kh := NewAPIKeysHandler(keys)

	owner, err := store.Create(context.Background(), "owner2@t1.com", user.RoleOperator, "t1", nil)
	if err != nil {
		t.Fatalf("Create owner: %v", err)
	}
	_, rec, err := keys.Generate(context.Background(), owner.ID, "ownerkey")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/users/"+owner.ID+"/api-keys/"+rec.ID, nil)
	delReq.SetPathValue("id", owner.ID)
	delReq.SetPathValue("keyId", rec.ID)
	// No-admin, y sin Identity inyectada (simula otro usuario no autenticado
	// como el dueño): debe rechazar con 403 igual que "no soy el dueño".
	delReq = WithAdminContextValue(delReq, AdminContext{})
	w := httptest.NewRecorder()
	kh.RevokeAPIKey(w, delReq)
	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403, obtuve %d: %s", w.Code, w.Body.String())
	}
}

// HU-EVO-022 AC1 / INT-usersme-identity: GET /users/me resuelve el perfil
// del usuario autenticado a partir de auth.Identity (no de AdminContext) --
// tanto admin como operator ven su propio perfil real desde PostgreSQL.
func TestUsersHTTP_Me_ReturnsOwnProfile(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	u, err := store.Create(context.Background(), "me@t1.com", user.RoleOperator, "t1", []string{"capability:chat"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	h := NewUsersHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(auth.WithIdentity(req.Context(), auth.Identity{Subject: u.ID, Tenant: "t1"}))
	w := httptest.NewRecorder()
	h.Me(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", w.Code, w.Body.String())
	}
	var got user.User
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Email != "me@t1.com" || got.ID != u.ID {
		t.Fatalf("perfil incorrecto: %+v", got)
	}
}

// HU-EVO-022: sin identidad resuelta en contexto, GET /users/me -> 401 (no
// filtra ni asume un usuario por defecto).
func TestUsersHTTP_Me_NoIdentity_Returns401(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()
	store, err := user.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	h := NewUsersHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	h.Me(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d: %s", w.Code, w.Body.String())
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// Package auth define la identidad del consumidor y su transporte por context,
// compartida por los métodos de autenticación (API key, OAuth2/OIDC, mTLS).
package auth

import "context"

// Identity es el consumidor autenticado, resuelto por cualquier Authenticator.
type Identity struct {
	Subject      string   // identificador del consumidor (para auditoría)
	Tenant       string   // tenant al que pertenece
	Scopes       []string // p.ej. "capability:coding", "capability:vision:trusted"
	VettedModels []string // modelos explícitamente vetados para esta credencial
	SessionID    string   // ID de sesión si la identidad proviene de un token de sesión
}

type ctxKey struct{}

// WithIdentity inyecta la identidad en el context.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext recupera la identidad inyectada.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// HasScope indica si la identidad porta el scope dado.
func (id Identity) HasScope(scope string) bool {
	for _, s := range id.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

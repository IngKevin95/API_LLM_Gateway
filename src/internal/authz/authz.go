// Package authz autoriza peticiones ya autenticadas por scope/RBAC y tenant,
// con aislamiento multi-tenant y restricciones de capacidad/modelo.
package authz

import (
	"log"
	"net/http"

	"api-llm-gateway/internal/auth"
)

// Access describe la intención de la petición a autorizar.
type Access struct {
	Capability string
	Model      string
	Tenant     string
}

// Middleware autoriza según la identidad del ctx y la intención extraída.
// `extract` obtiene la Access de la petición (en EP-005 parsea el body real).
func Middleware(extract func(*http.Request) Access) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := auth.FromContext(r.Context())
			if !ok {
				forbidden(w, r, "sin identidad autenticada")
				return
			}
			acc := extract(r)

			// Aislamiento de tenant: la petición no puede apuntar a otro tenant.
			if acc.Tenant != "" && acc.Tenant != id.Tenant {
				forbidden(w, r, "cruce de tenant")
				return
			}
			// Scope de capacidad; vision exige scope explícito de confianza.
			required := "capability:" + acc.Capability
			if acc.Capability == "vision" {
				required = "capability:vision:trusted"
			}
			if !id.HasScope(required) {
				forbidden(w, r, "capacidad fuera de scope")
				return
			}
			// Modelo vetado para esta credencial.
			for _, m := range id.VettedModels {
				if m == acc.Model && acc.Model != "" {
					forbidden(w, r, "modelo vetado")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func forbidden(w http.ResponseWriter, r *http.Request, reason string) {
	// Registra el intento (identidad + motivo), sin enumerar recursos de otros tenants.
	if id, ok := auth.FromContext(r.Context()); ok {
		log.Printf("WARN authz: 403 (%s) subject=%s tenant=%s", reason, id.Subject, id.Tenant)
	}
	http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
}

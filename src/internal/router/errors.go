package router

import (
	"errors"
	"fmt"
	"strings"
)

// Errores del enrutamiento. La capa HTTP (EP-005) los mapea a status codes.
var (
	// ErrUnknownCapability: capacidad no declarada en el routing (→ 400).
	ErrUnknownCapability = errors.New("router: capacidad no soportada")
	// ErrNoModelsAvailable: capacidad conocida pero sin candidatos aptos (→ 503).
	ErrNoModelsAvailable = errors.New("router: ningún modelo disponible para la capacidad")
	// ErrModelUnavailable: modelo explícito no-sano y sin fallback permitido (→ 503).
	ErrModelUnavailable = errors.New("router: modelo solicitado no disponible")
)

// ModelNotFoundError: modelo explícito inexistente en el Registry (→ 404).
type ModelNotFoundError struct {
	Requested string
	Valid     []string
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("router: modelo %q no encontrado; válidos: %s", e.Requested, strings.Join(e.Valid, ", "))
}

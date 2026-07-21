// Package failover recorre la cadena de fallback del router probando cada
// eslabón vía su Adapter, con política retry-vs-failover por código HTTP,
// degradación a local y timeouts dinámicos por capacidad.
package failover

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
)

// ChainResolver entrega la cadena de fallback ordenada (lo satisface *router.Router).
type ChainResolver interface {
	Resolve(capability string, estimatedTokens int) ([]registry.Model, error)
}

// CircuitBreaker protege a cada proveedor (lo satisface *breaker.Breaker).
type CircuitBreaker interface {
	Allow(provider string) bool
	Release(provider string)
	Trip(provider string)
}

// Error es un fallo terminal del failover (pool agotado → 502).
type Error struct {
	Status int
	Err    error
}

func (e *Error) Error() string { return fmt.Sprintf("failover: status %d: %v", e.Status, e.Err) }
func (e *Error) Unwrap() error { return e.Err }

// Engine orquesta la resolución + los adapters con política de failover.
type Engine struct {
	chain    ChainResolver
	adapters map[string]adapter.Adapter // por providerID

	// TTFT es el timeout pre-primer-token por capacidad (estricto chat/código,
	// relajado reasoning). Se aplica por intento. 0 = sin límite.
	TTFT map[string]time.Duration

	// Breaker (opcional) hace fast-fail de proveedores tripped y los penaliza
	// ante fallos, para prevenir Failover Suicide.
	Breaker CircuitBreaker
}

var errBreakerOpen = errors.New("circuit breaker abierto")

func (e *Engine) ttftFor(capability string) time.Duration {
	return e.TTFT[capability]
}

// New construye un Engine con la cadena y los adapters por proveedor.
func New(chain ChainResolver, adapters map[string]adapter.Adapter) *Engine {
	return &Engine{chain: chain, adapters: adapters}
}

// Complete resuelve la capacidad y prueba cada eslabón de la cadena. Ante error
// no-retryable (400) aborta y lo retorna; ante retryable (429/5xx/timeout) pasa
// al siguiente; al agotar el pool retorna *Error 502.
func (e *Engine) Complete(ctx context.Context, capability string, req adapter.Request) (adapter.Response, error) {
	chain, err := e.chain.Resolve(capability, tokensOf(req))
	if err != nil {
		return adapter.Response{}, err // ErrUnknownCapability / ErrNoModelsAvailable del router
	}

	var lastErr error
	for _, m := range chain {
		ad, ok := e.adapters[m.ProviderID]
		if !ok {
			log.Printf("WARN failover: sin adapter para proveedor %q, salteando", m.ProviderID)
			continue
		}
		// Circuit Breaker: fast-fail (0 I/O) si el proveedor está tripped o saturado.
		if e.Breaker != nil {
			if !e.Breaker.Allow(m.ProviderID) {
				lastErr = &adapter.ProviderError{Provider: m.ProviderID, Status: 503, Retryable: true, Err: errBreakerOpen}
				continue
			}
		}
		call := req
		call.Model = m.Name

		// TTFT por intento: cada eslabón recibe su propio presupuesto pre-token.
		attemptCtx := ctx
		var cancel context.CancelFunc
		if d := e.ttftFor(capability); d > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, d)
		}
		resp, cerr := ad.Chat(attemptCtx, call)
		if cancel != nil {
			cancel()
		}
		if e.Breaker != nil {
			e.Breaker.Release(m.ProviderID)
		}
		if cerr == nil {
			return resp, nil
		}
		// 400 (o cualquier no-retryable): abortar sin failover.
		var pe *adapter.ProviderError
		if errors.As(cerr, &pe) && !pe.Retryable {
			return adapter.Response{}, cerr
		}
		// Retryable: penaliza al proveedor (breaker) y prueba el siguiente eslabón.
		if e.Breaker != nil {
			e.Breaker.Trip(m.ProviderID)
		}
		lastErr = cerr
	}

	if lastErr == nil {
		lastErr = errors.New("cadena vacía o sin adapters")
	}
	return adapter.Response{}, &Error{Status: 502, Err: lastErr}
}

func tokensOf(req adapter.Request) int {
	// Estimación simple para la resolución; el tokenizador fino vive en el router.
	n := 0
	for _, m := range req.Messages {
		n += len(m.Content)
	}
	return n
}

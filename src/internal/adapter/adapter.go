// Package adapter define el contrato de la frontera externa: traduce el
// request interno al formato de cada proveedor LLM, envía la llamada y
// normaliza respuesta y errores. Es el único I/O no determinista del Gateway.
package adapter

import (
	"context"
	"fmt"
	"io"
)

// Message es un turno de conversación en formato interno (estilo OpenAI).
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Tool es una definición de herramienta (JSON Schema).
type Tool struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

// ToolCall es una invocación de herramienta devuelta por el modelo.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Request es el formato interno de una petición al proveedor.
type Request struct {
	Model     string
	Messages  []Message
	Tools     []Tool
	MaxTokens int
	Stream    bool
	Input     []string       // para embeddings
	Params    map[string]any // params extra estilo OpenAI (ej. seed); cada adapter decide qué soporta

	// StreamScannerFunc permite inspeccionar de forma asíncrona la respuesta en vuelo (ej. DLP).
	// Si está definido y la petición es streaming, el adaptador envuelve su cuerpo de respuesta
	// y llama a esta función con un reader conectado, pasándole una función cancelFunc
	// que interrumpirá el contexto del request si se detecta una violación.
	StreamScannerFunc func(ctx context.Context, stream io.Reader, cancelFunc context.CancelFunc)
}

// Response es la respuesta normalizada de chat.
type Response struct {
	Content   string
	ToolCalls []ToolCall
}

// Embedding es la respuesta normalizada de embeddings.
type Embedding struct {
	Vectors [][]float64
}

// TokenStream emite tokens de una respuesta en streaming.
type TokenStream interface {
	// Next devuelve el próximo token; ok=false al terminar; error ante fallo.
	Next() (token string, ok bool, err error)
	Close() error
}

// Adapter traduce y ejecuta peticiones contra un proveedor concreto.
type Adapter interface {
	Chat(ctx context.Context, req Request) (Response, error)
	Stream(ctx context.Context, req Request) (TokenStream, error)
	Embed(ctx context.Context, req Request) (Embedding, error)
}

// ProviderError normaliza un fallo del proveedor para que el Failover Engine
// decida (retry vs failover vs abort) sin conocer el formato del proveedor.
type ProviderError struct {
	Provider  string
	Status    int
	Retryable bool // true para 5xx/429/timeout; false para 400 (cliente)
	Err       error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("adapter[%s]: status %d (retryable=%v): %v", e.Provider, e.Status, e.Retryable, e.Err)
}

func (e *ProviderError) Unwrap() error { return e.Err }

// RetryableStatus indica si un código HTTP amerita failover (no retry al mismo).
func RetryableStatus(status int) bool {
	return status == 429 || status >= 500
}

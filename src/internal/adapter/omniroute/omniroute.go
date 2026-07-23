// Package omniroute implementa el Adapter para OmniRoute, un proxy externo
// configurable con contrato OpenAI-compatible.
package omniroute

import (
	"context"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/openai"
)

// Config contiene la configuración del adapter OmniRoute.
type Config struct {
	BaseURL string
	APIKey  string
}

// Adapter implementa adapter.Adapter para OmniRoute.
type Adapter struct {
	inner *openai.Adapter
}

// New crea un adapter OmniRoute. BaseURL y APIKey se obtienen del registry YAML.
func New(cfg Config) *Adapter {
	a := openai.New(cfg.BaseURL, cfg.APIKey)
	a.Name = "omniroute"
	return &Adapter{inner: a}
}

func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	return a.inner.Chat(ctx, req)
}

func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return a.inner.Stream(ctx, req)
}

func (a *Adapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return a.inner.Embed(ctx, req)
}

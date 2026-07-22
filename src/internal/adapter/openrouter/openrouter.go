// Package openrouter implementa el Adapter para OpenRouter.
// OpenRouter es OpenAI-compatible; solo requiere headers adicionales obligatorios.
package openrouter

import (
	"context"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/openai"
)

const defaultBaseURL = "https://openrouter.ai/api"

// Config contiene la configuración del adapter.
type Config struct {
	BaseURL string
	APIKey  string
	// Referer se inyecta como HTTP-Referer (obligatorio por OpenRouter ToS).
	Referer string
	// Title se inyecta como X-Title.
	Title string
}

// Adapter implementa adapter.Adapter para OpenRouter.
type Adapter struct {
	inner   *openai.Adapter
	referer string
	title   string
}

// New crea un adapter OpenRouter con los headers requeridos.
func New(cfg Config) *Adapter {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Referer == "" {
		cfg.Referer = "https://api-llm-gateway"
	}
	if cfg.Title == "" {
		cfg.Title = "API LLM Gateway"
	}
	a := openai.New(cfg.BaseURL, cfg.APIKey)
	a.Name = "openrouter"
	a.ExtraHeaders = map[string]string{
		"HTTP-Referer": cfg.Referer,
		"X-Title":      cfg.Title,
	}
	return &Adapter{inner: a, referer: cfg.Referer, title: cfg.Title}
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

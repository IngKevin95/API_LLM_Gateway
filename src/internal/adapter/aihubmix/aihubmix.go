// Package aihubmix implementa el Adapter para AIHubMix, cuya API es 100%
// compatible con OpenAI. Es un thin wrapper que solo sobreescribe el BaseURL.
package aihubmix

import (
	"context"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/openai"
)

const defaultBaseURL = "https://aihubmix.com/v1"

// Adapter implementa adapter.Adapter para AIHubMix.
type Adapter struct {
	inner *openai.Adapter
}

// New crea un adapter AIHubMix. Si baseURL está vacío, usa el endpoint oficial.
func New(baseURL, apiKey string) *Adapter {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	a := openai.New(baseURL, apiKey)
	a.Name = "aihubmix"
	return &Adapter{inner: a}
}

func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	// Params no soportados se omiten — openai.Adapter solo usa los campos
	// definidos en chatRequest; cualquier key extra en req.Params es ignorado.
	return a.inner.Chat(ctx, req)
}

func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return a.inner.Stream(ctx, req)
}

func (a *Adapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return a.inner.Embed(ctx, req)
}

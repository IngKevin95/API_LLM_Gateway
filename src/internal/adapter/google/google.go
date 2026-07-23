// Package google implementa el Adapter para la API nativa de Google Gemini.
// Traduce el formato interno (estilo OpenAI) al protocolo generateContent de Gemini.
package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"api-llm-gateway/internal/adapter"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com"

// Adapter implementa adapter.Adapter para Google Gemini.
type Adapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New crea un adapter Gemini. Si baseURL está vacío, usa el endpoint oficial.
func New(baseURL, apiKey string) *Adapter {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Adapter{baseURL: baseURL, apiKey: apiKey, client: http.DefaultClient}
}

// --- Tipos de wire Gemini ---

type geminiPart struct {
	Text       string         `json:"text,omitempty"`
	InlineData *geminiInline  `json:"inlineData,omitempty"`
}

type geminiInline struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Chat traduce el request al formato Gemini, lo envía y normaliza la respuesta.
func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	greq := a.buildRequest(req)

	// Use Authorization header instead of query param to avoid exposing API key in URLs
	path := fmt.Sprintf("/v1beta/models/%s:generateContent", req.Model)
	buf, err := json.Marshal(greq)
	if err != nil {
		return adapter.Response{}, a.protocolError(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return adapter.Response{}, &adapter.ProviderError{Provider: "google", Retryable: true, Err: err}
	}
	defer resp.Body.Close()

	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Response{}, perr
	}

	var gr geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return adapter.Response{}, a.protocolError(fmt.Errorf("respuesta sin candidates"))
	}
	return adapter.Response{Content: gr.Candidates[0].Content.Parts[0].Text}, nil
}

// Stream no está soportado en esta versión — devuelve error explícito.
func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, &adapter.ProviderError{Provider: "google", Status: 501, Retryable: false,
		Err: fmt.Errorf("streaming Gemini diferido a Fase 2")}
}

// Embed no está soportado en esta versión.
func (a *Adapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, &adapter.ProviderError{Provider: "google", Status: 501, Retryable: false,
		Err: fmt.Errorf("embeddings Gemini diferido a Fase 2")}
}

// buildRequest traduce el formato interno al de Gemini.
// - messages[role=system] → systemInstruction
// - messages[role=user/assistant] → contents (role=user/model)
// - content con prefijo "data:<mime>;base64," → inlineData
func (a *Adapter) buildRequest(req adapter.Request) geminiRequest {
	var greq geminiRequest

	for _, m := range req.Messages {
		if m.Role == "system" {
			greq.SystemInstruction = &geminiSystemInstruction{
				Parts: []geminiPart{{Text: m.Content}},
			}
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		part := geminiPart{Text: m.Content}
		greq.Contents = append(greq.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{part},
		})
	}
	return greq
}

func (a *Adapter) checkStatus(resp *http.Response) *adapter.ProviderError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	msg, _ := io.ReadAll(resp.Body)
	return &adapter.ProviderError{
		Provider:  "google",
		Status:    resp.StatusCode,
		Retryable: adapter.RetryableStatus(resp.StatusCode),
		Err:       fmt.Errorf("%s", bytes.TrimSpace(msg)),
	}
}

func (a *Adapter) protocolError(err error) *adapter.ProviderError {
	return &adapter.ProviderError{Provider: "google", Status: 502, Retryable: true, Err: err}
}

// Package anthropic implementa el Adapter contra la Messages API de Anthropic,
// traduciendo desde/hacia el formato interno estilo OpenAI.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"api-llm-gateway/internal/adapter"
)

const (
	defaultMaxTokens = 4096
	apiVersion       = "2023-06-01"
)

// unsupportedParams: params estilo OpenAI que Anthropic no soporta; se ignoran con WARN.
var unsupportedParams = []string{"seed", "logit_bias", "frequency_penalty", "presence_penalty"}

// Adapter habla el protocolo Messages de Anthropic.
type Adapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
	name    string

	StreamIdle int64 // ns; reservado para stream.go
}

// New crea un adapter Anthropic.
func New(baseURL, apiKey string) *Adapter {
	return &Adapter{baseURL: baseURL, apiKey: apiKey, client: http.DefaultClient, name: "anthropic"}
}

type wireTool struct {
	Name        string          `json:"name"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type messagesRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system,omitempty"`
	Messages  []map[string]string `json:"messages"`
	Tools     []wireTool          `json:"tools,omitempty"`
	Stream    bool                `json:"stream,omitempty"`
}

type messagesResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
}

// buildRequest traduce el Request interno al formato Messages de Anthropic.
func (a *Adapter) buildRequest(req adapter.Request, stream bool) messagesRequest {
	out := messagesRequest{Model: req.Model, Stream: stream}
	out.MaxTokens = req.MaxTokens
	if out.MaxTokens <= 0 {
		out.MaxTokens = defaultMaxTokens // AC4: max_tokens obligatorio en Anthropic
	}
	for _, m := range req.Messages {
		if m.Role == "system" {
			out.System = m.Content // AC1: extrae el system del array de messages
			continue
		}
		out.Messages = append(out.Messages, map[string]string{"role": m.Role, "content": m.Content})
	}
	for _, t := range req.Tools {
		out.Tools = append(out.Tools, wireTool{Name: t.Name, InputSchema: json.RawMessage(t.Schema)}) // AC2: → tool_use
	}
	// AC3: params no soportados se ignoran con WARN, sin abortar.
	for _, k := range unsupportedParams {
		if _, ok := req.Params[k]; ok {
			log.Printf("WARN anthropic: parámetro %q no soportado, ignorado", k)
		}
	}
	return out
}

// Chat traduce, envía a /v1/messages y normaliza respuesta y errores.
func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	resp, err := a.post(ctx, "/v1/messages", a.buildRequest(req, false))
	if err != nil {
		return adapter.Response{}, err
	}
	defer resp.Body.Close()
	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Response{}, perr
	}

	var mr messagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	var out adapter.Response
	for _, c := range mr.Content {
		switch c.Type {
		case "text":
			out.Content += c.Text
		case "tool_use":
			out.ToolCalls = append(out.ToolCalls, adapter.ToolCall{ID: c.ID, Name: c.Name, Arguments: string(c.Input)})
		}
	}
	return out, nil
}

// Stream y Embed: Stream en stream.go (SS2); Embed no aplica a Anthropic Messages.
func (a *Adapter) Embed(context.Context, adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, &adapter.ProviderError{Provider: a.name, Status: 400, Retryable: false, Err: errors.New("anthropic: embeddings no soportado")}
}

func (a *Adapter) post(ctx context.Context, path string, payload any) (*http.Response, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, a.protocolError(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, a.protocolError(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &adapter.ProviderError{Provider: a.name, Status: 0, Retryable: true, Err: err}
	}
	return resp, nil
}

func (a *Adapter) checkStatus(resp *http.Response) *adapter.ProviderError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	msg, _ := io.ReadAll(resp.Body)
	return &adapter.ProviderError{
		Provider:  a.name,
		Status:    resp.StatusCode,
		Retryable: adapter.RetryableStatus(resp.StatusCode),
		Err:       fmt.Errorf("%s", bytes.TrimSpace(msg)),
	}
}

func (a *Adapter) protocolError(err error) *adapter.ProviderError {
	return &adapter.ProviderError{Provider: a.name, Status: 502, Retryable: true, Err: err}
}

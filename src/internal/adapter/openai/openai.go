// Package openai implementa el Adapter contra la API OpenAI (y compatibles).
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

// Adapter habla el protocolo OpenAI /v1/chat/completions y /v1/embeddings.
type Adapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
	Name    string

	// StreamIdle es el Stream Idle Timeout; 0 usa el default.
	StreamIdle time.Duration
	// MaxBatch es el tope de textos por petición de embeddings; 0 usa el default.
	MaxBatch int
}

const defaultMaxBatch = 2048

// New crea un adapter OpenAI-compatible apuntando a baseURL.
func New(baseURL, apiKey string) *Adapter {
	return &Adapter{baseURL: baseURL, apiKey: apiKey, client: http.DefaultClient, Name: "openai"}
}

// --- formato de cable OpenAI (subconjunto) ---

type wireMessage struct {
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	ToolCalls []wireToolCall     `json:"tool_calls,omitempty"`
}

type wireToolCall struct {
	ID       string `json:"id"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []wireMessage `json:"messages"`
	Tools    []wireTool    `json:"tools,omitempty"`
	Stream   bool          `json:"stream,omitempty"`
}

type wireTool struct {
	Type     string          `json:"type"`
	Function json.RawMessage `json:"function"`
}

type chatResponse struct {
	Choices []struct {
		Message wireMessage `json:"message"`
	} `json:"choices"`
}

// Chat traduce el request al formato OpenAI, lo envía y normaliza la respuesta.
func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	body := chatRequest{Model: req.Model, Stream: false}
	for _, m := range req.Messages {
		body.Messages = append(body.Messages, wireMessage{Role: m.Role, Content: m.Content})
	}
	for _, t := range req.Tools {
		// Preserva el schema del tool intacto en el request.
		fn := fmt.Sprintf(`{"name":%q,"parameters":%s}`, t.Name, t.Schema)
		body.Tools = append(body.Tools, wireTool{Type: "function", Function: json.RawMessage(fn)})
	}

	resp, err := a.post(ctx, "/v1/chat/completions", body)
	if err != nil {
		return adapter.Response{}, err
	}
	defer resp.Body.Close()

	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Response{}, perr
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	if len(cr.Choices) == 0 {
		return adapter.Response{}, a.protocolError(errors.New("respuesta sin choices"))
	}

	msg := cr.Choices[0].Message
	out := adapter.Response{Content: msg.Content}
	for _, tc := range msg.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, adapter.ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments})
	}
	return out, nil
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// Embed redirige a /v1/embeddings, respeta el límite de batch sin truncar en
// silencio, y normaliza la respuesta y los errores.
func (a *Adapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	max := a.MaxBatch
	if max <= 0 {
		max = defaultMaxBatch
	}
	if len(req.Input) > max {
		return adapter.Embedding{}, fmt.Errorf("openai: lote de %d excede el límite de batch %d; particiona la petición", len(req.Input), max)
	}

	resp, err := a.post(ctx, "/v1/embeddings", embedRequest{Model: req.Model, Input: req.Input})
	if err != nil {
		return adapter.Embedding{}, err
	}
	defer resp.Body.Close()
	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Embedding{}, perr
	}

	var er embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return adapter.Embedding{}, a.protocolError(err)
	}
	out := adapter.Embedding{Vectors: make([][]float64, len(er.Data))}
	for i, d := range er.Data {
		out.Vectors[i] = d.Embedding
	}
	return out, nil
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
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		// Fallo de red/timeout: retryable para failover.
		return nil, &adapter.ProviderError{Provider: a.Name, Status: 0, Retryable: true, Err: err}
	}
	return resp, nil
}

// checkStatus mapea un status no-2xx a *ProviderError normalizado.
func (a *Adapter) checkStatus(resp *http.Response) *adapter.ProviderError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	msg, _ := io.ReadAll(resp.Body)
	return &adapter.ProviderError{
		Provider:  a.Name,
		Status:    resp.StatusCode,
		Retryable: adapter.RetryableStatus(resp.StatusCode),
		Err:       fmt.Errorf("%s", bytes.TrimSpace(msg)),
	}
}

// protocolError: respuesta 2xx pero no interpretable → falla de proveedor retryable.
func (a *Adapter) protocolError(err error) *adapter.ProviderError {
	return &adapter.ProviderError{Provider: a.Name, Status: 502, Retryable: true, Err: err}
}

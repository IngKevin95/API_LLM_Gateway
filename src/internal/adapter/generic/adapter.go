// Package generic implementa un Adapter data-driven parametrizado por
// ProviderSpec: permite declarar proveedores OpenAI-compatibles o
// Claude-compatibles (Groq, Cerebras, Mistral, Gemini, Cloudflare AI, etc.)
// sin escribir un paquete Go dedicado por proveedor.
package generic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-llm-gateway/internal/adapter"
)

// FormatOpenAI y FormatClaude son los dos wire formats soportados; cualquier
// otro valor en ProviderSpec.Format es rechazado como spec inválido.
const (
	FormatOpenAI = "openai"
	FormatClaude = "claude"

	defaultClaudeMaxTokens = 4096
	defaultAnthropicVersion = "2023-06-01"
)

// ErrInvalidProviderSpec: el ProviderSpec no cumple los campos obligatorios
// o declara un Format no soportado (AC3 HU-EVO-001).
var ErrInvalidProviderSpec = errors.New("generic: provider spec inválido")

// ProviderSpec declara cómo hablar con un proveedor OpenAI/Claude-compatible.
type ProviderSpec struct {
	BaseURL    string
	AuthHeader string // header de autenticación, ej. "Authorization" o "x-api-key"; default "Authorization"
	Format     string // "openai" | "claude"
	Headers    map[string]string // headers extra, no puede sobrescribir AuthHeader
	TimeoutMs  int               // timeout propio del adapter; 0 usa el default del http.Client
}

// Validate comprueba los campos obligatorios de spec.
func Validate(spec ProviderSpec) error {
	if spec.BaseURL == "" {
		return fmt.Errorf("%w: base_url vacío", ErrInvalidProviderSpec)
	}
	if spec.Format != FormatOpenAI && spec.Format != FormatClaude {
		return fmt.Errorf("%w: format %q no soportado (esperaba %q o %q)", ErrInvalidProviderSpec, spec.Format, FormatOpenAI, FormatClaude)
	}
	return nil
}

// Adapter traduce y ejecuta peticiones contra un proveedor declarado vía spec.
type Adapter struct {
	spec       ProviderSpec
	apiKey     string
	authHeader string
	client     *http.Client
	Name       string
}

// New crea un Adapter genérico. Devuelve ErrInvalidProviderSpec si spec no es válido (AC3).
func New(spec ProviderSpec, apiKey string) (*Adapter, error) {
	if err := Validate(spec); err != nil {
		return nil, err
	}
	authHeader := spec.AuthHeader
	if authHeader == "" {
		authHeader = "Authorization"
	}
	client := http.DefaultClient
	if spec.TimeoutMs > 0 {
		client = &http.Client{Timeout: time.Duration(spec.TimeoutMs) * time.Millisecond} // AC5: timeout propio
	}
	return &Adapter{spec: spec, apiKey: apiKey, authHeader: authHeader, client: client, Name: "generic"}, nil
}

// --- wire OpenAI (subconjunto, espejo de adapter/openai) ---

type owireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ochatRequest struct {
	Model    string         `json:"model"`
	Messages []owireMessage `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
}

type ochatResponse struct {
	Choices []struct {
		Message owireMessage `json:"message"`
	} `json:"choices"`
}

type oembedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type oembedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// --- wire Claude (subconjunto, espejo de adapter/anthropic) ---

type cmessagesRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system,omitempty"`
	Messages  []map[string]string `json:"messages"`
	Stream    bool                `json:"stream,omitempty"`
}

type cmessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Chat traduce req al wire del Format declarado, envía y normaliza.
func (a *Adapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	switch a.spec.Format {
	case FormatOpenAI:
		return a.chatOpenAI(ctx, req)
	case FormatClaude:
		return a.chatClaude(ctx, req)
	default:
		return adapter.Response{}, fmt.Errorf("%w: format %q", ErrInvalidProviderSpec, a.spec.Format)
	}
}

func (a *Adapter) chatOpenAI(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	body := ochatRequest{Model: req.Model, Stream: false}
	for _, m := range req.Messages {
		body.Messages = append(body.Messages, owireMessage{Role: m.Role, Content: m.Content})
	}
	resp, err := a.post(ctx, "/v1/chat/completions", body)
	if err != nil {
		return adapter.Response{}, err
	}
	defer resp.Body.Close()
	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Response{}, perr
	}
	var cr ochatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	if len(cr.Choices) == 0 {
		return adapter.Response{}, a.protocolError(errors.New("respuesta sin choices"))
	}
	return adapter.Response{Content: cr.Choices[0].Message.Content}, nil
}

func (a *Adapter) chatClaude(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	body := a.buildClaudeRequest(req, false)
	resp, err := a.post(ctx, "/v1/messages", body)
	if err != nil {
		return adapter.Response{}, err
	}
	defer resp.Body.Close()
	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Response{}, perr
	}
	var mr cmessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return adapter.Response{}, a.protocolError(err)
	}
	var out adapter.Response
	for _, c := range mr.Content {
		if c.Type == "text" {
			out.Content += c.Text
		}
	}
	return out, nil
}

func (a *Adapter) buildClaudeRequest(req adapter.Request, stream bool) cmessagesRequest {
	out := cmessagesRequest{Model: req.Model, Stream: stream, MaxTokens: req.MaxTokens}
	if out.MaxTokens <= 0 {
		out.MaxTokens = defaultClaudeMaxTokens
	}
	for _, m := range req.Messages {
		if m.Role == "system" {
			out.System = m.Content
			continue
		}
		out.Messages = append(out.Messages, map[string]string{"role": m.Role, "content": m.Content})
	}
	return out
}

// Embed sólo aplica a format openai; claude no soporta embeddings.
func (a *Adapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	if a.spec.Format != FormatOpenAI {
		return adapter.Embedding{}, &adapter.ProviderError{Provider: a.Name, Status: 400, Retryable: false, Err: errors.New("generic: embeddings no soportado para format " + a.spec.Format)}
	}
	resp, err := a.post(ctx, "/v1/embeddings", oembedRequest{Model: req.Model, Input: req.Input})
	if err != nil {
		return adapter.Embedding{}, err
	}
	defer resp.Body.Close()
	if perr := a.checkStatus(resp); perr != nil {
		return adapter.Embedding{}, perr
	}
	var er oembedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return adapter.Embedding{}, a.protocolError(err)
	}
	out := adapter.Embedding{Vectors: make([][]float64, len(er.Data))}
	for i, d := range er.Data {
		out.Vectors[i] = d.Embedding
	}
	return out, nil
}

// post envía payload a a.spec.BaseURL+path inyectando el header de auth y los
// headers extra de spec (sin sobrescribir el de auth, AC4), aplicando el
// timeout propio del adapter (AC5) vía a.client.
func (a *Adapter) post(ctx context.Context, path string, payload any) (*http.Response, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, a.protocolError(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.spec.BaseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, a.protocolError(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	a.setAuth(httpReq)
	if a.spec.Format == FormatClaude {
		httpReq.Header.Set("anthropic-version", defaultAnthropicVersion)
	}
	for k, v := range a.spec.Headers {
		if strings.EqualFold(k, a.authHeader) {
			continue // AC4: headers extra nunca sobrescriben el header de auth
		}
		httpReq.Header.Set(k, v)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &adapter.ProviderError{Provider: a.Name, Status: 0, Retryable: true, Err: err}
	}
	return resp, nil
}

func (a *Adapter) setAuth(httpReq *http.Request) {
	if strings.EqualFold(a.authHeader, "Authorization") {
		httpReq.Header.Set(a.authHeader, "Bearer "+a.apiKey)
		return
	}
	httpReq.Header.Set(a.authHeader, a.apiKey)
}

func (a *Adapter) checkStatus(resp *http.Response) *adapter.ProviderError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	msg, _ := io.ReadAll(resp.Body)
	var retryAfter time.Duration
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
	}
	return &adapter.ProviderError{
		Provider:   a.Name,
		Status:     resp.StatusCode,
		Retryable:  adapter.RetryableStatus(resp.StatusCode),
		Err:        fmt.Errorf("%s", bytes.TrimSpace(msg)),
		RetryAfter: retryAfter,
	}
}

// parseRetryAfter interpreta el header Retry-After como segundos (formato
// delta-seconds, el único usado por los proveedores free-tier curados).
// Devuelve 0 si el header está vacío o no es un entero de segundos válido.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	secs, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func (a *Adapter) protocolError(err error) *adapter.ProviderError {
	return &adapter.ProviderError{Provider: a.Name, Status: 502, Retryable: true, Err: err}
}

package generic

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"api-llm-gateway/internal/adapter"
)

const defaultStreamIdle = 5 * time.Second

type odeltaChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type ceventChunk struct {
	Type  string `json:"type"`
	Delta struct {
		Text string `json:"text"`
	} `json:"delta"`
}

// Stream abre un streaming SSE contra el proveedor, en el Format declarado.
// El fallo pre-primer-token retorna *ProviderError (failover); el corte
// mid-stream se detecta por Stream Idle Timeout (AC4 HU-EVO-004: sin
// failover transparente una vez iniciado el stream).
func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	switch a.spec.Format {
	case FormatOpenAI:
		body := ochatRequest{Model: req.Model, Stream: true}
		for _, m := range req.Messages {
			body.Messages = append(body.Messages, owireMessage{Role: m.Role, Content: m.Content})
		}
		return a.openStream(ctx, "/v1/chat/completions", body, parseOpenAILine)
	case FormatClaude:
		body := a.buildClaudeRequest(req, true)
		return a.openStream(ctx, "/v1/messages", body, parseClaudeLine)
	default:
		return nil, errors.New("generic: format no soportado para stream")
	}
}

// lineParser interpreta una línea SSE y devuelve (token, done, ok-parseado).
type lineParser func(line string) (token string, done bool, matched bool)

func parseOpenAILine(line string) (string, bool, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return "", false, false
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if payload == "[DONE]" {
		return "", true, true
	}
	var ch odeltaChunk
	if err := json.Unmarshal([]byte(payload), &ch); err != nil {
		return "", false, false
	}
	if len(ch.Choices) == 0 || ch.Choices[0].Delta.Content == "" {
		return "", false, false
	}
	return ch.Choices[0].Delta.Content, false, true
}

func parseClaudeLine(line string) (string, bool, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return "", false, false
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	var ch ceventChunk
	if err := json.Unmarshal([]byte(payload), &ch); err != nil {
		return "", false, false
	}
	if ch.Type == "message_stop" {
		return "", true, true
	}
	if ch.Type != "content_block_delta" || ch.Delta.Text == "" {
		return "", false, false
	}
	return ch.Delta.Text, false, true
}

func (a *Adapter) openStream(ctx context.Context, path string, payload any, parse lineParser) (adapter.TokenStream, error) {
	resp, err := a.post(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	if perr := a.checkStatus(resp); perr != nil {
		resp.Body.Close()
		return nil, perr
	}

	idle := time.Duration(a.spec.TimeoutMs) * time.Millisecond
	if idle <= 0 {
		idle = defaultStreamIdle
	}

	s := &sseStream{
		closer:   resp.Body,
		tokens:   make(chan string),
		errc:     make(chan error, 1),
		done:     make(chan struct{}),
		idle:     idle,
		provider: a.Name,
	}
	go s.read(bufio.NewScanner(resp.Body), parse)
	return s, nil
}

type sseStream struct {
	closer   interface{ Close() error }
	tokens   chan string
	errc     chan error
	done     chan struct{}
	idle     time.Duration
	provider string
}

// streamErrorChunk detecta un evento de error SSE mid-stream, formato
// compartido por OpenAI y Claude: `data: {"error": {"type": "...", "message": "..."}}`.
// HU-EVO-0010 AC4: un 429/rate-limit llegado a mitad del stream aborta sin
// intentar failover transparente (a diferencia del pre-stream, que sí lo hace
// vía checkStatus antes de abrir la conexión).
type streamErrorChunk struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func detectStreamError(payload, provider string) *adapter.ProviderError {
	payload = strings.TrimSpace(payload)
	if payload == "" || payload == "[DONE]" {
		return nil
	}
	var ch streamErrorChunk
	if err := json.Unmarshal([]byte(payload), &ch); err != nil {
		return nil
	}
	if ch.Error.Type == "" && ch.Error.Message == "" {
		return nil
	}
	status := 502
	if strings.Contains(strings.ToLower(ch.Error.Type), "rate_limit") {
		status = 429
	}
	msg := ch.Error.Message
	if msg == "" {
		msg = ch.Error.Type
	}
	return &adapter.ProviderError{
		Provider:  provider,
		Status:    status,
		Retryable: false, // AC4: sin failover transparente mid-stream
		Err:       errors.New(msg),
	}
}

func (s *sseStream) read(sc *bufio.Scanner, parse lineParser) {
	defer close(s.tokens)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "data:") {
			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if perr := detectStreamError(payload, s.provider); perr != nil {
				select {
				case s.errc <- perr:
				case <-s.done:
				}
				return
			}
		}
		token, done, matched := parse(line)
		if done {
			return
		}
		if !matched {
			continue
		}
		select {
		case s.tokens <- token:
		case <-s.done:
			return
		}
	}
	if err := sc.Err(); err != nil {
		select {
		case s.errc <- err:
		case <-s.done:
		}
	}
}

func (s *sseStream) Next() (string, bool, error) {
	// errc se chequea primero, sin bloquear: si read() ya cerró tokens *y*
	// dejó un error pendiente (ej. abort mid-stream, HU-EVO-0010 AC4), el
	// error tiene prioridad sobre el cierre silencioso del canal de tokens.
	select {
	case err := <-s.errc:
		return "", false, err
	default:
	}
	select {
	case tok, ok := <-s.tokens:
		if !ok {
			select {
			case err := <-s.errc:
				return "", false, err
			default:
				return "", false, nil
			}
		}
		return tok, true, nil
	case err := <-s.errc:
		return "", false, err
	case <-time.After(s.idle):
		return "", false, &adapter.ProviderError{Provider: s.provider, Status: 0, Retryable: false, Err: errors.New("stream idle timeout")}
	}
}

func (s *sseStream) Close() error {
	select {
	case <-s.done:
	default:
		close(s.done)
	}
	return s.closer.Close()
}

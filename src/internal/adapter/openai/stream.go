package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"api-llm-gateway/internal/adapter"
)

// defaultStreamIdle es el Stream Idle Timeout por defecto (PRD: 5s).
const defaultStreamIdle = 5 * time.Second

// Stream abre un streaming SSE. El fallo pre-primer-token retorna *ProviderError
// (permite failover); el corte mid-stream se detecta por Stream Idle Timeout.
func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	body := chatRequest{Model: req.Model, Stream: true}
	for _, m := range req.Messages {
		body.Messages = append(body.Messages, wireMessage{Role: m.Role, Content: m.Content})
	}

	resp, err := a.post(ctx, "/v1/chat/completions", body)
	if err != nil {
		return nil, err // ProviderError de red, pre-primer-token
	}
	if perr := a.checkStatus(resp); perr != nil {
		resp.Body.Close()
		return nil, perr // fallo pre-primer-token, nada emitido → failover
	}

	idle := a.StreamIdle
	if idle <= 0 {
		idle = defaultStreamIdle
	}

	bodyReader := resp.Body
	if req.StreamScannerFunc != nil {
		pr, pw := io.Pipe()
		// TeeReader duplica el cuerpo que lee el scanner SSE hacia el PipeWriter.
		// NOTA: Es crucial que req.StreamScannerFunc no bloquee la lectura si el PipeWriter
		// se llena, lo cual aseguramos haciéndolo asíncrono y descartando si se cancela.
		bodyReader = struct {
			io.Reader
			io.Closer
		}{
			Reader: io.TeeReader(resp.Body, pw),
			Closer: resp.Body,
		}

		// Ejecutar la inspección asíncrona pasándole el contexto y el PipeReader.
		// Pasamos un cancelFunc wrapper que internamente cierra el cuerpo de respuesta
		// HTTP y notifica al cliente, cortando la conexión TCP abruptamente.
		ctx, cancel := context.WithCancel(ctx)

		killSwitch := func() {
			resp.Body.Close() // Corta la lectura del SSE.
			pw.Close()
			cancel() // Notifica cancelación del contexto
		}

		go req.StreamScannerFunc(ctx, pr, killSwitch)

		// Un goroutine de limpieza en caso de fin normal
		go func() {
			<-ctx.Done()
			pw.Close()
		}()
	}

	s := &sseStream{
		closer:   bodyReader,
		tokens:   make(chan string),
		errc:     make(chan error, 1),
		done:     make(chan struct{}),
		idle:     idle,
		provider: a.Name,
	}
	go s.read(bufio.NewScanner(bodyReader))
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

type deltaChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (s *sseStream) read(sc *bufio.Scanner) {
	defer close(s.tokens)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			return
		}
		var ch deltaChunk
		if err := json.Unmarshal([]byte(payload), &ch); err != nil {
			continue // ignora chunks no interpretables
		}
		if len(ch.Choices) == 0 || ch.Choices[0].Delta.Content == "" {
			continue
		}
		select {
		case s.tokens <- ch.Choices[0].Delta.Content:
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

// Next devuelve el próximo token, o error si se excede el Stream Idle Timeout.
func (s *sseStream) Next() (string, bool, error) {
	select {
	case tok, ok := <-s.tokens:
		if !ok {
			return "", false, nil
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

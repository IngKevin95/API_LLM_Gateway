package anthropic

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

const defaultStreamIdle = 5 * time.Second

// Stream abre un streaming SSE traduciendo los eventos nativos de Anthropic
// (content_block_delta) a tokens. Fallo pre-primer-token → *ProviderError;
// corte mid-stream → Stream Idle Timeout.
func (a *Adapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	resp, err := a.post(ctx, "/v1/messages", a.buildRequest(req, true))
	if err != nil {
		return nil, err
	}
	if perr := a.checkStatus(resp); perr != nil {
		resp.Body.Close()
		return nil, perr
	}

	idle := time.Duration(a.StreamIdle)
	if idle <= 0 {
		idle = defaultStreamIdle
	}

	bodyReader := resp.Body
	if req.StreamScannerFunc != nil {
		pr, pw := io.Pipe()
		bodyReader = struct {
			io.Reader
			io.Closer
		}{
			Reader: io.TeeReader(resp.Body, pw),
			Closer: resp.Body,
		}

		ctx, cancel := context.WithCancel(ctx)

		killSwitch := func() {
			resp.Body.Close()
			pw.Close()
			cancel()
		}

		go req.StreamScannerFunc(ctx, pr, killSwitch)

		go func() {
			<-ctx.Done()
			pw.Close()
		}()
	}

	s := &sseStream{closer: bodyReader, tokens: make(chan string), errc: make(chan error, 1), done: make(chan struct{}), idle: idle}
	go s.read(bufio.NewScanner(bodyReader))
	return s, nil
}

type sseStream struct {
	closer interface{ Close() error }
	tokens chan string
	errc   chan error
	done   chan struct{}
	idle   time.Duration
}

type deltaEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Text string `json:"text"`
	} `json:"delta"`
}

func (s *sseStream) read(sc *bufio.Scanner) {
	defer close(s.tokens)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "data:") {
			continue // ignora líneas "event:" y en blanco
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		var ev deltaEvent
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		if ev.Type != "content_block_delta" || ev.Delta.Text == "" {
			continue
		}
		select {
		case s.tokens <- ev.Delta.Text:
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
	select {
	case tok, ok := <-s.tokens:
		if !ok {
			return "", false, nil
		}
		return tok, true, nil
	case err := <-s.errc:
		return "", false, err
	case <-time.After(s.idle):
		return "", false, &adapter.ProviderError{Provider: "anthropic", Status: 0, Retryable: false, Err: errors.New("stream idle timeout")}
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

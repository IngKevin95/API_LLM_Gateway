// Command gateway is the entrypoint del API LLM Gateway.
// Scaffold mínimo runnable: levanta un servidor HTTP con /health y /metrics.
// Registry, Router, Adapters y demás componentes se cablean por slice (EP-XXX).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-llm-gateway/internal/adapter"
	adapteraihubmix "api-llm-gateway/internal/adapter/aihubmix"
	adapteranthropc "api-llm-gateway/internal/adapter/anthropic"
	adaptergoogle "api-llm-gateway/internal/adapter/google"
	adapterlocal "api-llm-gateway/internal/adapter/local"
	adapteromniroute "api-llm-gateway/internal/adapter/omniroute"
	adapteropenai "api-llm-gateway/internal/adapter/openai"
	apianthropic "api-llm-gateway/internal/api/anthropic"
	apiopenai "api-llm-gateway/internal/api/openai"
	"api-llm-gateway/internal/failover"
	"api-llm-gateway/internal/registry"
	"api-llm-gateway/internal/router"
	"api-llm-gateway/internal/tokenizer"
)

func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	// Carga del Registry en boot. Fail-fast si el config existe pero es inválido;
	// si no hay config declarado, arranca en modo scaffold (solo /health).
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		if _, err := os.Stat("config.yaml"); err == nil {
			cfgPath = "config.yaml"
		}
	}

	var processor *GatewayProcessor
	if cfgPath != "" {
		var err error
		reg, err := registry.Load(cfgPath, nil)
		if err != nil {
			log.Fatalf("registry: %v", err) // fail-fast, no arranca en estado parcial
		}

		// Build Router (EP-001)
		rt := router.New(reg, router.StaticHealth{}, router.StaticQuota{}, tokenizer.NewHeuristic())

		// Build Adapters (EP-002, EP-008)
		adapters := buildAdapters(reg)

		// Build Failover Engine (EP-002)
		fe := failover.New(rt, adapters)

		// Create Processor that uses Failover
		processor = NewGatewayProcessor(fe)
	} else {
		log.Printf("WARN gateway: sin config.yaml, arrancando en modo scaffold (solo /health)")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	// HU-060: /metrics endpoint con datos operacionales (MVP: datos duros)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		metrics := map[string]interface{}{
			"uptime_seconds": 3600,
			"requests": map[string]interface{}{
				"total":      42,
				"by_handler": map[string]int{"openai": 30, "anthropic": 12},
				"errors":     3,
			},
			"providers": map[string]interface{}{
				"openai":     map[string]interface{}{"available": true, "last_success": "2026-07-23T13:52:08Z"},
				"anthropic":  map[string]interface{}{"available": true, "last_error": nil},
				"omniroute":  map[string]interface{}{"available": true, "circuit_breaker_open": false},
			},
			"latency": map[string]int{"p50_ms": 450, "p95_ms": 1200, "p99_ms": 2100},
		}
		json.NewEncoder(w).Encode(metrics)
	})

	// Register OpenAI-compatible endpoints (HU-012a, HU-012b, HU-012c)
	if processor != nil {
		openaiHandler := apiopenai.NewHandler(processor)
		mux.HandleFunc("POST /v1/chat/completions", openaiHandler.HandleChatCompletions)
		mux.HandleFunc("POST /v1/embeddings", openaiHandler.HandleEmbeddings)

		// Register Anthropic-compatible endpoints (HU-013, HU-016)
		anthropicHandler := apianthropic.NewHandler(processor)
		mux.HandleFunc("POST /v1/messages", anthropicHandler.HandleMessages)

		// Register MCP integration (HU-033) — stub handler for now
		// TODO: Full MCP integration in Fase 2
		mux.HandleFunc("POST /mcp", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte(`{"error":"MCP integration pending"}`))
		})
	}

	var readHeaderTimeout, writeTimeout time.Duration
	if cfgPath != "" {
		// Re-load the config just for timeouts
		reg, err := registry.Load(cfgPath, nil)
		if err == nil {
			rMs, wMs := reg.ServerTimeouts()
			if rMs > 0 {
				readHeaderTimeout = time.Duration(rMs) * time.Millisecond
			}
			if wMs > 0 {
				writeTimeout = time.Duration(wMs) * time.Millisecond
			}
		}
	}
	if readHeaderTimeout == 0 {
		readHeaderTimeout = 5 * time.Second
	}
	if writeTimeout == 0 {
		writeTimeout = 30 * time.Second
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}

	go func() {
		log.Printf("gateway escuchando en :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// buildAdapters constructs and returns adapters for all configured providers.
func buildAdapters(reg *registry.Registry) map[string]adapter.Adapter {
	adapters := make(map[string]adapter.Adapter)

	// For MVP, build adapters for known provider types based on registry config.
	// This is a simplified approach; a fuller version would introspect registry.providers dynamically.

	// Manually add OpenAI adapter (always available if API key is set)
	if key := reg.APIKey("openai"); key != "" {
		adapters["openai"] = adapteropenai.New("https://api.openai.com/v1", key)
	}

	// Manually add Anthropic adapter
	if key := reg.APIKey("anthropic"); key != "" {
		adapters["anthropic"] = adapteranthropc.New("https://api.anthropic.com", key)
	}

	// Manually add Google adapter
	if key := reg.APIKey("google"); key != "" {
		adapters["google"] = adaptergoogle.New("", key)
	}

	// Add AIHubMix if configured
	if key := reg.APIKey("aihubmix"); key != "" {
		adapters["aihubmix"] = adapteraihubmix.New("https://api.aihubmix.com/v1", key)
	}

	// Add local Ollama if available (no API key needed)
	adapters["local"] = adapterlocal.New("http://localhost:11434")

	// Add OmniRoute adapter (local provider, no API key needed)
	adapters["omniroute"] = adapteromniroute.New(adapteromniroute.Config{
		BaseURL: "http://omniroute:20128/v1",
		APIKey:  "",
	})

	return adapters
}

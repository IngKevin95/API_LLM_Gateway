// Command gateway is the entrypoint del API LLM Gateway.
// Scaffold mínimo runnable: levanta un servidor HTTP con /health y /metrics.
// Registry, Router, Adapters y demás componentes se cablean por slice (EP-XXX).
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
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
	if cfgPath != "" {
		if _, err := registry.Load(cfgPath, nil); err != nil {
			log.Fatalf("registry: %v", err) // fail-fast, no arranca en estado parcial
		}
	} else {
		log.Printf("WARN gateway: sin config.yaml, arrancando en modo scaffold (solo /health)")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	// ponytail: /metrics es un stub JSON hasta HU-017/HU-023 (EP-007).
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	var readHeaderTimeout, writeTimeout time.Duration
	if cfgPath != "" {
		// Re-load the config just for timeouts (since registry isn't returned globally yet in scaffold)
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

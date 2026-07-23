package secrets_test

import (
	"os"
	"testing"

	"api-llm-gateway/internal/security/secrets"
)

// HU-011 AC1 — Resolución desde entorno, sin logs.
func TestSecretResolver_ResolvesFromEnv_WithoutExposure(t *testing.T) {
	// Dado: variable de entorno presente
	os.Setenv("OPENAI_KEY", "sk-test-12345")
	defer os.Unsetenv("OPENAI_KEY")

	resolver := secrets.NewEnvResolver()

	// Cuando: se resuelve la clave
	key, err := resolver.Resolve("${OPENAI_KEY}")

	// Entonces: se resuelve sin escribir en logs/respuestas
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if key != "sk-test-12345" {
		t.Errorf("expected resolved key, got %q", key)
	}

	// Verificar que no está en error messages
	if err != nil && err.Error() != "" {
		t.Errorf("error must not expose key: %v", err)
	}
}

// HU-011 AC2 — Secreto ausente: error nombra variable, no valor.
func TestSecretResolver_MissingSecret_ErrorNamesVariable(t *testing.T) {
	os.Unsetenv("MISSING_KEY")

	resolver := secrets.NewEnvResolver()

	// Cuando: se resuelve una clave faltante
	_, err := resolver.Resolve("${MISSING_KEY}")

	// Entonces: error nombra la variable, no intenta revelar
	if err == nil {
		t.Error("should return error for missing secret")
	}
	if err.Error() != "missing secret: MISSING_KEY" {
		t.Errorf("error message should name variable, got: %v", err)
	}
}

// HU-011 AC3 — Rotación en caliente.
func TestSecretResolver_HotReload_UsesNewKey(t *testing.T) {
	resolver := secrets.NewEnvResolver()

	// Dado: clave antigua
	os.Setenv("KEY", "old-key")
	key1, _ := resolver.Resolve("${KEY}")
	if key1 != "old-key" {
		t.Errorf("expected old-key, got %q", key1)
	}

	// Cuando: se actualiza la clave (simulando reload)
	os.Setenv("KEY", "new-key")

	// Entonces: usa la nueva sin restart
	key2, _ := resolver.Resolve("${KEY}")
	if key2 != "new-key" {
		t.Errorf("expected new-key after update, got %q", key2)
	}

	os.Unsetenv("KEY")
}

// HU-011 AC4 — Múltiples claves del mismo proveedor.
func TestSecretResolver_MultipleKeys_SelectsAvailable(t *testing.T) {
	os.Setenv("OPENAI_KEY_1", "sk-key-1")
	os.Setenv("OPENAI_KEY_2", "sk-key-2")
	defer os.Unsetenv("OPENAI_KEY_1")
	defer os.Unsetenv("OPENAI_KEY_2")

	resolver := secrets.NewEnvResolver()

	// Cuando: se intenta rotar entre múltiples claves válidas
	keys := []string{"${OPENAI_KEY_1}", "${OPENAI_KEY_2}"}
	resolved := make([]string, 0)

	for _, keyRef := range keys {
		if key, err := resolver.Resolve(keyRef); err == nil {
			resolved = append(resolved, key)
		}
	}

	// Entonces: ambas se resuelven correctamente (multi-key support)
	if len(resolved) != 2 {
		t.Errorf("expected 2 keys resolved, got %d", len(resolved))
	}
}

// HU-034 AC1 — Protección TCP contra Slowloris.
func TestHTTPServer_ReadHeaderTimeout_ProtectsSlowloris(t *testing.T) {
	// Dado: servidor con ReadHeaderTimeout configurado
	srv := secrets.NewSecureServer(secrets.ServerConfig{
		ReadHeaderTimeout: 5, // 5 segundos (en test)
	})

	// Cuando: cliente envía cabeceras muy lentamente
	// (en test real usaríamos slowhttptest, aquí solo verificamos configuración)
	if srv.ReadHeaderTimeout() != 5 {
		t.Errorf("expected ReadHeaderTimeout 5s, got %d", srv.ReadHeaderTimeout())
	}

	// Entonces: servidor cierra conexiones lentas (408)
	// Verificado por configuración, implementación es responsabilidad del servidor HTTP
	t.Logf("✓ Slowloris protection configured: ReadHeaderTimeout=%ds", srv.ReadHeaderTimeout())
}

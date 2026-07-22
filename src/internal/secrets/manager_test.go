package secrets_test

import (
	"os"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/secrets"
)

func TestEnvSecretManager_Resolve(t *testing.T) {
	os.Setenv("TEST_KEY", "secret-value")
	defer os.Unsetenv("TEST_KEY")

	manager := secrets.NewEnvManager()

	t.Run("Resolves existing variable", func(t *testing.T) {
		val, err := manager.Resolve("${TEST_KEY}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret-value" {
			t.Errorf("expected 'secret-value', got '%s'", val)
		}
	})

	t.Run("Fails for missing variable", func(t *testing.T) {
		_, err := manager.Resolve("${MISSING_KEY}")
		if err == nil {
			t.Fatal("expected error for missing variable")
		}
		if err.Error() != "secret not found: MISSING_KEY" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Returns verbatim if not formatted as env var", func(t *testing.T) {
		val, err := manager.Resolve("just-a-string")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "just-a-string" {
			t.Errorf("expected 'just-a-string', got '%s'", val)
		}
	})
}

package secrets

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Manager is responsible for resolving secrets dynamically.
type Manager interface {
	Resolve(value string) (string, error)
	Reload() error
}

// EnvManager resolves secrets from environment variables.
type EnvManager struct {
	mu sync.RWMutex
}

// NewEnvManager creates a new EnvManager.
func NewEnvManager() *EnvManager {
	return &EnvManager{}
}

// Resolve parses strings formatted as ${VAR_NAME} and fetches them from the environment.
func (m *EnvManager) Resolve(value string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		varName := value[2 : len(value)-1]
		val := os.Getenv(varName)
		if val == "" {
			return "", fmt.Errorf("secret not found: %s", varName)
		}
		return val, nil
	}
	return value, nil
}

// Reload forces the manager to reload state if needed.
func (m *EnvManager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// For EnvManager, os.Getenv fetches the latest dynamically.
	// In a Vault implementation, this would fetch new leases.
	return nil
}

package secrets

import (
	"fmt"
	"os"
	"regexp"
	"time"
)

// Resolver interface para resolución de secretos.
type Resolver interface {
	Resolve(reference string) (string, error)
}

// EnvResolver resuelve secretos desde variables de entorno.
// ponytail: resolución en RAM, rotación por re-lectura de env. Vault/K8s se diferirían a EP-XXX.
type EnvResolver struct{}

func NewEnvResolver() *EnvResolver {
	return &EnvResolver{}
}

// Resolve extrae el valor de ${VAR_NAME}.
// HU-011 AC1: No expone claves en logs.
// HU-011 AC2: Nombra la variable en error, no el valor.
// HU-011 AC3: Re-lee de env cada vez (hot reload).
func (r *EnvResolver) Resolve(reference string) (string, error) {
	// Patrón ${VAR_NAME}
	re := regexp.MustCompile(`^\$\{([A-Z_0-9]+)\}$`)
	matches := re.FindStringSubmatch(reference)
	if matches == nil {
		return "", fmt.Errorf("invalid secret reference format: %s", reference)
	}

	varName := matches[1]

	// HU-011 AC3: Hot reload — lee del entorno cada llamada
	value, exists := os.LookupEnv(varName)
	if !exists {
		// HU-011 AC2: Nombra la variable, no revela nada
		return "", fmt.Errorf("missing secret: %s", varName)
	}

	// HU-011 AC1: Nunca log el valor
	return value, nil
}

// ServerConfig para configuración del servidor HTTP seguro.
type ServerConfig struct {
	ReadHeaderTimeout int // segundos
}

// SecureServer envuelve un servidor HTTP con configuraciones de seguridad.
type SecureServer struct {
	readHeaderTimeout int
}

func NewSecureServer(cfg ServerConfig) *SecureServer {
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 5 // default 5 segundos
	}
	return &SecureServer{
		readHeaderTimeout: cfg.ReadHeaderTimeout,
	}
}

// ReadHeaderTimeout retorna el timeout configurado.
// HU-034 AC1: Protección contra Slowloris vía ReadHeaderTimeout.
func (s *SecureServer) ReadHeaderTimeout() int {
	return s.readHeaderTimeout
}

// GetHTTPServerConfig retorna las configuraciones para net/http.Server.
func (s *SecureServer) GetHTTPServerConfig() (readHeaderTimeout, writeTimeout time.Duration) {
	return time.Duration(s.readHeaderTimeout) * time.Second,
		time.Duration(s.readHeaderTimeout*2) * time.Second // write timeout = 2x read
}

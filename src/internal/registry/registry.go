// Package registry carga el catálogo declarativo (providers, models, routing)
// desde config.yaml a memoria RAM y lo expone al resto del Gateway.
package registry

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/IngKevin95/API_LLM_Gateway/internal/secrets"
)

// envRef matchea una referencia a variable de entorno: ${NOMBRE}.
var envRef = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)

// Model es un modelo concreto de un provider con sus atributos de score.
type Model struct {
	Name            string   `yaml:"name"`
	Capabilities    []string `yaml:"capabilities"`
	QualityScore    int      `yaml:"quality_score"`
	LatencyP50ms    int      `yaml:"latency_p50_ms"`
	CostPer1M       int      `yaml:"cost_per_1m_tokens"`
	MaxContextToks  int      `yaml:"max_context_tokens"`
	Disabled        bool     `yaml:"disabled"`

	// Provider al que pertenece (rellenado en Load).
	ProviderID string `yaml:"-"`
}

// Provider agrupa modelos y parámetros de red físicos.
type Provider struct {
	ID          string  `yaml:"id"`
	Type        string  `yaml:"type"`
	BaseURL     string  `yaml:"base_url"`
	APIKey      string  `yaml:"api_key"`
	MaxInFlight int     `yaml:"max_in_flight"`
	Models      []Model `yaml:"models"`
}

// CapabilityRoute declara la cadena de providers de una capacidad.
type CapabilityRoute struct {
	Providers           []string `yaml:"providers"`
	StreamIdleTimeoutMs int      `yaml:"stream_idle_timeout_ms"`
}

type routing struct {
	Capabilities map[string]CapabilityRoute `yaml:"capabilities"`
}

type serverConfig struct {
	ReadHeaderTimeoutMs int `yaml:"read_header_timeout_ms"`
	WriteTimeoutMs      int `yaml:"write_timeout_ms"`
}

type rawConfig struct {
	Server    serverConfig `yaml:"server"`
	Providers []Provider   `yaml:"providers"`
	Routing   routing      `yaml:"routing"`
}

type Registry struct {
	server    serverConfig
	providers map[string]Provider
	routing   map[string]CapabilityRoute
	available map[string]bool
	sm        secrets.Manager
}

// Load lee y valida config.yaml, cargándolo en RAM. Si sm es nil, usa env por defecto.
func Load(path string, sm secrets.Manager) (*Registry, error) {
	if sm == nil {
		sm = secrets.NewEnvManager()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("registry: leer %s: %w", path, err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("registry: parsear %s: %w", path, err)
	}

	if err := validate(raw, path); err != nil {
		return nil, err
	}

	reg := &Registry{
		server:    raw.Server,
		providers: make(map[string]Provider, len(raw.Providers)),
		routing:   raw.Routing.Capabilities,
		sm:        sm,
	}
	count := 0
	for _, p := range raw.Providers {
		// Intenta resolver la API key. Si falla, el provider se deshabilita silenciosamente en RAM.
		if m := envRef.FindStringSubmatch(p.APIKey); m != nil {
			val, err := sm.Resolve(p.APIKey)
			if err != nil {
				log.Printf("WARN registry: omitiendo provider %q por fallo de secretos: %v", p.ID, err)
				continue
			}
			p.APIKey = val
		}
		for i := range p.Models {
			p.Models[i].ProviderID = p.ID
			count++
		}
		reg.providers[p.ID] = p
	}

	// Marca disponibilidad por capacidad; WARN si alguna queda sin modelos.
	reg.available = make(map[string]bool, len(reg.routing))
	for cap := range reg.routing {
		ok := len(reg.ModelsFor(cap)) > 0
		reg.available[cap] = ok
		if !ok {
			log.Printf("WARN registry: capacidad %q sin modelos habilitados, marcada no disponible", cap)
		}
	}

	fmt.Printf("registry: %d modelos cargados desde %s\n", count, path)
	return reg, nil
}

// ModelsFor devuelve los modelos habilitados para una capacidad, en el orden
// de la cadena de providers de esa capacidad.
func (r *Registry) ModelsFor(capability string) []Model {
	route, ok := r.routing[capability]
	if !ok {
		return nil
	}
	var out []Model
	for _, pid := range route.Providers {
		p, ok := r.providers[pid]
		if !ok {
			continue
		}
		for _, m := range p.Models {
			if hasCapability(m, capability) {
				out = append(out, m)
			}
		}
	}
	return out
}

// ModelCount es el total de modelos cargados en el catálogo.
func (r *Registry) ModelCount() int {
	n := 0
	for _, p := range r.providers {
		n += len(p.Models)
	}
	return n
}

// validate comprueba los campos obligatorios del catálogo, fail-fast.
func validate(raw rawConfig, path string) error {
	if len(raw.Providers) == 0 {
		return fmt.Errorf("registry: %s: campo obligatorio faltante: 'providers' vacío", path)
	}
	for i, p := range raw.Providers {
		if p.ID == "" {
			return fmt.Errorf("registry: %s: provider #%d: campo obligatorio faltante: 'id'", path, i+1)
		}
		// Secreto literal prohibido: exige referencia ${VAR}. No imprime el valor.
		if p.APIKey != "" && !envRef.MatchString(p.APIKey) {
			return fmt.Errorf("registry: %s: provider %q: api_key literal prohibido; usa una referencia de entorno ${VAR}", path, p.ID)
		}
		for j, m := range p.Models {
			if m.Name == "" {
				return fmt.Errorf("registry: %s: provider %q modelo #%d: campo obligatorio faltante: 'name'", path, p.ID, j+1)
			}
		}
	}
	return nil
}

// MaxInFlight es el tope de peticiones concurrentes de un provider (Circuit Breaker).
func (r *Registry) MaxInFlight(providerID string) int {
	return r.providers[providerID].MaxInFlight
}

// StreamIdleTimeoutMs es el timeout de stream ocioso de una capacidad.
func (r *Registry) StreamIdleTimeoutMs(capability string) int {
	return r.routing[capability].StreamIdleTimeoutMs
}

// Available indica si una capacidad tiene al menos un modelo habilitado.
func (r *Registry) Available(capability string) bool {
	return r.available[capability]
}

// HasCapability indica si la capacidad está declarada en el routing (aunque
// no tenga modelos habilitados). Distingue "desconocida" de "sin modelos".
func (r *Registry) HasCapability(capability string) bool {
	_, ok := r.routing[capability]
	return ok
}

// FindModel busca un modelo por nombre en todo el catálogo.
func (r *Registry) FindModel(name string) (Model, bool) {
	for _, p := range r.providers {
		for _, m := range p.Models {
			if m.Name == name {
				return m, true
			}
		}
	}
	return Model{}, false
}

// ModelNames devuelve los nombres de todos los modelos del catálogo (orden estable).
func (r *Registry) ModelNames() []string {
	var out []string
	for _, p := range r.providers {
		for _, m := range p.Models {
			out = append(out, m.Name)
		}
	}
	sort.Strings(out)
	return out
}

// APIKey devuelve la API key resuelta de un provider (vacía si no existe).
func (r *Registry) APIKey(providerID string) string {
	// Intentar resolver dinámicamente si es una referencia en vivo
	val := r.providers[providerID].APIKey
	if r.sm != nil && envRef.MatchString(val) {
		res, err := r.sm.Resolve(val)
		if err == nil {
			return res
		}
	}
	return val
}

// ServerTimeouts devuelve los timeouts en milisegundos para protección TCP.
func (r *Registry) ServerTimeouts() (readHeaderMs, writeMs int) {
	return r.server.ReadHeaderTimeoutMs, r.server.WriteTimeoutMs
}

func hasCapability(m Model, capability string) bool {
	for _, c := range m.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

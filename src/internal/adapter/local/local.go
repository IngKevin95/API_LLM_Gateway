// Package local implementa el Adapter para servidores locales OpenAI-compatibles
// (Ollama / vLLM / LM Studio), reutilizando el adapter OpenAI contra un base_url
// local. Es el último eslabón de degradación de la cadena de failover.
package local

import "github.com/IngKevin95/API_LLM_Gateway/internal/adapter/openai"

// New crea un adapter local OpenAI-compatible apuntando a baseURL (sin api_key).
// El timeout se aplica vía el context de la petición (TTFT/timeout dinámico del
// Failover Engine); las respuestas no OpenAI-compatibles se normalizan a
// *adapter.ProviderError por el adapter OpenAI subyacente, sin crashear.
func New(baseURL string) *openai.Adapter {
	ad := openai.New(baseURL, "")
	ad.Name = "local"
	return ad
}

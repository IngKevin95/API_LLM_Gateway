# Arquitectura de Fábrica de IA Local + Gateway Inteligente
## Consolidación de conversaciones

**Autor:** Kevin Beltrán  
**Versión:** 1.1

---

# Objetivo

Construir una fábrica de agentes de IA completamente desacoplada del proveedor de modelos, donde un Gateway Inteligente decida automáticamente qué LLM utilizar según:

- Calidad
- Costo
- Latencia
- Disponibilidad
- Cuota restante
- Tipo de tarea
- Contexto
- Herramientas disponibles

Todo pensado para que cualquier agente pueda cambiar de modelo sin modificar una sola línea de código.

---

# Filosofía

Los agentes NO conocen ningún modelo.

Los agentes únicamente conocen una interfaz.

```
Agent
    ↓
Gateway IA
    ↓
Proveedor seleccionado
```

El Gateway se convierte en el cerebro de la plataforma.

---

# Objetivos del Gateway

El Gateway deberá:

- Balancear modelos
- Cambiar automáticamente de proveedor
- Detectar fallos
- Controlar cuotas
- Medir calidad
- Medir latencia
- Elegir el modelo óptimo
- Mantener estadísticas históricas
- Registrar costos
- Administrar múltiples API Keys

---

# Componentes principales

## Agentes

Ejemplos:

- Agente Comercial
- Agente RRHH
- Agente Jurídico
- Agente Contable
- Agente CRM
- Agente Zoho
- Agente RAG
- Agente Programador
- Agente Documentación
- Agente DevOps
- Agente QA

Todos llaman únicamente al Gateway.

---

## Gateway IA

Responsable de:

- Router
- Caché
- Telemetría
- Reintentos
- Failover
- Costos
- Estadísticas
- Cuotas

---

## Adaptadores

Cada proveedor tiene su propio adaptador.

Ejemplo:

```
OpenAI Adapter

Google Adapter

Anthropic Adapter

OpenRouter Adapter

AIHubMix Adapter

OpenCode Adapter

Local Adapter

Ollama Adapter
```

---

# OpenCode

Durante la conversación se discutió utilizar **OpenCode** como capa de abstracción para desarrollo.

La idea NO es reemplazar el Gateway, sino aprovechar OpenCode como cliente inteligente para tareas de programación y desarrollo.

Arquitectura:

```
IDE

↓

OpenCode

↓

Gateway

↓

LLMs
```

Ventajas:

- Compatible con múltiples proveedores.
- Excelente para desarrollo de software.
- Permite cambiar de modelo sin modificar el flujo del desarrollador.
- Puede integrarse con herramientas MCP.
- Facilita el uso de modelos locales y remotos.

El Gateway seguiría siendo el encargado de decidir qué modelo utilizar.

---

# Modelos Locales

También se discutió ejecutar modelos localmente.

Ejemplos:

- Qwen
- DeepSeek
- Llama
- Mistral
- Gemma
- GLM
- Phi
- Nemotron

Mediante:

- Ollama
- vLLM
- SGLang
- LM Studio

El Gateway debe tratarlos exactamente igual que un proveedor remoto.

---

# Arquitectura híbrida

```
                Gateway IA
                     │
        ┌────────────┼────────────┐
        │            │            │
     OpenAI       Google      Anthropic
        │            │            │
     OpenRouter    AIHubMix     Local
                                  │
                              Ollama
                                  │
                           Qwen / Llama
```

Para el Gateway todos son simplemente proveedores.

---

# Modelos gratuitos encontrados

Durante la investigación se encontraron múltiples modelos gratuitos en AIHubMix.

Entre ellos:

## OpenAI

- GPT-5.5 Free
- GPT-4.1 Free
- GPT Image 2 Free

---

## Google

- Gemini 3 Flash Preview
- Gemini 3.1 Flash Image Preview

---

## GLM

- GLM 4.7 Flash
- Coding GLM 4.7
- Coding GLM 5
- Coding GLM 5.1
- GLM 5

---

## MiniMax

- MiniMax M2.5
- MiniMax M2.7
- MiniMax M3

---

## Xiaomi

- Xiaomi Mimo V2 Pro

---

## Video

- Seedance

---

Todos poseen cuotas gratuitas diarias.

Generalmente:

- 5 requests por minuto
- 500 requests por día
- hasta 1 millón de tokens diarios (según el modelo)

---

# Router Inteligente

La idea es que el Gateway seleccione automáticamente el modelo.

Ejemplo:

```
¿Es código?

↓

Sí

↓

GPT-5.5
↓

Si está ocupado

↓

GLM 5.1

↓

Si falla

↓

MiniMax

↓

Si falla

↓

Qwen Local
```

Otro ejemplo:

```
Generar imágenes

↓

GPT Image

↓

Si falla

↓

Gemini Image

↓

Si falla

↓

Stable Diffusion Local
```

---

# Configuración YAML

La conversación llevó a la idea de mantener TODA la configuración en YAML.

Ejemplo conceptual:

```yaml
providers:

  openai:
    enabled: true
    api_key: ${OPENAI_KEY}

  google:
    enabled: true
    api_key: ${GOOGLE_KEY}

  aihubmix:
    enabled: true
    api_key: ${AIHUB_KEY}

models:

  gpt55:
    provider: openai
    quality: 10
    latency: 8
    cost: 10

  glm5:
    provider: aihubmix
    quality: 8
    latency: 9
    cost: 10

routing:

  coding:
    - gpt55
    - glm5
    - minimax

  reasoning:
    - gpt55
    - gemini
    - llama-local

  image:
    - gpt-image
    - gemini-image
```

Todo configurable sin recompilar.

---

# API Keys

La conversación también exploró la posibilidad de utilizar múltiples API Keys.

Ejemplo:

```
Proveedor A

API Key 1

API Key 2

API Key 3
```

El Gateway podría seleccionar automáticamente cuál utilizar.

Sin embargo, es importante considerar que:

- Muchos proveedores agrupan cuotas por cuenta u organización.
- Crear múltiples cuentas únicamente para ampliar límites puede violar los Términos de Servicio del proveedor.
- La arquitectura debe diseñarse para soportar múltiples claves legítimas (por ejemplo, de distintos proveedores o proyectos), pero no asumir que pueden usarse para eludir cuotas.

---

# Métricas que almacenará el Gateway

Cada modelo tendrá métricas dinámicas.

Ejemplo:

```yaml
gpt55:

 latency_ms:

 avg_response:

 success_rate:

 daily_tokens:

 requests_today:

 quota_remaining:

 availability:

 quality_score:

 reasoning_score:

 coding_score:

 image_score:

 cost_per_token:
```

Todo actualizado automáticamente.

---

# Router basado en puntajes

El Gateway calculará una puntuación.

Ejemplo:

```
Score =
Calidad
+
Velocidad
+
Disponibilidad
+
Cuota restante
+
Costo
```

El modelo con mayor puntuación será seleccionado.

---

# Telemetría

Cada llamada registrará:

- Modelo utilizado
- Tokens
- Tiempo
- Costo
- Usuario
- Agente
- Herramientas utilizadas
- Errores
- Retries
- Cache hit

---

# Dashboard

Se propuso visualizar:

## Disponibilidad

```
GPT-5.5

98%

Gemini

96%

GLM

99%

MiniMax

95%
```

---

## Latencia

```
GPT

1200 ms

Gemini

980 ms

GLM

750 ms
```

---

## Costos

Costo diario

Costo mensual

Costo por proveedor

Costo por agente

---

## Consumo

Tokens

Requests

Errores

Quotas restantes

---

# Escalabilidad

Agregar un nuevo proveedor debería requerir únicamente:

1.

Crear Adapter

2.

Agregar YAML

3.

Registrar Provider

Nada más.

---

# Beneficios

- Totalmente desacoplado
- Independiente del proveedor
- Escalable
- Fácil de mantener
- Configurable
- Alta disponibilidad
- Balanceo automático
- Failover
- Optimización de costos
- Optimización de latencia
- Soporte para modelos locales y remotos
- Integración sencilla con OpenCode y otros clientes

---

# Visión futura

La meta es construir una plataforma donde decenas o cientos de agentes especializados puedan operar de forma coordinada sobre un Gateway Inteligente que actúe como un "Sistema Operativo de IA". Este Gateway decidirá en tiempo real el mejor modelo para cada tarea, integrando proveedores comerciales, modelos gratuitos y modelos locales, con observabilidad completa, telemetría avanzada y capacidad de evolución sin modificar el código de los agentes.

A futuro, el Gateway también podría incorporar:

- Evaluación automática de calidad de respuestas.
- A/B testing entre modelos.
- Autoaprendizaje del enrutamiento según resultados históricos.
- Presupuestos por proyecto o cliente.
- Integración con MCP (Model Context Protocol).
- Ejecución de herramientas y workflows.
- Memoria compartida entre agentes.
- Orquestación multiagente.
- Políticas de seguridad y gobernanza por organización.

El objetivo final es disponer de una infraestructura de IA modular, resiliente y preparada para incorporar cualquier modelo que aparezca en el mercado sin depender de un único proveedor.
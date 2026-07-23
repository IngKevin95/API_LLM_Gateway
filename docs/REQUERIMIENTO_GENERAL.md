# Arquitectura General de la Fábrica Agéntica
## Documento de Diseño Técnico
Versión 1.0

---

# Objetivo

Construir una plataforma de agentes inteligentes completamente desacoplada del proveedor de modelos de IA.

El objetivo principal es que:

- ningún agente conozca qué modelo utiliza.
- ningún agente dependa de OpenAI, Anthropic o Google.
- cualquier modelo pueda ser reemplazado sin modificar el código.

Toda la inteligencia de selección será responsabilidad del AI Gateway.

---

# Filosofía

Los agentes NO consumen modelos.

Los agentes consumen capacidades.

Por ejemplo:

```
Necesito programar

Necesito razonar

Necesito analizar imágenes

Necesito OCR

Necesito embeddings
```

El Gateway decidirá cuál modelo satisface mejor esa necesidad.

---

# Arquitectura General

```
                    Usuario
                       │
                       ▼
                 CLI / IDE
                       │
                       ▼
               Agente IA (OpenCode)
                       │
                       ▼
                  AI Gateway
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
   OpenAI        Anthropic       Google
        │              │              │
        ▼              ▼              ▼
      GPT          Claude         Gemini

                 ▼

             AIHubMix

                 ▼

          GLM / MiniMax /
          Xiaomi / GPT /
          Gemini / etc.

                 ▼

          Modelos Locales

        Ollama
        vLLM
        LM Studio
```

---

# Diferencia entre CLI, Agente, Modelo y Gateway

## CLI

La CLI únicamente recibe comandos.

Ejemplos:

```
claude

opencode

factory

openhands
```

No piensa.

No genera código.

No utiliza IA.

Simplemente entrega instrucciones al agente.

---

## Agente

El agente es quien implementa el ciclo:

```
Pensar

↓

Usar herramientas

↓

Leer archivos

↓

Ejecutar terminal

↓

Analizar resultados

↓

Pensar nuevamente
```

Ejemplos:

- OpenCode
- Claude Code
- OpenHands
- Roo Code
- Goose
- Aider

---

## Modelo

Es el cerebro.

Ejemplos:

- GPT-5.5
- Claude Opus
- Gemini
- GLM
- MiniMax
- DeepSeek
- Kimi
- Llama

El modelo únicamente responde.

No sabe abrir archivos.

No usa Git.

No ejecuta Docker.

---

## AI Gateway

Es el componente encargado de decidir:

- qué proveedor usar
- qué modelo utilizar
- qué API Key consumir
- cuándo cambiar de proveedor
- cuándo hacer failover
- cuánto dinero gastar

---

# Comparación de Agentes

## Claude Code

Ventajas

- Excelente experiencia.
- Herramientas muy maduras.
- Muy buena calidad.

Desventajas

- Depende de Anthropic.
- Cuotas limitadas.
- Poco configurable.

---

## Free Claude Code

Idea principal

Reutiliza el cliente de Claude Code.

En lugar de conectarse directamente a Anthropic puede conectarse a:

- AIHubMix
- OpenRouter
- Gateway propio

Arquitectura

```
Claude Code

↓

Free Claude Code

↓

Gateway

↓

Proveedor
```

Es una excelente alternativa para conservar la experiencia de Claude Code sin depender exclusivamente de Anthropic.

---

## OpenCode

Actualmente es una de las mejores alternativas para programación.

Ventajas

- CLI moderna.
- Compatible con múltiples proveedores.
- MCP.
- Muy flexible.
- Fácil integración con Gateway.

Arquitectura

```
VS Code

↓

OpenCode

↓

Gateway

↓

Proveedor
```

Para este proyecto se considera el mejor candidato como interfaz principal de desarrollo.

---

## OpenHands

Orientado a tareas autónomas.

Ideal para:

- resolver issues
- refactorizaciones grandes
- automatización

No está pensado como CLI principal del desarrollador.

---

## Roo Code

Muy fuerte dentro de VS Code.

Excelente para:

- Arquitectura
- Debug
- Refactorización

No posee una CLI independiente.

---

# AI Gateway

Es el cerebro de toda la plataforma.

Responsabilidades

- Router
- Balanceador
- Caché
- Observabilidad
- Costos
- Cuotas
- Failover
- Telemetría

---

# Adaptadores

Cada proveedor tendrá un Adapter.

Ejemplos

```
OpenAI Adapter

Anthropic Adapter

Google Adapter

OpenRouter Adapter

AIHubMix Adapter

Ollama Adapter

vLLM Adapter
```

Agregar un proveedor nuevo únicamente requerirá crear un Adapter.

---

# Registro YAML

Toda la configuración será declarativa.

Ejemplo

```yaml
providers:

  - id: aihubmix

    endpoint: https://...

    api_key: ...

    enabled: true

    models:

      - gpt55
      - glm5
      - minimax

  - id: openai

    endpoint: https://...

    api_key: ...

    models:

      - gpt55

  - id: google

    endpoint: https://...

    api_key: ...

    models:

      - gemini
```

---

# Catálogo de Modelos

Cada modelo tendrá atributos.

```yaml
gpt55

quality:100

coding:100

reasoning:100

speed:75

vision:95

cost:10
```

Otro ejemplo

```yaml
glm5

quality:92

coding:97

speed:95

cost:0
```

---

# Model Router

Los agentes nunca pedirán modelos.

Solicitarán capacidades.

```
router.coding()

router.reasoning()

router.vision()

router.embedding()

router.image()
```

El Gateway decidirá el modelo.

---

# Router Inteligente

Ejemplo

```
Programación Java

↓

GPT-5.5

↓

si falla

↓

GLM 5

↓

si falla

↓

DeepSeek

↓

si falla

↓

Llama Local
```

---

# Modelos Gratuitos encontrados

Durante la investigación se encontraron:

## GPT

- GPT-5.5 Free
- GPT-4.1 Free
- GPT Image

---

## Gemini

- Gemini Flash
- Gemini Image

---

## GLM

- GLM 4.7
- GLM 5
- GLM Coding
- GLM 5.1

---

## MiniMax

- M2.5
- M2.7
- M3

---

## Xiaomi

- Mimo V2

---

## Seedance

Modelo para Video.

---

# Administración de API Keys

La idea es registrar múltiples proveedores.

Ejemplo

```
OpenAI

Key 1

Key 2

Google

Key 1

AIHubMix

Key 1

Key 2
```

El Gateway seleccionará automáticamente cuál utilizar.

**Nota:** aunque técnicamente es posible registrar varias cuentas o credenciales, el diseño debe respetar los términos de uso de cada proveedor y no asumir que es válido crear cuentas adicionales para eludir cuotas gratuitas.

---

# Estado dinámico

Cada proveedor tendrá información viva.

```
Latencia

Errores

Disponibilidad

Quota

Tokens

Costo

Estado
```

---

# Health Monitor

Cada minuto ejecutará pruebas.

Medirá

- disponibilidad
- latencia
- throughput
- errores
- tiempo promedio

---

# Quota Manager

Controlará

- requests
- tokens
- cuota diaria
- cuota mensual
- costo

Cuando una API se agote dejará de utilizarse automáticamente.

---

# Failover

```
GPT

↓

429

↓

Gemini

↓

500

↓

GLM

↓

Timeout

↓

Llama Local
```

Todo transparente.

---

# Learning Engine

Registrará

```
Modelo

Prompt

Tokens

Costo

Tiempo

Errores

Calificación

Caso de uso
```

Con miles de registros aprenderá automáticamente.

---

# Dashboard

Mostrará

Disponibilidad

Latencia

Costos

Consumo

Errores

Ranking de modelos

Ranking de proveedores

Ranking de agentes

---

# Escalabilidad

Agregar un nuevo proveedor requiere únicamente

1.

Crear Adapter

2.

Agregar YAML

3.

Registrar capacidades

Nada más.

---

# Roadmap futuro

Fase 1

- Gateway básico
- YAML
- Router
- OpenCode
- AIHubMix

---

Fase 2

- Health Checks
- Dashboard
- Cache
- Failover

---

Fase 3

- IA que aprenda automáticamente
- Ajuste dinámico de pesos
- Predicción de costos
- Selección automática del mejor modelo

---

Fase 4

- Integración con MCP
- Multiagente
- Memoria compartida
- Workflows
- RAG
- Orquestación

---

# Objetivo Final

Construir una plataforma equivalente a un "Sistema Operativo de IA", donde:

- Los agentes sean independientes de los modelos.
- El Gateway centralice toda la inteligencia de selección.
- Los proveedores puedan cambiarse sin modificar los agentes.
- Se combinen modelos comerciales, gratuitos y locales.
- El sistema optimice automáticamente costo, calidad, latencia y disponibilidad.

El resultado será una infraestructura preparada para escalar desde unos pocos agentes hasta una fábrica completa de cientos de agentes especializados trabajando en paralelo sobre un único Gateway Inteligente.
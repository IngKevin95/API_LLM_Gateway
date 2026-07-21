# Diseño del AI Gateway Inteligente para una Fábrica de Agentes

## Objetivo

Diseñar un AI Gateway capaz de abstraer completamente el acceso a múltiples proveedores y modelos de IA, permitiendo que los agentes consuman capacidades (razonamiento, programación, visión, generación de imágenes, etc.) sin depender de un modelo específico.

La arquitectura debe ser escalable, tolerante a fallos, autoaprendizaje y optimizada por costo, latencia, calidad y disponibilidad.

---

# Motivación

Actualmente existen gateways como:

- LiteLLM
- Portkey
- OpenRouter
- Vercel AI Gateway

Sin embargo, el objetivo de este proyecto es construir un gateway propio que permita:

- Control total del enrutamiento.
- Administración de múltiples proveedores.
- Balanceo inteligente.
- Failover.
- Aprendizaje basado en métricas reales.
- Integración con la Fábrica de Agentes.
- Compatibilidad con modelos locales y remotos.

---

# Proveedores soportados

Ejemplo:

- OpenAI
- Anthropic
- Google Gemini
- AIHubMix
- GLM (Zhipu)
- MiniMax
- Xiaomi Mimo
- DeepSeek
- Ollama
- vLLM
- LM Studio
- OpenRouter

---

# Configuración mediante YAML

Toda la configuración de proveedores será declarativa.

Ejemplo:

```yaml
providers:

  - id: aihubmix_prod_1

    endpoint: https://api.aihubmix.com/v1

    api_key: ${AIHUBMIX_KEY}

    enabled: true

    models:

      - gpt-5.5-free
      - gpt-4.1-free
      - gemini-3-flash-preview-free
      - glm-5
      - minimax-m3
      - gpt-image-2-free

  - id: aihubmix_prod_2

    endpoint: https://api.aihubmix.com/v1

    api_key: ${AIHUBMIX_KEY_2}

    enabled: true

    models:

      - gpt-5.5-free
      - gpt-4.1-free
```

Este YAML únicamente describe:

- credenciales
- endpoint
- modelos accesibles
- estado

No contiene información dinámica.

---

# ¿Una API Key por modelo?

No es necesario.

En plataformas como AIHubMix normalmente una misma API Key puede acceder a múltiples modelos habilitados.

El gateway solamente cambia el parámetro:

```json
{
    "model":"gpt-5.5-free"
}
```

o

```json
{
    "model":"gemini-3-flash-preview-free"
}
```

---

# ¿Es posible usar varias cuentas?

Técnicamente sí.

Si el proveedor asigna la cuota por cuenta, el gateway podría administrar múltiples credenciales.

Ejemplo:

Cuenta A

- GPT-5.5
- GPT-4.1
- Gemini

Cuenta B

- GPT-5.5
- GPT-4.1

Cuenta C

- GPT-5.5

Cuando una cuenta alcanza su cuota diaria o devuelve HTTP 429, el gateway puede redirigir automáticamente las solicitudes hacia otra credencial disponible.

**Importante:** esta estrategia debe respetar siempre los términos de uso del proveedor. Muchos servicios consideran un incumplimiento crear múltiples cuentas con el único fin de eludir límites gratuitos.

---

# Estado dinámico

El gateway mantiene información en memoria (o Redis).

Ejemplo:

```python
ProviderState

provider = AIHubMix_01

status = HEALTHY

latency = 720 ms

quota_remaining = 870000 tokens

requests_remaining = 420

throughput = 14 req/min

last_failure = 3 min

average_response = 2.3 sec

cost = FREE
```

Este estado cambia constantemente.

No se almacena en YAML.

---

# Catálogo de modelos

Cada modelo posee una ficha técnica.

Ejemplo:

```yaml
gpt-5.5:

  reasoning: 100

  coding: 100

  writing: 98

  vision: 95

  speed: 72

  cost: 5
```

Otro ejemplo:

```yaml
glm-5:

  reasoning: 92

  coding: 98

  writing: 88

  speed: 90

  cost: 0
```

Gemini:

```yaml
gemini-3-flash:

  vision: 100

  reasoning: 90

  speed: 100

  cost: 0
```

Estos valores pueden comenzar siendo manuales y luego ajustarse automáticamente con base en el desempeño observado.

---

# Model Router

Los agentes nunca conocen el modelo.

En lugar de decir:

```python
usar GPT55
```

Simplemente solicitan una capacidad.

Ejemplo:

```python
router.chat()

router.reasoning()

router.coding()

router.vision()

router.image()

router.embedding()
```

El router selecciona automáticamente el modelo óptimo.

---

# Algoritmo de decisión

Para cada solicitud se calcula un Score.

Ejemplo:

GPT-5.5

Calidad:

100

Velocidad:

75

Costo:

20

Latencia:

60

Score:

82

Gemini Flash

Calidad:

90

Velocidad:

100

Costo:

100

Latencia:

95

Score:

96

Resultado:

Seleccionar Gemini.

---

# Health Check

El gateway ejecuta pruebas periódicas sobre todos los proveedores.

Verifica:

- disponibilidad
- latencia
- throughput
- errores HTTP
- errores 429
- errores 500
- tiempo promedio de respuesta

Con ello siempre conoce el estado real de cada proveedor.

---

# Failover

Si el proveedor principal falla:

```
GPT-5.5

↓

HTTP 429

↓

GPT-4.1

↓

HTTP 500

↓

Gemini

↓

Timeout

↓

GLM-5

↓

MiniMax
```

El cambio debe ser transparente para los agentes.

---

# Quota Manager

El gateway administra todas las cuotas.

Ejemplo:

- solicitudes por minuto
- solicitudes por día
- tokens diarios
- tokens mensuales
- cuota restante
- costo acumulado

Cuando una credencial se agota, automáticamente deja de utilizarse hasta su recuperación.

---

# Learning Engine

Uno de los componentes diferenciadores.

Después de cada interacción se registra:

```text
Modelo

Prompt Tokens

Completion Tokens

Tiempo

Costo

Errores

Usuario satisfecho

Calificación

Caso de uso
```

Con miles de registros el sistema aprende:

- cuál modelo programa mejor
- cuál resume mejor
- cuál responde más rápido
- cuál funciona mejor para RAG
- cuál genera mejores documentos
- cuál funciona mejor con OCR
- cuál funciona mejor con PDFs
- cuál funciona mejor para visión
- cuál tiene mejor relación calidad/costo

---

# Selección basada en contexto

El router podrá decidir según:

- tipo de tarea
- costo permitido
- latencia máxima
- cuota restante
- complejidad
- calidad requerida
- proveedor disponible

Ejemplo:

Generar código

↓

GLM-5

Analizar imágenes

↓

Gemini

Razonamiento complejo

↓

GPT-5.5

Documentación técnica

↓

GPT-5.5

Extracción rápida

↓

Gemini Flash

Respuesta económica

↓

MiniMax

---

# Componentes del Gateway

## 1. Registry

Mantiene el inventario de:

- proveedores
- endpoints
- modelos
- API Keys
- capacidades

---

## 2. Health Monitor

Monitorea continuamente:

- latencia
- disponibilidad
- errores
- rendimiento

---

## 3. Model Router

Selecciona el modelo óptimo considerando:

- calidad
- velocidad
- costo
- especialidad
- cuota disponible

---

## 4. Quota Manager

Controla:

- consumo
- límites
- disponibilidad
- rotación de credenciales

---

## 5. Learning Engine

Aprende automáticamente del uso real del sistema.

Actualiza continuamente la estrategia de selección.

---

# Beneficios

- Independencia del proveedor.
- Agentes desacoplados de modelos específicos.
- Cambio transparente entre modelos.
- Optimización automática por costo y calidad.
- Alta disponibilidad.
- Failover automático.
- Balanceo inteligente.
- Soporte para modelos locales y remotos.
- Aprendizaje continuo.
- Arquitectura preparada para escalar a cientos de agentes.

---

# Visión a futuro

El objetivo final es que el AI Gateway funcione como el "cerebro de enrutamiento" de toda la Fábrica de Agentes.

Los agentes nunca decidirán qué modelo utilizar. Simplemente declararán la capacidad requerida (razonamiento, programación, visión, OCR, generación de imágenes, búsqueda, embeddings, etc.) y el gateway elegirá dinámicamente la mejor combinación de proveedor, modelo y credencial disponible según métricas en tiempo real, cuotas, costo, latencia y desempeño histórico.
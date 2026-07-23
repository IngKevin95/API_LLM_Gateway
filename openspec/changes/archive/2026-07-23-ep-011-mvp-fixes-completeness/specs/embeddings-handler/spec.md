# Spec: Embeddings Handler (Modified)

## MODIFIED Requirements

### Requirement: ProcessEmbedding returns HTTP 200 with valid embeddings
The system SHALL fix ProcessEmbedding() to return HTTP 200 status (instead of 503) with embedding vectors populated from provider.

#### Scenario: Valid embedding request returns 200
- **WHEN** ProcessEmbedding receives valid request (input text + model)
- **THEN** response is HTTP 200 with {"data": [{"embedding": [0.1, 0.2, ...]}], "usage": {...}}

#### Scenario: Invalid request returns 400
- **WHEN** ProcessEmbedding receives invalid request (missing input, invalid encoding)
- **THEN** response is HTTP 400 with error message

#### Scenario: Model not found returns 404
- **WHEN** ProcessEmbedding requested with model not in router registry
- **THEN** response is HTTP 404 "Model not found"

### Requirement: Embeddings handler emits structured logging
The system SHALL emit JSON logs from embeddings handler with request_id, model, provider, vector dimension.

#### Scenario: Successful embedding logs completion
- **WHEN** ProcessEmbedding completes successfully
- **THEN** log includes {"level": "info", "request_id": "...", "provider": "...", "embedding_dim": N, "text_tokens": N}

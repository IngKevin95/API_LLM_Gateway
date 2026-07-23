## 1. Exposición de Métricas

- [ ] 1.1 Implementar endpoints `/v1/metrics` y `/v1/metrics/dashboard`
- [ ] 1.2 Agregar agregación de datos por modelo y proveedor
- [ ] 1.3 Incorporar middleware de autorización para scope de operador
- [ ] 1.4 Manejar error states (Empty state y BD no disponible)

## 2. Histórico de Peticiones

- [ ] 2.1 Definir estructura del log histórico (modelo, latencia, resultado, tokens)
- [ ] 2.2 Implementar persistencia asíncrona de las peticiones (sin bloquear hilo principal)
- [ ] 2.3 Implementar endpoint o mecanismo de calificación/feedback
- [ ] 2.4 Incorporar la lógica de redacción de PII antes de guardar

## 3. Learning Engine

- [ ] 3.1 Implementar job/cron heurístico de ajuste de pesos
- [ ] 3.2 Añadir validación de muestra mínima (>100 peticiones)
- [ ] 3.3 Configurar guardrails para topes de asignación de tráfico
- [ ] 3.4 Implementar rollback en caso de degradación de success rate

## 4. Caché Semántica

- [ ] 4.1 Integrar librería ligera zero-deps de Vector Search local
- [ ] 4.2 Implementar bypass de prompts cortos en la caché
- [ ] 4.3 Agregar verificación de timeout de <50ms para el lookup local
- [ ] 4.4 Integrar la validación de umbral de similitud (>0.98) para interceptar request

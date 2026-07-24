# Technical PRD: Fase 2 - Autonomía, Base de Datos y Caché Semántica

## 1. Visión General
Transformar el API LLM Gateway de un MVP estático (estado in-memory) a una plataforma Enterprise Autónoma. Se habilitará persistencia robusta, caché de ahorro de costos (semántica), optimización de pesos (Learning Engine) y gobernanza total desde una interfaz de usuario.

## 2. Decisiones Arquitectónicas Core
- **Stack de Datos**: PostgreSQL servirá como fuente de verdad única para Quotas, Auditoría y Tenants.
- **Stack Vectorial**: Extensión `pgvector` sobre PostgreSQL para agilizar el stack (evita desplegar Milvus/Weaviate en esta fase).
- **Frontend**: React + Vite (Dashboard) consumiendo nuevos endpoints de administración (Admin API).
- **Asincronía Segura**: Los logs de auditoría continúan enrutándose asíncronamente con un Write-Ahead Log (WAL) para sobrevivir caídas, pero ahora aterrizan en PostgreSQL cifrados con KMS.

## 3. Especificaciones Funcionales (Épicas)
- **EP-012**: Persistencia Real (Quotas/Auth), Caché Semántica (pgvector) para latencia de respuesta < 20ms, y Learning Engine (retroalimentación de latencias al router).
- **EP-013**: Interfaz de Dashboard (Admin UI) donde se crean tokens de proveedores in-house, se inyectan cifrados al KMS, y se visualiza la cuota y gasto exacto *por modelo* (ej. Claude 3.5 Sonnet = 25%, GPT-4o = 75%).

## 4. Criterios de Aceptación (Release Gate)
1. **Gobernanza Completa**: No debe existir necesidad de reiniciar el gateway para rotar llaves o actualizar cuotas de proveedores o tenants.
2. **Hit Rate**: La caché semántica debe poder atrapar al menos el 20% de consultas repetitivas de agentes (ej. "Resume this error log").
3. **Resiliencia**: Si PostgreSQL se cae temporalmente, el WAL y la caché in-memory deben sostener la operación del Gateway por al menos 5 minutos sin degradación visible para el cliente (salvo Admin UI).

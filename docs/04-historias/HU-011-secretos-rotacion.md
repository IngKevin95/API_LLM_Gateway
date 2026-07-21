---
id: HU-011
titulo: Gestionar y rotar secretos sin exponerlos
epica: EP-004B
prioridad: Should
complejidad: M
estado: lista
---

# Gestionar y rotar secretos sin exponerlos

Como **operador de seguridad**, quiero **que las API keys de proveedor se lean de variables de entorno/secret manager y puedan rotarse sin reiniciar ni aparecer en logs**, para **reducir el riesgo de fuga de credenciales**.

Contexto: gestión de secretos avanzada. Soporta multi-key legítimo. Actividad 5.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — resolución desde entorno | Dado que un provider con `api_key: ${OPENAI_KEY}` y la variable presente | Cuando la Gateway hace una llamada al proveedor | Entonces usa la clave resuelta sin escribirla nunca en logs ni respuestas |
| 2 | Error — secreto ausente | Dado que un provider que referencia `${OPENAI_KEY}` no definida | Cuando la Gateway intenta usar ese provider | Entonces marca el provider como no configurable y lo excluye, con un error que nombra la variable faltante (no su valor) |
| 3 | Happy — rotación en caliente | Dado que una clave que se actualiza en el secret manager | Cuando se dispara la recarga de secretos | Entonces las nuevas peticiones usan la clave nueva sin reiniciar la Gateway |
| 4 | Edge — múltiples claves del mismo proveedor | Dado que un provider con varias claves legítimas | Cuando una clave alcanza su límite | Entonces la Gateway rota a otra clave válida del mismo provider (sin crear cuentas para eludir cuotas) |

## Checklist INVEST

- [x] Independent — se apoya en HU-001 (Registry)
- [x] Negotiable — backend de secretos abierto (env/Vault)
- [x] Valuable — reduce riesgo de fuga
- [x] Estimable — resolución + recarga
- [x] Small — un sprint
- [x] Testable — variables presentes/ausentes/rotadas

## Notas técnicas

Nunca serializar secretos en errores. Rotación por señal o poll. Respetar ToS: multi-key solo para claves legítimas.

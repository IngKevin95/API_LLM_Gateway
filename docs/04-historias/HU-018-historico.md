---
id: HU-018
titulo: Registrar histórico de peticiones para aprendizaje
epica: EP-007
prioridad: Could
complejidad: M
estado: lista
---

# Registrar histórico de peticiones para aprendizaje

Como **operador de la plataforma**, quiero **que la Gateway persista un histórico de cada petición (modelo, capacidad, tokens, costo, tiempo, errores, calificación)**, para **poder auditar y consultar decisiones de enrutamiento pasadas de inmediato y, además, alimentar el ajuste automático futuro con datos reales**.

Contexto: base del Learning Engine (HU-019). Actividad 5, release v2.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — registro histórico | Dado que una petición completada | Cuando termina | Entonces se persiste un registro con modelo, capacidad, tokens, costo, tiempo, resultado y caso de uso, consultable después |
| 2 | Happy — calificación posterior | Dado que un registro existente | Cuando llega feedback de calidad de la respuesta | Entonces el registro se enriquece con la calificación sin duplicarse |
| 3 | Error — almacén lleno/no disponible | Dado que el almacén histórico sin espacio o caído | Cuando se completa una petición | Entonces la petición no falla por esto; el registro se degrada/encola y se emite alerta |
| 4 | Edge — redacción | Dado que un prompt con datos sensibles | Cuando se registra en el histórico | Entonces los datos sensibles se redactan igual que en auditoría; el histórico no es una fuga |

## Checklist INVEST

- [x] Independent — reutiliza el pipeline de redacción de HU-010; no depende de HU-017 ni HU-019 (ambas consumen este histórico después)
- [x] Negotiable — esquema de almacenamiento abierto
- [x] Valuable — habilita aprendizaje futuro
- [x] Estimable — persistencia acotada
- [x] Small — un sprint
- [x] Testable — registros simulados

## Notas técnicas

Retención configurable. Reutilizar la redacción de PII/secretos de HU-010.

---
id: HU-036
title: Integrar OmniRoute API
epic: EP-008
type: Should
---

# HU-036: Integrar OmniRoute API

## INVEST
- [x] Independent: el adapter funciona bajo la abstracción de enrutamiento ya implementada.
- [x] Negotiable: detalles de configuración de modelos específicos de OmniRoute se pueden ajustar en el YAML.
- [x] Valuable: añade un proveedor adicional que incrementa las opciones de modelos y rutas disponibles.
- [x] Estimable: misma complejidad que integrar OpenRouter, ya que también expone una API compatible o similar.
- [x] Small: solo implica escribir un nuevo adapter, sin afectar el core.
- [x] Testable: validable con llamadas reales y mocks.

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Petición exitosa | Un provider OmniRoute configurado en registry | La Gateway decide usar OmniRoute para la capacidad solicitada | Se transforma y reenvía el request, devolviendo respuesta exitosa unificada |
| 2 | Modelos de fallback | Falla el proveedor principal | La política de failover incluye OmniRoute en la cadena | La Gateway reintenta con OmniRoute y completa el request |
| 3 | Manejo de errores 429/500 | OmniRoute responde con error de capacidad o cuota | La petición se dirige a OmniRoute | El adapter lo traduce a un error estándar que dispara el failover hacia el siguiente proveedor |

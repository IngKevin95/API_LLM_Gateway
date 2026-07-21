# Diseño simple (Kent Beck) y code smells

`simple-design-reviewer` corre esta guía sobre código que **ya pasa sus tests** — el orden importa:
correcto primero, elegante después.

## Las 4 reglas de Beck, en orden de prioridad

1. **Pasa los tests.** Nada más cuenta si esto falla.
2. **Revela la intención.** Nombres y estructura cuentan la historia por sí solos; nadie debería
   tener que adivinar.
3. **Sin duplicación.** DRY de conocimiento, no de texto que casualmente se parece — aplica la regla
   de tres antes de extraer una abstracción.
4. **El mínimo de elementos** que satisfaga 1–3. YAGNI: nada especulativo, nada "por si se necesita
   después".

Cuando dos reglas chocan: la 1 siempre gana. Entre la 2 y la 3, prioriza eliminar duplicación. Entre
la 4 y las otras dos, prioriza intención clara y cero duplicación — la concisión nunca justifica
código oscuro.

## Smells comunes y su refactor

| Smell | Señal | Refactor |
|---|---|---|
| Función larga | hace demasiadas cosas a la vez | Extract Function; una responsabilidad por función |
| Lista de parámetros larga | 4+ parámetros | Introduce Parameter Object |
| Números/strings mágicos | literales sueltos en el código | Constante con nombre, o configuración versionada |
| Duplicación | misma lógica repetida en 2+ sitios | Extraer a función o módulo compartido |
| Feature envy | un módulo usa más los datos de otro que los propios | Mover el comportamiento donde viven esos datos |
| Componente "Dios" | mezcla UI + estado + lógica de negocio | Separar la capa de decisión de la capa de presentación |
| `any` / casts sueltos | se perdió el tipado | Tipar con interfaces o zod |
| Comentario-muleta | explica código que en realidad es confuso | Renombrar/extraer hasta que el comentario sobre |
| Código muerto | nada lo referencia | Borrarlo |

## Puntos específicos del dominio

- La capa de decisión determinista vive en su propio módulo puro: testeable, sin UI, sin IA — misma
  entrada, misma salida, explicable.
- Umbrales, pesos y bandas de decisión son configuración versionada, jamás literales repartidos por
  el código.
- La capa de servicios externos (el servicio externo/IA de la frontera que declara el consumidor)
  queda aislada tras una interfaz — el resto del código no sabe que existe.
- Toda salida de un servicio externo/IA se trata como entrada no confiable: un único esquema
  centralizado la valida antes de que llegue a la capa de decisión.

Clasifica cada hallazgo en BLOQUEANTE / RECOMENDADO / NIT. Solo lo BLOQUEANTE impide `gates.smell`.

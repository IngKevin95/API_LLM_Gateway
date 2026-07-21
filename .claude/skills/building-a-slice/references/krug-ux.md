# Usabilidad Krug ("Don't Make Me Think")

`ux-krug-reviewer` aplica estas heurísticas a slices con UI. La meta es simple: quien opera el
producto no debería tener que pararse a pensar en cómo usarlo.

## Las leyes que se revisan

1. **No me hagas pensar.** Todo debería ser autoevidente; lo que no lo sea, al menos autoexplicativo.
   Cero acertijos de interfaz.
2. **La página se escanea, no se lee.** Diseño para el vistazo rápido: jerarquía visual clara,
   encabezados que orientan, fragmentos cortos, lo importante arriba y grande.
3. **Satisficing**: la gente toma la primera opción razonable, no compara exhaustivamente. Ofrece
   caminos obvios en vez de forzar comparación.
4. **Recorta las palabras que sobran.** Y luego recórtalas otra vez — el ruido textual compite con lo
   que sí importa.
5. **Convención por encima de originalidad.** Botones que parecen botones, enlaces que parecen
   enlaces, navegación donde el usuario ya espera encontrarla.
6. **Lo clicable se ve clicable**, y toda acción da retroalimentación inmediata: hover, loading,
   disabled, foco visible.

## Aplicado a las decisiones de alto impacto del dominio

- La decisión de alto impacto (el ejemplo concreto vive en el domain-pack del consumidor) y su
  resultado/banda/clasificación se leen de un vistazo, sin scroll ni clics extra.
- Los drivers o factores detrás del resultado son visibles y comprensibles sin abrir documentación
  aparte.
- Cualquier inconsistencia o discrepancia detectada queda resaltada y atribuida a su origen real, sin
  dejar al usuario adivinando.
- Todo estado se declara explícitamente: cargando, vacío, error, sin permisos. Nunca una pantalla
  muda que deja al usuario sin saber qué pasó.
- Los errores se recuperan con un mensaje accionable — qué pasó, qué hacer — sin exponer PII/datos
  sensibles ni stack traces.

## Verificar, no opinar

Cuando la app corre, usa el MCP **chrome-devtools**:
- `take_snapshot` para inspeccionar el árbol de accesibilidad (roles, labels, orden de foco).
- `lighthouse_audit` para puntuaciones de Accessibility y Best Practices, citando fallos concretos.

Clasifica cada hallazgo como BLOQUEANTE, RECOMENDADO o NIT, indicando la pantalla afectada y el fix
sugerido.

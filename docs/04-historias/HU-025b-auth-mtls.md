---
id: HU-025b
titulo: Autenticación mTLS
epica: EP-004A
prioridad: Could
complejidad: M
estado: lista
---

# Autenticación mTLS

Como **ingeniero de seguridad**, quiero **soportar mTLS para comunicaciones servicio-a-servicio**, para **cumplir con estándares zero-trust de red**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — mTLS auth | Dado que un servicio interno con certificado cliente válido | Cuando llama a la gateway | Entonces la conexión se establece y extrae el scope del certificado |
| 2 | Error — Certificado revocado | Dado que un servicio con certificado revocado/expirado | Cuando llama a la gateway | Entonces el handshake falla y la petición es abortada |
| 3 | Error — Sin certificado | Dado que un servicio sin certificado cliente | Cuando llama al puerto seguro | Entonces el Gateway rechaza la conexión en capa TCP/TLS |
| 4 | Edge — CA no confiable | Dado que un servicio presenta un certificado cliente válido pero emitido por una CA no incluida en el trust store del Gateway | Cuando llama a la gateway | Entonces el handshake mTLS falla con `unknown certificate authority` y la petición es abortada |

## Checklist INVEST

- [x] Independent — Se resuelve a nivel terminación TLS o middleware temprano.
- [x] Negotiable — CA interna vs externa se puede decidir en la marcha.
- [x] Valuable — Permite comunicación segura servicio-a-servicio (ej. microservicios on-premise).
- [x] Estimable — TLS stack es estándar y directo de estimar.
- [x] Small — Es principalmente configuración de certificados y un parser de subject.
- [x] Testable — Probable con curl pasando --cert y --key.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.

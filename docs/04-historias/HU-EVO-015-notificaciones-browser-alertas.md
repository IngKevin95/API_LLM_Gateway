---
id: HU-EVO-015
titulo: Notificaciones browser cuando remaining < umbral configurable
epica: EP-EVO-003
prioridad: Should
complejidad: M
estado: lista
---

# Notificaciones browser cuando remaining < umbral configurable

Como **usuario del Gateway**, quiero **recibir notificaciones push/toast en el navegador cuando mi cuota está baja**, para **no perder el estado de alertas sin revisar el dashboard constantemente**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — notificación toast | Dado que remaining de un proveedor cae < 10% | Cuando GET `/alerts` lo detecta | Entonces dashboard muestra toast (esquina inferior derecha, 5s fade-out) con mensaje "Groq cuota baja: 8% restante" |
| 2 | Happy — notificación sonora | Dado que usuario tiene sonido habilitado en settings | Cuando alerta critical generada | Entonces playback sonido corto (beep) para atraer atención |
| 3 | Happy — browser notification (Notification API) | Dado que usuario permitió notifications | Cuando alerta critical se genera | Entonces `Notification.requestPermission()` y muestra notificación del sistema (fuera del navegador) |
| 4 | Edge — deduplicación | Dado que ya hay toast visible de Groq | Cuando llega otra alerta Groq < 10s después | Entonces no duplica toast, solo actualiza existing |
| 5 | Edge — configuración de umbral | Dado que usuario quiere alertar en 25% en lugar de 10% | Cuando edita settings | Entonces usa nuevo umbral sin redeploy del dashboard |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-012/014 (alerts, dashboard)
- [x] Negotiable — notificación tipo y sonido configurables
- [x] Valuable — UX proactiva, reduce polling
- [x] Estimable — toast lib (react-toastify) + Notification API
- [x] Small — 1-2 días
- [x] Testable — mock alerts, verifica toast rendering

## Notas técnicas

Notifications en `src/ui/dashboard/hooks/useAlerts.js`:

```jsx
import { toast } from 'react-toastify';

const useAlerts = () => {
  useEffect(() => {
    const pollAlerts = async () => {
      const res = await fetch('/alerts');
      const alerts = await res.json();
      
      alerts.forEach(alert => {
        if (alert.severity === 'warning') {
          toast.warning(`${alert.provider} cuota baja: ${alert.message}`, {
            autoClose: 5000,
          });
        } else if (alert.severity === 'critical') {
          toast.error(`CRÍTICO: ${alert.provider} ${alert.message}`, {
            autoClose: 10000,
          });
          playBeep();
          notifyBrowser(alert);
        }
      });
    };
    
    const interval = setInterval(pollAlerts, 30000); // Check cada 30s
    return () => clearInterval(interval);
  }, []);
};

function playBeep() {
  new Audio('data:audio/wav;base64,...').play();
}

function notifyBrowser(alert) {
  if ('Notification' in window && Notification.permission === 'granted') {
    new Notification('API LLM Gateway', { body: alert.message });
  }
}
```

---

## Relación con existentes

- Integra: Dashboard React (HU-EVO-014)
- Usa: GET `/alerts` (HU-EVO-013)
- Complementa: HU-EVO-012 (Alert Manager)

import { useEffect, useRef, useState } from 'react';
import { toast } from 'react-toastify';
import { authHeaders, getNotifyThreshold, isSoundEnabled } from '../authConfig';

const POLL_INTERVAL_MS = 30000;
const DEDUP_WINDOW_MS = 10000;

const BEEP_DATA_URI =
  'data:audio/wav;base64,UklGRiQAAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQAAAAA=';

function defaultPlayBeep() {
  try {
    const audio = new Audio(BEEP_DATA_URI);
    audio.play().catch(() => {});
  } catch {
    // entorno sin soporte de Audio (jsdom/tests) — no bloquea la UI
  }
}

function defaultNotifyBrowser(alert) {
  if (typeof window === 'undefined' || !('Notification' in window)) return;
  if (window.Notification.permission === 'granted') {
    // eslint-disable-next-line no-new
    new window.Notification('API LLM Gateway', { body: alert.message });
  }
}

function alertKey(alert) {
  return `${alert.provider}::${alert.model}`;
}

function remainingRatio(alert) {
  if (typeof alert.remaining_ratio === 'number') return alert.remaining_ratio;
  if (typeof alert.limit === 'number' && alert.limit > 0 && typeof alert.remaining === 'number') {
    return alert.remaining / alert.limit;
  }
  return null;
}

export default function useAlerts({
  intervalMs = POLL_INTERVAL_MS,
  fetchImpl = fetch,
  toastImpl = toast,
  playBeep = defaultPlayBeep,
  notifyBrowser = defaultNotifyBrowser,
  now = () => Date.now(),
} = {}) {
  const [alerts, setAlerts] = useState([]);
  const seenRef = useRef(new Map());

  useEffect(() => {
    let cancelled = false;

    const evaluate = (list) => {
      const threshold = getNotifyThreshold();
      list.forEach((alert) => {
        const ratio = remainingRatio(alert);
        if (alert.severity !== 'critical' && ratio !== null && ratio >= threshold) {
          return;
        }

        const key = alertKey(alert);
        const seen = seenRef.current.get(key);
        const timestamp = now();
        const isDuplicate = seen && timestamp - seen.lastSeenAt < DEDUP_WINDOW_MS;

        const message =
          alert.severity === 'critical'
            ? `CRÍTICO: ${alert.provider} ${alert.message}`
            : `${alert.provider} cuota baja: ${alert.message}`;

        if (isDuplicate) {
          toastImpl.update(seen.toastId, { render: message });
          seenRef.current.set(key, { toastId: seen.toastId, lastSeenAt: timestamp, notified: seen.notified });
          return;
        }

        let toastId;
        if (alert.severity === 'critical') {
          toastId = toastImpl.error(message, { autoClose: 10000 });
        } else {
          toastId = toastImpl.warning(message, { autoClose: 5000 });
        }

        const alreadyNotified = Boolean(seen && seen.notified);
        if (alert.severity === 'critical' && !alreadyNotified) {
          if (isSoundEnabled()) {
            playBeep();
          }
          notifyBrowser(alert);
        }

        seenRef.current.set(key, {
          toastId,
          lastSeenAt: timestamp,
          notified: alreadyNotified || alert.severity === 'critical',
        });
      });
    };

    const poll = async () => {
      try {
        const res = await fetchImpl('/alerts', { headers: authHeaders() });
        if (!res.ok) return;
        const payload = await res.json();
        const list = Array.isArray(payload) ? payload : payload.data || [];
        if (cancelled) return;
        setAlerts(list);
        evaluate(list);
      } catch {
        // fallo de red no bloquea el resto del dashboard
      }
    };

    poll();
    const interval = setInterval(poll, intervalMs);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [intervalMs, fetchImpl, toastImpl, playBeep, notifyBrowser, now]);

  return { alerts };
}

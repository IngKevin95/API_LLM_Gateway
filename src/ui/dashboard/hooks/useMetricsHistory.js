import { useEffect, useRef, useState } from 'react';

const MAX_POINTS = 20;

/**
 * Acumula en memoria los ultimos N snapshots de /metrics recibidos via useMetrics.
 * Decision de diseno: el backend expone solo un snapshot puntual (no series
 * historicas), asi que la serie temporal para los graficos de Overview/Providers
 * se construye del lado del cliente a partir de cada poll real - no son datos
 * inventados, son la acumulacion de lecturas reales del propio polling.
 */
export default function useMetricsHistory(metrics) {
  const [history, setHistory] = useState([]);
  const lastUptimeRef = useRef(null);

  useEffect(() => {
    if (!metrics) return;
    // evita duplicar el mismo snapshot si uptime_seconds no cambio entre polls
    if (lastUptimeRef.current === metrics.uptime_seconds) return;
    lastUptimeRef.current = metrics.uptime_seconds;

    setHistory((prev) => {
      const point = {
        t: new Date().toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit' }),
        requests: metrics.requests?.total ?? 0,
        p50: metrics.latency?.p50_ms ?? 0,
        p95: metrics.latency?.p95_ms ?? 0,
        p99: metrics.latency?.p99_ms ?? 0,
        providers: (metrics.providers || []).reduce((acc, p) => {
          acc[p.name] = p.circuit_breaker_open ? 0 : 1;
          return acc;
        }, {}),
      };
      const next = [...prev, point];
      return next.length > MAX_POINTS ? next.slice(next.length - MAX_POINTS) : next;
    });
  }, [metrics]);

  return history;
}

import { useEffect, useRef, useState } from 'react';
import { authHeaders } from '../authConfig';

const POLL_INTERVAL_MS = 5000;

export default function useMetrics({ intervalMs = POLL_INTERVAL_MS, fetchImpl = fetch } = {}) {
  const [metrics, setMetrics] = useState(null);
  const [error, setError] = useState(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;

    const load = async () => {
      try {
        const res = await fetchImpl('/metrics', { headers: authHeaders() });
        if (!res.ok) {
          throw new Error(`GET /metrics -> ${res.status}`);
        }
        const data = await res.json();
        if (mountedRef.current) {
          setMetrics(data);
          setError(null);
        }
      } catch (err) {
        if (mountedRef.current) {
          setError(err);
        }
      }
    };

    load();
    const interval = setInterval(load, intervalMs);

    return () => {
      mountedRef.current = false;
      clearInterval(interval);
    };
  }, [intervalMs, fetchImpl]);

  return { metrics, error };
}

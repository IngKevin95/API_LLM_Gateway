import { useEffect, useRef, useState } from 'react';
import { authHeaders } from '../authConfig';

// useCurrentUser resuelve el perfil del usuario autenticado consumiendo
// GET /users/me (HU-EVO-022 AC1), usado tanto por ProfileSecurity (perfil
// propio) como por Dashboard (decidir si la tab "Team" es visible, HU-EVO-021
// AC4 -- el frontend no muestra controles que el backend igual rechazaria).
export default function useCurrentUser({ fetchImpl = fetch } = {}) {
  const [user, setUser] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;

    const load = async () => {
      try {
        const res = await fetchImpl('/users/me', { headers: authHeaders() });
        if (!res.ok) {
          throw new Error(`GET /users/me -> ${res.status}`);
        }
        const data = await res.json();
        if (mountedRef.current) {
          setUser(data);
          setError(null);
        }
      } catch (err) {
        if (mountedRef.current) {
          setError(err);
        }
      } finally {
        if (mountedRef.current) {
          setLoading(false);
        }
      }
    };

    load();

    return () => {
      mountedRef.current = false;
    };
  }, [fetchImpl]);

  return { user, error, loading };
}

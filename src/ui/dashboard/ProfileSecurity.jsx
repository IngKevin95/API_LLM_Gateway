import { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import { authHeaders, getNotifyThreshold, isSoundEnabled } from './authConfig';
import useCurrentUser from './hooks/useCurrentUser';

function maskPrefix(prefix) {
  return `${prefix}••••••••`;
}

// ProfileSecurity (HU-EVO-022): tab "Profile & Security", visible para
// cualquier usuario autenticado. Consume GET /users/me (perfil), GET/POST/
// DELETE /users/:id/api-keys (HU-EVO-018), GET/DELETE /sessions (HU-EVO-019)
// y POST /auth/mfa/enroll|verify (HU-EVO-020).
export default function ProfileSecurity({ fetchImpl = fetch }) {
  const { user, loading: userLoading } = useCurrentUser({ fetchImpl });
  const [keys, setKeys] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [newKeyName, setNewKeyName] = useState('');
  const [revealedKey, setRevealedKey] = useState(null);
  const [mfaEnabled, setMfaEnabled] = useState(false);
  useEffect(() => {
    if (user) setMfaEnabled(!!user.mfa_enabled);
  }, [user]);
  const [mfaSetup, setMfaSetup] = useState(null); // { secret, uri }
  const [mfaCode, setMfaCode] = useState('');
  const [mfaError, setMfaError] = useState(null);

  const loadKeys = async (userId) => {
    if (!userId) return;
    const res = await fetchImpl(`/users/${userId}/api-keys`, { headers: authHeaders() });
    if (res.ok) {
      const data = await res.json();
      setKeys(data.data || []);
    }
  };

  const loadSessions = async () => {
    const res = await fetchImpl('/sessions', { headers: authHeaders() });
    if (res.ok) {
      const data = await res.json();
      setSessions(data.data || []);
    }
  };

  useEffect(() => {
    if (user?.id) {
      loadKeys(user.id);
    }
    loadSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.id]);

  const generateKey = async () => {
    if (!user?.id || !newKeyName.trim()) return;
    const res = await fetchImpl(`/users/${user.id}/api-keys`, {
      method: 'POST',
      headers: { ...authHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newKeyName.trim() }),
    });
    if (!res.ok) {
      toast.error(`No se pudo generar la key (HTTP ${res.status})`);
      return;
    }
    const created = await res.json();
    setRevealedKey(created.key);
    setNewKeyName('');
    await loadKeys(user.id);
  };

  const revokeKey = async (keyId) => {
    if (!user?.id) return;
    const res = await fetchImpl(`/users/${user.id}/api-keys/${keyId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!res.ok) {
      toast.error(`No se pudo revocar la key (HTTP ${res.status})`);
      return;
    }
    setKeys((prev) => prev.filter((k) => k.id !== keyId));
  };

  const closeSession = async (sessionId) => {
    const res = await fetchImpl(`/sessions/${sessionId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!res.ok) {
      toast.error(`No se pudo cerrar la sesión (HTTP ${res.status})`);
      return;
    }
    setSessions((prev) => prev.filter((s) => s.ID !== sessionId));
  };

  const startMfaEnroll = async () => {
    const res = await fetchImpl('/auth/mfa/enroll', { method: 'POST', headers: authHeaders() });
    if (!res.ok) {
      toast.error(`No se pudo iniciar 2FA (HTTP ${res.status})`);
      return;
    }
    const data = await res.json();
    setMfaSetup(data);
    setMfaError(null);
  };

  const confirmMfaEnroll = async (e) => {
    e.preventDefault();
    const res = await fetchImpl('/auth/mfa/verify', {
      method: 'POST',
      headers: { ...authHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: mfaCode }),
    });
    if (!res.ok) {
      setMfaError('Código incorrecto. Reintentá.');
      return;
    }
    setMfaEnabled(true);
    setMfaSetup(null);
    setMfaCode('');
    setMfaError(null);
    toast.success('2FA activado');
  };

  if (userLoading) {
    return <section data-testid="profile-security-tab"><p>Cargando perfil…</p></section>;
  }

  return (
    <section data-testid="profile-security-tab" aria-label="Profile & Security">
      <div className="page-heading">
        <h2>Profile &amp; Security</h2>
      </div>

      {user && (
        <div className="profile-card" data-testid="profile-card">
          <span className="provider-avatar" aria-hidden="true">{(user.email || '?').charAt(0).toUpperCase()}</span>
          <div>
            <div className="profile-email">{user.email}</div>
            <span className="badge badge-blue">{user.role}</span>
          </div>
        </div>
      )}

      {/* Notification preferences: reusa el estado de authConfig.js ya
          editable en el modal de Ajustes (HU-EVO-015) -- solo lectura acá
          para no duplicar la fuente de verdad, con nota técnica de
          HU-EVO-022 explícita en ese sentido. */}
      <div className="section-card" data-testid="notification-preferences">
        <h3>Notification Preferences</h3>
        <p className="page-subtitle">
          Umbral de alerta: {(getNotifyThreshold() * 100).toFixed(0)}% · Sonido: {isSoundEnabled() ? 'habilitado' : 'deshabilitado'}
          {' '}(editable en Ajustes ⚙)
        </p>
      </div>

      <div className="section-card">
        <h3>API Keys</h3>
        <div className="api-keys-generate">
          <input
            type="text"
            placeholder="Nombre de la key"
            data-testid="new-key-name"
            value={newKeyName}
            onChange={(e) => setNewKeyName(e.target.value)}
          />
          <button type="button" className="btn-primary" data-testid="generate-key" onClick={generateKey}>
            Generate new key
          </button>
        </div>

        {revealedKey && (
          <div role="dialog" data-testid="revealed-key-modal" aria-label="API key generada">
            <p>Copiá esta key ahora — no se volverá a mostrar:</p>
            <code data-testid="revealed-key-value">{revealedKey}</code>
            <button type="button" onClick={() => setRevealedKey(null)}>Listo, ya la copié</button>
          </div>
        )}

        <table className="mono-table" data-testid="api-keys-table">
          <thead>
            <tr>
              <th>Nombre</th>
              <th>Prefijo</th>
              <th>Scopes</th>
              <th>Creada</th>
              <th>Último uso</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {keys.map((k) => (
              <tr key={k.id} data-testid={`api-key-row-${k.id}`}>
                <td>{k.name}</td>
                <td>{maskPrefix(k.prefix)}</td>
                <td>{(k.scopes || []).join(', ') || '—'}</td>
                <td>{k.created_at}</td>
                <td>{k.last_used_at || 'nunca'}</td>
                <td>
                  <button type="button" data-testid={`revoke-key-${k.id}`} onClick={() => revokeKey(k.id)}>
                    Revoke
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section-card">
        <h3>Sesiones activas</h3>
        <table className="mono-table" data-testid="sessions-table">
          <thead>
            <tr>
              <th>User Agent</th>
              <th>IP</th>
              <th>Última actividad</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {sessions.map((s) => (
              <tr key={s.ID} data-testid={`session-row-${s.ID}`}>
                <td>{s.UserAgent}</td>
                <td>{s.IP}</td>
                <td>{s.LastUsedAt}</td>
                <td>
                  <button type="button" data-testid={`close-session-${s.ID}`} onClick={() => closeSession(s.ID)}>
                    Cerrar sesión
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section-card">
        <h3>Autenticación de dos factores</h3>
        {!mfaEnabled && !mfaSetup && (
          <button type="button" className="btn-primary" data-testid="mfa-start" onClick={startMfaEnroll}>
            Activar 2FA
          </button>
        )}
        {mfaSetup && (
          <form onSubmit={confirmMfaEnroll} data-testid="mfa-verify-form">
            <p>Escaneá con tu app TOTP (pegá esta URI) o ingresá el secret manualmente, y confirmá el código:</p>
            {/* ponytail: sin lib de QR nueva -- se muestra la otpauth:// URI
                como texto copiable, la mayoria de apps TOTP aceptan pegarla
                directo. Agregar renderizado de QR real si se pide. */}
            <code data-testid="mfa-uri">{mfaSetup.uri}</code>
            <code data-testid="mfa-secret">{mfaSetup.secret}</code>
            <input
              type="text"
              placeholder="Código de 6 dígitos"
              data-testid="mfa-code"
              value={mfaCode}
              onChange={(e) => setMfaCode(e.target.value)}
            />
            {mfaError && <p role="alert" data-testid="mfa-error">{mfaError}</p>}
            <button type="submit" data-testid="mfa-confirm">Confirmar</button>
          </form>
        )}
        {mfaEnabled && <span className="badge badge-green" data-testid="mfa-active-badge">2FA activo</span>}
      </div>
    </section>
  );
}

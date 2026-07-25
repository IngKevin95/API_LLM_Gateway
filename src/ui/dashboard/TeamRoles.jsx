import { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import { authHeaders } from './authConfig';

const ROLE_OPTIONS = ['admin', 'operator'];

function initial(name) {
  return (name || '?').charAt(0).toUpperCase();
}

function statusBadgeClass(status) {
  if (status === 'active') return 'badge-green';
  if (status === 'invited') return 'badge-yellow';
  return 'badge-red'; // suspended
}

// TeamRoles (HU-EVO-021): tab "Team", admin-only. Consume GET/POST/PATCH
// /users (HU-EVO-017) — la visibilidad de la tab en si la decide Dashboard
// (AC4), no este componente.
export default function TeamRoles({ fetchImpl = fetch }) {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [form, setForm] = useState({ email: '', role: 'operator', scopes: '' });
  const [formError, setFormError] = useState(null);
  const [submitting, setSubmitting] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const res = await fetchImpl('/users', { headers: authHeaders() });
      if (!res.ok) throw new Error(`GET /users -> ${res.status}`);
      const data = await res.json();
      setUsers(data.data || []);
      setError(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const submitInvite = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setFormError(null);
    try {
      const res = await fetchImpl('/users', {
        method: 'POST',
        headers: { ...authHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: form.email,
          role: form.role,
          scopes: form.scopes
            .split(',')
            .map((s) => s.trim())
            .filter(Boolean),
        }),
      });
      if (res.status === 409) {
        setFormError('Ya existe un usuario con ese email.');
        return;
      }
      if (!res.ok) {
        setFormError(`No se pudo invitar (HTTP ${res.status}).`);
        return;
      }
      const created = await res.json();
      setUsers((prev) => [...prev, created]);
      toast.success(`Invitación enviada a ${created.email}`);
      setInviteOpen(false);
      setForm({ email: '', role: 'operator', scopes: '' });
    } catch (err) {
      setFormError('Error de red al invitar. Reintentá.');
    } finally {
      setSubmitting(false);
    }
  };

  const updateUser = async (id, patch) => {
    try {
      const res = await fetchImpl(`/users/${id}`, {
        method: 'PATCH',
        headers: { ...authHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify(patch),
      });
      if (!res.ok) {
        toast.error(`No se pudo actualizar el usuario (HTTP ${res.status})`);
        return;
      }
      const updated = await res.json();
      setUsers((prev) => prev.map((u) => (u.id === id ? updated : u)));
    } catch {
      toast.error('Error de red al actualizar el usuario');
    }
  };

  return (
    <section data-testid="team-roles-tab" aria-label="Team & Roles">
      <div className="page-heading">
        <h2>Team &amp; Roles</h2>
        <p className="page-subtitle">Gestión de usuarios, roles y acceso RBAC</p>
        <button
          type="button"
          className="btn-primary"
          data-testid="invite-member-open"
          onClick={() => setInviteOpen(true)}
        >
          Invite member
        </button>
      </div>

      {loading && <p data-testid="team-loading">Cargando equipo…</p>}
      {error && <p data-testid="team-error" role="alert">No se pudo cargar el equipo.</p>}

      {!loading && !error && (
        <>
          {/* ponytail: cards derivadas de GET /users ya cargado, sin
              endpoint nuevo -- "Security Alerts"/"Cluster Quota" del mockup
              Stitch no tienen fuente de dato real en el dominio actual, se
              omiten en vez de inventar numeros. */}
          <div className="stat-grid stat-grid-compact">
            <div className="stat-card">
              <span className="stat-label">Total Users</span>
              <span className="stat-value mono">{users.length}</span>
            </div>
            <div className="stat-card">
              <span className="stat-label">Admins</span>
              <span className="stat-value mono">{users.filter((u) => u.role === 'admin').length}</span>
            </div>
            <div className="stat-card">
              <span className="stat-label">Active</span>
              <span className="stat-value mono">{users.filter((u) => u.status === 'active').length}</span>
            </div>
          </div>
          <table className="mono-table" data-testid="team-table">
          <thead>
            <tr>
              <th>Usuario</th>
              <th>Rol</th>
              <th>Scopes / Tenant</th>
              <th>Estado</th>
              <th>Última actividad</th>
              <th>Acciones</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} data-testid={`team-row-${u.id}`}>
                <td>
                  <span className="provider-avatar" aria-hidden="true">{initial(u.email)}</span>
                  {u.email}
                </td>
                <td>
                  <span className={`badge ${u.role === 'admin' ? 'badge-blue' : 'badge-gray'}`}>{u.role}</span>
                </td>
                <td>{(u.scopes || []).join(', ') || u.tenant}</td>
                <td>
                  <span className={`badge ${statusBadgeClass(u.status)}`}>{u.status}</span>
                </td>
                <td>{u.updated_at || '—'}</td>
                <td>
                  <select
                    aria-label={`Cambiar rol de ${u.email}`}
                    value={u.role}
                    data-testid={`team-role-select-${u.id}`}
                    onChange={(e) => updateUser(u.id, { role: e.target.value })}
                  >
                    {ROLE_OPTIONS.map((r) => (
                      <option key={r} value={r}>{r}</option>
                    ))}
                  </select>
                  {u.status !== 'suspended' ? (
                    <button
                      type="button"
                      data-testid={`team-suspend-${u.id}`}
                      onClick={() => updateUser(u.id, { status: 'suspended' })}
                    >
                      Suspend
                    </button>
                  ) : (
                    <button
                      type="button"
                      data-testid={`team-activate-${u.id}`}
                      onClick={() => updateUser(u.id, { status: 'active' })}
                    >
                      Activate
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </>
      )}

      {inviteOpen && (
        <div className="modal-backdrop" role="dialog" aria-label="Invite member" data-testid="invite-modal">
          <form className="modal" onSubmit={submitInvite}>
            <h3>Invite member</h3>
            <label>
              Email
              <input
                type="email"
                required
                data-testid="invite-email"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
              />
            </label>
            <label>
              Rol
              <select
                data-testid="invite-role"
                value={form.role}
                onChange={(e) => setForm({ ...form, role: e.target.value })}
              >
                {ROLE_OPTIONS.map((r) => (
                  <option key={r} value={r}>{r}</option>
                ))}
              </select>
            </label>
            <label>
              Scopes (coma-separados)
              <input
                type="text"
                data-testid="invite-scopes"
                value={form.scopes}
                onChange={(e) => setForm({ ...form, scopes: e.target.value })}
              />
            </label>
            {formError && <p role="alert" data-testid="invite-error">{formError}</p>}
            <div className="modal-actions">
              <button type="button" onClick={() => setInviteOpen(false)}>Cancel</button>
              <button type="submit" disabled={submitting} data-testid="invite-submit">
                {submitting ? 'Invitando…' : 'Invite'}
              </button>
            </div>
          </form>
        </div>
      )}
    </section>
  );
}

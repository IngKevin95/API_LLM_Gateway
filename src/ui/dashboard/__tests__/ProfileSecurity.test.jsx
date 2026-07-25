import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import ProfileSecurity from '../ProfileSecurity';

const mePayload = { id: '1', email: 'me@t1.com', role: 'operator', tenant: 't1' };
const keysPayload = {
  data: [{ id: 'k1', name: 'ci-key', prefix: 'sk_ab', created_at: '2026-07-24', last_used_at: null }],
};
const sessionsPayload = {
  data: [{ ID: 's1', UserAgent: 'jsdom', IP: '127.0.0.1', LastUsedAt: '2026-07-24' }],
};

function baseFetch(overrides = {}) {
  return vi.fn((url, opts) => {
    if (url === '/users/me') return Promise.resolve({ ok: true, status: 200, json: async () => mePayload });
    if (url === '/users/1/api-keys' && (!opts || opts.method === undefined)) {
      return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
    }
    if (url === '/sessions' && (!opts || opts.method === undefined)) {
      return Promise.resolve({ ok: true, status: 200, json: async () => sessionsPayload });
    }
    if (overrides[url]) return overrides[url](opts);
    return Promise.resolve({ ok: true, status: 200, json: async () => ({ data: [] }) });
  });
}

function flush() {
  return act(async () => {
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe('ProfileSecurity (HU-EVO-022)', () => {
  afterEach(() => vi.restoreAllMocks());

  it('AC1: muestra perfil y tabla de API keys desde GET /users/:id/api-keys', async () => {
    render(<ProfileSecurity fetchImpl={baseFetch()} />);
    await flush();

    expect(screen.getByTestId('profile-card')).toHaveTextContent('me@t1.com');
    expect(screen.getByText('ci-key')).toBeInTheDocument();
  });

  it('AC2: generar key muestra el valor en claro una sola vez', async () => {
    const fetchImpl = baseFetch({
      '/users/1/api-keys': (opts) => {
        if (opts?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            status: 201,
            json: async () => ({ id: 'k2', name: 'nueva', key: 'sk_live_PLAINTEXT', prefix: 'sk_pl' }),
          });
        }
        return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
      },
    });
    render(<ProfileSecurity fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.change(screen.getByTestId('new-key-name'), { target: { value: 'nueva' } });
    fireEvent.click(screen.getByTestId('generate-key'));

    await waitFor(() => expect(screen.getByTestId('revealed-key-modal')).toBeInTheDocument());
    expect(screen.getByTestId('revealed-key-value')).toHaveTextContent('sk_live_PLAINTEXT');
  });

  it('AC3: revocar key y cerrar sesión quitan la fila sin reload', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (url === '/users/me') return Promise.resolve({ ok: true, status: 200, json: async () => mePayload });
      if (url === '/users/1/api-keys') return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
      if (url === '/users/1/api-keys/k1' && opts?.method === 'DELETE') {
        return Promise.resolve({ ok: true, status: 204, json: async () => ({}) });
      }
      if (url === '/sessions') return Promise.resolve({ ok: true, status: 200, json: async () => sessionsPayload });
      if (url === '/sessions/s1' && opts?.method === 'DELETE') {
        return Promise.resolve({ ok: true, status: 204, json: async () => ({}) });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ data: [] }) });
    });
    render(<ProfileSecurity fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.click(screen.getByTestId('revoke-key-k1'));
    await waitFor(() => expect(screen.queryByTestId('api-key-row-k1')).not.toBeInTheDocument());

    fireEvent.click(screen.getByTestId('close-session-s1'));
    await waitFor(() => expect(screen.queryByTestId('session-row-s1')).not.toBeInTheDocument());
  });

  it('AC4: activar 2FA llama enroll y luego verify', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (url === '/users/me') return Promise.resolve({ ok: true, status: 200, json: async () => mePayload });
      if (url === '/users/1/api-keys') return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
      if (url === '/sessions') return Promise.resolve({ ok: true, status: 200, json: async () => sessionsPayload });
      if (url === '/auth/mfa/enroll') {
        return Promise.resolve({ ok: true, status: 200, json: async () => ({ secret: 'SECRET123', uri: 'otpauth://x' }) });
      }
      if (url === '/auth/mfa/verify') {
        return Promise.resolve({ ok: true, status: 200, json: async () => ({}) });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ data: [] }) });
    });
    render(<ProfileSecurity fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.click(screen.getByTestId('mfa-start'));
    await waitFor(() => expect(screen.getByTestId('mfa-secret')).toBeInTheDocument());

    fireEvent.change(screen.getByTestId('mfa-code'), { target: { value: '123456' } });
    fireEvent.click(screen.getByTestId('mfa-confirm'));

    await waitFor(() => expect(screen.getByTestId('mfa-active-badge')).toBeInTheDocument());
  });

  it('AC4: 2FA ya activo en el backend se refleja al montar (sin re-enrolar)', async () => {
    const fetchImpl = vi.fn((url) => {
      if (url === '/users/me') {
        return Promise.resolve({ ok: true, status: 200, json: async () => ({ ...mePayload, mfa_enabled: true }) });
      }
      if (url === '/users/1/api-keys') return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
      if (url === '/sessions') return Promise.resolve({ ok: true, status: 200, json: async () => sessionsPayload });
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ data: [] }) });
    });
    render(<ProfileSecurity fetchImpl={fetchImpl} />);
    await flush();

    expect(screen.getByTestId('mfa-active-badge')).toBeInTheDocument();
    expect(screen.queryByTestId('mfa-start')).not.toBeInTheDocument();
  });

  it('AC5: código 2FA incorrecto muestra error sin activar el toggle', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (url === '/users/me') return Promise.resolve({ ok: true, status: 200, json: async () => mePayload });
      if (url === '/users/1/api-keys') return Promise.resolve({ ok: true, status: 200, json: async () => keysPayload });
      if (url === '/sessions') return Promise.resolve({ ok: true, status: 200, json: async () => sessionsPayload });
      if (url === '/auth/mfa/enroll') {
        return Promise.resolve({ ok: true, status: 200, json: async () => ({ secret: 'SECRET123', uri: 'otpauth://x' }) });
      }
      if (url === '/auth/mfa/verify') {
        return Promise.resolve({ ok: false, status: 400, json: async () => ({ error: 'invalid code' }) });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ data: [] }) });
    });
    render(<ProfileSecurity fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.click(screen.getByTestId('mfa-start'));
    await waitFor(() => expect(screen.getByTestId('mfa-secret')).toBeInTheDocument());

    fireEvent.change(screen.getByTestId('mfa-code'), { target: { value: '000000' } });
    fireEvent.click(screen.getByTestId('mfa-confirm'));

    await waitFor(() => expect(screen.getByTestId('mfa-error')).toBeInTheDocument());
    expect(screen.queryByTestId('mfa-active-badge')).not.toBeInTheDocument();
  });
});

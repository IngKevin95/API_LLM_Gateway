import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import TeamRoles from '../TeamRoles';

const usersPayload = {
  data: [
    { id: '1', email: 'admin@t1.com', role: 'admin', status: 'active', tenant: 't1', scopes: [], updated_at: '2026-07-24' },
    { id: '2', email: 'op@t1.com', role: 'operator', status: 'invited', tenant: 't1', scopes: ['capability:chat'], updated_at: '2026-07-24' },
  ],
};

function flush() {
  return act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe('TeamRoles (HU-EVO-021)', () => {
  afterEach(() => vi.restoreAllMocks());

  it('AC1: renderiza la tabla de equipo desde GET /users', async () => {
    const fetchImpl = vi.fn(() => Promise.resolve({ ok: true, status: 200, json: async () => usersPayload }));
    render(<TeamRoles fetchImpl={fetchImpl} />);
    await flush();

    expect(fetchImpl).toHaveBeenCalledWith('/users', expect.anything());
    expect(screen.getByText('admin@t1.com')).toBeInTheDocument();
    expect(screen.getByText('op@t1.com')).toBeInTheDocument();
    expect(screen.getByTestId('team-table')).toBeInTheDocument();
  });

  it('AC2: invitar miembro llama POST /users y agrega la fila', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (opts?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          status: 201,
          json: async () => ({ id: '3', email: 'new@t1.com', role: 'operator', status: 'invited', tenant: 't1', scopes: [] }),
        });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => usersPayload });
    });
    render(<TeamRoles fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.click(screen.getByTestId('invite-member-open'));
    fireEvent.change(screen.getByTestId('invite-email'), { target: { value: 'new@t1.com' } });
    fireEvent.click(screen.getByTestId('invite-submit'));

    await waitFor(() => expect(screen.getByText('new@t1.com')).toBeInTheDocument());
    expect(fetchImpl).toHaveBeenCalledWith('/users', expect.objectContaining({ method: 'POST' }));
  });

  it('AC3: cambiar rol llama PATCH /users/:id', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (opts?.method === 'PATCH') {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ ...usersPayload.data[1], role: 'admin' }),
        });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => usersPayload });
    });
    render(<TeamRoles fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.change(screen.getByTestId('team-role-select-2'), { target: { value: 'admin' } });

    await waitFor(() =>
      expect(fetchImpl).toHaveBeenCalledWith('/users/2', expect.objectContaining({ method: 'PATCH' }))
    );
  });

  it('AC5: invitar email duplicado muestra error sin perder el formulario', async () => {
    const fetchImpl = vi.fn((url, opts) => {
      if (opts?.method === 'POST') {
        return Promise.resolve({ ok: false, status: 409, json: async () => ({ error: 'email already exists' }) });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => usersPayload });
    });
    render(<TeamRoles fetchImpl={fetchImpl} />);
    await flush();

    fireEvent.click(screen.getByTestId('invite-member-open'));
    fireEvent.change(screen.getByTestId('invite-email'), { target: { value: 'admin@t1.com' } });
    fireEvent.click(screen.getByTestId('invite-submit'));

    await waitFor(() => expect(screen.getByTestId('invite-error')).toBeInTheDocument());
    // El formulario sigue abierto con los datos tipeados, no se pierde
    expect(screen.getByTestId('invite-email').value).toBe('admin@t1.com');
  });
});

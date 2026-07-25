import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import Dashboard from '../Dashboard';

const metricsPayload = {
  uptime_seconds: 36000,
  requests: { total: 120, by_handler: {}, errors: 2 },
  latency: { p50_ms: 80, p95_ms: 200, p99_ms: 400 },
  quota: [
    { provider: 'groq', model: 'mixtral', limit: 1000, remaining: 900, reset_at: null, healthy: true },
  ],
  providers: [{ name: 'groq', available: true, last_success: '', circuit_breaker_open: false }],
};

const alertsPayload = {
  data: [
    { provider: 'groq', model: 'mixtral', severity: 'critical', message: 'agotado', remaining: 0, limit: 1000 },
  ],
};

function stubFetch(meRole) {
  return vi.fn((url) => {
    if (url === '/metrics') {
      return Promise.resolve({ ok: true, status: 200, json: async () => metricsPayload });
    }
    if (url === '/alerts') {
      return Promise.resolve({ ok: true, status: 200, json: async () => alertsPayload });
    }
    if (url === '/users/me' && meRole) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ id: '1', email: 'u@t1.com', role: meRole }),
      });
    }
    return Promise.resolve({ ok: false, status: 404, json: async () => ({}) });
  });
}

describe('Dashboard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    window.localStorage.clear();
    global.fetch = stubFetch();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('AC1: renderiza 4 tabs interactivos al cargar', async () => {
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(screen.getByTestId('tab-overview')).toBeInTheDocument();
    expect(screen.getByTestId('tab-quotas')).toBeInTheDocument();
    expect(screen.getByTestId('tab-alerts')).toBeInTheDocument();
    expect(screen.getByTestId('tab-providers')).toBeInTheDocument();

    fireEvent.click(screen.getByTestId('tab-quotas'));
    expect(screen.getByTestId('quotas-tab')).toBeInTheDocument();
  });

  it('AC3: tab Quotas muestra tabla poblada desde /metrics', async () => {
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    fireEvent.click(screen.getByTestId('tab-quotas'));

    expect(screen.getByText('groq')).toBeInTheDocument();
    expect(screen.getByText('mixtral')).toBeInTheDocument();
  });

  it('AC4: tab Alerts muestra alerta critical resaltada', async () => {
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    fireEvent.click(screen.getByTestId('tab-alerts'));

    expect(screen.getByText('agotado')).toBeInTheDocument();
    expect(screen.getByText('critical')).toBeInTheDocument();
  });

  it('AC5: tab Providers muestra estado del proveedor', async () => {
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    fireEvent.click(screen.getByTestId('tab-providers'));

    expect(screen.getByText('groq', { selector: '.provider-name' })).toBeInTheDocument();
    expect(screen.getByTestId('providers-tab').textContent).toContain('closed');
  });

  it('HU-EVO-021 AC4: rol operator no ve la tab Team, pero si Profile & Security', async () => {
    global.fetch = stubFetch('operator');
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(screen.queryByTestId('tab-team')).not.toBeInTheDocument();
    expect(screen.getByTestId('tab-profile-security')).toBeInTheDocument();
  });

  it('HU-EVO-021: rol admin si ve la tab Team', async () => {
    global.fetch = stubFetch('admin');
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(screen.getByTestId('tab-team')).toBeInTheDocument();
  });

  it('abre la modal de settings', async () => {
    render(<Dashboard />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    fireEvent.click(screen.getByTestId('open-settings'));
    expect(screen.getByTestId('settings-modal')).toBeInTheDocument();
  });
});

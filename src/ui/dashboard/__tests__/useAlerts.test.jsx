import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import useAlerts from '../hooks/useAlerts';
import { setNotifyThreshold, setSoundEnabled } from '../authConfig';

function makeFetch(payload) {
  return vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => payload });
}

function makeToast() {
  let idCounter = 0;
  return {
    warning: vi.fn(() => `toast-${idCounter++}`),
    error: vi.fn(() => `toast-${idCounter++}`),
    update: vi.fn(),
  };
}

describe('useAlerts', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    window.localStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('AC1: muestra toast warning cuando remaining cae bajo el umbral', async () => {
    setNotifyThreshold(0.10);
    const alert = { provider: 'groq', model: 'mixtral', severity: 'warning', message: '8% restante', remaining: 8, limit: 100 };
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();

    renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep: vi.fn(), notifyBrowser: vi.fn() }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(toastImpl.warning).toHaveBeenCalledTimes(1);
    expect(toastImpl.warning.mock.calls[0][0]).toContain('groq');
  });

  it('AC2: reproduce sonido en alerta critical cuando el sonido esta habilitado', async () => {
    setSoundEnabled(true);
    const alert = { provider: 'groq', model: 'mixtral', severity: 'critical', message: 'agotado', remaining: 0, limit: 100 };
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();
    const playBeep = vi.fn();

    renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep, notifyBrowser: vi.fn() }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(playBeep).toHaveBeenCalledTimes(1);
  });

  it('AC2b: no reproduce sonido si el usuario lo deshabilito', async () => {
    setSoundEnabled(false);
    const alert = { provider: 'groq', model: 'mixtral', severity: 'critical', message: 'agotado', remaining: 0, limit: 100 };
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();
    const playBeep = vi.fn();

    renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep, notifyBrowser: vi.fn() }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(playBeep).not.toHaveBeenCalled();
  });

  it('AC3: dispara notifyBrowser (Notification API) en alerta critical', async () => {
    const alert = { provider: 'groq', model: 'mixtral', severity: 'critical', message: 'agotado', remaining: 0, limit: 100 };
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();
    const notifyBrowser = vi.fn();

    renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep: vi.fn(), notifyBrowser }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(notifyBrowser).toHaveBeenCalledWith(alert);
  });

  it('AC4: deduplica toast de la misma alerta dentro de la ventana de 10s (actualiza en vez de crear otro)', async () => {
    setNotifyThreshold(0.10);
    const alert = { provider: 'groq', model: 'mixtral', severity: 'warning', message: '8% restante', remaining: 8, limit: 100 };
    let tick = 0;
    const now = () => tick;
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();

    renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep: vi.fn(), notifyBrowser: vi.fn(), now, intervalMs: 30000 }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(toastImpl.warning).toHaveBeenCalledTimes(1);

    tick = 5000;
    await act(async () => {
      await vi.advanceTimersByTimeAsync(30000);
    });

    expect(toastImpl.warning).toHaveBeenCalledTimes(1);
    expect(toastImpl.update.mock.calls.length).toBeGreaterThanOrEqual(1);
    toastImpl.update.mock.calls.forEach((call) => expect(call[0]).toBe('toast-0'));
  });

  it('AC5: umbral configurable en localStorage cambia si se notifica o no', async () => {
    const alert = { provider: 'groq', model: 'mixtral', severity: 'warning', message: '20% restante', remaining: 20, limit: 100 };
    const fetchImpl = makeFetch({ data: [alert] });
    const toastImpl = makeToast();

    setNotifyThreshold(0.10);
    const { unmount } = renderHook(() => useAlerts({ fetchImpl, toastImpl, playBeep: vi.fn(), notifyBrowser: vi.fn() }));
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(toastImpl.warning).not.toHaveBeenCalled();
    unmount();

    setNotifyThreshold(0.25);
    const toastImpl2 = makeToast();
    renderHook(() => useAlerts({ fetchImpl, toastImpl: toastImpl2, playBeep: vi.fn(), notifyBrowser: vi.fn() }));
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(toastImpl2.warning).toHaveBeenCalledTimes(1);
  });
});

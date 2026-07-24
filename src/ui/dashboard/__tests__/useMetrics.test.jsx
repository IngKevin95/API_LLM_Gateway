import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import useMetrics from '../hooks/useMetrics';

const samplePayload = {
  uptime_seconds: 36000,
  total_requests: 120,
  errors: 2,
  latency: { p50: 80, p95: 200, p99: 400 },
  quota: [
    { provider: 'groq', model: 'mixtral', limit: 1000, remaining: 900, reset_at: null, healthy: true },
  ],
  providers: [{ provider: 'groq', healthy: true, last_response_at: null, circuit_breaker: 'closed' }],
};

function makeFetch(payload) {
  return vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => payload,
  });
}

describe('useMetrics', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('carga /metrics al montar (AC2/AC3/AC5 HU-EVO-014)', async () => {
    const fetchImpl = makeFetch(samplePayload);
    const { result } = renderHook(() => useMetrics({ fetchImpl }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(fetchImpl).toHaveBeenCalledWith('/metrics', { headers: {} });
    expect(result.current.metrics.quota[0].provider).toBe('groq');
  });

  it('refresca automaticamente cada 5s sin intervencion del usuario (AC6 HU-EVO-014)', async () => {
    const fetchImpl = makeFetch(samplePayload);
    renderHook(() => useMetrics({ fetchImpl }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(fetchImpl).toHaveBeenCalledTimes(1);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });
    expect(fetchImpl).toHaveBeenCalledTimes(2);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });
    expect(fetchImpl).toHaveBeenCalledTimes(3);
  });

  it('limpia el interval al desmontar (sin memory leak)', async () => {
    const fetchImpl = makeFetch(samplePayload);
    const { unmount } = renderHook(() => useMetrics({ fetchImpl }));
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(fetchImpl).toHaveBeenCalledTimes(1);

    unmount();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(20000);
    });
    expect(fetchImpl).toHaveBeenCalledTimes(1);
  });

  it('expone error si /metrics falla', async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: false, status: 401, json: async () => ({}) });
    const { result } = renderHook(() => useMetrics({ fetchImpl }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(result.current.error).not.toBeNull();
  });
});

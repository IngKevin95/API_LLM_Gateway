import { describe, it, expect, beforeEach } from 'vitest';
import {
  getToken,
  setToken,
  authHeaders,
  getNotifyThreshold,
  setNotifyThreshold,
  isSoundEnabled,
  setSoundEnabled,
} from '../authConfig';

describe('authConfig', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('devuelve string vacio si no hay token guardado', () => {
    expect(getToken()).toBe('');
    expect(authHeaders()).toEqual({});
  });

  it('persiste y recupera el token, y arma el header Authorization', () => {
    setToken('abc123');
    expect(getToken()).toBe('abc123');
    expect(authHeaders()).toEqual({ Authorization: 'Bearer abc123' });
  });

  it('umbral de notificacion tiene default 10% y es configurable sin redeploy (AC5 HU-EVO-015)', () => {
    expect(getNotifyThreshold()).toBe(0.10);
    setNotifyThreshold(0.25);
    expect(getNotifyThreshold()).toBe(0.25);
  });

  it('sonido habilitado por default y configurable', () => {
    expect(isSoundEnabled()).toBe(true);
    setSoundEnabled(false);
    expect(isSoundEnabled()).toBe(false);
  });
});

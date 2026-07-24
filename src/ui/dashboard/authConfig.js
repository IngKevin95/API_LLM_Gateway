const TOKEN_KEY = 'gateway_dashboard_token';
const THRESHOLD_KEY = 'gateway_dashboard_notify_threshold';
const SOUND_KEY = 'gateway_dashboard_notify_sound';
const DEFAULT_THRESHOLD = 0.10;

export function getToken() {
  try {
    return window.localStorage.getItem(TOKEN_KEY) || '';
  } catch {
    return '';
  }
}

export function setToken(token) {
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function authHeaders() {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export function getNotifyThreshold() {
  try {
    const raw = window.localStorage.getItem(THRESHOLD_KEY);
    const parsed = raw === null ? DEFAULT_THRESHOLD : Number(raw);
    return Number.isFinite(parsed) ? parsed : DEFAULT_THRESHOLD;
  } catch {
    return DEFAULT_THRESHOLD;
  }
}

export function setNotifyThreshold(threshold) {
  window.localStorage.setItem(THRESHOLD_KEY, String(threshold));
}

export function isSoundEnabled() {
  try {
    return window.localStorage.getItem(SOUND_KEY) !== 'false';
  } catch {
    return true;
  }
}

export function setSoundEnabled(enabled) {
  window.localStorage.setItem(SOUND_KEY, String(enabled));
}

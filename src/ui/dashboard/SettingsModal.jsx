import { useState } from 'react';
import {
  getToken,
  setToken,
  getNotifyThreshold,
  setNotifyThreshold,
  isSoundEnabled,
  setSoundEnabled,
} from './authConfig';

export default function SettingsModal({ onClose }) {
  const [token, setTokenValue] = useState(getToken());
  const [threshold, setThresholdValue] = useState(Math.round(getNotifyThreshold() * 100));
  const [sound, setSoundValue] = useState(isSoundEnabled());
  const [permission, setPermission] = useState(
    typeof window !== 'undefined' && 'Notification' in window ? window.Notification.permission : 'unsupported',
  );

  const save = () => {
    setToken(token.trim());
    setNotifyThreshold(Math.max(0, Math.min(100, Number(threshold) || 0)) / 100);
    setSoundEnabled(sound);
    onClose();
  };

  const requestPermission = async () => {
    if (typeof window === 'undefined' || !('Notification' in window)) return;
    const result = await window.Notification.requestPermission();
    setPermission(result);
  };

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true" data-testid="settings-modal">
      <div className="modal">
        <h2>Configuracion</h2>

        <label>
          API key / Bearer token
          <input
            data-testid="settings-token"
            type="password"
            value={token}
            onChange={(e) => setTokenValue(e.target.value)}
          />
        </label>

        <label>
          Umbral de notificacion (%)
          <input
            data-testid="settings-threshold"
            type="number"
            min="0"
            max="100"
            value={threshold}
            onChange={(e) => setThresholdValue(e.target.value)}
          />
        </label>

        <label>
          <input
            data-testid="settings-sound"
            type="checkbox"
            checked={sound}
            onChange={(e) => setSoundValue(e.target.checked)}
          />
          Sonido habilitado
        </label>

        <div>
          <button type="button" onClick={requestPermission} data-testid="settings-request-permission">
            Permitir notificaciones del sistema ({permission})
          </button>
        </div>

        <div className="modal-actions">
          <button type="button" onClick={save} data-testid="settings-save">
            Guardar
          </button>
          <button type="button" onClick={onClose} data-testid="settings-cancel">
            Cancelar
          </button>
        </div>
      </div>
    </div>
  );
}

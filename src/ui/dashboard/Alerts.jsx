import { useState } from 'react';
import { isSoundEnabled, setSoundEnabled } from './authConfig';

export default function Alerts({ alerts }) {
  const list = alerts || [];
  const [filter, setFilter] = useState('');
  const [soundOn, setSoundOn] = useState(isSoundEnabled());

  const criticalCount = list.filter((a) => a.severity === 'critical').length;
  const warningCount = list.filter((a) => a.severity !== 'critical').length;

  const filtered = filter.trim()
    ? list.filter((a) => (a.provider || '').toLowerCase().includes(filter.trim().toLowerCase()))
    : list;

  const toggleSound = () => {
    const next = !soundOn;
    setSoundOn(next);
    setSoundEnabled(next);
  };

  return (
    <section data-testid="alerts-tab" aria-label="Alerts">
      <div className="page-heading-row">
        <div className="page-heading">
          <h2>Active Alerts</h2>
          <p className="page-subtitle">Real-time incident monitoring for the gateway</p>
        </div>
        <label className="inline-toggle" data-testid="alerts-mute-toggle">
          <input type="checkbox" checked={!soundOn} onChange={toggleSound} />
          Mute sound notifications
        </label>
      </div>

      <div className="alert-summary" data-testid="alert-summary">
        <div className="summary-card summary-card-red">
          <span className="summary-value">{criticalCount}</span>
          <span className="summary-label">criticas</span>
        </div>
        <div className="summary-card summary-card-yellow">
          <span className="summary-value">{warningCount}</span>
          <span className="summary-label">warnings totales</span>
        </div>
      </div>

      <div className="alert-toolbar">
        <input
          type="text"
          name="alert-provider-filter"
          id="alert-provider-filter"
          className="alert-filter-input"
          placeholder="Filter by provider..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          data-testid="alert-filter-input"
          aria-label="Filtrar alertas por proveedor"
        />
      </div>

      <table className="mono-table" aria-label="Lista de alertas">
        <thead>
          <tr>
            <th>Severity</th>
            <th>Provider</th>
            <th>Model</th>
            <th>Message</th>
            <th>AlertTime</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((alert, idx) => (
            <tr
              key={`${alert.provider}-${alert.model}-${idx}`}
              className={alert.severity === 'critical' ? 'row-critical' : 'row-warning'}
            >
              <td>
                <span className={`badge ${alert.severity === 'critical' ? 'badge-red' : 'badge-yellow'}`}>
                  {alert.severity}
                </span>
              </td>
              <td>{alert.provider}</td>
              <td>{alert.model}</td>
              <td>{alert.message}</td>
              <td className="mono">{alert.alert_time ?? '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

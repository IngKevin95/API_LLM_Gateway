function healthBadge(healthy) {
  return (
    <span className={`badge badge-status ${healthy ? 'badge-green' : 'badge-red'}`}>
      {healthy ? 'healthy' : 'unhealthy'}
    </span>
  );
}

function remainingBar(remaining, limit) {
  const pct = limit > 0 ? Math.max(0, Math.min(100, (remaining / limit) * 100)) : 0;
  const level = pct < 15 ? 'bar-red' : pct < 40 ? 'bar-yellow' : 'bar-green';
  return (
    <div className="remaining-cell">
      <span className="mono">{remaining}</span>
      <div className="progress-track" role="progressbar" aria-valuenow={Math.round(pct)} aria-valuemin={0} aria-valuemax={100}>
        <div className={`progress-fill ${level}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

export default function Quotas({ metrics }) {
  const quota = metrics?.quota || [];
  const providers = metrics?.providers || [];

  const activeProviders = providers.filter((p) => p.available).length;
  const avgLatency = metrics?.latency?.p50_ms ?? 0;
  // "Cost Burn" del mockup se omite: el backend no expone datos de costo,
  // no se inventa una cifra en dinero.

  return (
    <section data-testid="quotas-tab" aria-label="Quotas">
      <div className="page-heading">
        <h2>Resource Quota Management</h2>
        <p className="page-subtitle">Real-time monitoring of provider limits and consumption across active gateways</p>
      </div>

      <div className="stat-grid stat-grid-compact">
        <div className="stat-card">
          <span className="stat-label">Active Providers</span>
          <span className="stat-value mono">{activeProviders}</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Avg Latency (p50)</span>
          <span className="stat-value mono">{avgLatency}ms</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Tracked Quotas</span>
          <span className="stat-value mono">{quota.length}</span>
        </div>
      </div>

      <table className="mono-table" aria-label="Tabla de cuotas por proveedor y modelo">
        <thead>
          <tr>
            <th>Provider</th>
            <th>Model</th>
            <th>Limit</th>
            <th>Remaining</th>
            <th>ResetAt</th>
            <th>HealthStatus</th>
          </tr>
        </thead>
        <tbody>
          {quota.map((row) => (
            <tr key={`${row.provider}-${row.model}`}>
              <td>{row.provider}</td>
              <td>{row.model}</td>
              <td className="mono">{row.limit}</td>
              <td>{remainingBar(row.remaining, row.limit)}</td>
              <td className="mono">{row.reset_at ?? '—'}</td>
              <td>{healthBadge(row.healthy)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

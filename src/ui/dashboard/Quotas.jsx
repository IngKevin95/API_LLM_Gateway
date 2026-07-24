function healthBadge(healthy) {
  return (
    <span className={`badge ${healthy ? 'badge-green' : 'badge-red'}`}>
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

  return (
    <section data-testid="quotas-tab" aria-label="Quotas">
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

function healthBadge(healthy) {
  return (
    <span className={`badge ${healthy ? 'badge-green' : 'badge-red'}`}>
      {healthy ? 'healthy' : 'unhealthy'}
    </span>
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
              <td className="mono">{row.remaining}</td>
              <td className="mono">{row.reset_at ?? '—'}</td>
              <td>{healthBadge(row.healthy)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

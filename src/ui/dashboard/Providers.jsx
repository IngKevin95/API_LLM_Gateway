import useMetricsHistory from './hooks/useMetricsHistory';

function initial(name) {
  return (name || '?').charAt(0).toUpperCase();
}

// Mini barra de disponibilidad derivada de los snapshots reales acumulados por
// useMetricsHistory (circuit_breaker_open por poll) — no hay endpoint de series
// historicas por provider en el backend, asi que se deriva del propio polling.
function AvailabilitySparkline({ name, history }) {
  const points = history.map((h) => h.providers?.[name]);
  if (points.every((p) => p === undefined)) return null;
  return (
    <div className="sparkline" aria-hidden="true">
      {points.map((v, i) => (
        <span
          key={i}
          className={`sparkline-bar ${v === 0 ? 'sparkline-down' : 'sparkline-up'}`}
        />
      ))}
    </div>
  );
}

export default function Providers({ metrics }) {
  const providers = metrics?.providers || [];
  const history = useMetricsHistory(metrics);

  return (
    <section data-testid="providers-tab" aria-label="Providers">
      <div className="page-heading">
        <h2>Upstream Providers</h2>
        <p className="page-subtitle">Real-time health status and circuit breaker orchestration for active LLM backends</p>
      </div>

      <div className="provider-grid">
        {providers.map((p) => (
          <div className="provider-card" key={p.name}>
            <header>
              <div className="provider-identity">
                <span className="provider-avatar" aria-hidden="true">{initial(p.name)}</span>
                <span className="provider-name">{p.name}</span>
              </div>
              <span className={`badge ${p.available ? 'badge-green' : 'badge-red'}`}>
                {p.available ? 'healthy' : 'unhealthy'}
              </span>
            </header>
            <div className="provider-body mono">
              <div>last response: {p.last_success || '—'}</div>
              <div className="circuit-breaker-row">
                circuit breaker:{' '}
                <span className={`badge badge-cb ${p.circuit_breaker_open ? 'badge-red' : 'badge-green'}`}>
                  {p.circuit_breaker_open ? 'open' : 'closed'}
                </span>
              </div>
            </div>
            <AvailabilitySparkline name={p.name} history={history} />
          </div>
        ))}
      </div>
    </section>
  );
}

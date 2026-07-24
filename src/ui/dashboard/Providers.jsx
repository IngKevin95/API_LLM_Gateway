export default function Providers({ metrics }) {
  const providers = metrics?.providers || [];

  return (
    <section data-testid="providers-tab" aria-label="Providers">
      <div className="provider-grid">
        {providers.map((p) => (
          <div className="provider-card" key={p.provider}>
            <header>
              <span className="provider-name">{p.provider}</span>
              <span className={`badge ${p.healthy ? 'badge-green' : 'badge-red'}`}>
                {p.healthy ? 'healthy' : 'unhealthy'}
              </span>
            </header>
            <div className="provider-body mono">
              <div>last response: {p.last_response_at ?? '—'}</div>
              <div>circuit breaker: {p.circuit_breaker ?? 'unknown'}</div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

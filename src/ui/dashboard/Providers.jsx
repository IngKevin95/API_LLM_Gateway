export default function Providers({ metrics }) {
  const providers = metrics?.providers || [];

  return (
    <section data-testid="providers-tab" aria-label="Providers">
      <div className="provider-grid">
        {providers.map((p) => (
          <div className="provider-card" key={p.name}>
            <header>
              <span className="provider-name">{p.name}</span>
              <span className={`badge ${p.available ? 'badge-green' : 'badge-red'}`}>
                {p.available ? 'healthy' : 'unhealthy'}
              </span>
            </header>
            <div className="provider-body mono">
              <div>last response: {p.last_success || '—'}</div>
              <div>circuit breaker: {p.circuit_breaker_open ? 'open' : 'closed'}</div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

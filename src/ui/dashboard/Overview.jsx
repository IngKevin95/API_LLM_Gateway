export default function Overview({ metrics }) {
  if (!metrics) {
    return <p data-testid="overview-loading">Cargando metricas...</p>;
  }

  const latency = metrics.latency || {};
  const requests = metrics.requests || {};

  return (
    <section data-testid="overview-tab" aria-label="Overview">
      <div className="stat-grid">
        <div className="stat-card">
          <span className="stat-label">Uptime</span>
          <span className="stat-value mono">{metrics.uptime_seconds ?? 0}s</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Total Requests</span>
          <span className="stat-value mono">{requests.total ?? 0}</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Errors</span>
          <span className="stat-value mono">{requests.errors ?? 0}</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Latency p50</span>
          <span className="stat-value mono">{latency.p50_ms ?? 0}ms</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Latency p95</span>
          <span className="stat-value mono">{latency.p95_ms ?? 0}ms</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Latency p99</span>
          <span className="stat-value mono">{latency.p99_ms ?? 0}ms</span>
        </div>
      </div>
    </section>
  );
}

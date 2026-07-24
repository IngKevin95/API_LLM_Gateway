import { useMemo } from 'react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import useMetricsHistory from './hooks/useMetricsHistory';

const CHART_COLORS = { requests: '#58a6ff', p50: '#3fb950', p95: '#58a6ff', p99: '#f85149' };

function StatCard({ icon, label, value, sub, badge }) {
  return (
    <div className="stat-card">
      <div className="stat-card-top">
        <span className="stat-icon" aria-hidden="true">{icon}</span>
        {badge}
      </div>
      <span className="stat-label">{label}</span>
      <span className="stat-value mono">{value}</span>
      {sub && <span className="stat-sub">{sub}</span>}
    </div>
  );
}

function formatUptime(seconds) {
  if (!seconds) return '0h 0m';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${h}h ${m}m`;
}

export default function Overview({ metrics }) {
  const history = useMetricsHistory(metrics);

  // "Requests per minute" derivado de deltas reales entre snapshots consecutivos
  // del propio polling (intervalo ~5s) — no es un endpoint de series historicas,
  // es una tasa calculada a partir de datos reales.
  const requestsSeries = useMemo(() => {
    const points = [];
    for (let i = 1; i < history.length; i += 1) {
      const delta = Math.max(0, history[i].requests - history[i - 1].requests);
      points.push({ t: history[i].t, rpm: delta });
    }
    return points;
  }, [history]);

  const latencySeries = useMemo(
    () => history.map((h) => ({ t: h.t, p50: h.p50, p95: h.p95, p99: h.p99 })),
    [history],
  );

  if (!metrics) {
    return <p data-testid="overview-loading">Cargando metricas...</p>;
  }

  const latency = metrics.latency || {};
  const requests = metrics.requests || {};
  const errorRate = requests.total > 0 ? ((requests.errors ?? 0) / requests.total) * 100 : 0;

  return (
    <section data-testid="overview-tab" aria-label="Overview">
      <div className="page-heading">
        <h2>System Infrastructure Overview</h2>
        <p className="page-subtitle">Real-time gateway health and performance telemetry</p>
      </div>

      <div className="stat-grid">
        <StatCard icon="⏱" label="Uptime" value={`${formatUptime(metrics.uptime_seconds)}`} sub={`${metrics.uptime_seconds ?? 0}s`} />
        <StatCard icon="↗" label="Total Requests" value={(requests.total ?? 0).toLocaleString('es')} />
        <StatCard
          icon="⚠"
          label="Error Rate"
          value={`${errorRate.toFixed(1)}%`}
          badge={<span className={`badge ${errorRate < 1 ? 'badge-green' : 'badge-red'}`}>{errorRate < 1 ? 'HEALTHY' : 'ELEVATED'}</span>}
        />
        <StatCard icon="⚡" label="Latency p95" value={`${latency.p95_ms ?? 0}ms`} sub={`p50 ${latency.p50_ms ?? 0}ms · p99 ${latency.p99_ms ?? 0}ms`} />
      </div>

      <div className="chart-grid">
        <div className="chart-panel">
          <h3>Requests per minute (RPM)</h3>
          {requestsSeries.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={requestsSeries}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="t" stroke="var(--text-dim)" fontSize={11} />
                <YAxis stroke="var(--text-dim)" fontSize={11} />
                <Tooltip contentStyle={{ background: 'var(--bg-panel)', border: '1px solid var(--border)' }} />
                <Bar dataKey="rpm" fill={CHART_COLORS.requests} radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="chart-empty">Acumulando datos del polling ({history.length}/2 snapshots)...</p>
          )}
        </div>

        <div className="chart-panel">
          <h3>Latency (ms)</h3>
          {latencySeries.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={latencySeries}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="t" stroke="var(--text-dim)" fontSize={11} />
                <YAxis stroke="var(--text-dim)" fontSize={11} />
                <Tooltip contentStyle={{ background: 'var(--bg-panel)', border: '1px solid var(--border)' }} />
                <Legend wrapperStyle={{ fontSize: 11 }} />
                <Line type="monotone" dataKey="p50" stroke={CHART_COLORS.p50} dot={false} strokeWidth={2} />
                <Line type="monotone" dataKey="p95" stroke={CHART_COLORS.p95} dot={false} strokeWidth={2} />
                <Line type="monotone" dataKey="p99" stroke={CHART_COLORS.p99} dot={false} strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="chart-empty">Acumulando datos del polling...</p>
          )}
        </div>
      </div>

      {/* Nota: el mockup incluye un panel "Live Gateway Logs" — se omite a proposito
          porque el backend no expone un endpoint de logs; mockear entradas de log
          seria mas enganoso que no mostrarlas. */}
    </section>
  );
}

export default function Alerts({ alerts }) {
  const list = alerts || [];
  const criticalCount = list.filter((a) => a.severity === 'critical').length;
  const warningCount = list.filter((a) => a.severity !== 'critical').length;

  return (
    <section data-testid="alerts-tab" aria-label="Alerts">
      <div className="alert-summary" data-testid="alert-summary">
        <span className="badge badge-red">{criticalCount} criticas</span>
        <span className="badge badge-yellow">{warningCount} warnings</span>
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
          {list.map((alert, idx) => (
            <tr
              key={`${alert.provider}-${alert.model}-${idx}`}
              className={alert.severity === 'critical' ? 'row-critical' : ''}
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

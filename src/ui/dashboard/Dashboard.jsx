import { useState } from 'react';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import useMetrics from './hooks/useMetrics';
import useAlerts from './hooks/useAlerts';
import Overview from './Overview';
import Quotas from './Quotas';
import Alerts from './Alerts';
import Providers from './Providers';
import SettingsModal from './SettingsModal';

const TABS = ['Overview', 'Quotas', 'Alerts', 'Providers'];

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState('Overview');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const { metrics } = useMetrics();
  const { alerts } = useAlerts();

  return (
    <div className="gateway-ops-dark" data-testid="dashboard-root">
      <header className="dashboard-header">
        <div className="header-brand">
          <span className="header-logo" aria-hidden="true">◆</span>
          <h1>API LLM Gateway</h1>
        </div>
        <div className="header-actions">
          <span className="tenant-pill" data-testid="tenant-pill" title="RBAC de tenant real fuera de alcance de este payload">
            Tenant: Production
          </span>
          <span className="refresh-indicator" data-testid="refresh-indicator" title="Los datos se refrescan automaticamente cada 5s">
            <span className="refresh-dot" aria-hidden="true" />
            auto-refresh: 5s
          </span>
          <button
            type="button"
            className="icon-button"
            onClick={() => window.location.reload()}
            data-testid="manual-refresh"
            title="Refrescar manualmente"
            aria-label="Refrescar manualmente"
          >
            ⟳
          </button>
          <button
            type="button"
            className="icon-button"
            onClick={() => setSettingsOpen(true)}
            data-testid="open-settings"
            title="Ajustes"
            aria-label="Ajustes"
          >
            ⚙
          </button>
        </div>
      </header>

      <nav className="tabs" role="tablist" aria-label="Dashboard tabs">
        {TABS.map((tab) => (
          <button
            key={tab}
            type="button"
            role="tab"
            aria-selected={activeTab === tab}
            className={activeTab === tab ? 'tab-active' : 'tab'}
            data-testid={`tab-${tab.toLowerCase()}`}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>

      <main>
        {activeTab === 'Overview' && <Overview metrics={metrics} />}
        {activeTab === 'Quotas' && <Quotas metrics={metrics} />}
        {activeTab === 'Alerts' && <Alerts alerts={alerts} />}
        {activeTab === 'Providers' && <Providers metrics={metrics} />}
      </main>

      {settingsOpen && <SettingsModal onClose={() => setSettingsOpen(false)} />}
      <ToastContainer position="bottom-right" theme="dark" />
    </div>
  );
}

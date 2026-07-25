import { useState } from 'react';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import useMetrics from './hooks/useMetrics';
import useAlerts from './hooks/useAlerts';
import useCurrentUser from './hooks/useCurrentUser';
import Overview from './Overview';
import Quotas from './Quotas';
import Alerts from './Alerts';
import Providers from './Providers';
import TeamRoles from './TeamRoles';
import ProfileSecurity from './ProfileSecurity';
import SettingsModal from './SettingsModal';

const BASE_TABS = ['Overview', 'Quotas', 'Alerts', 'Providers'];

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState('Overview');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const { metrics } = useMetrics();
  const { alerts } = useAlerts();
  const { user: currentUser } = useCurrentUser();

  // HU-EVO-021 AC4: "Team" solo se muestra si el usuario autenticado es
  // admin -- el frontend no expone controles que el backend igual va a
  // rechazar con 403. "Profile & Security" es visible para cualquier
  // usuario autenticado (HU-EVO-022).
  const TABS = [
    ...BASE_TABS,
    ...(currentUser?.role === 'admin' ? ['Team'] : []),
    'Profile & Security',
  ];

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
            data-testid={`tab-${tab.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}
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
        {activeTab === 'Team' && <TeamRoles />}
        {activeTab === 'Profile & Security' && <ProfileSecurity />}
      </main>

      {settingsOpen && <SettingsModal onClose={() => setSettingsOpen(false)} />}
      <ToastContainer position="bottom-right" theme="dark" />
    </div>
  );
}

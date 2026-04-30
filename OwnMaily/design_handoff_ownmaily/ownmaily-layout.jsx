// OwnMaily Layout — Sidebar + TopBar + AppShell

const NAV_ITEMS = [
  { id: 'dashboard',    label: 'Dashboard',    icon: 'Home' },
  { id: 'subscribers',  label: 'Subscribers',  icon: 'Users' },
  { id: 'lists',        label: 'Lists',        icon: 'Layers' },
  { id: 'tags',         label: 'Tags',         icon: 'Tag' },
  { id: 'campaigns',    label: 'Campaigns',    icon: 'Mail' },
  { id: 'analytics',   label: 'Analytics',    icon: 'BarChart' },
];

const NAV_BOTTOM = [
  { id: 'settings',    label: 'Settings',     icon: 'Settings' },
];

const NavItem = ({ item, active, onClick, collapsed }) => {
  const [hov, setHov] = React.useState(false);
  const isActive = active === item.id || active?.startsWith(item.id + '/');
  const IconComp = Icons[item.icon];
  return (
    <button onClick={() => onClick(item.id)}
      onMouseEnter={() => setHov(true)} onMouseLeave={() => setHov(false)}
      title={collapsed ? item.label : undefined}
      style={{
        display: 'flex', alignItems: 'center', gap: 10,
        padding: collapsed ? '9px 0' : '9px 12px',
        justifyContent: collapsed ? 'center' : 'flex-start',
        width: '100%', background: isActive ? '#1e1e26' : hov ? '#161619' : 'transparent',
        border: 'none', borderRadius: 7, cursor: 'pointer',
        color: isActive ? '#fff' : hov ? '#ccc' : '#8a8a9a',
        fontFamily: 'inherit', fontSize: 13, fontWeight: isActive ? 500 : 400,
        transition: 'all 100ms', position: 'relative',
        borderLeft: `2px solid ${isActive ? '#e07d3c' : 'transparent'}`,
        marginLeft: -2, paddingLeft: collapsed ? undefined : 10,
      }}>
      <IconComp size={15} color={isActive ? '#e07d3c' : 'currentColor'} />
      {!collapsed && <span>{item.label}</span>}
    </button>
  );
};

const Sidebar = ({ current, navigate, collapsed }) => {
  const sidebarW = collapsed ? 56 : 216;
  return (
    <div style={{
      width: sidebarW, flexShrink: 0, background: '#0a0a0c',
      borderRight: '1px solid #1a1a22', display: 'flex', flexDirection: 'column',
      height: '100vh', position: 'sticky', top: 0, transition: 'width 200ms ease',
      overflow: 'hidden',
    }}>
      {/* Logo */}
      <div style={{
        padding: collapsed ? '18px 0' : '18px 16px', borderBottom: '1px solid #1a1a22',
        display: 'flex', alignItems: 'center', gap: 9, justifyContent: collapsed ? 'center' : 'flex-start',
        flexShrink: 0,
      }}>
        <div style={{
          width: 26, height: 26, borderRadius: 7, background: '#e07d3c',
          display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
        }}>
          <Icons.Inbox size={14} color="#fff" />
        </div>
        {!collapsed && (
          <span style={{ fontSize: 14, fontWeight: 700, color: '#fff', letterSpacing: '-0.02em', whiteSpace: 'nowrap' }}>
            OwnMaily
          </span>
        )}
      </div>

      {/* Main nav */}
      <nav style={{ flex: 1, padding: '10px 8px', display: 'flex', flexDirection: 'column', gap: 2, overflowY: 'auto' }}>
        {NAV_ITEMS.map(item => (
          <NavItem key={item.id} item={item} active={current} onClick={navigate} collapsed={collapsed} />
        ))}
      </nav>

      {/* Bottom nav */}
      <div style={{ padding: '10px 8px', borderTop: '1px solid #1a1a22', display: 'flex', flexDirection: 'column', gap: 2 }}>
        {NAV_BOTTOM.map(item => (
          <NavItem key={item.id} item={item} active={current} onClick={navigate} collapsed={collapsed} />
        ))}
        {/* User */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 8,
          padding: collapsed ? '9px 0' : '9px 12px',
          justifyContent: collapsed ? 'center' : 'flex-start',
          marginTop: 4, borderTop: '1px solid #1a1a22', paddingTop: 12,
        }}>
          <div style={{
            width: 26, height: 26, borderRadius: '50%', background: '#2a2a3a',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
            fontSize: 11, fontWeight: 700, color: '#e07d3c',
          }}>A</div>
          {!collapsed && (
            <div style={{ minWidth: 0 }}>
              <div style={{ fontSize: 12, fontWeight: 500, color: '#ccc', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>admin@example.com</div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const TopBar = ({ title, subtitle, actions, navigate, current }) => {
  // Breadcrumb
  const crumbs = (() => {
    if (!current) return [title];
    const parts = current.split('/');
    return parts.map((p, i) => {
      const labels = { dashboard: 'Dashboard', subscribers: 'Subscribers', lists: 'Lists', tags: 'Tags', campaigns: 'Campaigns', analytics: 'Analytics', settings: 'Settings', general: 'General', smtp: 'SMTP', apikey: 'API Key', suppression: 'Suppression List', create: 'New Campaign', stats: 'Stats', detail: 'Detail', wizard: 'Setup Wizard' };
      return labels[p] || p;
    });
  })();

  return (
    <div style={{
      height: 52, background: '#fff', borderBottom: '1px solid #ece9e1',
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '0 24px', flexShrink: 0, position: 'sticky', top: 0, zIndex: 10,
    }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {crumbs.map((c, i) => (
            <React.Fragment key={i}>
              {i > 0 && <Icons.ChevronRight size={12} color="#ccc" />}
              <span style={{
                fontSize: i === crumbs.length - 1 ? 14 : 13,
                fontWeight: i === crumbs.length - 1 ? 600 : 400,
                color: i === crumbs.length - 1 ? '#111' : '#999',
              }}>{c}</span>
            </React.Fragment>
          ))}
        </div>
        {subtitle && <span style={{ fontSize: 11, color: '#aaa' }}>{subtitle}</span>}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {actions}
      </div>
    </div>
  );
};

const PageContent = ({ children, style: extra }) => (
  <div style={{ padding: '24px 28px', flex: 1, overflowY: 'auto', ...extra }}>
    {children}
  </div>
);

const AppShell = ({ current, navigate, topActions, topSubtitle, children, tweaksOpen, tweaksPanel, collapsed }) => (
  <div style={{ display: 'flex', height: '100vh', background: '#f4f3ef', fontFamily: "'DM Sans', sans-serif", overflow: 'hidden' }}>
    <Sidebar current={current} navigate={navigate} collapsed={collapsed} />
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, overflow: 'hidden' }}>
      <TopBar title={current} current={current} actions={topActions} subtitle={topSubtitle} navigate={navigate} />
      <div style={{ flex: 1, overflowY: 'auto', position: 'relative' }}>
        {children}
        {tweaksPanel}
      </div>
    </div>
  </div>
);

Object.assign(window, { AppShell, Sidebar, TopBar, PageContent, NAV_ITEMS });

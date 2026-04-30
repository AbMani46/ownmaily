// OwnMaily Root App — routing + tweaks

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accent": "#e07d3c",
  "sidebarCollapsed": false,
  "density": "comfortable"
}/*EDITMODE-END*/;

const App = () => {
  const tweaksResult = (typeof useTweaks === 'function' ? useTweaks(TWEAK_DEFAULTS) : null) || {};
  const tweaks = tweaksResult.tweaks || TWEAK_DEFAULTS;
  const setTweak = tweaksResult.setTweak || (() => {});
  const [screen, setScreen] = React.useState('login');

  // Inject accent color CSS variable
  React.useEffect(() => {
    document.documentElement.style.setProperty('--accent', tweaks.accent);
  }, [tweaks.accent]);

  const navigate = (s) => setScreen(s);

  // Parse screen to route parts
  const parts = screen.split('/');
  const root = parts[0];

  // Screens without shell
  if (screen === 'login') return <LoginScreen navigate={navigate} />;
  if (screen === 'setup-wizard') return <SetupWizard navigate={navigate} />;

  // Determine top bar actions based on screen
  const topActions = (() => {
    if (root === 'dashboard') return <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />} onClick={() => navigate('campaigns/create/new')}>New Campaign</Button>;
    if (root === 'subscribers' && parts.length === 1) return (
      <div style={{ display: 'flex', gap: 6 }}>
        <Button variant="secondary" size="sm" icon={<Icons.Upload size={12} />}>Import</Button>
        <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />}>Add Subscriber</Button>
      </div>
    );
    if (root === 'campaigns' && parts.length === 1) return <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />} onClick={() => navigate('campaigns/create/new')}>Create Campaign</Button>;
    if (root === 'lists' && parts.length === 1) return <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />}>Create List</Button>;
    if (root === 'tags') return <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />}>Create Tag</Button>;
    return null;
  })();

  const renderScreen = () => {
    if (root === 'dashboard') return <DashboardScreen navigate={navigate} />;
    if (root === 'subscribers') {
      if (parts[1] === 'detail') return <SubscriberDetailScreen id={parts[2]} navigate={navigate} />;
      return <SubscribersScreen navigate={navigate} />;
    }
    if (root === 'lists') {
      if (parts[1] === 'detail') return <ListDetailScreen id={parts[2]} navigate={navigate} />;
      return <ListsScreen navigate={navigate} />;
    }
    if (root === 'tags') return <TagsScreen navigate={navigate} />;
    if (root === 'campaigns') {
      if (parts[1] === 'create') return <CampaignCreateScreen id={parts[2]} navigate={navigate} />;
      if (parts[1] === 'stats') return <CampaignStatsScreen id={parts[2]} navigate={navigate} />;
      return <CampaignsScreen navigate={navigate} />;
    }
    if (root === 'analytics') return <AnalyticsScreen navigate={navigate} />;
    if (root === 'settings') {
      const sub = parts[1] || 'general';
      if (sub === 'general') return <SettingsGeneralScreen navigate={navigate} />;
      if (sub === 'smtp') return <SettingsSMTPScreen navigate={navigate} />;
      if (sub === 'apikey') return <SettingsAPIKeyScreen navigate={navigate} />;
      if (sub === 'suppression') return <SettingsSuppressionScreen navigate={navigate} />;
      return <SettingsGeneralScreen navigate={navigate} />;
    }
    return <DashboardScreen navigate={navigate} />;
  };

  const tweaksPanel = (
    <TweaksPanel>
      <TweakSection title="Brand">
        <TweakColor id="accent" label="Accent color" value={tweaks.accent} onChange={v => setTweak('accent', v)} />
      </TweakSection>
      <TweakSection title="Layout">
        <TweakToggle id="sidebarCollapsed" label="Collapse sidebar" value={tweaks.sidebarCollapsed} onChange={v => setTweak('sidebarCollapsed', v)} />
        <TweakRadio id="density" label="Density" value={tweaks.density} onChange={v => setTweak('density', v)} options={[{ value: 'compact', label: 'Compact' }, { value: 'comfortable', label: 'Comfortable' }]} />
      </TweakSection>
      <TweakSection title="Navigate">
        {[
          ['Dashboard', 'dashboard'],
          ['Subscribers', 'subscribers'],
          ['Subscriber Detail', 'subscribers/detail/1'],
          ['Lists', 'lists'],
          ['List Detail', 'lists/detail/1'],
          ['Tags', 'tags'],
          ['Campaigns', 'campaigns'],
          ['New Campaign', 'campaigns/create/new'],
          ['Campaign Stats', 'campaigns/stats/1'],
          ['Analytics', 'analytics'],
          ['Settings: General', 'settings/general'],
          ['Settings: SMTP', 'settings/smtp'],
          ['Settings: API Key', 'settings/apikey'],
          ['Settings: Suppression', 'settings/suppression'],
          ['Login', 'login'],
          ['Setup Wizard', 'setup-wizard'],
        ].map(([label, id]) => (
          <TweakButton key={id} label={label} onClick={() => navigate(id)}
            style={{ fontWeight: screen === id ? 600 : 400, color: screen === id ? '#e07d3c' : undefined }} />
        ))}
      </TweakSection>
    </TweaksPanel>
  );

  return (
    <AppShell
      current={screen}
      navigate={navigate}
      topActions={topActions}
      collapsed={tweaks.sidebarCollapsed}
      tweaksPanel={tweaksPanel}
    >
      <div style={{ '--density-gap': tweaks.density === 'compact' ? '14px' : '22px' }}>
        {renderScreen()}
      </div>
    </AppShell>
  );
};

ReactDOM.createRoot(document.getElementById('root')).render(<App />);

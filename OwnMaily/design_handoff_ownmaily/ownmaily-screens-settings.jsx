// Settings screens: General, SMTP, API Key, Suppression List

const SettingsSidebar = ({ active, onChange }) => {
  const items = [
    { id: 'settings/general', label: 'General', icon: 'Globe' },
    { id: 'settings/smtp', label: 'SMTP', icon: 'Server' },
    { id: 'settings/apikey', label: 'API Key', icon: 'Key' },
    { id: 'settings/suppression', label: 'Suppression List', icon: 'Shield' },
  ];
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 2, width: 200, flexShrink: 0 }}>
      {items.map(item => {
        const IconComp = Icons[item.icon];
        const isActive = active === item.id;
        return (
          <button key={item.id} onClick={() => onChange(item.id)} style={{
            display: 'flex', alignItems: 'center', gap: 9, padding: '9px 12px',
            borderRadius: 7, border: 'none', background: isActive ? '#fff8f4' : 'transparent',
            color: isActive ? '#e07d3c' : '#555', fontFamily: 'inherit', fontSize: 13,
            fontWeight: isActive ? 600 : 400, cursor: 'pointer', transition: 'all 100ms',
            borderLeft: `2px solid ${isActive ? '#e07d3c' : 'transparent'}`, paddingLeft: 10,
          }}>
            <IconComp size={14} color={isActive ? '#e07d3c' : '#aaa'} />
            {item.label}
          </button>
        );
      })}
    </div>
  );
};

const SettingsLayout = ({ current, navigate, children }) => (
  <div style={{ padding: '24px 28px', display: 'flex', gap: 32 }}>
    <SettingsSidebar active={current} onChange={navigate} />
    <div style={{ flex: 1, minWidth: 0 }}>
      {children}
    </div>
  </div>
);

const FormSection = ({ title, description, children }) => (
  <div style={{ marginBottom: 28 }}>
    <div style={{ marginBottom: 16 }}>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: '#111', margin: '0 0 4px' }}>{title}</h3>
      {description && <p style={{ fontSize: 13, color: '#888', margin: 0, lineHeight: 1.5 }}>{description}</p>}
    </div>
    <Card>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {children}
      </div>
    </Card>
  </div>
);

const SettingsGeneralScreen = ({ navigate }) => {
  const [saved, setSaved] = React.useState(false);
  const handleSave = () => { setSaved(true); setTimeout(() => setSaved(false), 2000); };
  return (
    <SettingsLayout current="settings/general" navigate={navigate}>
      <div style={{ maxWidth: 560 }}>
        <FormSection title="Site Information" description="Basic details about your installation.">
          <Input label="Site name" defaultValue="My Newsletter" fullWidth />
          <Input label="Installation URL" prefix="https://" defaultValue="mail.example.com" fullWidth />
          <Select label="Timezone" options={['UTC', 'America/New_York', 'America/Chicago', 'America/Los_Angeles', 'Europe/London', 'Europe/Berlin', 'Asia/Tokyo', 'Australia/Sydney']} fullWidth />
        </FormSection>
        <FormSection title="Sender Defaults" description="Used as defaults for new campaigns.">
          <Input label="From name" defaultValue="OwnMaily Team" fullWidth />
          <Input label="From email" defaultValue="hello@example.com" fullWidth />
          <Input label="Reply-to" defaultValue="hello@example.com" fullWidth />
        </FormSection>
        <FormSection title="Compliance" description="Required for CAN-SPAM and GDPR compliance.">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
            <label style={{ fontSize: 12, fontWeight: 600, color: '#444', letterSpacing: '0.02em' }}>Physical address</label>
            <textarea defaultValue="123 Main St, Suite 100&#10;San Francisco, CA 94102&#10;United States" rows={3}
              style={{ padding: '8px 11px', border: '1px solid #ddd', borderRadius: 7, fontSize: 13, fontFamily: 'inherit', outline: 'none', resize: 'vertical', color: '#111', lineHeight: 1.5 }} />
            <span style={{ fontSize: 11, color: '#aaa' }}>Shown in the footer of every email you send.</span>
          </div>
        </FormSection>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <Button variant="primary" onClick={handleSave}>Save changes</Button>
          {saved && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#059669', fontSize: 13, fontWeight: 500 }}>
              <Icons.Check size={14} color="#059669" /> Saved
            </div>
          )}
        </div>
      </div>
    </SettingsLayout>
  );
};

const SMTP_FIELDS = {
  Resend: [
    { key: 'apiKey', label: 'API Key', type: 'password', placeholder: 're_••••••••••••••••', mono: true },
    { key: 'fromDomain', label: 'Sending domain', placeholder: 'mg.example.com' },
  ],
  Mailgun: [
    { key: 'apiKey', label: 'API Key', type: 'password', placeholder: 'key-••••••••••••••••', mono: true },
    { key: 'domain', label: 'Mailgun domain', placeholder: 'mg.example.com' },
    { key: 'region', label: 'Region', type: 'select', options: ['US', 'EU'] },
  ],
  'Amazon SES': [
    { key: 'accessKey', label: 'AWS Access Key ID', placeholder: 'AKIA••••••••••••••••', mono: true },
    { key: 'secretKey', label: 'AWS Secret Key', type: 'password', placeholder: '••••••••••••••••••••••••••••••••••••••••', mono: true },
    { key: 'region', label: 'AWS Region', type: 'select', options: ['us-east-1', 'us-west-2', 'eu-west-1', 'ap-southeast-1'] },
  ],
  'Custom SMTP': [
    { key: 'host', label: 'SMTP host', placeholder: 'smtp.example.com' },
    { key: 'port', label: 'Port', placeholder: '587' },
    { key: 'user', label: 'Username', placeholder: 'smtp@example.com' },
    { key: 'pass', label: 'Password', type: 'password', placeholder: '••••••••' },
    { key: 'encryption', label: 'Encryption', type: 'select', options: ['TLS', 'SSL', 'None'] },
  ],
};

const SettingsSMTPScreen = ({ navigate }) => {
  const [provider, setProvider] = React.useState('Resend');
  const [tested, setTested] = React.useState(null);
  const fields = SMTP_FIELDS[provider] || [];
  const handleTest = () => { setTested('loading'); setTimeout(() => setTested('success'), 1200); };

  return (
    <SettingsLayout current="settings/smtp" navigate={navigate}>
      <div style={{ maxWidth: 560 }}>
        <FormSection title="Email Provider" description="Configure how OwnMaily sends emails on your behalf.">
          <Select label="Provider" value={provider} onChange={e => { setProvider(e.target.value); setTested(null); }}
            options={Object.keys(SMTP_FIELDS)} fullWidth />
          {fields.map(f => f.type === 'select'
            ? <Select key={f.key} label={f.label} options={f.options} fullWidth />
            : <Input key={f.key} label={f.label} type={f.type || 'text'} placeholder={f.placeholder} mono={f.mono} fullWidth />
          )}
        </FormSection>

        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <Button variant="primary">Save configuration</Button>
          <Button variant="secondary" onClick={handleTest} icon={<Icons.Zap size={13} />}>
            {tested === 'loading' ? 'Testing…' : 'Test connection'}
          </Button>
          {tested === 'success' && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#059669', fontSize: 13, fontWeight: 500 }}>
              <Icons.Check size={14} color="#059669" /> Connection successful
            </div>
          )}
        </div>
      </div>
    </SettingsLayout>
  );
};

const SettingsAPIKeyScreen = ({ navigate }) => {
  const FULL_KEY = 'om_live_sk_a4f2b8c3d9e7f1a0b5c6d2e8f3a9b4c7d0e5f1a2b3c4d5e6f7a8b9c0d1e2f3';
  const MASKED = 'om_live_sk_' + '•'.repeat(48);
  const [revealed, setRevealed] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const [regenerated, setRegenerated] = React.useState(false);
  const [showWarning, setShowWarning] = React.useState(false);

  const handleCopy = () => {
    navigator.clipboard?.writeText(FULL_KEY);
    setCopied(true); setTimeout(() => setCopied(false), 2000);
  };

  const handleRegen = () => {
    setShowWarning(true);
  };

  return (
    <SettingsLayout current="settings/apikey" navigate={navigate}>
      <div style={{ maxWidth: 560 }}>
        <FormSection title="API Key" description="Use this key to interact with the OwnMaily REST API.">
          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: '#444', letterSpacing: '0.02em', display: 'block', marginBottom: 5 }}>Secret key</label>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <div style={{
                flex: 1, padding: '8px 12px', background: '#0c0c0f', border: '1px solid #222',
                borderRadius: 7, fontFamily: "'JetBrains Mono', monospace", fontSize: 12,
                color: '#9aefbc', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
              }}>
                {revealed ? FULL_KEY : MASKED}
              </div>
              <button onClick={() => setRevealed(v => !v)} style={{ padding: '8px 10px', border: '1px solid #ddd', borderRadius: 7, background: '#fff', cursor: 'pointer', color: '#555' }}>
                {revealed ? <Icons.EyeOff size={14} /> : <Icons.Eye size={14} />}
              </button>
              <button onClick={handleCopy} style={{ padding: '8px 10px', border: '1px solid #ddd', borderRadius: 7, background: '#fff', cursor: 'pointer', color: copied ? '#059669' : '#555' }}>
                {copied ? <Icons.Check size={14} color="#059669" /> : <Icons.Copy size={14} />}
              </button>
            </div>
          </div>

          <div style={{ background: '#fffbeb', border: '1px solid #fde68a', borderRadius: 8, padding: '12px 14px', display: 'flex', gap: 10 }}>
            <Icons.AlertCircle size={16} color="#d97706" style={{ flexShrink: 0, marginTop: 1 }} />
            <p style={{ margin: 0, fontSize: 12, color: '#92400e', lineHeight: 1.5 }}>
              This key grants full access to your OwnMaily installation. Keep it secret. It is shown in full only once after generation.
            </p>
          </div>
        </FormSection>

        {showWarning ? (
          <div style={{ background: '#fff', border: '1px solid #fca5a5', borderRadius: 10, padding: '20px 22px' }}>
            <h4 style={{ margin: '0 0 8px', color: '#991b1b', fontSize: 14, fontWeight: 600 }}>Regenerate API key?</h4>
            <p style={{ margin: '0 0 14px', fontSize: 13, color: '#666', lineHeight: 1.5 }}>
              Your current key will be immediately invalidated. Any integrations using it will stop working until updated.
            </p>
            <div style={{ display: 'flex', gap: 8 }}>
              <Button variant="danger" size="sm" onClick={() => { setRegenerated(true); setShowWarning(false); setRevealed(true); }}>
                Yes, regenerate
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setShowWarning(false)}>Cancel</Button>
            </div>
          </div>
        ) : (
          <Button variant="outline" icon={<Icons.RefreshCw size={13} />} onClick={handleRegen}>
            {regenerated ? 'Key regenerated' : 'Regenerate key'}
          </Button>
        )}

        <div style={{ marginTop: 28 }}>
          <FormSection title="API Documentation" description="Use the API to manage subscribers, campaigns, and more programmatically.">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <CodeBlock code={`curl https://mail.example.com/api/subscribers \\
  -H "Authorization: Bearer om_live_sk_••••••••"`} />
              <a href="#" style={{ fontSize: 12, color: '#6366f1', textDecoration: 'none' }}>View full API reference →</a>
            </div>
          </FormSection>
        </div>
      </div>
    </SettingsLayout>
  );
};

const SUPPRESSION_DATA = [
  { email: 'bounce@invalid.com', reason: 'Hard bounce', date: 'Apr 20, 2026' },
  { email: 'noreply@corporate.com', reason: 'Hard bounce', date: 'Apr 15, 2026' },
  { email: 'spam@complaint.net', reason: 'Spam complaint', date: 'Apr 10, 2026' },
  { email: 'manual@removed.io', reason: 'Manually added', date: 'Mar 28, 2026' },
  { email: 'typo@gmial.com', reason: 'Hard bounce', date: 'Mar 12, 2026' },
];

const SettingsSuppressionScreen = ({ navigate }) => {
  const [search, setSearch] = React.useState('');
  const [list, setList] = React.useState(SUPPRESSION_DATA);
  const filtered = list.filter(r => !search || r.email.includes(search.toLowerCase()));

  const cols = [
    { key: 'email', label: 'Email', render: v => <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{v}</span> },
    { key: 'reason', label: 'Reason', render: v => <Badge status="suppressed" label={v} /> },
    { key: 'date', label: 'Added', muted: true },
    { key: 'email', label: '', render: (v) => (
      <Button variant="ghost" size="sm" icon={<Icons.Trash size={11} color="#dc2626" />}
        style={{ color: '#dc2626' }}
        onClick={e => { e.stopPropagation(); setList(l => l.filter(r => r.email !== v)); }}>
        Remove
      </Button>
    )},
  ];

  return (
    <SettingsLayout current="settings/suppression" navigate={navigate}>
      <div>
        <div style={{ display: 'flex', gap: 8, marginBottom: 18, alignItems: 'center' }}>
          <div style={{ flex: 1, position: 'relative' }}>
            <Icons.Search size={14} style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: '#bbb' }} />
            <input value={search} onChange={e => setSearch(e.target.value)} placeholder="Search suppressed emails…"
              style={{ width: '100%', padding: '8px 12px 8px 32px', border: '1px solid #ddd', borderRadius: 7, fontSize: 13, fontFamily: 'inherit', outline: 'none', boxSizing: 'border-box' }} />
          </div>
          <Button variant="secondary" icon={<Icons.Download size={13} />}>Export</Button>
          <Button variant="primary" icon={<Icons.Plus size={13} />}>Add Email</Button>
        </div>

        <div style={{ background: '#fff8f4', border: '1px solid #fed7aa', borderRadius: 8, padding: '12px 14px', display: 'flex', gap: 10, marginBottom: 18 }}>
          <Icons.AlertCircle size={15} color="#c2410c" style={{ flexShrink: 0, marginTop: 1 }} />
          <p style={{ margin: 0, fontSize: 12, color: '#9a3412', lineHeight: 1.5 }}>
            Emails on this list will never be contacted, regardless of list membership or campaign target. This protects your sender reputation.
          </p>
        </div>

        <Card padding={null}>
          <Table columns={cols} rows={filtered} />
          {filtered.length === 0 && (
            <div style={{ padding: '32px', textAlign: 'center', color: '#bbb', fontSize: 13 }}>No suppressed emails match your search.</div>
          )}
        </Card>
      </div>
    </SettingsLayout>
  );
};

Object.assign(window, { SettingsGeneralScreen, SettingsSMTPScreen, SettingsAPIKeyScreen, SettingsSuppressionScreen });

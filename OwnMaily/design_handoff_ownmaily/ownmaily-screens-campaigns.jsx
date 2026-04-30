// Campaign screens: List, Create/Edit, Stats

const CampaignsScreen = ({ navigate }) => {
  const [tab, setTab] = React.useState('all');
  const tabs = [
    { id: 'all', label: 'All', count: MOCK_CAMPAIGNS.length },
    { id: 'sent', label: 'Sent', count: MOCK_CAMPAIGNS.filter(c => c.status === 'sent').length },
    { id: 'scheduled', label: 'Scheduled', count: MOCK_CAMPAIGNS.filter(c => c.status === 'scheduled').length },
    { id: 'draft', label: 'Drafts', count: MOCK_CAMPAIGNS.filter(c => c.status === 'draft').length },
  ];
  const filtered = tab === 'all' ? MOCK_CAMPAIGNS : MOCK_CAMPAIGNS.filter(c => c.status === tab);

  const cols = [
    { key: 'name', label: 'Campaign', render: (v, row) => (
      <div>
        <div style={{ fontWeight: 500, color: '#111' }}>{v}</div>
        <div style={{ fontSize: 11, color: '#aaa' }}>{row.list}</div>
      </div>
    )},
    { key: 'status', label: 'Status', render: v => <Badge status={v} /> },
    { key: 'date', label: 'Date', muted: true },
    { key: 'sent', label: 'Sent', render: v => v ? v.toLocaleString() : '—', muted: true },
    { key: 'opens', label: 'Open Rate', render: v => <span style={{ fontWeight: v !== '—' ? 600 : 400, color: v !== '—' ? '#059669' : '#ccc' }}>{v}</span> },
    { key: 'clicks', label: 'Click Rate', render: v => <span style={{ fontWeight: v !== '—' ? 600 : 400, color: v !== '—' ? '#6366f1' : '#ccc' }}>{v}</span> },
    { key: 'id', label: '', render: (v, row) => (
      <div style={{ display: 'flex', gap: 4 }} onClick={e => e.stopPropagation()}>
        {row.status !== 'sent' && <Button variant="ghost" size="sm" onClick={() => navigate('campaigns/create/' + v)}>Edit</Button>}
        {row.status === 'sent' && <Button variant="ghost" size="sm" onClick={() => navigate('campaigns/stats/' + v)}>Stats</Button>}
      </div>
    )},
  ];

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <Button variant="primary" icon={<Icons.Plus size={13} />} onClick={() => navigate('campaigns/create/new')}>Create Campaign</Button>
      </div>
      <Card padding={null}>
        <div style={{ padding: '4px 14px 0' }}>
          <Tabs tabs={tabs} active={tab} onChange={setTab} />
        </div>
        <Table columns={cols} rows={filtered} onRowClick={row => row.status === 'sent' ? navigate('campaigns/stats/' + row.id) : navigate('campaigns/create/' + row.id)} />
      </Card>
    </div>
  );
};

const TEMPLATES = [
  { id: 'blank', label: 'Blank', icon: '□' },
  { id: 'newsletter', label: 'Newsletter', icon: '▦' },
  { id: 'announcement', label: 'Announcement', icon: '◉' },
  { id: 'digest', label: 'Weekly Digest', icon: '≡' },
];

const CampaignCreateScreen = ({ id, navigate }) => {
  const isNew = id === 'new';
  const existing = !isNew ? MOCK_CAMPAIGNS.find(c => c.id === Number(id)) : null;
  const [form, setForm] = React.useState({
    name: existing?.name || '',
    subject: existing ? `${existing.name} — updates for you` : '',
    preview: '',
    fromName: 'OwnMaily Team',
    fromEmail: 'hello@example.com',
    replyTo: 'hello@example.com',
    target: existing?.list || 'All Subscribers',
    template: 'newsletter',
  });
  const set = (k, v) => setForm(f => ({ ...f, [k]: v }));
  const [previewMode, setPreviewMode] = React.useState(false);

  const editorContent = `<h2>Hello {first_name | "there"},</h2>
<p>We've been working hard behind the scenes, and we're excited to share what's new this month.</p>
<h3>🚀 What's new</h3>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
<p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>
<hr />
<p style="color:#888;font-size:13px">You're receiving this because you subscribed at example.com. <a href="#">Unsubscribe</a></p>`;

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 20, height: 'calc(100vh - 52px)', boxSizing: 'border-box' }}>
      {/* Template selector */}
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <span style={{ fontSize: 12, color: '#888', fontWeight: 500 }}>Template:</span>
        {TEMPLATES.map(t => (
          <button key={t.id} onClick={() => set('template', t.id)} style={{
            padding: '5px 12px', borderRadius: 7, border: `1px solid ${form.template === t.id ? '#e07d3c' : '#ddd'}`,
            background: form.template === t.id ? '#fff8f4' : '#fff', color: form.template === t.id ? '#e07d3c' : '#666',
            fontSize: 12, fontWeight: form.template === t.id ? 600 : 400, cursor: 'pointer', fontFamily: 'inherit',
          }}>{t.icon} {t.label}</button>
        ))}
      </div>

      {/* Two-column editor */}
      <div style={{ display: 'grid', gridTemplateColumns: '360px 1fr', gap: 20, flex: 1, minHeight: 0 }}>
        {/* Left: Form */}
        <Card style={{ overflowY: 'auto' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Input label="Campaign name" placeholder="e.g. May Newsletter" value={form.name} onChange={e => set('name', e.target.value)} fullWidth />
            <div style={{ height: 1, background: '#f0ede6' }} />
            <Input label="Subject line" placeholder="What's in it for them?" value={form.subject} onChange={e => set('subject', e.target.value)} fullWidth />
            <Input label="Preview text" placeholder="Short teaser shown in inbox" value={form.preview} onChange={e => set('preview', e.target.value)} hint="Appears after the subject in most email clients" fullWidth />
            <div style={{ height: 1, background: '#f0ede6' }} />
            <Input label="From name" value={form.fromName} onChange={e => set('fromName', e.target.value)} fullWidth />
            <Input label="From email" value={form.fromEmail} onChange={e => set('fromEmail', e.target.value)} fullWidth />
            <Input label="Reply-to" value={form.replyTo} onChange={e => set('replyTo', e.target.value)} fullWidth />
            <div style={{ height: 1, background: '#f0ede6' }} />
            <Select label="Send to" value={form.target} onChange={e => set('target', e.target.value)} fullWidth
              options={MOCK_LISTS.map(l => ({ value: l.name, label: `${l.name} (${l.count.toLocaleString()})` }))} />
            <div style={{ background: '#faf8f4', borderRadius: 8, padding: '10px 12px', fontSize: 12, color: '#666' }}>
              <strong style={{ color: '#333' }}>{(MOCK_LISTS.find(l => l.name === form.target)?.count || 4218).toLocaleString()}</strong> subscribers will receive this campaign
            </div>
          </div>
        </Card>

        {/* Right: Editor */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 0, background: '#fff', border: '1px solid #e8e5de', borderRadius: 10, overflow: 'hidden' }}>
          {/* Toolbar */}
          <div style={{ display: 'flex', gap: 2, padding: '8px 12px', borderBottom: '1px solid #f0ede6', alignItems: 'center', background: '#faf9f7', flexWrap: 'wrap' }}>
            {[
              ['B', 'font-weight:bold', 'Bold'],
              ['I', 'font-style:italic', 'Italic'],
              ['U', 'text-decoration:underline', 'Underline'],
            ].map(([label, style, title]) => (
              <button key={label} title={title} style={{ width: 28, height: 28, border: '1px solid #e8e5de', borderRadius: 5, background: '#fff', cursor: 'pointer', fontSize: 13, fontFamily: 'inherit', fontWeight: label === 'B' ? 700 : 400, fontStyle: label === 'I' ? 'italic' : 'normal', textDecoration: label === 'U' ? 'underline' : 'none', color: '#555' }}>{label}</button>
            ))}
            <div style={{ width: 1, height: 20, background: '#e8e5de', margin: '0 4px' }} />
            {['H1', 'H2', 'H3'].map(h => (
              <button key={h} style={{ padding: '4px 8px', border: '1px solid #e8e5de', borderRadius: 5, background: '#fff', cursor: 'pointer', fontSize: 11, fontWeight: 600, color: '#555', fontFamily: 'inherit' }}>{h}</button>
            ))}
            <div style={{ width: 1, height: 20, background: '#e8e5de', margin: '0 4px' }} />
            <button title="Link" style={{ width: 28, height: 28, border: '1px solid #e8e5de', borderRadius: 5, background: '#fff', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icons.Link size={13} color="#555" /></button>
            <button title="Image" style={{ padding: '4px 8px', border: '1px solid #e8e5de', borderRadius: 5, background: '#fff', cursor: 'pointer', fontSize: 11, color: '#555', fontFamily: 'inherit' }}>Image</button>
            <button title="List" style={{ width: 28, height: 28, border: '1px solid #e8e5de', borderRadius: 5, background: '#fff', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icons.List size={13} color="#555" /></button>
            <div style={{ flex: 1 }} />
            <button onClick={() => setPreviewMode(v => !v)} style={{ padding: '4px 10px', borderRadius: 6, border: '1px solid #e8e5de', background: previewMode ? '#f0ede6' : '#fff', color: '#555', fontSize: 11, cursor: 'pointer', fontFamily: 'inherit', display: 'flex', alignItems: 'center', gap: 5 }}>
              <Icons.Eye size={12} /> {previewMode ? 'Edit' : 'Preview'}
            </button>
          </div>

          {/* Content area */}
          <div style={{ flex: 1, overflowY: 'auto', padding: '24px', background: previewMode ? '#f4f3ef' : '#fff' }}>
            {previewMode ? (
              <div style={{ maxWidth: 600, margin: '0 auto', background: '#fff', borderRadius: 8, padding: '32px', boxShadow: '0 4px 24px rgba(0,0,0,0.08)', fontFamily: 'Georgia, serif', lineHeight: 1.7 }}>
                <div style={{ fontSize: 12, color: '#aaa', marginBottom: 16 }}>From: {form.fromName} &lt;{form.fromEmail}&gt;</div>
                <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 6, fontFamily: 'inherit' }}>{form.subject || 'No subject yet'}</div>
                <div style={{ color: '#aaa', fontSize: 12, marginBottom: 24 }}>{form.preview || 'No preview text'}</div>
                <div dangerouslySetInnerHTML={{ __html: editorContent }} />
              </div>
            ) : (
              <div contentEditable suppressContentEditableWarning style={{ minHeight: 400, outline: 'none', fontSize: 14, lineHeight: 1.7, color: '#222', fontFamily: "'DM Sans', sans-serif" }} dangerouslySetInnerHTML={{ __html: editorContent }} />
            )}
          </div>

          {/* Bottom actions */}
          <div style={{ padding: '12px 16px', borderTop: '1px solid #f0ede6', display: 'flex', gap: 8, justifyContent: 'flex-end', background: '#faf9f7' }}>
            <Button variant="secondary" size="md">Save Draft</Button>
            <Button variant="outline" size="md" icon={<Icons.Calendar size={13} />}>Schedule</Button>
            <Button variant="primary" size="md" icon={<Icons.Send size={13} />} onClick={() => navigate('campaigns')}>Send Now</Button>
          </div>
        </div>
      </div>
    </div>
  );
};

// Timeline chart as inline SVG
const TimelineChart = ({ width = 600 }) => {
  const opens = [42, 180, 310, 390, 420, 440, 445, 451, 453, 455, 456, 456];
  const clicks = [8, 40, 72, 88, 95, 100, 103, 106, 107, 108, 108, 108];
  const hours = ['0h', '1h', '2h', '4h', '6h', '8h', '10h', '12h', '16h', '20h', '24h', '48h'];
  const h = 120, maxV = 460, pts = (data) => data.map((v, i) => `${(i / (data.length - 1)) * width},${h - (v / maxV) * (h - 8) - 4}`).join(' ');
  return (
    <svg width="100%" viewBox={`0 0 ${width} ${h + 24}`} style={{ display: 'block' }}>
      <polygon points={`${pts(opens)} ${width},${h} 0,${h}`} fill="#e07d3c" opacity="0.1" />
      <polyline points={pts(opens)} fill="none" stroke="#e07d3c" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round" />
      <polygon points={`${pts(clicks)} ${width},${h} 0,${h}`} fill="#6366f1" opacity="0.1" />
      <polyline points={pts(clicks)} fill="none" stroke="#6366f1" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round" />
      {hours.map((label, i) => (
        <text key={i} x={(i / (hours.length - 1)) * width} y={h + 18} textAnchor="middle" fontSize={9} fill="#bbb">{label}</text>
      ))}
    </svg>
  );
};

const CampaignStatsScreen = ({ id, navigate }) => {
  const camp = MOCK_CAMPAIGNS.find(c => c.id === Number(id)) || MOCK_CAMPAIGNS[0];
  const links = [
    { url: 'https://example.com/product', clicks: 52, pct: '48.1%' },
    { url: 'https://example.com/blog/april', clicks: 31, pct: '28.7%' },
    { url: 'https://example.com/pricing', clicks: 18, pct: '16.7%' },
    { url: 'https://example.com/unsubscribe', clicks: 7, pct: '6.5%' },
  ];
  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <Button variant="ghost" size="sm" onClick={() => navigate('campaigns')} icon={<Icons.ChevronRight size={12} style={{ transform: 'rotate(180deg)' }} />}>Back to Campaigns</Button>

      {/* Header */}
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h2 style={{ fontSize: 18, fontWeight: 700, color: '#111', margin: '0 0 6px', letterSpacing: '-0.02em' }}>{camp.name}</h2>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', fontSize: 13, color: '#888' }}>
              <Badge status={camp.status} />
              <span>Sent {camp.date}</span>
              <span>·</span>
              <span>To: {camp.list}</span>
              <span>·</span>
              <span>{camp.sent?.toLocaleString() || 0} recipients</span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button variant="secondary" size="sm" icon={<Icons.Copy size={12} />}>Duplicate</Button>
            <Button variant="secondary" size="sm">View Email</Button>
          </div>
        </div>
      </Card>

      {/* Stat cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 12 }}>
        {[
          { label: 'Sent', value: camp.sent?.toLocaleString() || '3,842' },
          { label: 'Delivered', value: '3,828', sub: '99.6%' },
          { label: 'Opens', value: '1,812', sub: camp.opens },
          { label: 'Unique Opens', value: '1,681', sub: '43.8%' },
          { label: 'Clicks', value: '311', sub: camp.clicks },
          { label: 'Bounces', value: '14', sub: '0.4%' },
        ].map(s => <StatCard key={s.label} label={s.label} value={s.value} sub={s.sub} />)}
      </div>

      {/* Timeline chart */}
      <Card>
        <SectionTitle>
          Opens &amp; Clicks Over Time
          <div style={{ display: 'flex', gap: 14, float: 'right', marginTop: -2 }}>
            <span style={{ fontSize: 12, color: '#e07d3c', fontWeight: 500, display: 'flex', alignItems: 'center', gap: 5 }}><span style={{ width: 12, height: 2, background: '#e07d3c', display: 'inline-block', borderRadius: 2 }} /> Opens</span>
            <span style={{ fontSize: 12, color: '#6366f1', fontWeight: 500, display: 'flex', alignItems: 'center', gap: 5 }}><span style={{ width: 12, height: 2, background: '#6366f1', display: 'inline-block', borderRadius: 2 }} /> Clicks</span>
          </div>
        </SectionTitle>
        <TimelineChart />
      </Card>

      {/* Link breakdown */}
      <Card>
        <SectionTitle>Link Click Breakdown</SectionTitle>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {links.map((link, i) => (
            <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                <a href="#" style={{ color: '#6366f1', textDecoration: 'none', fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{link.url}</a>
                <div style={{ display: 'flex', gap: 12, color: '#555' }}>
                  <span style={{ fontWeight: 600 }}>{link.clicks} clicks</span>
                  <span style={{ color: '#aaa', minWidth: 40, textAlign: 'right' }}>{link.pct}</span>
                </div>
              </div>
              <div style={{ height: 5, background: '#f0ede6', borderRadius: 3 }}>
                <div style={{ height: '100%', width: link.pct, background: i === 0 ? '#e07d3c' : i === 1 ? '#6366f1' : i === 2 ? '#059669' : '#e8e5de', borderRadius: 3, transition: 'width 600ms ease' }} />
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
};

Object.assign(window, { CampaignsScreen, CampaignCreateScreen, CampaignStatsScreen });

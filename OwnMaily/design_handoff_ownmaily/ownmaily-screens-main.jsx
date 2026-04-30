// Dashboard + Analytics screens

const MOCK_CAMPAIGNS = [
  { id: 1, name: 'April Product Update', status: 'sent', list: 'All Subscribers', date: 'Apr 22, 2026', opens: '47.2%', clicks: '8.1%', sent: 3842 },
  { id: 2, name: 'Weekly Digest #18', status: 'sent', list: 'Weekly Digest', date: 'Apr 15, 2026', opens: '51.8%', clicks: '11.4%', sent: 2109 },
  { id: 3, name: 'May Launch Teaser', status: 'scheduled', list: 'Early Access', date: 'May 1, 2026', opens: '—', clicks: '—', sent: 0 },
  { id: 4, name: 'Re-engagement Series', status: 'draft', list: 'Inactive 90d', date: '—', opens: '—', clicks: '—', sent: 0 },
  { id: 5, name: 'March Roundup', status: 'sent', list: 'All Subscribers', date: 'Mar 29, 2026', opens: '44.1%', clicks: '7.3%', sent: 3801 },
];

const MOCK_SUBSCRIBERS = [
  { id: 1, email: 'ada@lovelace.io', name: 'Ada Lovelace', status: 'active', created: 'Jan 12, 2026', tags: ['Early Access', 'Beta'], opens: 38, clicks: 14 },
  { id: 2, email: 'grace@hopper.dev', name: 'Grace Hopper', status: 'active', created: 'Feb 3, 2026', tags: ['Weekly Digest'], opens: 27, clicks: 9 },
  { id: 3, email: 'alan@turing.net', name: 'Alan Turing', status: 'unsubscribed', created: 'Nov 8, 2025', tags: [], opens: 5, clicks: 1 },
  { id: 4, email: 'linus@kernel.org', name: 'Linus Torvalds', status: 'active', created: 'Mar 19, 2026', tags: ['Early Access'], opens: 42, clicks: 19 },
  { id: 5, email: 'margaret@hamilton.io', name: 'Margaret Hamilton', status: 'bounced', created: 'Dec 1, 2025', tags: ['Beta'], opens: 0, clicks: 0 },
  { id: 6, email: 'tim@berners.web', name: 'Tim Berners-Lee', status: 'active', created: 'Apr 2, 2026', tags: ['Weekly Digest', 'Early Access'], opens: 15, clicks: 6 },
  { id: 7, email: 'katherine@johnson.space', name: 'Katherine Johnson', status: 'active', created: 'Apr 10, 2026', tags: ['Beta'], opens: 8, clicks: 3 },
];

const MOCK_LISTS = [
  { id: 1, name: 'All Subscribers', count: 4218, optin: 'single_optin', created: 'Jan 1, 2026' },
  { id: 2, name: 'Weekly Digest', count: 2109, optin: 'double_optin', created: 'Jan 5, 2026' },
  { id: 3, name: 'Early Access', count: 312, optin: 'double_optin', created: 'Feb 14, 2026' },
  { id: 4, name: 'Inactive 90d', count: 489, optin: 'single_optin', created: 'Mar 1, 2026' },
  { id: 5, name: 'Beta Testers', count: 78, optin: 'double_optin', created: 'Apr 1, 2026' },
];

const MOCK_TAGS = [
  { id: 1, name: 'Early Access', count: 312, color: '#e07d3c' },
  { id: 2, name: 'Beta', count: 78, color: '#8b5cf6' },
  { id: 3, name: 'Weekly Digest', count: 2109, color: '#0ea5e9' },
  { id: 4, name: 'VIP', count: 44, color: '#f59e0b' },
  { id: 5, name: 'Inactive', count: 489, color: '#6b7280' },
  { id: 6, name: 'Conference 2026', count: 203, color: '#10b981' },
];

// Mini sparkline SVG
const Sparkline = ({ data, color = '#e07d3c', width = 120, height = 36 }) => {
  const max = Math.max(...data), min = Math.min(...data);
  const range = max - min || 1;
  const pts = data.map((v, i) => `${(i / (data.length - 1)) * width},${height - ((v - min) / range) * (height - 4) - 2}`).join(' ');
  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} style={{ overflow: 'visible' }}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <polygon points={`${pts} ${width},${height} 0,${height}`} fill={color} opacity="0.08" />
    </svg>
  );
};

// Bar chart for analytics
const BarChart = ({ data, color = '#e07d3c', height = 120 }) => {
  const max = Math.max(...data.map(d => d.value));
  return (
    <svg width="100%" height={height} viewBox={`0 0 ${data.length * 32} ${height}`} preserveAspectRatio="none" style={{ display: 'block' }}>
      {data.map((d, i) => {
        const barH = (d.value / max) * (height - 20);
        return (
          <g key={i}>
            <rect x={i * 32 + 4} y={height - barH - 20} width={24} height={barH} rx={3} fill={color} opacity={0.8} />
            <text x={i * 32 + 16} y={height - 4} textAnchor="middle" fontSize={9} fill="#aaa">{d.label}</text>
          </g>
        );
      })}
    </svg>
  );
};

const openSparkData = [28, 35, 42, 38, 51, 47, 44, 52, 48, 55, 47, 51];
const growthData = [3801, 3842, 3900, 3940, 3990, 4050, 4100, 4130, 4180, 4210, 4218];

const DashboardScreen = ({ navigate }) => {
  const recentCols = [
    { key: 'name', label: 'Campaign', render: (v, row) => (
      <div>
        <div style={{ fontWeight: 500, color: '#111', marginBottom: 1 }}>{v}</div>
        <div style={{ fontSize: 11, color: '#aaa' }}>{row.list}</div>
      </div>
    )},
    { key: 'status', label: 'Status', render: v => <Badge status={v} /> },
    { key: 'date', label: 'Sent', muted: true },
    { key: 'opens', label: 'Opens', render: v => <span style={{ fontWeight: v !== '—' ? 600 : 400, color: v !== '—' ? '#111' : '#ccc' }}>{v}</span> },
    { key: 'clicks', label: 'Clicks', render: v => <span style={{ fontWeight: v !== '—' ? 600 : 400, color: v !== '—' ? '#111' : '#ccc' }}>{v}</span> },
  ];

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 22 }}>
      {/* Stat cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
        <StatCard label="Total Subscribers" value="4,218" delta="+88" sub="this month" icon={<Icons.Users size={16} />} />
        <StatCard label="Campaigns Sent" value="24" sub="all time" icon={<Icons.Mail size={16} />} />
        <StatCard label="Avg. Open Rate" value="47.7%" delta="+2.3%" sub="vs last month" icon={<Icons.Eye size={16} />} />
        <StatCard label="Avg. Click Rate" value="8.9%" delta="+0.4%" sub="vs last month" icon={<Icons.Activity size={16} />} />
      </div>

      {/* Charts row */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <Card>
          <SectionTitle>Subscriber Growth</SectionTitle>
          <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16 }}>
            <div>
              <div style={{ fontSize: 28, fontWeight: 700, color: '#111', letterSpacing: '-0.03em' }}>4,218</div>
              <div style={{ fontSize: 12, color: '#059669', fontWeight: 500 }}>↑ 88 new this month</div>
            </div>
            <div style={{ flex: 1, overflow: 'hidden' }}>
              <Sparkline data={growthData} width={220} height={52} />
            </div>
          </div>
        </Card>
        <Card>
          <SectionTitle>Opens vs Clicks <span style={{ color: '#bbb', fontSize: 11, fontWeight: 400 }}>(last 12 campaigns)</span></SectionTitle>
          <div style={{ display: 'flex', gap: 18, marginBottom: 8 }}>
            <div><span style={{ fontSize: 20, fontWeight: 700, color: '#e07d3c' }}>47.7%</span> <span style={{ fontSize: 11, color: '#aaa' }}>opens</span></div>
            <div><span style={{ fontSize: 20, fontWeight: 700, color: '#6366f1' }}>8.9%</span> <span style={{ fontSize: 11, color: '#aaa' }}>clicks</span></div>
          </div>
          <Sparkline data={openSparkData} color="#e07d3c" width="100%" height={44} />
        </Card>
      </div>

      {/* Recent campaigns */}
      <Card padding={null}>
        <div style={{ padding: '16px 22px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <SectionTitle>Recent Campaigns</SectionTitle>
          <Button variant="ghost" size="sm" onClick={() => navigate('campaigns')}>View all →</Button>
        </div>
        <Table columns={recentCols} rows={MOCK_CAMPAIGNS.slice(0, 4)} onRowClick={row => navigate('campaigns/stats/' + row.id)} />
      </Card>
    </div>
  );
};

const weeklyBarData = [
  { label: 'Mon', value: 420 }, { label: 'Tue', value: 380 }, { label: 'Wed', value: 510 },
  { label: 'Thu', value: 450 }, { label: 'Fri', value: 490 }, { label: 'Sat', value: 200 }, { label: 'Sun', value: 160 },
];

const AnalyticsScreen = ({ navigate }) => {
  const [period, setPeriod] = React.useState('30d');
  const periodOpts = [{ id: '7d', label: '7 days' }, { id: '30d', label: '30 days' }, { id: '90d', label: '90 days' }, { id: 'all', label: 'All time' }];

  const campCols = [
    { key: 'name', label: 'Campaign', render: (v, row) => <span style={{ fontWeight: 500 }}>{v}</span> },
    { key: 'status', label: 'Status', render: v => <Badge status={v} /> },
    { key: 'sent', label: 'Sent', render: v => v ? v.toLocaleString() : '—', muted: true },
    { key: 'opens', label: 'Open Rate', render: v => <span style={{ fontWeight: 600, color: v !== '—' ? '#059669' : '#ccc' }}>{v}</span> },
    { key: 'clicks', label: 'Click Rate', render: v => <span style={{ fontWeight: 600, color: v !== '—' ? '#6366f1' : '#ccc' }}>{v}</span> },
  ];

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 22 }}>
      {/* Period selector */}
      <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
        {periodOpts.map(o => (
          <button key={o.id} onClick={() => setPeriod(o.id)} style={{
            padding: '5px 12px', borderRadius: 6, border: '1px solid',
            borderColor: period === o.id ? '#e07d3c' : '#ddd',
            background: period === o.id ? '#e07d3c' : '#fff',
            color: period === o.id ? '#fff' : '#666', fontSize: 12, fontWeight: 500,
            cursor: 'pointer', fontFamily: 'inherit', transition: 'all 120ms',
          }}>{o.label}</button>
        ))}
      </div>

      {/* Stat cards — expanded */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
        <StatCard label="Total Subscribers" value="4,218" delta="+88" sub="this period" icon={<Icons.Users size={16} />} />
        <StatCard label="Emails Sent" value="9,751" sub="this period" icon={<Icons.Mail size={16} />} />
        <StatCard label="Avg. Open Rate" value="47.7%" delta="+2.3%" sub="vs prior period" icon={<Icons.Eye size={16} />} />
        <StatCard label="Avg. Click Rate" value="8.9%" delta="+0.4%" sub="vs prior period" icon={<Icons.Activity size={16} />} />
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
        <StatCard label="Unsubscribes" value="23" delta="-8" sub="vs prior period" icon={<Icons.X size={16} />} />
        <StatCard label="Bounces" value="14" delta="-3" sub="vs prior period" icon={<Icons.AlertCircle size={16} />} />
        <StatCard label="New Subscribers" value="312" delta="+41" sub="vs prior period" icon={<Icons.TrendingUp size={16} />} />
        <StatCard label="Deliverability" value="99.6%" sub="delivery rate" icon={<Icons.Zap size={16} />} />
      </div>

      {/* Charts */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 14 }}>
        <Card>
          <SectionTitle>Opens by Day of Week</SectionTitle>
          <BarChart data={weeklyBarData} height={120} />
        </Card>
        <Card>
          <SectionTitle>Subscriber Breakdown</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 4 }}>
            {[['Active', 3842, '#059669'], ['Unsubscribed', 312, '#6b7280'], ['Bounced', 64, '#dc2626']].map(([label, count, color]) => (
              <div key={label} style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                  <span style={{ color: '#555' }}>{label}</span>
                  <span style={{ fontWeight: 600, color: '#333' }}>{count.toLocaleString()}</span>
                </div>
                <div style={{ height: 6, background: '#f0ede6', borderRadius: 3 }}>
                  <div style={{ height: '100%', width: `${(count / 4218) * 100}%`, background: color, borderRadius: 3 }} />
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Campaign perf table */}
      <Card padding={null}>
        <div style={{ padding: '16px 22px 0' }}>
          <SectionTitle>Campaign Performance</SectionTitle>
        </div>
        <Table columns={campCols} rows={MOCK_CAMPAIGNS.filter(c => c.status === 'sent')} onRowClick={row => navigate('campaigns/stats/' + row.id)} />
      </Card>
    </div>
  );
};

Object.assign(window, { DashboardScreen, AnalyticsScreen, MOCK_CAMPAIGNS, MOCK_SUBSCRIBERS, MOCK_LISTS, MOCK_TAGS, Sparkline, BarChart });

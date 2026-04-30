// Subscriber screens: List, Detail, Lists, ListDetail, Tags

const SubscribersScreen = ({ navigate }) => {
  const [search, setSearch] = React.useState('');
  const [tab, setTab] = React.useState('all');
  const [page, setPage] = React.useState(1);
  const PER_PAGE = 5;

  const filtered = MOCK_SUBSCRIBERS.filter(s => {
    const matchTab = tab === 'all' || s.status === tab;
    const q = search.toLowerCase();
    const matchSearch = !q || s.email.includes(q) || s.name.toLowerCase().includes(q);
    return matchTab && matchSearch;
  });

  const paged = filtered.slice((page - 1) * PER_PAGE, page * PER_PAGE);

  const cols = [
    { key: 'email', label: 'Email / Name', render: (v, row) => (
      <div>
        <div style={{ fontWeight: 500, color: '#111' }}>{v}</div>
        <div style={{ fontSize: 11, color: '#aaa' }}>{row.name}</div>
      </div>
    )},
    { key: 'status', label: 'Status', render: v => <Badge status={v} /> },
    { key: 'tags', label: 'Tags', render: v => (
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        {v.map(t => <Pill key={t} label={t} />)}
      </div>
    )},
    { key: 'created', label: 'Added', muted: true },
  ];

  const tabs = [
    { id: 'all', label: 'All', count: MOCK_SUBSCRIBERS.length },
    { id: 'active', label: 'Active', count: MOCK_SUBSCRIBERS.filter(s => s.status === 'active').length },
    { id: 'unsubscribed', label: 'Unsubscribed', count: MOCK_SUBSCRIBERS.filter(s => s.status === 'unsubscribed').length },
    { id: 'bounced', label: 'Bounced', count: MOCK_SUBSCRIBERS.filter(s => s.status === 'bounced').length },
  ];

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>
      {/* Actions bar */}
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <Icons.Search size={14} style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: '#bbb' }} />
          <input value={search} onChange={e => { setSearch(e.target.value); setPage(1); }}
            placeholder="Search by email or name…" style={{
              width: '100%', padding: '8px 12px 8px 32px', border: '1px solid #ddd',
              borderRadius: 7, fontSize: 13, fontFamily: 'inherit', outline: 'none',
              background: '#fff', color: '#111', boxSizing: 'border-box',
            }} />
        </div>
        <Button variant="secondary" size="md" icon={<Icons.Upload size={13} />}>Import CSV</Button>
        <Button variant="secondary" size="md" icon={<Icons.Download size={13} />}>Export</Button>
        <Button variant="primary" size="md" icon={<Icons.Plus size={13} />} onClick={() => navigate('subscribers/detail/new')}>Add Subscriber</Button>
      </div>

      <Card padding={null}>
        <div style={{ padding: '4px 14px 0' }}>
          <Tabs tabs={tabs} active={tab} onChange={t => { setTab(t); setPage(1); }} />
        </div>
        <Table columns={cols} rows={paged} onRowClick={row => navigate('subscribers/detail/' + row.id)} />
        {filtered.length > PER_PAGE && (
          <div style={{ padding: '0 14px 14px' }}>
            <Pagination page={page} total={filtered.length} perPage={PER_PAGE} onChange={setPage} />
          </div>
        )}
      </Card>
    </div>
  );
};

const SubscriberDetailScreen = ({ id, navigate }) => {
  const sub = MOCK_SUBSCRIBERS.find(s => s.id === Number(id)) || MOCK_SUBSCRIBERS[0];
  const timeline = [
    { date: 'Apr 22, 2026', event: 'Opened "April Product Update"', type: 'open' },
    { date: 'Apr 22, 2026', event: 'Clicked link in "April Product Update"', type: 'click' },
    { date: 'Apr 15, 2026', event: 'Opened "Weekly Digest #18"', type: 'open' },
    { date: 'Apr 1, 2026', event: 'Subscribed via landing page', type: 'join' },
    { date: 'Apr 1, 2026', event: 'Confirmed email (double opt-in)', type: 'confirm' },
  ];
  const typeColors = { open: '#e07d3c', click: '#6366f1', join: '#059669', confirm: '#0ea5e9' };
  const typeIcons = { open: Icons.Eye, click: Icons.Link, join: Icons.Users, confirm: Icons.Check };

  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <Button variant="ghost" size="sm" onClick={() => navigate('subscribers')} icon={<Icons.ChevronRight size={12} style={{ transform: 'rotate(180deg)' }} />}>Back to Subscribers</Button>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        {/* Subscriber info */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 18 }}>
              <div style={{ width: 48, height: 48, borderRadius: '50%', background: '#f0ede6', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, fontWeight: 700, color: '#e07d3c' }}>
                {sub.name.charAt(0)}
              </div>
              <div>
                <div style={{ fontSize: 16, fontWeight: 600, color: '#111' }}>{sub.name}</div>
                <div style={{ fontSize: 13, color: '#888' }}>{sub.email}</div>
              </div>
              <div style={{ marginLeft: 'auto' }}><Badge status={sub.status} /></div>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
              {[['Added', sub.created], ['Last Active', 'Apr 22, 2026'], ['Source', 'Landing page'], ['IP Country', 'US']].map(([k, v]) => (
                <div key={k} style={{ background: '#faf8f4', borderRadius: 7, padding: '9px 12px' }}>
                  <div style={{ fontSize: 10, color: '#aaa', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 2 }}>{k}</div>
                  <div style={{ fontSize: 13, color: '#333', fontWeight: 500 }}>{v}</div>
                </div>
              ))}
            </div>
          </Card>

          {/* Tags */}
          <Card>
            <SectionTitle action={<Button variant="ghost" size="sm" icon={<Icons.Plus size={11} />}>Add tag</Button>}>Tags</SectionTitle>
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
              {sub.tags.length ? sub.tags.map(t => {
                const tag = MOCK_TAGS.find(tg => tg.name === t);
                return <Pill key={t} label={t} color={tag?.color || '#e07d3c'} />;
              }) : <span style={{ fontSize: 13, color: '#bbb' }}>No tags assigned</span>}
            </div>
          </Card>

          {/* List memberships */}
          <Card>
            <SectionTitle>List Memberships</SectionTitle>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {MOCK_LISTS.slice(0, 2).map(list => (
                <div key={list.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 10px', background: '#faf8f4', borderRadius: 7 }}>
                  <div>
                    <div style={{ fontSize: 13, fontWeight: 500, color: '#333' }}>{list.name}</div>
                    <div style={{ fontSize: 11, color: '#aaa' }}>Subscribed {sub.created}</div>
                  </div>
                  <Badge status={list.optin} />
                </div>
              ))}
            </div>
          </Card>
        </div>

        {/* Right column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Engagement stats */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
            <StatCard label="Campaigns" value="12" sub="received" icon={<Icons.Mail size={14} />} />
            <StatCard label="Opens" value={sub.opens} sub="total" icon={<Icons.Eye size={14} />} />
            <StatCard label="Clicks" value={sub.clicks} sub="total" icon={<Icons.Activity size={14} />} />
          </div>

          {/* Activity timeline */}
          <Card style={{ flex: 1 }}>
            <SectionTitle>Activity Timeline</SectionTitle>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
              {timeline.map((evt, i) => {
                const IconComp = typeIcons[evt.type];
                return (
                  <div key={i} style={{ display: 'flex', gap: 12, paddingBottom: i < timeline.length - 1 ? 16 : 0, position: 'relative' }}>
                    {i < timeline.length - 1 && (
                      <div style={{ position: 'absolute', left: 15, top: 28, bottom: 0, width: 1, background: '#ece9e1' }} />
                    )}
                    <div style={{ width: 30, height: 30, borderRadius: '50%', background: typeColors[evt.type] + '18', border: `1px solid ${typeColors[evt.type]}30`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                      <IconComp size={12} color={typeColors[evt.type]} />
                    </div>
                    <div style={{ paddingTop: 4 }}>
                      <div style={{ fontSize: 13, color: '#333' }}>{evt.event}</div>
                      <div style={{ fontSize: 11, color: '#aaa', marginTop: 2 }}>{evt.date}</div>
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

const ListsScreen = ({ navigate }) => {
  const cols = [
    { key: 'name', label: 'List Name', render: (v) => <span style={{ fontWeight: 500, color: '#111' }}>{v}</span> },
    { key: 'count', label: 'Subscribers', render: v => <span style={{ fontWeight: 600 }}>{v.toLocaleString()}</span> },
    { key: 'optin', label: 'Opt-in Type', render: v => <Badge status={v} /> },
    { key: 'created', label: 'Created', muted: true },
  ];
  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <Button variant="primary" icon={<Icons.Plus size={13} />}>Create List</Button>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 8 }}>
        {MOCK_LISTS.slice(0, 3).map(list => (
          <div key={list.id} onClick={() => navigate('lists/detail/' + list.id)}
            style={{ background: '#fff', border: '1px solid #e8e5de', borderRadius: 10, padding: '18px 20px', cursor: 'pointer', transition: 'all 120ms' }}
            onMouseEnter={e => { e.currentTarget.style.borderColor = '#e07d3c'; e.currentTarget.style.boxShadow = '0 4px 16px rgba(224,125,60,0.08)'; }}
            onMouseLeave={e => { e.currentTarget.style.borderColor = '#e8e5de'; e.currentTarget.style.boxShadow = ''; }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
              <div style={{ width: 36, height: 36, borderRadius: 9, background: '#f0ede6', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Icons.Layers size={16} color="#e07d3c" />
              </div>
              <Badge status={list.optin} />
            </div>
            <div style={{ fontSize: 15, fontWeight: 600, color: '#111', marginBottom: 4 }}>{list.name}</div>
            <div style={{ fontSize: 22, fontWeight: 700, color: '#e07d3c' }}>{list.count.toLocaleString()}</div>
            <div style={{ fontSize: 11, color: '#aaa', marginTop: 2 }}>subscribers · created {list.created}</div>
          </div>
        ))}
      </div>
      <Card padding={null}>
        <div style={{ padding: '16px 22px 0' }}><SectionTitle>All Lists</SectionTitle></div>
        <Table columns={cols} rows={MOCK_LISTS} onRowClick={row => navigate('lists/detail/' + row.id)} />
      </Card>
    </div>
  );
};

const ListDetailScreen = ({ id, navigate }) => {
  const list = MOCK_LISTS.find(l => l.id === Number(id)) || MOCK_LISTS[0];
  const embedCode = `<form action="https://mail.example.com/lists/${list.id}/subscribe" method="POST">
  <input type="email" name="email" placeholder="you@example.com" required />
  <input type="text" name="name" placeholder="Your name" />
  <button type="submit">Subscribe</button>
</form>`;
  const cols = [
    { key: 'email', label: 'Email / Name', render: (v, row) => <div><div style={{ fontWeight: 500 }}>{v}</div><div style={{ fontSize: 11, color: '#aaa' }}>{row.name}</div></div> },
    { key: 'status', label: 'Status', render: v => <Badge status={v} /> },
    { key: 'created', label: 'Added', muted: true },
  ];
  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <Button variant="ghost" size="sm" onClick={() => navigate('lists')} icon={<Icons.ChevronRight size={12} style={{ transform: 'rotate(180deg)' }} />}>Back to Lists</Button>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 14 }}>
        <StatCard label="Total Subscribers" value={list.count.toLocaleString()} icon={<Icons.Users size={14} />} />
        <StatCard label="Opt-in Type" value={list.optin === 'double_optin' ? 'Double' : 'Single'} sub="opt-in" icon={<Icons.Shield size={14} />} />
        <StatCard label="Created" value={list.created} icon={<Icons.Calendar size={14} />} />
      </div>
      <Card padding={null}>
        <div style={{ padding: '16px 22px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <SectionTitle>{list.name} — Subscribers</SectionTitle>
          <Button variant="primary" size="sm" icon={<Icons.Plus size={12} />}>Add Subscriber</Button>
        </div>
        <Table columns={cols} rows={MOCK_SUBSCRIBERS.filter(s => s.status === 'active').slice(0, 5)} onRowClick={row => navigate('subscribers/detail/' + row.id)} />
      </Card>
      <Card>
        <SectionTitle action={<Button variant="ghost" size="sm" icon={<Icons.Copy size={11} />}>Copy</Button>}>Embed Subscribe Form</SectionTitle>
        <CodeBlock code={embedCode} />
        <p style={{ fontSize: 12, color: '#aaa', marginTop: 10, marginBottom: 0 }}>Paste this snippet into any HTML page. Style it to match your brand.</p>
      </Card>
    </div>
  );
};

const TagsScreen = ({ navigate }) => {
  const cols = [
    { key: 'name', label: 'Tag Name', render: (v, row) => (
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <div style={{ width: 8, height: 8, borderRadius: '50%', background: row.color, flexShrink: 0 }} />
        <span style={{ fontWeight: 500, color: '#111' }}>{v}</span>
      </div>
    )},
    { key: 'count', label: 'Subscribers', render: v => <span style={{ fontWeight: 600 }}>{v.toLocaleString()}</span> },
    { key: 'id', label: 'Actions', render: (v, row) => (
      <div style={{ display: 'flex', gap: 6 }}>
        <Button variant="ghost" size="sm" onClick={e => e.stopPropagation()}>Edit</Button>
        <Button variant="ghost" size="sm" onClick={e => e.stopPropagation()} icon={<Icons.Trash size={12} color="#dc2626" />} style={{ color: '#dc2626' }}>Delete</Button>
      </div>
    )},
  ];
  return (
    <div style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <Button variant="primary" icon={<Icons.Plus size={13} />}>Create Tag</Button>
      </div>
      <Card padding={null}>
        <div style={{ padding: '16px 22px 0' }}><SectionTitle>All Tags</SectionTitle></div>
        <Table columns={cols} rows={MOCK_TAGS} />
      </Card>
    </div>
  );
};

Object.assign(window, { SubscribersScreen, SubscriberDetailScreen, ListsScreen, ListDetailScreen, TagsScreen });

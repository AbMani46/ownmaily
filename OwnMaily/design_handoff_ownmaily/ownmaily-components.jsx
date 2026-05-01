// OwnMaily Shared UI Components

const BADGE_STYLES = {
  active:      { bg: '#d1fae5', color: '#065f46', label: 'Active' },
  sent:        { bg: '#d1fae5', color: '#065f46', label: 'Sent' },
  draft:       { bg: '#fef3c7', color: '#92400e', label: 'Draft' },
  scheduled:   { bg: '#dbeafe', color: '#1e40af', label: 'Scheduled' },
  pending:     { bg: '#fef3c7', color: '#92400e', label: 'Pending' },
  bounced:     { bg: '#fee2e2', color: '#991b1b', label: 'Bounced' },
  failed:      { bg: '#fee2e2', color: '#991b1b', label: 'Failed' },
  unsubscribed:{ bg: '#f3f4f6', color: '#4b5563', label: 'Unsubscribed' },
  double_optin:{ bg: '#ede9fe', color: '#5b21b6', label: 'Double Opt-in' },
  single_optin:{ bg: '#e0f2fe', color: '#075985', label: 'Single Opt-in' },
  suppressed:  { bg: '#fee2e2', color: '#991b1b', label: 'Suppressed' },
};

const Badge = ({ status, label, custom }) => {
  const style = BADGE_STYLES[status] || { bg: '#f3f4f6', color: '#4b5563', label: status };
  const text = label || custom || style.label;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      background: style.bg, color: style.color,
      fontSize: 11, fontWeight: 600, letterSpacing: '0.03em',
      padding: '2px 8px', borderRadius: 20, whiteSpace: 'nowrap',
      textTransform: 'uppercase',
    }}>
      <span style={{ width: 5, height: 5, borderRadius: '50%', background: style.color, flexShrink: 0 }} />
      {text}
    </span>
  );
};

const Button = ({ children, variant = 'primary', size = 'md', onClick, icon, disabled, style: extraStyle }) => {
  const [hov, setHov] = React.useState(false);
  const base = {
    display: 'inline-flex', alignItems: 'center', gap: 6,
    fontFamily: 'inherit', fontWeight: 500, cursor: disabled ? 'not-allowed' : 'pointer',
    border: 'none', borderRadius: 7, transition: 'all 120ms ease',
    opacity: disabled ? 0.5 : 1, outline: 'none', whiteSpace: 'nowrap',
  };
  const sizes = {
    sm: { fontSize: 12, padding: '5px 10px' },
    md: { fontSize: 13, padding: '7px 14px' },
    lg: { fontSize: 14, padding: '9px 18px' },
  };
  const variants = {
    primary: { background: hov ? '#c96a2e' : '#e07d3c', color: '#fff' },
    secondary: { background: hov ? '#f0ede6' : '#eae7e0', color: '#1a1a1a', border: '1px solid #d6d2c8' },
    ghost: { background: hov ? '#f0ede6' : 'transparent', color: '#4a4a4a', border: '1px solid transparent' },
    danger: { background: hov ? '#b91c1c' : '#dc2626', color: '#fff' },
    outline: { background: 'transparent', color: hov ? '#e07d3c' : '#4a4a4a', border: '1px solid ' + (hov ? '#e07d3c' : '#d6d2c8') },
  };
  return (
    <button style={{ ...base, ...sizes[size], ...variants[variant], ...extraStyle }}
      onClick={onClick} disabled={disabled}
      onMouseEnter={() => setHov(true)} onMouseLeave={() => setHov(false)}>
      {icon && icon}
      {children}
    </button>
  );
};

const StatCard = ({ label, value, sub, icon, delta, accent }) => (
  <div style={{
    background: '#fff', border: '1px solid #e8e5de',
    borderRadius: 10, padding: '20px 22px',
    display: 'flex', flexDirection: 'column', gap: 4,
  }}>
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
      <span style={{ fontSize: 12, color: '#888', fontWeight: 500, letterSpacing: '0.02em', textTransform: 'uppercase' }}>{label}</span>
      {icon && <span style={{ color: '#ccc' }}>{icon}</span>}
    </div>
    <div style={{ fontSize: 30, fontWeight: 700, color: accent || '#111', lineHeight: 1.1, marginTop: 4 }}>{value}</div>
    {(sub || delta) && (
      <div style={{ fontSize: 12, color: delta?.startsWith('+') ? '#059669' : delta ? '#dc2626' : '#888', marginTop: 2 }}>
        {delta && <span style={{ fontWeight: 600 }}>{delta} </span>}
        {sub}
      </div>
    )}
  </div>
);

const Input = ({ label, type = 'text', placeholder, value, onChange, hint, prefix, suffix, mono, disabled, fullWidth }) => {
  const [focused, setFocused] = React.useState(false);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 5, width: fullWidth ? '100%' : undefined }}>
      {label && <label style={{ fontSize: 12, fontWeight: 600, color: '#444', letterSpacing: '0.02em' }}>{label}</label>}
      <div style={{
        display: 'flex', alignItems: 'center',
        border: `1px solid ${focused ? '#e07d3c' : '#ddd'}`,
        borderRadius: 7, background: disabled ? '#f9f9f7' : '#fff',
        transition: 'border-color 120ms', overflow: 'hidden',
        boxShadow: focused ? '0 0 0 3px rgba(224,125,60,0.12)' : 'none',
      }}>
        {prefix && <span style={{ padding: '0 10px', color: '#999', fontSize: 13, borderRight: '1px solid #eee', background: '#fafaf8' }}>{prefix}</span>}
        <input type={type} placeholder={placeholder} value={value} onChange={onChange} disabled={disabled}
          style={{
            flex: 1, padding: '8px 11px', border: 'none', outline: 'none', fontSize: 13,
            fontFamily: mono ? "'JetBrains Mono', monospace" : 'inherit',
            background: 'transparent', color: '#111', minWidth: 0,
          }}
          onFocus={() => setFocused(true)} onBlur={() => setFocused(false)} />
        {suffix && <span style={{ padding: '0 10px', color: '#999', fontSize: 13, borderLeft: '1px solid #eee', background: '#fafaf8' }}>{suffix}</span>}
      </div>
      {hint && <span style={{ fontSize: 11, color: '#999' }}>{hint}</span>}
    </div>
  );
};

const Select = ({ label, value, onChange, options, fullWidth }) => {
  const [focused, setFocused] = React.useState(false);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 5, width: fullWidth ? '100%' : undefined }}>
      {label && <label style={{ fontSize: 12, fontWeight: 600, color: '#444', letterSpacing: '0.02em' }}>{label}</label>}
      <div style={{ position: 'relative' }}>
        <select value={value} onChange={onChange}
          style={{
            width: '100%', padding: '8px 32px 8px 11px', border: `1px solid ${focused ? '#e07d3c' : '#ddd'}`,
            borderRadius: 7, fontSize: 13, fontFamily: 'inherit', background: '#fff', color: '#111',
            outline: 'none', appearance: 'none', cursor: 'pointer',
            boxShadow: focused ? '0 0 0 3px rgba(224,125,60,0.12)' : 'none',
          }}
          onFocus={() => setFocused(true)} onBlur={() => setFocused(false)}>
          {options.map(o => <option key={o.value || o} value={o.value || o}>{o.label || o}</option>)}
        </select>
        <Icons.ChevronDown size={14} style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', color: '#999', pointerEvents: 'none' }} />
      </div>
    </div>
  );
};

const Table = ({ columns, rows, onRowClick }) => (
  <div style={{ overflowX: 'auto' }}>
    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
      <thead>
        <tr style={{ borderBottom: '1px solid #ece9e1' }}>
          {columns.map(col => (
            <th key={col.key} style={{
              textAlign: 'left', padding: '9px 14px', fontSize: 11, fontWeight: 600,
              color: '#888', letterSpacing: '0.05em', textTransform: 'uppercase',
              whiteSpace: 'nowrap',
            }}>{col.label}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, i) => (
          <tr key={i} onClick={() => onRowClick && onRowClick(row)}
            style={{
              borderBottom: '1px solid #f0ede6',
              cursor: onRowClick ? 'pointer' : 'default',
              transition: 'background 80ms',
            }}
            onMouseEnter={e => e.currentTarget.style.background = '#faf8f4'}
            onMouseLeave={e => e.currentTarget.style.background = ''}>
            {columns.map(col => (
              <td key={col.key} style={{ padding: '11px 14px', color: col.muted ? '#888' : '#222', ...col.style }}>
                {col.render ? col.render(row[col.key], row) : row[col.key]}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);

const Card = ({ children, style: extra, padding = '20px 22px' }) => (
  <div style={{ background: '#fff', border: '1px solid #e8e5de', borderRadius: 10, overflow: 'hidden', ...extra }}>
    {padding ? <div style={{ padding }}>{children}</div> : children}
  </div>
);

const SectionTitle = ({ children, action }) => (
  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
    <h3 style={{ fontSize: 13, fontWeight: 600, color: '#555', letterSpacing: '0.04em', textTransform: 'uppercase', margin: 0 }}>{children}</h3>
    {action}
  </div>
);

const Tabs = ({ tabs, active, onChange }) => (
  <div style={{ display: 'flex', gap: 2, borderBottom: '1px solid #e8e5de', marginBottom: 18 }}>
    {tabs.map(t => (
      <button key={t.id} onClick={() => onChange(t.id)} style={{
        padding: '8px 14px', fontSize: 13, fontWeight: active === t.id ? 600 : 400,
        color: active === t.id ? '#e07d3c' : '#666',
        background: 'none', border: 'none', cursor: 'pointer', fontFamily: 'inherit',
        borderBottom: `2px solid ${active === t.id ? '#e07d3c' : 'transparent'}`,
        marginBottom: -1, transition: 'all 120ms',
      }}>{t.label}{t.count !== undefined && <span style={{ marginLeft: 6, fontSize: 11, background: '#f0ede6', padding: '1px 6px', borderRadius: 10, color: '#888' }}>{t.count}</span>}</button>
    ))}
  </div>
);

const Pill = ({ label, color = '#e07d3c' }) => (
  <span style={{
    display: 'inline-block', padding: '2px 10px', borderRadius: 20,
    fontSize: 11, fontWeight: 500, background: color + '18', color: color,
    border: `1px solid ${color}30`,
  }}>{label}</span>
);

const CodeBlock = ({ code }) => (
  <pre style={{
    background: '#0c0c0f', color: '#9aefbc', borderRadius: 8, padding: '14px 16px',
    fontSize: 12, fontFamily: "'JetBrains Mono', monospace", overflowX: 'auto',
    lineHeight: 1.6, margin: 0, border: '1px solid #222',
  }}>{code}</pre>
);

const Pagination = ({ page, total, perPage, onChange }) => {
  const pages = Math.ceil(total / perPage);
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'flex-end', paddingTop: 14 }}>
      <span style={{ fontSize: 12, color: '#888' }}>Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}</span>
      <button onClick={() => onChange(page - 1)} disabled={page === 1} style={{ padding: '4px 10px', borderRadius: 6, border: '1px solid #ddd', background: '#fff', cursor: page === 1 ? 'not-allowed' : 'pointer', opacity: page === 1 ? 0.4 : 1, fontSize: 12 }}>←</button>
      <button onClick={() => onChange(page + 1)} disabled={page === pages} style={{ padding: '4px 10px', borderRadius: 6, border: '1px solid #ddd', background: '#fff', cursor: page === pages ? 'not-allowed' : 'pointer', opacity: page === pages ? 0.4 : 1, fontSize: 12 }}>→</button>
    </div>
  );
};

Object.assign(window, { Badge, Button, StatCard, Input, Select, Table, Card, SectionTitle, Tabs, Pill, CodeBlock, Pagination, BADGE_STYLES });

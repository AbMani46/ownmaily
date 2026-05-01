// Auth screens: Login + Setup Wizard

const LoginScreen = ({ navigate }) => {
  const [email, setEmail] = React.useState('admin@example.com');
  const [password, setPassword] = React.useState('');
  const [showPw, setShowPw] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  const handleLogin = () => {
    setLoading(true);
    setTimeout(() => { setLoading(false); navigate('dashboard'); }, 800);
  };

  return (
    <div style={{
      minHeight: '100vh', background: '#f4f3ef', display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontFamily: "'DM Sans', sans-serif",
    }}>
      <div style={{ width: '100%', maxWidth: 380 }}>
        {/* Logo */}
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 44, height: 44, borderRadius: 12, background: '#e07d3c',
            display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 12px',
          }}>
            <Icons.Inbox size={22} color="#fff" />
          </div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#111', margin: '0 0 4px', letterSpacing: '-0.03em' }}>OwnMaily</h1>
          <p style={{ fontSize: 13, color: '#888', margin: 0 }}>Your self-hosted email command center</p>
        </div>

        {/* Card */}
        <div style={{ background: '#fff', border: '1px solid #e8e5de', borderRadius: 12, padding: '28px 28px', boxShadow: '0 4px 24px rgba(0,0,0,0.06)' }}>
          <h2 style={{ fontSize: 16, fontWeight: 600, color: '#111', margin: '0 0 22px', letterSpacing: '-0.02em' }}>Sign in to your account</h2>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Input label="Email address" type="email" placeholder="you@example.com" value={email} onChange={e => setEmail(e.target.value)} fullWidth />

            <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
              <label style={{ fontSize: 12, fontWeight: 600, color: '#444', letterSpacing: '0.02em' }}>Password</label>
              <div style={{ position: 'relative' }}>
                <input type={showPw ? 'text' : 'password'} placeholder="••••••••" value={password}
                  onChange={e => setPassword(e.target.value)}
                  style={{ width: '100%', padding: '8px 36px 8px 11px', border: '1px solid #ddd', borderRadius: 7, fontSize: 13, fontFamily: 'inherit', outline: 'none', boxSizing: 'border-box', color: '#111' }} />
                <button onClick={() => setShowPw(v => !v)} style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: '#aaa', padding: 0 }}>
                  {showPw ? <Icons.EyeOff size={14} /> : <Icons.Eye size={14} />}
                </button>
              </div>
            </div>

            <Button variant="primary" size="lg" onClick={handleLogin} disabled={loading} style={{ width: '100%', justifyContent: 'center', marginTop: 4 }}>
              {loading ? 'Signing in…' : 'Sign in'}
            </Button>
          </div>
        </div>

        <p style={{ textAlign: 'center', fontSize: 12, color: '#bbb', marginTop: 20 }}>
          OwnMaily v1.0 · Self-hosted · <span style={{ color: '#e07d3c', cursor: 'pointer' }} onClick={() => navigate('setup-wizard')}>Run setup wizard</span>
        </p>
      </div>
    </div>
  );
};

const WIZARD_STEPS = [
  { id: 'welcome', title: 'Welcome', icon: 'Sparkles' },
  { id: 'account', title: 'Owner Account', icon: 'Users' },
  { id: 'general', title: 'General Settings', icon: 'Globe' },
  { id: 'smtp', title: 'Connect SMTP', icon: 'Server' },
  { id: 'test', title: 'Send Test Email', icon: 'Send' },
];

const SetupWizard = ({ navigate }) => {
  const [step, setStep] = React.useState(0);
  const [done, setDone] = React.useState(false);
  const current = WIZARD_STEPS[step];

  const next = () => { if (step < WIZARD_STEPS.length - 1) setStep(s => s + 1); else { setDone(true); setTimeout(() => navigate('dashboard'), 1200); } };
  const prev = () => setStep(s => s - 1);

  const StepContent = () => {
    if (step === 0) return (
      <div style={{ textAlign: 'center', padding: '20px 0' }}>
        <div style={{ fontSize: 48, marginBottom: 16 }}>
          <Icons.Sparkles size={52} color="#e07d3c" />
        </div>
        <h3 style={{ fontSize: 20, fontWeight: 700, color: '#111', margin: '0 0 10px' }}>Welcome to OwnMaily</h3>
        <p style={{ color: '#666', fontSize: 14, lineHeight: 1.6, maxWidth: 360, margin: '0 auto' }}>
          You're minutes away from owning your email list. No SaaS, no limits, no monthly fees. Let's get you set up.
        </p>
      </div>
    );
    if (step === 1) return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Input label="Full name" placeholder="Ada Lovelace" fullWidth />
        <Input label="Email address" placeholder="ada@example.com" fullWidth />
        <Input label="Password" type="password" placeholder="Choose a strong password" fullWidth />
        <Input label="Confirm password" type="password" placeholder="Repeat password" fullWidth />
      </div>
    );
    if (step === 2) return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Input label="Site name" placeholder="My Newsletter" defaultValue="My Newsletter" fullWidth />
        <Input label="Installation URL" prefix="https://" placeholder="mail.example.com" fullWidth />
        <Select label="Timezone" options={['UTC', 'America/New_York', 'America/Los_Angeles', 'Europe/London', 'Europe/Paris', 'Asia/Tokyo']} fullWidth />
        <Input label="Physical address" placeholder="123 Main St, City, Country" hint="Required for CAN-SPAM compliance" fullWidth />
      </div>
    );
    if (step === 3) return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Select label="SMTP provider" options={['Resend', 'Mailgun', 'Amazon SES', 'Custom SMTP']} fullWidth />
        <Input label="API key / SMTP password" type="password" placeholder="re_••••••••" mono fullWidth />
        <Input label="From email" placeholder="hello@example.com" fullWidth />
        <Input label="From name" placeholder="My Newsletter" fullWidth />
      </div>
    );
    if (step === 4) return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14, alignItems: 'center', textAlign: 'center' }}>
        <Icons.Send size={40} color="#e07d3c" />
        <p style={{ color: '#555', fontSize: 14, margin: 0 }}>We'll send a test email to your address to confirm everything is working.</p>
        <Input label="Send test to" placeholder="ada@example.com" fullWidth />
        {done && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: '#059669', fontSize: 13, fontWeight: 500 }}>
            <Icons.Check size={16} color="#059669" /> Test email sent! Taking you to your dashboard…
          </div>
        )}
      </div>
    );
  };

  return (
    <div style={{ minHeight: '100vh', background: '#f4f3ef', display: 'flex', alignItems: 'center', justifyContent: 'center', fontFamily: "'DM Sans', sans-serif", padding: 20 }}>
      <div style={{ width: '100%', maxWidth: 520 }}>
        {/* Logo */}
        <div style={{ textAlign: 'center', marginBottom: 28 }}>
          <div style={{ width: 36, height: 36, borderRadius: 9, background: '#e07d3c', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 10px' }}>
            <Icons.Inbox size={17} color="#fff" />
          </div>
          <span style={{ fontSize: 14, fontWeight: 700, color: '#111', letterSpacing: '-0.02em' }}>OwnMaily Setup</span>
        </div>

        {/* Step progress */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 0, marginBottom: 28 }}>
          {WIZARD_STEPS.map((s, i) => {
            const isComplete = i < step;
            const isCurrent = i === step;
            return (
              <React.Fragment key={s.id}>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
                  <div style={{
                    width: 32, height: 32, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center',
                    background: isComplete ? '#e07d3c' : isCurrent ? '#fff' : '#e8e5de',
                    border: `2px solid ${isComplete ? '#e07d3c' : isCurrent ? '#e07d3c' : '#e8e5de'}`,
                    color: isComplete ? '#fff' : isCurrent ? '#e07d3c' : '#aaa',
                    fontSize: 12, fontWeight: 700, transition: 'all 200ms',
                  }}>
                    {isComplete ? <Icons.Check size={14} color="#fff" /> : i + 1}
                  </div>
                  <span style={{ fontSize: 10, color: isCurrent ? '#e07d3c' : '#aaa', fontWeight: isCurrent ? 600 : 400, whiteSpace: 'nowrap' }}>{s.title}</span>
                </div>
                {i < WIZARD_STEPS.length - 1 && (
                  <div style={{ height: 2, flex: 1, background: i < step ? '#e07d3c' : '#e8e5de', margin: '0 4px', marginBottom: 18, transition: 'background 300ms' }} />
                )}
              </React.Fragment>
            );
          })}
        </div>

        {/* Card */}
        <div style={{ background: '#fff', border: '1px solid #e8e5de', borderRadius: 12, padding: '28px', boxShadow: '0 4px 24px rgba(0,0,0,0.06)' }}>
          <div style={{ marginBottom: 22 }}>
            <h2 style={{ fontSize: 16, fontWeight: 600, color: '#111', margin: '0 0 4px', letterSpacing: '-0.02em' }}>
              Step {step + 1}: {current.title}
            </h2>
            <div style={{ height: 2, background: '#f0ede6', borderRadius: 2, marginTop: 10 }}>
              <div style={{ height: '100%', width: `${((step + 1) / WIZARD_STEPS.length) * 100}%`, background: '#e07d3c', borderRadius: 2, transition: 'width 300ms ease' }} />
            </div>
          </div>

          <StepContent />

          <div style={{ display: 'flex', gap: 8, marginTop: 24, justifyContent: 'space-between' }}>
            <Button variant="ghost" onClick={step === 0 ? () => navigate('login') : prev}>{step === 0 ? 'Back to login' : '← Previous'}</Button>
            <Button variant="primary" onClick={next}>{step === WIZARD_STEPS.length - 1 ? 'Finish Setup' : 'Continue →'}</Button>
          </div>
        </div>
      </div>
    </div>
  );
};

Object.assign(window, { LoginScreen, SetupWizard });

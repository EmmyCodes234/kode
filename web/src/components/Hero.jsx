import { useState } from 'react'

const snippets = {
  curl: 'curl -fsSL https://trykode.xyz/install.sh | bash',
  go: 'go install github.com/kode/kode/cmd/kode@latest',
  npm: 'npm install -g kode',
}

export default function Hero() {
  const [tab, setTab] = useState('curl')
  const [copied, setCopied] = useState(null)

  const handleCopy = (key) => {
    navigator.clipboard?.writeText(snippets[key]).then(() => {
      setCopied(key)
      setTimeout(() => setCopied(null), 1500)
    })
  }

  return (
    <section style={{ padding: '152px 0 96px' }}>
      <div className="wrapper">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 48, alignItems: 'start' }}>
          <div style={{ minWidth: 0 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 8,
              background: 'var(--surface-dark)', color: 'var(--on-dark)',
              padding: '2px 8px', borderRadius: 4,
              fontSize: 14, lineHeight: 2, marginBottom: 24,
            }}>[+] Desktop beta available on macOS, Windows, Linux</span>
            <h1 className="display-xl" style={{ marginBottom: 16 }}>
              The safe AI<br />coding agent
            </h1>
            <p className="body-md" style={{ marginBottom: 32, maxWidth: '90%' }}>
              Connect any model from any provider, including Claude, GPT, Gemini, local models, and more.
            </p>

            <div style={{ display: 'flex', gap: 0, marginBottom: 0 }}>
              {['curl', 'go', 'npm'].map(key => (
                <button key={key} onClick={() => setTab(key)}
                  style={{
                    padding: '8px 16px', fontSize: 16, fontWeight: 500, lineHeight: 2,
                    background: 'transparent', border: 'none', cursor: 'pointer',
                    color: tab === key ? 'var(--ink)' : 'var(--mute)',
                    fontFamily: 'var(--font-mono)',
                    borderBottom: tab === key ? '2px solid var(--ink)' : '1px solid var(--hairline)',
                  }}>
                  {key}
                </button>
              ))}
            </div>
            <div style={{
              background: 'var(--surface-card)', color: 'var(--ink)',
              padding: '12px 16px', borderRadius: 4,
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              fontSize: 15, marginBottom: 0,
            }}>
              <code>{snippets[tab]}</code>
              <button onClick={() => handleCopy(tab)}
                style={{
                  background: 'none', border: '1px solid var(--hairline)', borderRadius: 4,
                  padding: '4px 12px', cursor: 'pointer', fontFamily: 'var(--font-mono)',
                  fontSize: 14, color: 'var(--mute)', lineHeight: 2,
                }}
                onMouseEnter={e => { e.target.style.borderColor = 'var(--ink)'; e.target.style.color = 'var(--ink)' }}
                onMouseLeave={e => { e.target.style.borderColor = 'var(--hairline)'; e.target.style.color = 'var(--mute)' }}>
                {copied === tab ? 'Copied' : 'Copy'}
              </button>
            </div>
          </div>

          <div style={{ minWidth: 0 }}>
            <div style={{
              background: 'var(--surface-dark)',
              padding: '32px 24px',
            }}>
              <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
                <img src="/kode-logo.svg" alt="Kode" width="160" height="35" />
              </div>
              <div style={{
                background: 'var(--surface-dark-elevated)',
                padding: '8px 12px', borderRadius: 4,
                fontSize: 14, color: 'var(--on-dark)', marginBottom: 16,
              }}>
                <span style={{ color: 'var(--on-dark-mute)' }}>|</span>{' '}
                <span>kode loop &quot;add user authentication&quot;  </span>
                <span style={{ color: 'var(--accent)' }}>Claude Opus 4.5</span>
              </div>
              <div style={{ textAlign: 'center', fontSize: 13, color: 'var(--ash)', lineHeight: 2 }}>
                <span style={{ margin: '0 12px' }}>tab  switch agent</span>
                <span style={{ margin: '0 12px' }}>ctrl-p  commands</span>
                <span style={{ margin: '0 12px' }}>KODE_BIN  gatekeeper active</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

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
    <section className="hero-section" id="install">
      <div className="wrapper">
        <div className="hero-grid">
          <div className="hero-left">
            <span className="hero-badge">[+] Desktop beta available on macOS, Windows, Linux</span>
            <h1 className="display-xl" style={{ marginBottom: 16 }}>
              The safe AI<br />coding agent
            </h1>
            <p className="body-md hero-sub">
              Connect any model from any provider, including Claude, GPT, Gemini, local models, and more.
            </p>

            <div className="install-container">
              <div className="install-tabs">
                {['curl', 'go', 'npm'].map(key => (
                  <button
                    key={key}
                    onClick={() => setTab(key)}
                    className={`install-tab ${tab === key ? 'active' : ''}`}
                  >
                    {key}
                  </button>
                ))}
              </div>
              <div className="install-snippet">
                <code>{snippets[tab]}</code>
                <button
                  onClick={() => handleCopy(tab)}
                  className="copy-btn"
                >
                  {copied === tab ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>
          </div>

          <div className="hero-right">
            <div className="tui-mockup">
              {/* Logo Row */}
              <div className="tui-logo-row">
                <img src="/kode-logo.svg" alt="Kode" width="160" height="35" />
              </div>
              
              {/* Input Box */}
              <div className="tui-input-box">
                <div className="tui-prompt-text">
                  Ask anything... &quot;Draft a RFC for the proposal&quot;
                </div>
                <div className="tui-meta-row">
                  <span className="tui-action-tag">Build</span>
                  <span className="tui-dot-separator">&middot;</span>
                  <span className="tui-model-tag">Llama 3.1 8B Cerebras</span>
                </div>
              </div>
              
              {/* Hints Row */}
              <div className="tui-hints-row">
                <span className="tui-hint-key">tab</span>
                <span className="tui-hint-val">agents</span>
                <span className="tui-hint-key" style={{ marginLeft: 12 }}>ctrl+p</span>
                <span className="tui-hint-val">commands</span>
              </div>
              
              {/* Tip Row */}
              <div className="tui-tip-row">
                <span className="tui-tip-tag">&bull; Tip</span>
                <span className="tui-tip-text">
                  Run <strong>/init</strong> to auto-generate project rules based on your codebase
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

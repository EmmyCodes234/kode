export default function Privacy() {
  return (
    <section>
      <div className="wrapper">
        <div className="features-list">
          <span style={{ fontSize: 24, color: 'var(--ink)', marginBottom: 16, display: 'block' }}>[+]</span>
          <div className="heading-md" style={{ marginBottom: 16 }}>Zero-knowledge code privacy</div>
          <p className="body-md">
            Kode operates 100% on your local machine. We do not proxy, store, or view your code.
            Blindfold Mode takes it further — every identifier (package names, functions, types) is
            SHA-256 obfuscated before LLM submission and reversed on output. Your proprietary logic
            is never exposed in plaintext to external models.
          </p>
          <div style={{ marginTop: 16, display: 'flex', gap: 24, flexWrap: 'wrap' }}>
            <div className="privacy-point">
              <span className="marker">[+]</span>
              <span className="desc">BYOK — your API keys, your providers, zero intermediaries</span>
            </div>
            <div className="privacy-point">
              <span className="marker">[+]</span>
              <span className="desc">No telemetry, no cloud gateway, no session sharing</span>
            </div>
            <div className="privacy-point">
              <span className="marker">[+]</span>
              <span className="desc">Blindfold Mode — SHA-256 identifier obfuscation by default</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

export default function Architecture() {
  return (
    <section>
      <div className="wrapper">
        <div style={{ maxWidth: 680 }}>
          <div className="heading-md" style={{ marginBottom: 24 }}>[x] Architecture</div>
          <p className="body-md" style={{ marginBottom: 8 }}>
            The Go Gatekeeper sits between LLM output and disk, running 5 inline checks before any file write is permitted. This is Kode's competitive moat.
          </p>
          <div style={{
            background: 'var(--surface-card)', padding: 24, borderRadius: 4,
            fontSize: 16, color: 'var(--ink)', marginTop: 24,
          }}>
            <span style={{ color: 'var(--ink)' }}>Plan</span>
            <span style={{ color: 'var(--ash)', margin: '0 4px' }}>&rarr;</span>
            <span style={{ color: 'var(--ink)' }}>Critique</span>
            <span style={{ color: 'var(--ash)', margin: '0 4px' }}>&rarr;</span>
            <span style={{ color: 'var(--ink)' }}>Generate</span>
            <span style={{ color: 'var(--ash)', margin: '0 4px' }}>&rarr;</span>
            <span style={{ color: 'var(--accent)' }}>Verify</span>
            <span style={{ color: 'var(--ash)', margin: '0 4px' }}>&rarr;</span>
            <span style={{ color: 'var(--ink)' }}>Apply</span>
            <span style={{ color: 'var(--ash)', margin: '0 4px' }}>&rarr;</span>
            <span style={{ color: 'var(--ink)' }}>Test</span>
            <div style={{
              marginTop: 16, paddingTop: 16,
              borderTop: '1px solid var(--hairline)',
              color: 'var(--body)', fontSize: 14,
            }}>
              The <strong>Verify</strong> stage runs all 5 gates inline: syntax &rarr; imports &rarr; calls &rarr; blast radius &rarr; architecture + TDD. Block on any gate, and the write never reaches disk. Rollback is automatic.
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

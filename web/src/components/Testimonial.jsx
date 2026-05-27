export default function Testimonial() {
  const useCases = [
    {
      scenario: 'Enterprise Security',
      icon: '🛡️',
      desc: 'Your CISO banned Copilot. Kode\'s Blindfold Mode and local-only execution pass the security review. Verification gates prevent the LLM from injecting vulnerabilities.',
    },
    {
      scenario: 'Platform Engineering',
      icon: '🏗️',
      desc: 'Enforce architectural boundaries across 50 microservices. Gate 5 blocks the LLM from crossing service boundaries — automatically, deterministically, in <50ms.',
    },
    {
      scenario: 'Solo Founders',
      icon: '🚀',
      desc: 'You can\'t afford a broken production deploy at 2 AM. Ghost Branches run 3 strategies in parallel. The best one wins. Rollback is automatic if tests fail.',
    },
  ]

  return (
    <section>
      <div className="wrapper">
        <div className="features-list">
          <div className="heading-md" style={{ marginBottom: 16 }}>[x] Built for teams that can't afford a broken deploy</div>
          <p className="body-md" style={{ marginBottom: 24 }}>
            The entire AI coding market is sitting on the sidelines because enterprises don't trust the code these tools generate. Kode is the answer.
          </p>
          <div className="use-case-grid">
            {useCases.map(uc => (
              <div key={uc.scenario} className="use-case-card">
                <div style={{ fontSize: 28, marginBottom: 12 }}>{uc.icon}</div>
                <div className="label" style={{ marginBottom: 8 }}>{uc.scenario}</div>
                <div className="desc">{uc.desc}</div>
              </div>
            ))}
          </div>
          <div style={{ marginTop: 24 }}>
            <a href="https://github.com/sicario-labs/kode" className="btn-primary" target="_blank" rel="noopener noreferrer">View on GitHub &rarr;</a>
          </div>
        </div>
      </div>
    </section>
  )
}

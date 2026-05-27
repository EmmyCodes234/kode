export default function PricingCards() {
  const cards = [
    {
      name: 'Open Source',
      price: '$0',
      sub: 'Bring your own key (BYOK). 100% local execution.',
      features: ['5 deterministic verification gates', 'Support for OpenAI, Anthropic, Gemini', 'Full AST modification engine', 'Local context graph builder'],
      cta: 'Install Now',
      href: '#install',
      primary: false,
      popular: false,
    },
    {
      name: 'Kode Pro',
      price: '$15',
      sub: 'per user / month. Build autonomously.',
      desc: 'Advanced agentic loop tools for power users.',
      features: ['Daemon Mode (background watcher)', 'Ghost Branches (parallel speculation)', 'MCP IDE Server integration', 'Priority email support'],
      cta: 'Get Early Access \u2192',
      href: '#',
      primary: true,
      popular: true,
    },
    {
      name: 'Enterprise',
      price: 'Custom',
      sub: 'For massive codebases and secure teams.',
      features: ['Custom Gatekeeper policies', 'Self-hosted LLM integration', 'SSO / SAML authentication', 'VPC airgapped deployment'],
      cta: 'Contact Sales',
      href: '#',
      primary: false,
      popular: false,
    },
  ]

  return (
    <div className="pricing-grid">
      {cards.map(card => (
        <div key={card.name} className={`pricing-card ${card.popular ? 'popular' : ''}`}>
          <div className="pricing-card-header">
            {card.popular ? (
              <div className="pricing-card-tier">
                {card.name} <span style={{ color: 'var(--accent-neon)', fontWeight: 700 }}>Recommended</span>
              </div>
            ) : (
              <div className="pricing-card-tier">{card.name}</div>
            )}
            <div className="pricing-card-price">{card.price}</div>
            {card.popular && <div className="pricing-card-sub">{card.sub}</div>}
            <div className="pricing-card-desc">{card.desc || card.sub}</div>
          </div>
          <ul className="pricing-card-features">
            {card.features.map(f => <li key={f}>[+] {f}</li>)}
          </ul>
          <div className="pricing-card-action">
            <a
              href={card.href}
              className={card.popular ? 'btn-primary' : 'btn-secondary'}
              style={{ width: '100%', display: 'flex', justifyContent: 'center' }}
            >
              {card.cta}
            </a>
          </div>
        </div>
      ))}
    </div>
  )
}

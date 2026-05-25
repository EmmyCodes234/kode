export default function PricingCards() {
  const cards = [
    {
      name: 'Free', price: '$0', sub: 'Get started instantly with no API key needed.',
      features: ['3 free models included', '100 requests/day', 'Full gatekeeper engine', 'Community support'],
      cta: 'Get Started', href: '#install', primary: false,
    },
    {
      name: 'Go', price: '$5', sub: 'first month / then $10/mo',
      desc: 'Curated open models for daily coding.',
      features: ['10+ open models', '5hr/week usage limit', 'DeepSeek, Qwen, GLM, Kimi', 'Priority support'],
      cta: 'Subscribe \u2192', href: '/go', primary: true, popular: true,
    },
    {
      name: 'Zen', price: 'Pay as you go', sub: 'Premium models billed per token. Top up as needed.',
      features: ['GPT-4o, Claude, Gemini', '30+ premium models', 'Per-token billing', 'Auto-reload & alerts'],
      cta: 'Get API Key', href: '/zen', primary: false,
    },
  ]

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 32 }}>
      {cards.map(card => (
        <div key={card.name} style={{
          background: card.popular ? 'var(--surface-dark)' : 'var(--surface-card)',
          padding: 24, borderRadius: 4, position: 'relative',
        }}>
          {card.popular && <div style={{
            position: 'absolute', top: -1, left: 24, right: 24,
            height: 3, background: 'var(--accent)', borderRadius: '0 0 2px 2px',
          }} />}
          {card.popular ? (
            <div style={{ fontSize: 14, color: 'var(--on-dark-mute)', lineHeight: 2, marginBottom: 4 }}>
              {card.name} <span style={{ color: 'var(--accent-neon)', fontWeight: 700 }}>Most popular</span>
            </div>
          ) : (
            <div style={{ fontSize: 14, color: 'var(--mute)', lineHeight: 2, marginBottom: 4 }}>{card.name}</div>
          )}
          <div style={{
            fontSize: 38, fontWeight: 700, lineHeight: 1.2,
            color: card.popular ? 'var(--on-dark)' : 'var(--ink)',
            marginBottom: card.popular ? 4 : 16,
          }}>{card.price}</div>
          {card.popular && (
            <div style={{ fontSize: 14, color: 'var(--on-dark-mute)', lineHeight: 2, marginBottom: 16 }}>{card.sub}</div>
          )}
          <div style={{
            fontSize: 14, lineHeight: 1.6, marginBottom: 20,
            color: card.popular ? 'var(--on-dark)' : 'var(--body)',
          }}>{card.desc || card.sub}</div>
          <ul style={{ listStyle: 'none', padding: 0, fontSize: 14, lineHeight: 2, color: card.popular ? 'var(--on-dark)' : 'var(--body)' }}>
            {card.features.map(f => <li key={f}>[+] {f}</li>)}
          </ul>
          <div style={{ marginTop: 20 }}>
            <a href={card.href} className={card.primary ? 'btn-primary' : 'btn-secondary'}
              style={card.popular ? { width: '100%', justifyContent: 'center', background: 'var(--accent)' } : { width: '100%', justifyContent: 'center' }}>
              {card.cta}
            </a>
          </div>
        </div>
      ))}
    </div>
  )
}

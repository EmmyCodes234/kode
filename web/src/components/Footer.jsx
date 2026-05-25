export default function Footer() {
  return (
    <footer style={{ borderTop: '1px solid var(--hairline)', padding: '32px 0' }}>
      <div className="wrapper">
        <div style={{
          display: 'flex', flexWrap: 'wrap',
          borderBottom: '1px solid var(--hairline)',
          paddingBottom: 24, marginBottom: 16,
        }}>
          {[
            { label: 'GitHub', href: 'https://github.com/sicario-labs/kode' },
            { label: 'Docs', href: '/docs' },
            { label: 'Changelog', href: '/changelog' },
            { label: 'Pricing', href: '/pricing' },
            { label: 'X', href: 'https://x.com/trykode' },
          ].map(link => (
            <a key={link.label} href={link.href}
              style={{
                flex: 1, textAlign: 'center', padding: '8px 0',
                color: 'var(--mute)', textDecoration: 'none',
                fontSize: 14, lineHeight: 2,
                borderRight: '1px solid var(--hairline)',
              }}
              onMouseEnter={e => e.target.style.color = 'var(--ink)'}
              onMouseLeave={e => e.target.style.color = 'var(--mute)'}>
              {link.label}
            </a>
          ))}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 14, color: 'var(--mute)', lineHeight: 2 }}>
          <span>&copy; 2026 Kode</span>
          <span>Brand &middot; Privacy &middot; Terms</span>
        </div>
      </div>
    </footer>
  )
}

export default function Newsletter() {
  return (
    <section>
      <div className="wrapper">
        <div style={{ maxWidth: 680 }}>
          <div className="heading-md" style={{ marginBottom: 8 }}>Be the first to know</div>
          <p className="body-md">Join the waitlist for early access to new features, desktop releases, and the Kode Gateway.</p>
          <form style={{ display: 'flex', gap: 8, marginTop: 24 }}
            onSubmit={e => { e.preventDefault(); alert('Subscribed!') }}>
            <input type="email" placeholder="your@email.com" required
              style={{
                background: 'var(--surface-soft)', color: 'var(--ink)',
                border: '1px solid var(--hairline)', borderRadius: 4,
                padding: '8px 12px', height: 40,
                fontFamily: 'var(--font-mono)', fontSize: 16, lineHeight: 1.5,
                flex: 1, maxWidth: 360, outline: 'none',
              }}
              onFocus={e => { e.target.style.background = 'var(--canvas)'; e.target.style.borderColor = 'var(--ink)' }}
              onBlur={e => { e.target.style.background = 'var(--surface-soft)'; e.target.style.borderColor = 'var(--hairline)' }} />
            <button type="submit" className="btn-primary">Subscribe</button>
          </form>
        </div>
      </div>
    </section>
  )
}

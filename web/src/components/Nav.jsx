import { Link, useLocation, useNavigate } from 'react-router-dom'

export default function Nav() {
  const location = useLocation()
  const navigate = useNavigate()

  const handleHashClick = (e, hash) => {
    if (location.pathname !== '/') {
      e.preventDefault()
      navigate('/' + hash)
    }
  }

  return (
    <nav style={{
      background: 'var(--canvas)',
      borderBottom: '1px solid var(--hairline)',
      height: 56,
      display: 'flex',
      alignItems: 'center',
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      zIndex: 100,
    }}>
      <div className="wrapper" style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        width: '100%',
      }}>
        <Link to="/" style={{
          display: 'flex',
          alignItems: 'center',
          textDecoration: 'none',
          height: 32,
        }}>
          <img src="/kode-logo-light.svg" alt="Kode" style={{ height: 32, width: 'auto', display: 'block' }} />
        </Link>
        <ul style={{
          display: 'flex',
          alignItems: 'center',
          gap: 20,
          listStyle: 'none',
          margin: 0,
          padding: 0,
        }}>
          <li><a href="#features" onClick={(e) => handleHashClick(e, '#features')} style={{ textDecoration: 'none', color: 'var(--body)', fontSize: 15, fontWeight: 500 }} onMouseEnter={e => e.target.style.color = 'var(--ink)'} onMouseLeave={e => e.target.style.color = 'var(--body)'}>[+] Features</a></li>
          <li><Link to="/pricing" style={{ textDecoration: 'none', color: 'var(--body)', fontSize: 15, fontWeight: 500 }} onMouseEnter={e => e.target.style.color = 'var(--ink)'} onMouseLeave={e => e.target.style.color = 'var(--body)'}>Plans</Link></li>
          <li><a href="#faq" onClick={(e) => handleHashClick(e, '#faq')} style={{ textDecoration: 'none', color: 'var(--body)', fontSize: 15, fontWeight: 500 }} onMouseEnter={e => e.target.style.color = 'var(--ink)'} onMouseLeave={e => e.target.style.color = 'var(--body)'}>[x] FAQ</a></li>
          <li><Link to="/docs" style={{ textDecoration: 'none', color: 'var(--body)', fontSize: 15, fontWeight: 500 }} onMouseEnter={e => e.target.style.color = 'var(--ink)'} onMouseLeave={e => e.target.style.color = 'var(--body)'}>Docs</Link></li>
          <li><a href="https://github.com/sicario-labs/kode" target="_blank" rel="noreferrer" style={{ textDecoration: 'none', color: 'var(--body)', fontSize: 15, fontWeight: 500 }} onMouseEnter={e => e.target.style.color = 'var(--ink)'} onMouseLeave={e => e.target.style.color = 'var(--body)'}>GitHub</a></li>
        </ul>
        <a href="#install" className="btn-primary" style={{ height: 32, fontSize: 15, padding: '4px 20px', lineHeight: 2, textDecoration: 'none', fontFamily: 'var(--font-mono)' }}>Download</a>
      </div>
    </nav>
  )
}

export default function Stats() {
  return (
    <section>
      <div className="wrapper">
        <div style={{ maxWidth: 680 }}>
          <div className="heading-md" style={{ marginBottom: 8 }}>The safe AI coding agent</div>
          <p className="body-md" style={{ maxWidth: 600 }}>
            With <strong>5 verification gates</strong> running in a compiled Go binary, Kode brings deterministic safety to AI-assisted coding — a paradigm shift from generate-and-pray.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24, marginTop: 32 }}>
            {[
              { points: '0,44 20,36 40,40 60,28 80,32 100,20 120,24 140,12 160,16 180,8 200,12', label: 'Fig 1. 5 Verification Gates' },
              { points: '0,44 30,42 60,36 90,30 120,22 150,14 180,10 200,8', label: 'Fig 2. Go Engine (~10MB)' },
              { points: '0,36 25,34 50,30 75,26 100,20 125,16 150,12 175,8 200,6', label: 'Fig 3. 121+ Go Tests Passing' },
            ].map((chart, i) => (
              <div key={i} style={{ padding: '16px 0' }}>
                <svg viewBox="0 0 200 48" preserveAspectRatio="none" style={{ display: 'block', marginBottom: 12, width: '100%', height: 48 }}>
                  {i === 2 && <line x1="0" y1="24" x2="200" y2="24" stroke="#201d1d" strokeWidth="0.5" strokeDasharray="4,4" />}
                  <polyline points={chart.points} fill="none" stroke="#201d1d" strokeWidth="0.8" vectorEffect="non-scaling-stroke" />
                  <circle cx="200" cy={[12, 8, 6][i]} r="1.5" fill="#201d1d" />
                </svg>
                <span style={{ color: 'var(--mute)', fontSize: 14, lineHeight: 2 }}>{chart.label}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}

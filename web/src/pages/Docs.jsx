export default function Docs() {
  return (
    <>
      <section style={{ padding: '152px 0 64px' }}>
        <div className="wrapper">
          <div style={{ maxWidth: 720 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 8,
              background: 'var(--surface-dark)', color: 'var(--on-dark)',
              padding: '2px 8px', borderRadius: 4,
              fontSize: 14, lineHeight: 2, marginBottom: 24,
            }}>[x] Documentation</span>
            <h1 className="display-xl" style={{ marginBottom: 24 }}>Getting started with Kode</h1>
            <p className="body-md" style={{ marginBottom: 48 }}>
              Kode is a Go-powered AI coding agent that verifies every generated patch through 5 deterministic gates before touching your filesystem.
            </p>

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Installation</div>
            <p className="body-md" style={{ marginBottom: 16 }}>Install Kode with one command:</p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 15 }}>
              <code style={{ color: 'var(--accent)' }}>$</code>{' '}
              <code>curl -fsSL https://trykode.xyz/install.sh | bash</code>
            </div>
            <p className="body-md" style={{ marginBottom: 16 }}>Or install via Go:</p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 15 }}>
              <code style={{ color: 'var(--accent)' }}>$</code>{' '}
              <code>go install github.com/kode/kode/cmd/kode@latest</code>
            </div>

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Quick start</div>
            <p className="body-md" style={{ marginBottom: 16 }}>Initialize Kode in your project:</p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 15 }}>
              <code style={{ color: 'var(--accent)' }}>$</code>{' '}
              <code>cd my-project &amp;&amp; kode init</code>
            </div>
            <p className="body-md" style={{ marginBottom: 16 }}>Run a full loop:</p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 15 }}>
              <code style={{ color: 'var(--accent)' }}>$</code>{' '}
              <code>kode loop &quot;add user authentication to the API&quot;</code>
            </div>

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Verification gates</div>
            <p className="body-md" style={{ marginBottom: 24 }}>
              Every generated patch passes through 5 gates before it reaches disk:
            </p>
            {[
              ['Syntax', 'Validates that the generated code compiles without syntax errors.'],
              ['Imports', 'Checks all imports resolve correctly and no unused imports exist.'],
              ['Calls', 'Verifies function signatures match across call sites.'],
              ['Blast Radius', 'Limits how many files can be modified per cycle. Walks the reverse dependency graph.'],
              ['Architecture + TDD', 'Enforces test-first workflow. Blocks prod writes without corresponding test files.'],
            ].map(([gate, desc]) => (
              <div key={gate} style={{ display: 'flex', gap: 16, padding: '8px 0', borderBottom: '1px solid var(--hairline)' }}>
                <span style={{ color: 'var(--ink)', fontWeight: 500, minWidth: 140 }}>[{gate}]</span>
                <span style={{ color: 'var(--body)' }}>{desc}</span>
              </div>
            ))}

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Using the TUI</div>
            <p className="body-md" style={{ marginBottom: 16 }}>Launch the interactive terminal UI:</p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 15 }}>
              <code style={{ color: 'var(--accent)' }}>$</code>{' '}
              <code>kode tui</code>
            </div>
            <p className="body-md" style={{ marginBottom: 16 }}>
              The TUI provides a split-panel interface with context graph, generation status, gatekeeper verdicts, and file diffs — all in your terminal.
            </p>

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Configuration</div>
            <p className="body-md" style={{ marginBottom: 16 }}>
              Kode is configured through <code style={{ background: 'var(--surface-card)', padding: '1px 4px', borderRadius: 2 }}>.kode/config.json</code>:
            </p>
            <div style={{ background: 'var(--surface-dark)', color: 'var(--on-dark)', padding: '16px 20px', borderRadius: 4, marginBottom: 32, fontSize: 14, lineHeight: 1.8 }}>
              <pre style={{ fontFamily: 'var(--font-mono)', whiteSpace: 'pre-wrap' }}>{`{
  "provider": "openai",
  "model": "gpt-4o",
  "tdd_mode": true,
  "max_blast_radius": 5,
  "token_budget_usd": 0.50,
  "blindfold_mode": false
}`}</pre>
            </div>

            <div className="heading-md" style={{ marginBottom: 16, paddingTop: 16, borderTop: '1px solid var(--hairline)' }}>Commands reference</div>
            {[
              ['kode init', 'Scaffold .kode/config.json with auto-detected test command'],
              ['kode plan <task>', 'Build a surgical 8K context graph for the given task'],
              ['kode generate <prompt>', 'Generate patches via LLM'],
              ['kode verify --input <file>', 'Run all 5 gate checks on a file'],
              ['kode run <prompt>', 'Full generate -> verify -> apply pipeline'],
              ['kode loop <task>', 'Full Plan -> Generate -> Verify -> Apply -> Test cycle'],
              ['kode revert', 'Undo the last applied hunk (surgical AST revert)'],
              ['kode stats', 'Analyze gatekeeper audit log'],
              ['kode tui', 'Launch interactive terminal UI'],
            ].map(([cmd, desc]) => (
              <div key={cmd} style={{ display: 'flex', gap: 16, padding: '8px 0', borderBottom: '1px solid var(--hairline)' }}>
                <code style={{ color: 'var(--ink)', minWidth: 220, fontSize: 15 }}>{cmd}</code>
                <span style={{ color: 'var(--body)', fontSize: 15 }}>{desc}</span>
              </div>
            ))}
          </div>
        </div>
      </section>
    </>
  )
}

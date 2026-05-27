const gates = [
  { num: '01', label: 'Syntax Gate', desc: 'Parses every modified file using a dual-engine architecture: full Tree-sitter AST validation when available, gracefully falling back to fast regex heuristics. Parse error = hard block.' },
  { num: '02', label: 'Imports Gate', desc: 'Validates every import path is resolvable against the project dependency graph. Catches hallucinated packages before they compile.' },
  { num: '03', label: 'Calls Gate', desc: 'Checks that every function and method call references a symbol that actually exists. The #1 source of LLM hallucinations — caught here.' },
  { num: '04', label: 'Blast Radius Gate', desc: 'Walks the dependency graph backward from every modified file. If downstream impact exceeds your configured threshold — blocked.' },
  { num: '05', label: 'Architecture Gate', desc: 'Enforces declared module boundaries. Prevents the LLM from crossing service boundaries or importing banned packages.' },
  { num: '06', label: 'Security Gate', desc: 'Automated vulnerability scanning on generated code. SQL injection, XSS, hardcoded secrets — caught before they\'re committed.' },
]

const capabilities = [
  { label: 'Ghost Branches', desc: 'Spawns N parallel git worktrees, each running a different strategy. Scores results by verification gates passed, blast radius, and time. Best branch wins.' },
  { label: 'Blindfold Mode', desc: 'SHA-256 identifier obfuscation before LLM submission. Your package names, functions, and types are never exposed in plaintext to external models.' },
  { label: 'Critique Lenses', desc: 'Pre-generation review that rejects bad patches before the LLM even runs. Coherence, convention, dependency, and blast radius checks.' },
  { label: 'Session Memory', desc: 'Per-project learning. Verification failures, blast radius trends, and ghost branch strategy success rates persist across sessions.' },
  { label: 'Golf Gate', desc: 'Benchmark regression detection. If a patch degrades performance beyond your configured threshold, it\'s rolled back automatically.' },
  { label: 'MCP Server', desc: 'Expose Kode as a verification oracle that other AI agents can call. Any MCP-compatible agent gets access to verified code generation.' },
]

export default function Features() {
  return (
    <section id="features">
      <div className="wrapper">
        <div className="features-list">
          <div className="heading-md" style={{ marginBottom: 24 }}>Six gates between the LLM and your filesystem</div>
          <p className="body-md" style={{ marginBottom: 32 }}>
            Every generated patch passes through a compiled Go binary that runs 6 deterministic checks in under 50ms. If any gate fails, the patch is rejected and the LLM self-corrects. No code reaches disk unverified.
          </p>
          {gates.map(g => (
            <div key={g.num} className="feature-row">
              <span className="marker" style={{ fontVariantNumeric: 'tabular-nums', minWidth: 32 }}>[{g.num}]</span>
              <div>
                <span className="label">{g.label}</span>
                <br />
                <span className="desc">{g.desc}</span>
              </div>
            </div>
          ))}
        </div>

        <div className="features-list" style={{ marginTop: 64 }}>
          <div className="heading-md" style={{ marginBottom: 24 }}>Beyond verification — capabilities no incumbent offers</div>
          <p className="body-md" style={{ marginBottom: 32 }}>
            Verification is the thesis. These are the weapons.
          </p>
          {capabilities.map(c => (
            <div key={c.label} className="feature-row">
              <span className="marker">[+]</span>
              <div>
                <span className="label">{c.label}</span>
                <br />
                <span className="desc">{c.desc}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

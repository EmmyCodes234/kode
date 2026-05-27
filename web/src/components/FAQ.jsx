import { useState } from 'react'

const faqs = [
  { q: 'What is Kode?', a: 'Kode is an open source AI coding agent with a compiled Go verification engine. Every generated patch must pass 5 deterministic gates (syntax, imports, calls, blast radius, architecture) before touching your filesystem.' },
  { q: 'How do I use Kode?', a: 'Install the binary, run kode init in your project, then kode loop "your task". The engine handles context gathering, LLM prompting, patch generation, verification, application, and testing in one cycle.' },
  { q: 'Do I need extra AI subscriptions?', a: 'Kode is a Bring Your Own Key (BYOK) platform. You must provide an API key for an OpenAI-compatible provider. We natively support Claude, GPT-4, Gemini, and local LLMs via Ollama/LMStudio.' },
  { q: 'Can I use my ChatGPT Plus or GitHub Copilot subscription?', a: 'No. Kode requires raw API access (e.g., Anthropic Console, OpenAI Platform) to function. Consumer subscriptions do not provide the raw API keys necessary for agentic execution.' },
  { q: 'Can I only use Kode in the terminal?', a: 'Kode works flawlessly in the terminal via kode tui. However, using the kode mcp serve command, you can also integrate Kode directly into IDEs like Cursor, Antigravity, and Claude Desktop.' },
  { q: 'How much does Kode cost?', a: 'The core Kode engine is 100% free and open source. You pay only the direct API costs to your LLM provider. Kode Pro offers advanced features like Daemon Mode and Ghost Branches for $15/mo.' },
  { q: 'What about data and privacy?', a: 'Kode operates entirely on your local machine. We do not proxy, store, or view your code. Your repository context is sent directly from your localhost to your chosen LLM API provider.' },
  { q: 'Is Kode open source?', a: 'Yes. The full source is at github.com/sicario-labs/kode. The Go Gatekeeper engine, TypeScript TUI, and MCP implementations are entirely open.' },
]

export default function FAQ() {
  const [open, setOpen] = useState(null)

  return (
    <section id="faq">
      <div className="wrapper">
        <div className="features-list">
          <div className="heading-md" style={{ marginBottom: 24 }}>FAQ</div>
          {faqs.map((item, i) => (
            <div key={i} className={`faq-item ${open === i ? 'open' : ''}`}>
              <button
                onClick={() => setOpen(open === i ? null : i)}
                className="faq-question"
              >
                {item.q}
                <span>{open === i ? '\u2212' : '+'}</span>
              </button>
              <div className="faq-answer">
                <p>{item.a}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

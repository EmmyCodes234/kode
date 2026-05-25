import PricingCards from '../components/PricingCards'
import FAQ from '../components/FAQ'
import Newsletter from '../components/Newsletter'

export default function Pricing() {
  return (
    <>
      <section style={{ padding: '152px 0 96px' }}>
        <div className="wrapper">
          <div style={{ maxWidth: 680 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 8,
              background: 'var(--surface-dark)', color: 'var(--on-dark)',
              padding: '2px 8px', borderRadius: 4,
              fontSize: 14, lineHeight: 2, marginBottom: 24,
            }}>[x] Plans & Pricing</span>
            <h1 className="display-xl" style={{ marginBottom: 16 }}>
              Choose your plan
            </h1>
            <p className="body-md" style={{ marginBottom: 32, maxWidth: '90%' }}>
              Start for free, upgrade when you need more. Every tier includes the full Kode Gatekeeper engine with all 5 verification gates.
            </p>
          </div>
          <PricingCards />
          <div style={{
            fontSize: 14, color: 'var(--body)', lineHeight: 1.6,
            borderTop: '1px solid var(--hairline)', paddingTop: 16, maxWidth: 680,
          }}>
            <strong>All plans include:</strong> Go Gatekeeper engine &middot; Blast Radius &middot; TDD Lockjaw &middot; Cost Budgeting &middot; Blindfold Mode &middot; Unlimited public projects
          </div>
        </div>
      </section>
      <section>
        <div className="wrapper">
          <div style={{ maxWidth: 680 }}>
            <div className="heading-md" style={{ marginBottom: 16 }}>Gateway model catalog</div>
            <p className="body-md" style={{ marginBottom: 24 }}>
              Every model in the Kode Gateway is benchmarked specifically for coding agent workloads. No random selection — each model passes a standardized eval suite before being added.
            </p>
            <div style={{
              background: 'var(--surface-card)', padding: 24, borderRadius: 4,
              fontSize: 14, color: 'var(--body)',
            }}>
              <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr', gap: 8, padding: '8px 0', borderBottom: '1px solid var(--hairline)', color: 'var(--ink)', fontWeight: 500 }}>
                <span>Model</span><span>Tier</span><span>Provider</span>
              </div>
              {[
                ['DeepSeek V4 Flash Free', 'Free', 'DeepSeek'],
                ['Llama 3 70B Free', 'Free', 'Meta'],
                ['Nemotron 3 Super Free', 'Free', 'NVIDIA'],
                ['DeepSeek V4', 'Go', 'DeepSeek'],
                ['Qwen 2.5 72B', 'Go', 'Alibaba'],
                ['GLM-4', 'Go', 'Zhipu'],
                ['Kimi V2', 'Go', 'Moonshot'],
                ['GPT-4o', 'Zen', 'OpenAI'],
                ['Claude Sonnet 4', 'Zen', 'Anthropic'],
                ['Gemini 2.5 Pro', 'Zen', 'Google'],
              ].map(row => (
                <div key={row[0]} style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr', gap: 8, padding: '6px 0', borderBottom: '1px solid var(--hairline)' }}>
                  <span>{row[0]}</span>
                  <span>{row[1]}</span>
                  <span>{row[2]}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
      <FAQ />
      <Newsletter />
    </>
  )
}

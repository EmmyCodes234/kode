package gateway

var DefaultCatalog = Catalog{
	Models: []Model{
		// Free tier (no API key needed, rate-limited)
		{ID: "deepseek-v4-flash-free", Name: "DeepSeek V4 Flash Free", Provider: "deepseek", InputCost: 0, OutputCost: 0, Tier: TierFree, Description: "Limited-time free, data may be used for model improvement"},
		{ID: "nemotron-3-super-free", Name: "Nemotron 3 Super Free", Provider: "nvidia", InputCost: 0, OutputCost: 0, Tier: TierFree, Description: "Nvidia trial terms apply"},
		{ID: "llama-3-70b-free", Name: "Llama 3 70B Free", Provider: "meta", InputCost: 0, OutputCost: 0, Tier: TierFree, Description: "Community tier, rate-limited"},

		// Go tier ($5/mo first month, $10/mo thereafter)
		{ID: "deepseek-v4", Name: "DeepSeek V4", Provider: "deepseek", InputCost: 15, OutputCost: 60, Tier: TierGo, Description: "Fast reasoning, strong coding benchmarks"},
		{ID: "qwen-2.5-72b", Name: "Qwen 2.5 72B", Provider: "alibaba", InputCost: 20, OutputCost: 80, Tier: TierGo, Description: "Strong multilingual coding"},
		{ID: "glm-4", Name: "GLM-4", Provider: "zhipu", InputCost: 15, OutputCost: 60, Tier: TierGo, Description: "Balanced general-purpose model"},
		{ID: "kimi-v2", Name: "Kimi V2", Provider: "moonshot", InputCost: 12, OutputCost: 50, Tier: TierGo, Description: "Long context window"},

		// Zen tier (pay-as-you-go, per-token pricing)
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", InputCost: 250, OutputCost: 1000, Tier: TierZen, Description: "OpenAI's flagship model"},
		{ID: "claude-sonnet-4", Name: "Claude Sonnet 4", Provider: "anthropic", InputCost: 300, OutputCost: 1500, Tier: TierZen, Description: "Best for complex coding tasks"},
		{ID: "claude-haiku-4", Name: "Claude Haiku 4", Provider: "anthropic", InputCost: 25, OutputCost: 125, Tier: TierZen, Description: "Fast, affordable coding"},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google", InputCost: 125, OutputCost: 500, Tier: TierZen, Description: "Large context, multimodal"},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai", InputCost: 15, OutputCost: 60, Tier: TierZen, Description: "Lightweight OpenAI model"},
	},
}

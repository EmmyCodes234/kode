package gateway

type Tier string

const (
	TierFree Tier = "free"
	TierGo   Tier = "go"
	TierZen  Tier = "zen"
)

type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	InputCost   int    `json:"input_cost"`   // per 1M tokens in cents
	OutputCost  int    `json:"output_cost"`  // per 1M tokens in cents
	Tier        Tier   `json:"tier"`
	Description string `json:"description,omitempty"`
}

type Catalog struct {
	Models []Model `json:"models"`
}

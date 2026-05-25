package gateway

type Tier string

const (
	TierLite Tier = "lite"
	TierPro  Tier = "pro"
)

type Protocol string

const (
	ProtocolMessages  Protocol = "messages"   // Anthropic Messages API
	ProtocolResponses Protocol = "responses"  // OpenAI Responses API
)

type Model struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	InputCost   float64  `json:"input_cost"`
	OutputCost  float64  `json:"output_cost"`
	Tier        Tier     `json:"tier"`
	Protocol    Protocol `json:"protocol,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Catalog struct {
	Models []Model `json:"models"`
}

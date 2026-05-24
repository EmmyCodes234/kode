package llm

import (
	"net/url"
	"os"
)

type Config struct {
	APIKey   string
	Endpoint string
	Model    string
}

func DefaultConfig() Config {
	key := os.Getenv("KODE_LLM_API_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	endpoint := os.Getenv("KODE_LLM_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	model := os.Getenv("KODE_LLM_MODEL")
	if model == "" {
		model = "gpt-4o"
	}
	return Config{
		APIKey:   key,
		Endpoint: endpoint,
		Model:    model,
	}
}

func (c Config) ChatURL() string {
	base, _ := url.JoinPath(c.Endpoint, "chat", "completions")
	return base
}

func (c Config) Valid() error {
	if c.APIKey == "" {
		return ErrMissingAPIKey
	}
	if c.Endpoint == "" {
		return ErrMissingEndpoint
	}
	if c.Model == "" {
		return ErrMissingModel
	}
	return nil
}

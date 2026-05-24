package llm

import (
	"encoding/json"
	"testing"
)

func TestDefaultConfig_EnvFallback(t *testing.T) {
	t.Setenv("KODE_LLM_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "sk-test123")
	t.Setenv("KODE_LLM_ENDPOINT", "")
	t.Setenv("KODE_LLM_MODEL", "")

	cfg := DefaultConfig()
	if cfg.APIKey != "sk-test123" {
		t.Fatalf("expected fallback to OPENAI_API_KEY, got %q", cfg.APIKey)
	}
	if cfg.Endpoint != "https://api.openai.com/v1" {
		t.Fatalf("expected default endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Model != "gpt-4o" {
		t.Fatalf("expected default model gpt-4o, got %q", cfg.Model)
	}
}

func TestDefaultConfig_ExplicitEnv(t *testing.T) {
	t.Setenv("KODE_LLM_API_KEY", "sk-kode-key")
	t.Setenv("KODE_LLM_ENDPOINT", "https://llm.example.com/v1")
	t.Setenv("KODE_LLM_MODEL", "claude-3-opus")

	cfg := DefaultConfig()
	if cfg.APIKey != "sk-kode-key" {
		t.Fatalf("expected KODE_LLM_API_KEY, got %q", cfg.APIKey)
	}
	if cfg.Endpoint != "https://llm.example.com/v1" {
		t.Fatalf("expected custom endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Model != "claude-3-opus" {
		t.Fatalf("expected custom model, got %q", cfg.Model)
	}
}

func TestConfig_ChatURL(t *testing.T) {
	cfg := Config{Endpoint: "https://api.openai.com/v1"}
	expected := "https://api.openai.com/v1/chat/completions"
	if got := cfg.ChatURL(); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}

	cfg2 := Config{Endpoint: "http://localhost:11434/v1"}
	expected2 := "http://localhost:11434/v1/chat/completions"
	if got := cfg2.ChatURL(); got != expected2 {
		t.Fatalf("expected %q, got %q", expected2, got)
	}
}

func TestConfig_Valid_MissingAPIKey(t *testing.T) {
	cfg := Config{Endpoint: "https://example.com", Model: "gpt-4o"}
	if err := cfg.Valid(); err != ErrMissingAPIKey {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestConfig_Valid_MissingEndpoint(t *testing.T) {
	cfg := Config{APIKey: "sk-test", Model: "gpt-4o"}
	if err := cfg.Valid(); err != ErrMissingEndpoint {
		t.Fatalf("expected ErrMissingEndpoint, got %v", err)
	}
}

func TestConfig_Valid_MissingModel(t *testing.T) {
	cfg := Config{APIKey: "sk-test", Endpoint: "https://example.com"}
	if err := cfg.Valid(); err != ErrMissingModel {
		t.Fatalf("expected ErrMissingModel, got %v", err)
	}
}

func TestConfig_Valid_OK(t *testing.T) {
	cfg := Config{APIKey: "sk-test", Endpoint: "https://example.com", Model: "gpt-4o"}
	if err := cfg.Valid(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestBuildGeneratePrompt_WithContext(t *testing.T) {
	p := BuildGeneratePrompt("add a function", "file: main.go\npackage main\nfunc main() {}")
	if p == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestBuildGeneratePrompt_WithoutContext(t *testing.T) {
	p := BuildGeneratePrompt("add a function", "")
	if p == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestNewClient_ConfigApplied(t *testing.T) {
	cfg := Config{APIKey: "sk-test", Endpoint: "https://example.com", Model: "gpt-4o"}
	c := NewClient(cfg)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.config.Model != "gpt-4o" {
		t.Fatalf("expected model gpt-4o, got %s", c.config.Model)
	}
}

func TestChatRequest_JSON_Marshal(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: RoleSystem, Content: "be concise"},
			{Role: RoleUser, Content: "hello"},
		},
		Temperature: 0.2,
		MaxTokens:   100,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

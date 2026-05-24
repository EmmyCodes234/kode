package llm

import "errors"

var (
	ErrMissingAPIKey    = errors.New("KODE_LLM_API_KEY or OPENAI_API_KEY not set")
	ErrMissingEndpoint  = errors.New("LLM endpoint not configured")
	ErrMissingModel     = errors.New("LLM model not configured")
	ErrAPIRequest       = errors.New("LLM API request failed")
	ErrRateLimit        = errors.New("LLM rate limit exceeded")
	ErrAuthFailed       = errors.New("LLM authentication failed (check API key)")
	ErrEmptyResponse    = errors.New("LLM returned empty response")
)

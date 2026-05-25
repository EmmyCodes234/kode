package router

import "errors"

var (
	ErrAllProvidersFailed = errors.New("all LLM providers failed")
	ErrNoProviders        = errors.New("no LLM providers configured")
)

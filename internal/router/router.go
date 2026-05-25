package router

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kode/kode/internal/llm"
)

type FallbackStrategy string

const (
	StrategyFailover   FallbackStrategy = "failover"
	StrategyRoundRobin FallbackStrategy = "round_robin"
	StrategyParallel   FallbackStrategy = "parallel"
)

type Provider struct {
	Name       string
	Client     *llm.Client
	Weight     int
	MaxRetries int
}

type RouteConfig struct {
	Providers        []Provider
	FallbackStrategy FallbackStrategy
	Timeout          time.Duration
}

type Router struct {
	mu      sync.Mutex
	rrIndex int
	config  RouteConfig
}

func NewRouter(config RouteConfig) *Router {
	return &Router{config: config}
}

func (r *Router) AddProvider(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.Providers = append(r.config.Providers, p)
}

func (r *Router) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if len(r.config.Providers) == 0 {
		return nil, ErrNoProviders
	}

	if r.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
	}

	switch r.config.FallbackStrategy {
	case StrategyRoundRobin:
		return r.chatRoundRobin(ctx, req)
	case StrategyParallel:
		return r.chatParallel(ctx, req)
	default:
		return r.chatFailover(ctx, req)
	}
}

func (r *Router) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := r.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: userPrompt},
		},
		Temperature: 0.2,
		MaxTokens:   4096,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", ErrAllProvidersFailed
	}
	return resp.Choices[0].Message.Content, nil
}

func (r *Router) chatFailover(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	var lastErr error
	for _, p := range r.config.Providers {
		resp, err := r.tryProvider(ctx, p, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, errors.Join(ErrAllProvidersFailed, lastErr)
}

func (r *Router) chatRoundRobin(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	r.mu.Lock()
	start := r.rrIndex
	r.rrIndex = (r.rrIndex + 1) % len(r.config.Providers)
	r.mu.Unlock()

	providers := r.config.Providers
	var lastErr error
	for i := 0; i < len(providers); i++ {
		idx := (start + i) % len(providers)
		resp, err := r.tryProvider(ctx, providers[idx], req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, errors.Join(ErrAllProvidersFailed, lastErr)
}

func (r *Router) chatParallel(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		resp *llm.ChatResponse
		err  error
	}

	results := make(chan result, len(r.config.Providers))

	var wg sync.WaitGroup
	for _, p := range r.config.Providers {
		wg.Add(1)
		p := p
		go func() {
			defer wg.Done()
			resp, err := r.tryProvider(ctx, p, req)
			select {
			case results <- result{resp, err}:
			case <-ctx.Done():
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var lastErr error
	for res := range results {
		if res.err == nil && res.resp != nil {
			return res.resp, nil
		}
		if res.err != nil {
			lastErr = res.err
		}
	}
	return nil, errors.Join(ErrAllProvidersFailed, lastErr)
}

func (r *Router) tryProvider(ctx context.Context, p Provider, req llm.ChatRequest) (*llm.ChatResponse, error) {
	var lastErr error
	for i := 0; i <= p.MaxRetries; i++ {
		resp, err := p.Client.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}
	return nil, fmt.Errorf("provider %s: %w", p.Name, lastErr)
}

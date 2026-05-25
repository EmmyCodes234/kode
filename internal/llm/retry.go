package llm

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimit) {
		return true
	}
	if errors.Is(err, ErrAPIRequest) {
		return true
	}
	return false
}

func (c *Client) ChatWithRetry(ctx context.Context, req ChatRequest, retryCfg RetryConfig) (*ChatResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= retryCfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := retryCfg.InitialBackoff * (1 << uint(attempt-1))
			if backoff > retryCfg.MaxBackoff {
				backoff = retryCfg.MaxBackoff
			}
			jitter := time.Duration(rand.Int63n(int64(backoff) / 4))
			backoff += jitter

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := c.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
	}

	return nil, lastErr
}

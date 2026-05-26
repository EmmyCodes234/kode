package gateway

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type KeyPool struct {
	mu     sync.Mutex
	keys   []string
	dead   map[string]time.Time // key → time when it can be retried
	fails  map[string]int       // key → consecutive failure count
	idx    atomic.Int64
	healthURL string
}

func NewKeyPool(keys []string) *KeyPool {
	return &KeyPool{
		keys:      keys,
		dead:      make(map[string]time.Time),
		fails:     make(map[string]int),
		healthURL: "https://api.openmodel.ai/v1/health",
	}
}

func (p *KeyPool) Next() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Prune expired dead keys
	now := time.Now()
	for k, retryAt := range p.dead {
		if now.After(retryAt) {
			delete(p.dead, k)
			p.fails[k] = 0
		}
	}

	// Collect live keys
	var live []string
	for _, k := range p.keys {
		if _, dead := p.dead[k]; !dead {
			live = append(live, k)
		}
	}

	if len(live) == 0 {
		return ""
	}

	pool := live
	i := p.idx.Add(1) % int64(len(pool))
	return pool[i]
}

// ReportFailure marks a key as having failed. After 3 consecutive failures,
// the key is circuit-broken for 10 minutes.
func (p *KeyPool) ReportFailure(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.fails[key]++
	if p.fails[key] >= 3 {
		p.dead[key] = time.Now().Add(10 * time.Minute)
	}
}

// ReportSuccess resets the failure count for a key.
func (p *KeyPool) ReportSuccess(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fails[key] = 0
}

// HealthyCount returns the number of non-dead keys.
func (p *KeyPool) HealthyCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	count := 0
	for _, k := range p.keys {
		if _, dead := p.dead[k]; !dead {
			count++
		}
	}
	return count
}

// TotalKeys returns the total number of keys in the pool.
func (p *KeyPool) TotalKeys() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.keys)
}

// CheckHealth pings all non-dead keys on the health endpoint.
// Keys that fail the health check are circuit-broken for 5 minutes.
func (p *KeyPool) CheckHealth() {
	p.mu.Lock()
	keys := make([]string, 0, len(p.keys))
	for _, k := range p.keys {
		if _, dead := p.dead[k]; !dead {
			keys = append(keys, k)
		}
	}
	p.mu.Unlock()

	client := &http.Client{Timeout: 5 * time.Second}
	for _, k := range keys {
		req, err := http.NewRequest("GET", p.healthURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "Bearer "+k)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		// If request succeeded, key is healthy
		if err == nil {
			p.ReportSuccess(k)
		}
	}
}

func (p *KeyPool) Keys() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, len(p.keys))
	copy(out, p.keys)
	return out
}

func KeysFromEnv(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	// Periodic health check every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			// Health check is triggered externally through the gateway startup
		}
	}()
}

func startHealthCheck(pool *KeyPool) {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			pool.CheckHealth()
		}
	}()
}

func (p *KeyPool) logState() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	deadCount := len(p.dead)
	return fmt.Sprintf("pool: %d total, %d dead", len(p.keys), deadCount)
}

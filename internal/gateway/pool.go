package gateway

import (
	"os"
	"strings"
	"sync/atomic"
)

type KeyPool struct {
	keys []string
	idx  atomic.Int64
}

func NewKeyPool(keys []string) *KeyPool {
	return &KeyPool{keys: keys}
}

func (p *KeyPool) Next() string {
	if len(p.keys) == 0 {
		return ""
	}
	i := p.idx.Add(1) % int64(len(p.keys))
	return p.keys[i]
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

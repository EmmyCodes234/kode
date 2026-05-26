package gateway

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type usageRecord struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	Model     string    `json:"model"`
	Tier      Tier      `json:"tier"`
	Tokens    int       `json:"tokens,omitempty"`
	Status    int       `json:"status"`
}

type UsageMonitor struct {
	mu    sync.Mutex
	ring  []usageRecord
	head  int
	count int
	cap   int
}

func NewUsageMonitor(capacity int) *UsageMonitor {
	return &UsageMonitor{
		ring: make([]usageRecord, capacity),
		cap:  capacity,
	}
}

func (m *UsageMonitor) Record(ip, model string, tier Tier, tokens, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ring[m.head] = usageRecord{
		Timestamp: time.Now(),
		IP:        ip,
		Model:     model,
		Tier:      tier,
		Tokens:    tokens,
		Status:    status,
	}
	m.head = (m.head + 1) % m.cap
	if m.count < m.cap {
		m.count++
	}
}

func (m *UsageMonitor) Recent() []usageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.count == 0 {
		return nil
	}

	out := make([]usageRecord, 0, m.count)
	start := m.head - m.count
	if start < 0 {
		start += m.cap
	}
	for i := 0; i < m.count; i++ {
		idx := (start + i) % m.cap
		out = append(out, m.ring[idx])
	}
	return out
}

func (m *UsageMonitor) Stats() map[string]any {
	recent := m.Recent()
	total := len(recent)

	var liteCount, proCount int
	modelCounts := make(map[string]int)
	statusCounts := make(map[int]int)

	for _, r := range recent {
		if r.Tier == TierLite {
			liteCount++
		} else {
			proCount++
		}
		modelCounts[r.Model]++
		statusCounts[r.Status]++
	}

	return map[string]any{
		"total_requests": total,
		"lite_count":     liteCount,
		"pro_count":      proCount,
		"by_model":       modelCounts,
		"by_status":      statusCounts,
	}
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.monitor.Stats())
}

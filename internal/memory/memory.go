package memory

import (
	"fmt"
	"sync"
	"time"
)

type RecordType string

const (
	TypeVerification RecordType = "verification"
	TypeGeneration   RecordType = "generation"
	TypeError        RecordType = "error"
	TypeContext      RecordType = "context"
)

type Record struct {
	ID        string            `json:"id"`
	TaskID    string            `json:"task_id"`
	Timestamp time.Time         `json:"timestamp"`
	Type      RecordType        `json:"type"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Store struct {
	mu      sync.RWMutex
	records []Record
	nextID  int
}

func NewStore() *Store {
	return &Store{
		records: make([]Record, 0),
		nextID:  1,
	}
}

func (s *Store) Store(r Record) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r.ID == "" {
		r.ID = fmt.Sprintf("%d", s.nextID)
		s.nextID++
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	if r.Metadata == nil {
		r.Metadata = make(map[string]string)
	}
	s.records = append(s.records, r)
}

func (s *Store) GetByTask(taskID string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Record
	for _, r := range s.records {
		if r.TaskID == taskID {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) GetByType(typ RecordType) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Record
	for _, r := range s.records {
		if r.Type == typ {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) GetByKey(key string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Record
	for _, r := range s.records {
		if r.Key == key {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) GetRecent(n int) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 || len(s.records) == 0 {
		return nil
	}
	start := len(s.records) - n
	if start < 0 {
		start = 0
	}
	out := make([]Record, len(s.records)-start)
	copy(out, s.records[start:])
	return out
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = make([]Record, 0)
	s.nextID = 1
}

func (s *Store) GetFailedHunks(taskID string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	// Iterate in reverse to find the most recent verification failure
	for i := len(s.records) - 1; i >= 0; i-- {
		r := s.records[i]
		if r.TaskID == taskID && r.Type == TypeVerification && r.Key != "" {
			result[r.Key] = r.Value
		}
	}
	return result
}

func (s *Store) GetPreviousAttempts(taskID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, r := range s.records {
		if r.TaskID == taskID && r.Type == TypeGeneration {
			count++
		}
	}
	return count
}

package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type VerificationFailure struct {
	Timestamp    time.Time `json:"timestamp"`
	TaskID       string    `json:"task_id"`
	ErrorMessage string    `json:"error_message"`
	FilePath     string    `json:"file_path"`
}

type BlastRadius struct {
	Timestamp    time.Time `json:"timestamp"`
	TaskID       string    `json:"task_id"`
	FilesChanged int       `json:"files_changed"`
}

type GhostStrategy struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID    string    `json:"task_id"`
	Strategy  string    `json:"strategy"`
	Score     float64   `json:"score"`
	IsWinner  bool      `json:"is_winner"`
}

type Database struct {
	Failures   []VerificationFailure `json:"verification_failures"`
	Radii      []BlastRadius         `json:"blast_radii"`
	Strategies []GhostStrategy       `json:"ghost_strategies"`
}

type Memory struct {
	dbPath string
	mu     sync.Mutex
	data   Database
}

// Open creates or connects to the JSON memory database for the project.
func Open(projectRoot string) (*Memory, error) {
	dbDir := filepath.Join(projectRoot, ".kode")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .kode dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "memory.json")
	m := &Memory{
		dbPath: dbPath,
		data: Database{
			Failures:   make([]VerificationFailure, 0),
			Radii:      make([]BlastRadius, 0),
			Strategies: make([]GhostStrategy, 0),
		},
	}

	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load memory db: %w", err)
	}

	return m, nil
}

func (m *Memory) Close() error {
	return m.save()
}

func (m *Memory) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, err := os.ReadFile(m.dbPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &m.data)
}

func (m *Memory) save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.dbPath, b, 0644)
}

// RecordVerificationFailure logs a failure so the agent can learn patterns to avoid.
func (m *Memory) RecordVerificationFailure(taskID, errorMessage, filePath string) error {
	m.mu.Lock()
	m.data.Failures = append(m.data.Failures, VerificationFailure{
		Timestamp:    time.Now(),
		TaskID:       taskID,
		ErrorMessage: errorMessage,
		FilePath:     filePath,
	})
	m.mu.Unlock()
	return m.save()
}

// RecordBlastRadius logs the number of files changed for trend analysis.
func (m *Memory) RecordBlastRadius(taskID string, filesChanged int) error {
	m.mu.Lock()
	m.data.Radii = append(m.data.Radii, BlastRadius{
		Timestamp:    time.Now(),
		TaskID:       taskID,
		FilesChanged: filesChanged,
	})
	m.mu.Unlock()
	return m.save()
}

// RecordGhostStrategy logs successful and failed strategies to optimize future speculative branching.
func (m *Memory) RecordGhostStrategy(taskID, strategy string, score float64, isWinner bool) error {
	m.mu.Lock()
	m.data.Strategies = append(m.data.Strategies, GhostStrategy{
		Timestamp: time.Now(),
		TaskID:    taskID,
		Strategy:  strategy,
		Score:     score,
		IsWinner:  isWinner,
	})
	m.mu.Unlock()
	return m.save()
}

package ghost

import (
	"fmt"
	"time"

	"github.com/kode/kode/internal/execution"
)

var ErrBranchPanic = fmt.Errorf("branch goroutine panic")

type BranchStrategy string

const (
	StrategyMinimal    BranchStrategy = "minimal"
	StrategyModular    BranchStrategy = "modular"
	StrategyAggressive BranchStrategy = "aggressive"
)

type BranchID string

const (
	BranchAlpha BranchID = "alpha"
	BranchBeta  BranchID = "beta"
	BranchGamma BranchID = "gamma"
)

type BranchSpec struct {
	ID       BranchID       `json:"id"`
	Strategy BranchStrategy `json:"strategy"`
	Label    string         `json:"label"`
	Prompt   string         `json:"prompt"`
}

type BranchResult struct {
	ID           BranchID                  `json:"id"`
	Strategy     BranchStrategy            `json:"strategy"`
	Status       execution.Status          `json:"status"`
	Summary      *execution.ExecutionSummary `json:"summary,omitempty"`
	TokenCost    float64                   `json:"token_cost"`
	Duration     time.Duration             `json:"duration"`
	BlastRadius  int                       `json:"blast_radius"`
	GatesPassed  int                       `json:"gates_passed"`
	Score        float64                   `json:"score"`
	Error        string                    `json:"error,omitempty"`
	WorktreePath string                    `json:"worktree_path"`
}

type GhostSummary struct {
	Task        string         `json:"task"`
	Branches    []*BranchResult `json:"branches"`
	Winner      *BranchResult  `json:"winner"`
	TotalTime   time.Duration  `json:"total_time"`
	TotalCost   float64        `json:"total_cost"`
}

func BranchStrategies(task string, count int) ([]BranchSpec, error) {
	specs := []BranchSpec{
		{
			ID:       BranchAlpha,
			Strategy: StrategyMinimal,
			Label:    "lightweight, minimal implementation",
			Prompt:   fmt.Sprintf("Implement the following with the MINIMAL code possible. No extra abstractions, no over-engineering. Single-file if reasonable:\n\n%s", task),
		},
		{
			ID:       BranchBeta,
			Strategy: StrategyModular,
			Label:    "robust, modular, well-structured",
			Prompt:   fmt.Sprintf("Implement the following with a CLEAN MODULAR architecture. Split into meaningful abstractions, use dependency injection, keep interfaces clean:\n\n%s", task),
		},
		{
			ID:       BranchGamma,
			Strategy: StrategyAggressive,
			Label:    "performance-optimized, aggressive",
			Prompt:   fmt.Sprintf("Implement the following with MAXIMUM PERFORMANCE and robustness. Use caching, async patterns, error handling at every boundary, defensive programming:\n\n%s", task),
		},
	}

	if count < 1 || count > 3 {
		count = 3
	}
	return specs[:count], nil
}

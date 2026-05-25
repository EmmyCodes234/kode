package ghost

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/llm"
	"github.com/kode/kode/internal/workflow"
)

type GhostEngine struct {
	repoDir     string
	worktrees   *WorktreeManager
	llmConfig   *llm.Config
	testCommand string
}

func NewGhostEngine(repoDir string, cfg *llm.Config, testCmd string) *GhostEngine {
	return &GhostEngine{
		repoDir:     repoDir,
		worktrees:   NewWorktreeManager(repoDir),
		llmConfig:   cfg,
		testCommand: testCmd,
	}
}

type GhostRunConfig struct {
	Task     string
	Branches int
}

func (e *GhostEngine) Run(ctx context.Context, cfg GhostRunConfig) (*GhostSummary, error) {
	start := time.Now()

	specs, err := BranchStrategies(cfg.Task, cfg.Branches)
	if err != nil {
		return nil, fmt.Errorf("branch strategies: %w", err)
	}

	e.worktrees.Cleanup()
	_ = e.worktrees.CurrentBranch()

	results := make([]*BranchResult, len(specs))
	type workItem struct {
		idx  int
		spec BranchSpec
	}
	work := make(chan workItem, len(specs))
	var wg sync.WaitGroup

	for i := 0; i < len(specs); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range work {
				results[item.idx] = e.executeBranch(ctx, item.spec, start)
			}
		}()
	}

	for i, spec := range specs {
		work <- workItem{idx: i, spec: spec}
	}
	close(work)
	wg.Wait()

	winner := SelectWinner(results)

	if winner != nil {
		if err := e.worktrees.MergeWinner(winner.ID); err != nil {
			fmt.Fprintf(os.Stderr, "  ghost: merge warning: %v\n", err)
		}
	}

	// Remove losing branches
	for _, r := range results {
		if winner != nil && r.ID == winner.ID {
			continue
		}
		e.worktrees.Remove(r.ID)
	}

	totalTime := time.Since(start)
	var totalCost float64
	for _, r := range results {
		totalCost += r.TokenCost
	}

	summary := &GhostSummary{
		Task:      cfg.Task,
		Branches:  results,
		Winner:    winner,
		TotalTime: totalTime,
		TotalCost: totalCost,
	}

	return summary, nil
}

func (e *GhostEngine) executeBranch(ctx context.Context, spec BranchSpec, globalStart time.Time) *BranchResult {
	branchStart := time.Now()
	result := &BranchResult{
		ID:       spec.ID,
		Strategy: spec.Strategy,
	}

	worktreePath, err := e.worktrees.Create(spec)
	if err != nil {
		result.Status = execution.StatusFail
		result.Error = fmt.Sprintf("worktree: %v", err)
		result.Duration = time.Since(branchStart)
		return result
	}
	result.WorktreePath = worktreePath

	pipe := workflow.NewPipeline(workflow.Config{
		LLMConfig:   e.llmConfig,
		TestCommand: e.testCommand,
		MaxRetries:  2,
	})

	pipe.BeforeStage(workflow.StageGenerate, func(s *workflow.State) {
		s.ProjectRoot = worktreePath
	})

	res, pipeErr := pipe.Run(ctx, spec.Prompt)
	result.Duration = time.Since(branchStart)

	if pipeErr != nil {
		result.Status = execution.StatusFail
		result.Error = pipeErr.Error()
		return result
	}

	if res == nil {
		result.Status = execution.StatusFail
		result.Error = "pipeline returned nil"
		return result
	}

	result.Status = res.Status
	result.Summary = res.State.Summary
	if res.State.Summary != nil {
		result.GatesPassed = len(res.State.Summary.AppliedHunks)
		result.BlastRadius = len(res.State.Summary.AppliedHunks)
	}

	return result
}

func (e *GhostEngine) Cleanup() {
	e.worktrees.Cleanup()
}

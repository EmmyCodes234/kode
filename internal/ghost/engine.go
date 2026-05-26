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

	e.worktrees.Cleanup(ctx)
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
				select {
				case <-ctx.Done():
					return
				default:
				}
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
		if err := e.worktrees.MergeWinner(ctx, winner.ID); err != nil {
			fmt.Fprintf(os.Stderr, "  ghost: merge warning: %v\n", err)
		}
	}

	// Remove losing branches
	for _, r := range results {
		if winner != nil && r.ID == winner.ID {
			continue
		}
		e.worktrees.Remove(ctx, r.ID)
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

func (e *GhostEngine) executeBranch(ctx context.Context, spec BranchSpec, globalStart time.Time) (result *BranchResult) {
	branchStart := time.Now()
	result = &BranchResult{
		ID:       spec.ID,
		Strategy: spec.Strategy,
	}

	defer func() {
		if r := recover(); r != nil {
			result.Status = execution.StatusFail
			result.Error = fmt.Sprintf("panic: %v", r)
			result.Duration = time.Since(branchStart)
		}
		if result.Status == execution.StatusFail && result.WorktreePath != "" {
			e.worktrees.Remove(ctx, spec.ID)
		}
	}()

	worktreePath, err := e.worktrees.Create(ctx, spec)
	if err != nil {
		result.Status = execution.StatusFail
		result.Error = fmt.Sprintf("worktree: %v", err)
		result.Duration = time.Since(branchStart)
		return result
	}
	result.WorktreePath = worktreePath

	// Self-healing retry loop: up to 3 attempts with escalating prompts
	maxAttempts := 3
	prompt := spec.Prompt

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Duration = time.Since(branchStart)

		pipe := workflow.NewPipeline(workflow.Config{
			LLMConfig:          e.llmConfig,
			TestCommand:        e.testCommand,
			MaxRetries:         2,
			EnableContextIndex: true,
		})

		pipe.BeforeStage(workflow.StageGenerate, func(s *workflow.State) {
			s.ProjectRoot = worktreePath
		})

		res, pipeErr := pipe.Run(ctx, prompt)

		if pipeErr == nil && res != nil && res.Status == execution.StatusPass {
			result.Status = res.Status
			result.Summary = res.State.Summary
			if res.State.Summary != nil {
				result.GatesPassed = len(res.State.Summary.AppliedHunks)
				result.BlastRadius = len(res.State.Summary.AppliedHunks)
			}
			return result
		}

		// Capture failure details for retry escalation
		var failureDetail string
		if pipeErr != nil {
			failureDetail = pipeErr.Error()
		} else if res != nil && len(res.State.Errors) > 0 {
			failureDetail = res.State.Errors[len(res.State.Errors)-1]
		} else {
			failureDetail = "unknown failure"
		}

		if attempt == maxAttempts {
			result.Status = execution.StatusFail
			result.Error = failureDetail
			return result
		}

		// Escalate prompt with failure context
		escalations := []string{
			"\n\n[RETRY] Previous attempt failed. Focus on simpler, more reliable code. Avoid complex patterns.",
			"\n\n[RETRY] Still failing. Try an even simpler approach. Break the change into smaller steps. Ensure all referenced types and functions exist.",
		}
		prompt = spec.Prompt + escalations[attempt-1] + "\n\nPrevious error: " + failureDetail
	}

	return result
}

func (e *GhostEngine) Cleanup() {
	e.worktrees.Cleanup(context.Background())
}

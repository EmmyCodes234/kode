package golf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/ghost"
	"github.com/kode/kode/internal/llm"
)

type GolfEngine struct {
	projectRoot string
	llmConfig   *llm.Config
	testCommand string
}

func NewGolfEngine(projectRoot string, cfg *llm.Config, testCmd string) *GolfEngine {
	if testCmd == "" {
		testCmd = execution.DetectTestCommand(projectRoot)
	}
	return &GolfEngine{
		projectRoot: projectRoot,
		llmConfig:   cfg,
		testCommand: testCmd,
	}
}

func (e *GolfEngine) Optimize(ctx context.Context, cfg GolfConfig) (*GolfSummary, error) {
	start := time.Now()

	if _, err := os.Stat(cfg.File); err != nil {
		return nil, fmt.Errorf("file not found: %s", cfg.File)
	}

	// Step 1: Establish baseline
	baseline, err := RunBenchmarks(e.projectRoot, e.testCommand)
	if err != nil {
		return nil, fmt.Errorf("baseline benchmark failed: %w", err)
	}

	// Step 2: Generate optimization prompts per strategy
	targetStrs := map[OptimizeTarget]string{
		OptimizeSpeed:      "speed / throughput — maximize operations per second",
		OptimizeMemory:     "memory efficiency — minimize allocations, reduce GC pressure",
		OptimizeComplexity: "algorithmic complexity — reduce Big-O, tighten loops",
	}

	goal, ok := targetStrs[cfg.Target]
	if !ok {
		goal = string(targetStrs[OptimizeSpeed])
	}

	specs := golfBranchSpecs(cfg.File, goal)

	// Step 3: Run in ghost worktrees
	ghostEngine := ghost.NewGhostEngine(e.projectRoot, e.llmConfig, e.testCommand)
	defer ghostEngine.Cleanup()

	ghostResults, err := ghostEngine.Run(ctx, ghost.GhostRunConfig{
		Task:     fmt.Sprintf("Optimize %s for %s. %s", cfg.File, cfg.Target, goal),
		Branches: len(specs),
	})
	if err != nil {
		return nil, fmt.Errorf("ghost run failed: %w", err)
	}

	// Step 4: Benchmark each branch
	var branchBenchs []BranchBench
	for _, r := range ghostResults.Branches {
		bb := BranchBench{
			Label:  string(r.Strategy),
			Branch: string(r.ID),
			Pass:   r.Status == execution.StatusPass,
		}

		if r.Status == execution.StatusPass && r.WorktreePath != "" {
			benchs, err := RunBenchmarksForFile(r.WorktreePath, e.testCommand, cfg.File)
			if err != nil {
				bb.Error = err.Error()
			} else {
				bb.Benchs = benchs
			}
		}

		if bb.Error == "" && len(bb.Benchs) == 0 && r.Status == execution.StatusPass {
			bb.Error = "no benchmark results"
		}

		branchBenchs = append(branchBenchs, bb)
	}

	// Step 5: Score and select winner
	winnerID, improvement := selectBenchWinner(baseline, branchBenchs, ghostResults.Branches)

	summary := &GolfSummary{
		File:        cfg.File,
		Target:      cfg.Target,
		Baseline:    baseline,
		Branches:    branchBenchs,
		Winner:      winnerID,
		Improvement: improvement,
		TotalTime:   time.Since(start),
	}

	return summary, nil
}

func golfBranchSpecs(file, goal string) []ghost.BranchSpec {
	return []ghost.BranchSpec{
		{
			ID:       ghost.BranchAlpha,
			Strategy: "concurrency",
			Label:    "concurrency & parallelism",
			Prompt: fmt.Sprintf(`Optimize this file (%s) for %s.

STRATEGY: CONCURRENCY & PARALLELISM
- Use goroutines, channels, or async patterns where applicable
- Parallelize independent loops and operations
- Use sync.Pool for reusable allocations
- Keep the API and exported signatures IDENTICAL
- All existing tests must pass

File to optimize:
%s`, file, goal, readFileContent(file)),
		},
		{
			ID:       ghost.BranchBeta,
			Strategy: "memory",
			Label:    "memory & allocation optimization",
			Prompt: fmt.Sprintf(`Optimize this file (%s) for %s.

STRATEGY: MEMORY & ALLOCATION OPTIMIZATION
- Pre-allocate slices and maps with known capacity
- Replace heap allocations with stack-allocated structs
- Use value receivers instead of pointers where appropriate
- Eliminate unnecessary copies and string conversions
- Keep the API and exported signatures IDENTICAL
- All existing tests must pass

File to optimize:
%s`, file, goal, readFileContent(file)),
		},
		{
			ID:       ghost.BranchGamma,
			Strategy: "algorithmic",
			Label:    "algorithmic complexity reduction",
			Prompt: fmt.Sprintf(`Optimize this file (%s) for %s.

STRATEGY: ALGORITHMIC COMPLEXITY REDUCTION
- Reduce Big-O time complexity (e.g., O(N^2) → O(N log N))
- Replace linear scans with hash maps or binary search
- Eliminate nested loops where possible
- Cache repeated computations
- Keep the API and exported signatures IDENTICAL
- All existing tests must pass

File to optimize:
%s`, file, goal, readFileContent(file)),
		},
	}
}

func selectBenchWinner(baseline []BenchResult, branches []BranchBench, ghostResults []*ghost.BranchResult) (winnerID string, improvement float64) {
	var bestID string
	var bestImprovement float64

	for _, bb := range branches {
		if !bb.Pass {
			continue
		}
		_, countImproved, countTotal := CompareBenchs(baseline, bb.Benchs)
		if countTotal == 0 {
			continue
		}
		pct := float64(countImproved) / float64(countTotal) * 100.0
		if pct > bestImprovement {
			bestImprovement = pct
			bestID = bb.Branch
		}
	}

	// Fallback: use ghost engine's scoring
	if bestID == "" {
		if ghostWinner := ghost.SelectWinner(ghostResults); ghostWinner != nil {
			bestID = string(ghostWinner.ID)
		}
	}

	return bestID, bestImprovement
}

func readFileContent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("(unable to read: %v)", err)
	}
	return string(data)
}

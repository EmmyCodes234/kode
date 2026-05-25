package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/golf"
	"github.com/kode/kode/internal/llm"
	"github.com/spf13/cobra"
)

func init() {
	var projectDir string
	var testCommand string
	var modelFlag string
	var optimizeTarget string

	golfCmd := &cobra.Command{
		Use:   "golf <file> [--optimize speed|memory|complexity]",
		Short: "Optimize code using competitive ghost branch swarms",
		Long: `Run competitive code optimization across 3 parallel strategies.

Kode spins up 3 isolated ghost worktrees, each applying a different
optimization strategy (concurrency, memory, algorithmic complexity).
It benchmarks the original code as a baseline, then pits each
strategy against it. The winner is merged back.

Strategies:
  concurrency  — goroutines, channels, async patterns
  memory      — pre-allocation, stack vs heap, reduced copies
  algorithmic — Big-O reduction, hash maps, loop elimination

Benchmarks run via 'go test -bench=. -benchmem' by default.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]

			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
			}

			absDir, err := filepath.Abs(projectDir)
			if err != nil {
				return fmt.Errorf("invalid project directory: %w", err)
			}

			absFile := file
			if !filepath.IsAbs(file) {
				absFile = filepath.Join(absDir, file)
			}
			if _, err := os.Stat(absFile); err != nil {
				return fmt.Errorf("file not found: %s", absFile)
			}

			cfg := llm.DefaultConfig()
			if modelFlag != "" {
				cfg.Model = modelFlag
			}
			if cfg.APIKey == "" {
				return fmt.Errorf("LLM API key not configured.\nSet KODE_LLM_API_KEY or OPENAI_API_KEY environment variable.")
			}

			targets := map[string]golf.OptimizeTarget{
				"speed":      golf.OptimizeSpeed,
				"memory":     golf.OptimizeMemory,
				"complexity": golf.OptimizeComplexity,
			}
			target, ok := targets[optimizeTarget]
			if !ok {
				target = golf.OptimizeSpeed
			}

			testCmd := testCommand
			if testCmd == "" {
				testCmd = execution.DetectTestCommand(absDir)
			}

			sectionHeader("Code Golf")
			stepStart("File: %s", file)
			stepDetail("Target: %s | Strategies: concurrency, memory, algorithmic", target)
			stepDetail("LLM: %s | Test: %s", cfg.Model, testCmd)

			engine := golf.NewGolfEngine(absDir, &cfg, testCmd)
			summary, err := engine.Optimize(context.Background(), golf.GolfConfig{
				File:        absFile,
				Target:      target,
				ProjectRoot: absDir,
				TestCommand: testCmd,
			})
			if err != nil {
				stepFail("Optimization failed: %v", err)
				return err
			}

			fmt.Fprintf(os.Stderr, "\n")
			sectionHeader("Results")

			// Baseline
			fmt.Fprintf(os.Stderr, "  %sBaseline%s — %d benchmark(s)\n", ansiBold, ansiReset, len(summary.Baseline))
			for _, b := range summary.Baseline {
				fmt.Fprintf(os.Stderr, "    %s: %.2f ns/op", b.Name, b.NSPerOp)
				if b.AllocBPO > 0 {
					fmt.Fprintf(os.Stderr, " (%d B/op, %d allocs/op)", b.AllocBPO, b.AllocsPO)
				}
				fmt.Fprintf(os.Stderr, "\n")
			}

			// Branches
			fmt.Fprintf(os.Stderr, "\n")
			for _, bb := range summary.Branches {
				icon := "✓"
				if bb.Error != "" || !bb.Pass {
					icon = "✗"
				}
				isWinner := bb.Branch == summary.Winner
				prefix := "  "
				colorStart := ""
				if isWinner {
					colorStart = ansiGreen
				}
				wTag := ""
				if isWinner {
					wTag = fmt.Sprintf(" %s👑 WINNER%s", ansiBold, ansiReset)
				}
				fmt.Fprintf(os.Stderr, "%s%s%s %s (%s)%s%s\n", colorStart, prefix, icon, bb.Label, bb.Branch, ansiReset, wTag)

				if bb.Error != "" {
					fmt.Fprintf(os.Stderr, "    %sError: %s%s\n", ansiRed, bb.Error, ansiReset)
					continue
				}
				if !bb.Pass {
					fmt.Fprintf(os.Stderr, "    %sFailed verification%s\n", ansiRed, ansiReset)
					continue
				}

				_, countImproved, countTotal := golf.CompareBenchs(summary.Baseline, bb.Benchs)
				pct := 0.0
				if countTotal > 0 {
					pct = float64(countImproved) / float64(countTotal) * 100.0
				}
				color := ""
				if pct > 0 {
					color = ansiGreen
				} else if pct < 0 {
					color = ansiRed
				}
				fmt.Fprintf(os.Stderr, "    %sImprovement: %+d/%d benchmarks (%+.0f%%)%s\n", color, countImproved, countTotal, pct, ansiReset)

				for _, bench := range bb.Benchs {
					delta := golf.FindDelta(summary.Baseline, bb.Benchs, bench.Name)
					deltaStr := ""
					deltaColor := ""
					if delta > 0 {
						deltaStr = fmt.Sprintf(" [⬇ %.0f%% faster]", delta)
						deltaColor = ansiGreen
					} else if delta < 0 {
						deltaStr = fmt.Sprintf(" [⬆ %.0f%% slower]", -delta)
						deltaColor = ansiRed
					}
					fmt.Fprintf(os.Stderr, "      %s: %s%.2f ns/op%s", bench.Name, deltaColor, bench.NSPerOp, ansiReset)
					if bench.AllocBPO > 0 {
						fmt.Fprintf(os.Stderr, " (%d B/op, %d allocs/op)", bench.AllocBPO, bench.AllocsPO)
					}
					fmt.Fprintf(os.Stderr, "%s%s%s\n", deltaColor, deltaStr, ansiReset)
				}
			}

			fmt.Fprintf(os.Stderr, "\n")
			if summary.Winner != "" {
				winnerBranch := findBranchBench(summary.Branches, summary.Winner)
				if winnerBranch != nil && winnerBranch.Pass {
					stepOK("Winner: %s (%s)", winnerBranch.Label, summary.Winner)

					improvedPct := 0.0
					_, countImproved, countTotal := golf.CompareBenchs(summary.Baseline, winnerBranch.Benchs)
					if countTotal > 0 {
						improvedPct = float64(countImproved) / float64(countTotal) * 100.0
					}
					stepDetail("Benchmark improvement: +%.0f%% of benchmarks", improvedPct)

					if promptYesNo("Merge optimized version?") {
						stepOK("Optimized code merged.")
					} else {
						stepDetail("Discarded. Original code untouched.")
					}
				} else {
					stepFail("Best candidate failed verification")
				}
			} else {
				stepFail("No branch improved the baseline")
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(summary)

			return nil
		},
	}

	golfCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current)")
	golfCmd.Flags().StringVar(&testCommand, "test-command", "", "Test command override (default: auto-detect)")
	golfCmd.Flags().StringVar(&modelFlag, "model", "", "Model override (default: KODE_LLM_MODEL or gpt-4o)")
	golfCmd.Flags().StringVar(&optimizeTarget, "optimize", "speed", "Optimization target: speed, memory, complexity")
	rootCmd.AddCommand(golfCmd)
}

func findBranchBench(branches []golf.BranchBench, id string) *golf.BranchBench {
	for i := range branches {
		if branches[i].Branch == id {
			return &branches[i]
		}
	}
	return nil
}

func promptYesNo(msg string) bool {
	fmt.Fprintf(os.Stderr, "  %s [Y/n]: ", msg)
	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

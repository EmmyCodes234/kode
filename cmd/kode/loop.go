package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kode/kode/internal/execution"
	ghostlib "github.com/kode/kode/internal/ghost"
	"github.com/kode/kode/internal/llm"
	"github.com/kode/kode/internal/workflow"
	"github.com/spf13/cobra"
)

func init() {
	var modelFlag string
	var contextFile string
	var testCommand string
	var projectDir string
	var branches int

	loopCmd := &cobra.Command{
		Use:   "loop <task>",
		Short: "Full Plan → Generate → Verify → Apply → Test cycle",
		Long: `Run the complete Kode workflow on a task:
   1. Generate patches via LLM
   2. Verify and apply to disk
   3. Run tests
   4. Rollback on test failure (restore from snapshot)

With --branches N, Kode runs the task in N parallel git worktrees
(Ghost Branch mode). Each branch uses a different strategy:
  Alpha  — lightweight, minimal implementation
  Beta   — robust, modular architecture
  Gamma  — performance-optimized, aggressive

Kode scores each result by blast radius, token cost, and speed,
then merges the winner and discards the rest.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.Join(args, " ")

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

			cfg := llm.DefaultConfig()
			if modelFlag != "" {
				cfg.Model = modelFlag
			}

			if cfg.APIKey == "" {
				return fmt.Errorf("LLM API key not configured.\nSet KODE_LLM_API_KEY or OPENAI_API_KEY environment variable.")
			}

			// Ghost Branch mode: run in parallel worktrees
			if branches > 1 {
				return runGhostLoop(cmd, absDir, task, branches, &cfg, testCommand)
			}

			pipe := workflow.NewPipeline(workflow.Config{
				LLMConfig:    &cfg,
				ContextFile:  contextFile,
				TestCommand:  testCommand,
			})

			start := time.Now()

			pipe.BeforeStage(workflow.StagePlan, func(s *workflow.State) {
				s.ProjectRoot = absDir
			})

			pipe.BeforeStage(workflow.StageGenerate, func(s *workflow.State) {
				sectionHeader("Generating patches")
				stepStart("Task: %s", task)
				stepDetail("LLM: %s", cfg.Model)
			})

			pipe.AfterStage(workflow.StageGenerate, func(s *workflow.State, err error) {
				if err == nil {
					genElapsed := time.Since(start).Seconds()
					stepOK("Generated (%.1fs): %d hunk(s)", genElapsed, len(s.Hunks))
				} else {
					stepFail("Generation failed: %v", err)
				}
			})

			pipe.BeforeStage(workflow.StageVerify, func(s *workflow.State) {
				sectionHeader("Verifying patches")
			})

			pipe.AfterStage(workflow.StageVerify, func(s *workflow.State, err error) {
				if err == nil {
					applyElapsed := time.Since(start).Seconds()
					stepOK("Applied %d hunk(s) (%.1fs)", len(s.Summary.AppliedHunks), applyElapsed)
				} else {
					stepFail("Verification failed: %d hunk(s) failed", len(s.Summary.FailedHunks))
					for id, msg := range s.Summary.FailedHunks {
						stepDetail("%s: %s", id, msg)
					}
				}
			})

			pipe.BeforeStage(workflow.StageTest, func(s *workflow.State) {
				sectionHeader("Running tests")
				testCmd := testCommand
				if testCmd == "" {
					testCmd = execution.DetectTestCommand(absDir)
				}
				stepStart("Command: %s", testCmd)
			})

			result, err := pipe.Run(context.Background(), task)
			if err != nil {
				stepFail("Pipeline failed: %v", err)
			}

			totalElapsed := time.Since(start).Seconds()
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			if result != nil && result.Status == execution.StatusPass {
				stepOK("All done (%.1fs)", totalElapsed)
				enc.Encode(map[string]interface{}{
					"status":           "PASS",
					"task":             task,
					"hunks_applied":    len(result.State.Summary.AppliedHunks),
					"test_command":     result.State.Summary,
					"duration_seconds": totalElapsed,
				})
				return nil
			}

			if result != nil {
				enc.Encode(map[string]interface{}{
					"status":           "FAIL",
					"task":             task,
					"hunks_applied":    len(result.State.Summary.AppliedHunks),
					"test_command":     testCommand,
					"test_error":       result.State.LastError(),
					"duration_seconds": totalElapsed,
				})
			}

			if err != nil {
				return err
			}
			return nil
		},
	}

	loopCmd.Flags().IntVar(&branches, "branches", 1, "Ghost Branch count (2-3 enables parallel speculation mode)")
	loopCmd.Flags().StringVar(&modelFlag, "model", "", "Model override (default: KODE_LLM_MODEL or gpt-4o)")
	loopCmd.Flags().StringVar(&contextFile, "context-file", "", "Path to context packet JSON from 'kode plan --packet'")
	loopCmd.Flags().StringVar(&testCommand, "test-command", "", "Test command override (default: auto-detect)")
	loopCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current working directory)")
	rootCmd.AddCommand(loopCmd)
}

func runGhostLoop(cmd *cobra.Command, absDir, task string, branches int, cfg *llm.Config, testCmd string) error {
	sectionHeader("Ghost Branch Mode")
	stepStart("Task: %s", task)
	stepDetail("Branches: %d strategies in parallel", branches)
	stepDetail("LLM: %s", cfg.Model)
	fmt.Fprintf(os.Stderr, "\n")

	engine := ghostlib.NewGhostEngine(absDir, cfg, testCmd)
	defer engine.Cleanup()

	summary, err := engine.Run(context.Background(), ghostlib.GhostRunConfig{
		Task:     task,
		Branches: branches,
	})
	if err != nil {
		stepFail("Ghost run failed: %v", err)
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	for _, b := range summary.Branches {
		icon := "[x]"
		statusColor := "PASS"
		if b.Status == execution.StatusFail {
			icon = "[!]"
			statusColor = "FAIL"
		}
		if summary.Winner != nil && summary.Winner.ID == b.ID {
			icon = "[+]"
			fmt.Fprintf(os.Stderr, "  %s %s (%s) — %s — Score: %.2f — WINNER\n",
				ansiGreen+icon+ansiReset, b.ID, b.Strategy, statusColor, b.Score)
		} else {
			fmt.Fprintf(os.Stderr, "  %s %s (%s) — %s — Score: %.2f\n",
				icon, b.ID, b.Strategy, statusColor, b.Score)
		}
		if b.Error != "" {
			stepDetail("Error: %s", b.Error)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	if summary.Winner != nil {
		stepOK("Winner: %s (%s) — score %.2f", summary.Winner.ID, summary.Winner.Strategy, summary.Winner.Score)
	}
	stepDetail("Total: %.1fs | $%.4f token cost", summary.TotalTime.Seconds(), summary.TotalCost)

	enc.Encode(map[string]interface{}{
		"status":     "PASS",
		"task":       task,
		"branches":   len(summary.Branches),
		"winner":     summary.Winner,
		"total_time": summary.TotalTime.Seconds(),
		"total_cost": summary.TotalCost,
	})
	return nil
}


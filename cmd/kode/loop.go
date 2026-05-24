package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/llm"
	"github.com/spf13/cobra"
)

func init() {
	var modelFlag string
	var contextFile string
	var testCommand string
	var projectDir string

	loopCmd := &cobra.Command{
		Use:   "loop <task>",
		Short: "Full Plan → Generate → Verify → Apply → Test cycle",
		Long: `Run the complete Kode workflow on a task:
  1. Generate patches via LLM
  2. Verify and apply to disk
  3. Run tests
  4. Rollback on test failure (restore from snapshot)

The test command is auto-detected (go test, npm test, cargo test)
or overridable with --test-command.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.Join(args, " ")
			start := time.Now()

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

			// Step 1: Read context if provided
			contextStr := ""
			if contextFile != "" {
				data, err := os.ReadFile(contextFile)
				if err != nil {
					return fmt.Errorf("cannot read context file: %w", err)
				}
				contextStr = string(data)
			}

			// Step 2: Generate patches
			fmt.Fprintf(os.Stderr, "  Generating patches for: %s\n", task)
			userPrompt := llm.BuildGeneratePrompt(task, contextStr)
			client := llm.NewClient(cfg)

			fmt.Fprintf(os.Stderr, "  LLM: %s\n", cfg.Model)
			content, err := client.Generate(context.Background(), llm.SystemPrompt, userPrompt)
			if err != nil {
				return fmt.Errorf("LLM call failed: %w", err)
			}

			genElapsed := time.Since(start).Seconds()
			fmt.Fprintf(os.Stderr, "  Generated (%.1fs). Parsing...\n", genElapsed)

			parser := execution.NewHunkParser()
			hunks, err := parser.ParseLLMResponse(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse LLM response:\n  %v\n\nRaw response:\n%s\n", err, content)
				return fmt.Errorf("parse error: %w", err)
			}

			fmt.Fprintf(os.Stderr, "  %d hunk(s) generated.\n", len(hunks))

			// Step 3: Take snapshot BEFORE applying
			var affectedFiles []string
			for _, h := range hunks {
				found := false
				for _, af := range affectedFiles {
					if af == h.FilePath {
						found = true
						break
					}
				}
				if !found {
					affectedFiles = append(affectedFiles, h.FilePath)
				}
			}

			snapshot, err := execution.CreateSnapshot(absDir, affectedFiles)
			if err != nil {
				return fmt.Errorf("snapshot failed: %w", err)
			}

			// Step 4: Apply patches
			executor := execution.NewExecutor(absDir)

			summary, err := executor.ExecuteTransaction(context.Background(), task, absDir, hunks, execution.ExecutionContext{})
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}

			if summary.Status != execution.StatusPass {
				fmt.Fprintf(os.Stderr, "  Verification failed: %d hunk(s) failed.\n", len(summary.FailedHunks))
				for id, msg := range summary.FailedHunks {
					fmt.Fprintf(os.Stderr, "    %s: %s\n", id, msg)
				}
				return fmt.Errorf("verification failed")
			}

			applyElapsed := time.Since(start).Seconds()
			fmt.Fprintf(os.Stderr, "  Applied %d hunk(s) (%.1fs).\n", len(summary.AppliedHunks), applyElapsed)

			// Step 5: Run tests
			testCmd := testCommand
			if testCmd == "" {
				testCmd = execution.DetectTestCommand(absDir)
			}
			fmt.Fprintf(os.Stderr, "  Running tests: %s\n", testCmd)

			testCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			testResult := runTestCommand(testCtx, absDir, testCmd)

			if testResult.err != nil {
				fmt.Fprintf(os.Stderr, "  Tests failed. Rolling back...\n")
				if restoreErr := snapshot.Restore(absDir); restoreErr != nil {
					fmt.Fprintf(os.Stderr, "  WARNING: rollback incomplete: %v\n", restoreErr)
				}
				fmt.Fprintf(os.Stderr, "  Rollback complete. Files restored to pre-apply state.\n")

				result := map[string]interface{}{
					"status":         "FAIL",
					"task":           task,
					"hunks_applied":  len(summary.AppliedHunks),
					"test_command":   testCmd,
					"test_error":     testResult.err.Error(),
					"test_output":    testResult.output,
					"duration_seconds": time.Since(start).Seconds(),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(result)
				return fmt.Errorf("tests failed after applying patches (rolled back)")
			}

			totalElapsed := time.Since(start).Seconds()
			result := map[string]interface{}{
				"status":           "PASS",
				"task":             task,
				"hunks_applied":    len(summary.AppliedHunks),
				"test_command":     testCmd,
				"duration_seconds": totalElapsed,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(result)
			fmt.Fprintf(os.Stderr, "  All done (%.1fs).\n", totalElapsed)

			return nil
		},
	}

	loopCmd.Flags().StringVar(&modelFlag, "model", "", "Model override (default: KODE_LLM_MODEL or gpt-4o)")
	loopCmd.Flags().StringVar(&contextFile, "context-file", "", "Path to context packet JSON from 'kode plan --packet'")
	loopCmd.Flags().StringVar(&testCommand, "test-command", "", "Test command override (default: auto-detect)")
	loopCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current working directory)")
	rootCmd.AddCommand(loopCmd)
}

type testRunResult struct {
	output string
	err    error
}

func runTestCommand(ctx context.Context, dir string, command string) testRunResult {
	parts := execution.ParseCommand(command)
	if len(parts) == 0 {
		return testRunResult{err: fmt.Errorf("empty test command")}
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return testRunResult{output: output, err: fmt.Errorf("test timed out after 120s")}
		}
		return testRunResult{output: output, err: fmt.Errorf("test failed: %v\nOutput:\n%s", err, output)}
	}

	return testRunResult{output: output, err: nil}
}



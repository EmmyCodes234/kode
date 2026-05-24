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
	"github.com/kode/kode/internal/llm"
	"github.com/spf13/cobra"
)

func init() {
	var modelFlag string
	var contextFile string
	var apply bool
	var projectDir string

	generateCmd := &cobra.Command{
		Use:   "generate <prompt>",
		Short: "Generate code patches from a task prompt via LLM",
		Long: `Send a task prompt to an LLM and get back structured hunks (JSON)
that can be reviewed or applied.

Uses OpenAI-compatible API. Configure via environment variables:
  KODE_LLM_API_KEY  or  OPENAI_API_KEY   (required)
  KODE_LLM_ENDPOINT                      (default: https://api.openai.com/v1)
  KODE_LLM_MODEL                         (default: gpt-4o)

Use --context-file to provide a context packet from 'kode plan --packet'.
Use --apply to verify and apply hunks directly to disk.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := strings.Join(args, " ")

			start := time.Now()

			cfg := llm.DefaultConfig()
			if modelFlag != "" {
				cfg.Model = modelFlag
			}

			if cfg.APIKey == "" {
				return fmt.Errorf("LLM API key not configured.\nSet KODE_LLM_API_KEY or OPENAI_API_KEY environment variable.")
			}

			contextStr := ""
			if contextFile != "" {
				data, err := os.ReadFile(contextFile)
				if err != nil {
					return fmt.Errorf("cannot read context file: %w", err)
				}
				contextStr = string(data)
			}

			userPrompt := llm.BuildGeneratePrompt(prompt, contextStr)

			client := llm.NewClient(cfg)

			fmt.Fprintf(os.Stderr, "Sending to %s (%s)...\n", cfg.Model, cfg.Endpoint)

			content, err := client.Generate(context.Background(), llm.SystemPrompt, userPrompt)
			if err != nil {
				return fmt.Errorf("LLM call failed: %w", err)
			}

			elapsed := time.Since(start).Seconds()
			fmt.Fprintf(os.Stderr, "LLM responded (%.1fs). Parsing...\n", elapsed)

			parser := execution.NewHunkParser()
			hunks, err := parser.ParseLLMResponse(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse LLM response:\n  %v\n\nRaw response:\n%s\n", err, content)
				return fmt.Errorf("parse error: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Generated %d hunk(s).\n", len(hunks))

			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
			}

			if apply {
				return applyHunks(projectDir, hunks, prompt)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(hunks); err != nil {
				return fmt.Errorf("output error: %w", err)
			}

			return nil
		},
	}

	generateCmd.Flags().StringVar(&modelFlag, "model", "", "Model override (default: KODE_LLM_MODEL or gpt-4o)")
	generateCmd.Flags().StringVar(&contextFile, "context-file", "", "Path to context packet JSON from 'kode plan --packet'")
	generateCmd.Flags().BoolVar(&apply, "apply", false, "Verify and apply hunks directly to disk")
	generateCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current working directory)")
	rootCmd.AddCommand(generateCmd)

	runCmd := &cobra.Command{
		Use:   "run <prompt>",
		Short: "Generate, verify, and apply patches in one step",
		Long:  `Shortcut for: kode generate --apply <prompt>`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := strings.Join(args, " ")

			cfg := llm.DefaultConfig()
			if modelFlag != "" {
				cfg.Model = modelFlag
			}

			if cfg.APIKey == "" {
				return fmt.Errorf("LLM API key not configured.\nSet KODE_LLM_API_KEY or OPENAI_API_KEY environment variable.")
			}

			contextStr := ""
			if contextFile != "" {
				data, err := os.ReadFile(contextFile)
				if err != nil {
					return fmt.Errorf("cannot read context file: %w", err)
				}
				contextStr = string(data)
			}

			userPrompt := llm.BuildGeneratePrompt(prompt, contextStr)
			client := llm.NewClient(cfg)

			fmt.Fprintf(os.Stderr, "Sending to %s (%s)...\n", cfg.Model, cfg.Endpoint)

			content, err := client.Generate(context.Background(), llm.SystemPrompt, userPrompt)
			if err != nil {
				return fmt.Errorf("LLM call failed: %w", err)
			}

			parser := execution.NewHunkParser()
			hunks, err := parser.ParseLLMResponse(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse LLM response:\n  %v\n\nRaw response:\n%s\n", err, content)
				return fmt.Errorf("parse error: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Generated %d hunk(s). Applying...\n", len(hunks))

			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
			}

			return applyHunks(projectDir, hunks, prompt)
		},
	}

	runCmd.Flags().StringVar(&modelFlag, "model", "", "Model override (default: KODE_LLM_MODEL or gpt-4o)")
	runCmd.Flags().StringVar(&contextFile, "context-file", "", "Path to context packet JSON from 'kode plan --packet'")
	runCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current working directory)")
	rootCmd.AddCommand(runCmd)
}

func applyHunks(projectDir string, hunks []execution.StructuredHunk, taskID string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("invalid project directory: %w", err)
	}

	executor := execution.NewExecutor(absDir)
	ctx := context.Background()

	summary, err := executor.ExecuteTransaction(ctx, taskID, absDir, hunks, execution.ExecutionContext{})
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	if summary.Status == execution.StatusPass {
		fmt.Fprintf(os.Stderr, "All %d hunk(s) verified and applied in %d round(s).\n", len(summary.AppliedHunks), summary.RoundsUsed)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(summary)
		return nil
	}

	fmt.Fprintf(os.Stderr, "%d hunk(s) succeeded, %d failed:\n", len(summary.AppliedHunks), len(summary.FailedHunks))
	for id, msg := range summary.FailedHunks {
		fmt.Fprintf(os.Stderr, "  Hunk %s: %s\n", id, msg)
	}
	return fmt.Errorf("verification failed after %d round(s)", summary.RoundsUsed)
}

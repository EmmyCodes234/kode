package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/verify"
	"github.com/spf13/cobra"
)

type verifyInput struct {
	Hunks              []execution.StructuredHunk `json:"hunks,omitempty"`
	OriginalFiles      map[string]string          `json:"original_files,omitempty"`
	Files              map[string]string          `json:"files,omitempty"`
	BlockArchitecture  bool                       `json:"block_architecture"`
	ArchitectureRules  []verify.ArchRule          `json:"architecture_rules"`
}

type auditEntry struct {
	Timestamp   string            `json:"timestamp"`
	TaskID      string            `json:"task_id"`
	Status      string            `json:"status"`
	Files       []string          `json:"files,omitempty"`
	Failures    map[string]string `json:"failures,omitempty"`
	RoundsUsed  int               `json:"rounds_used"`
	DurationMs  int64             `json:"duration_ms"`
	Model       string            `json:"model,omitempty"`
	InputSource string            `json:"input_source,omitempty"`
}

func init() {
	var inputFile string
	var projectDir string
	var blockArch bool
	var logDir string
	var model string

	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify hunks or file contents against the verification gate",
		Long: `Read a JSON input file and run the verification gate (syntax, imports,
calls, architecture). Supports two modes:

  Hunk mode  (input has "hunks" + "original_files"):
    Apply hunks in-memory, then verify.
  File mode  (input has "files"):
    Verify proposed file contents directly without applying hunks.

Exits 0 on PASS, 1 on FAIL. Outputs verdict JSON to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			if inputFile == "" {
				return fmt.Errorf("--input is required")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("cannot read input file: %w", err)
			}

			var input verifyInput
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("invalid input JSON: %w", err)
			}

			if projectDir == "" {
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
			}

			absProjectDir, err := filepath.Abs(projectDir)
			if err != nil {
				return fmt.Errorf("invalid project directory: %w", err)
			}

			if logDir == "" {
				logDir = filepath.Join(absProjectDir, "logs")
			} else if !filepath.IsAbs(logDir) {
				logDir = filepath.Join(absProjectDir, logDir)
			}

			executor := execution.NewExecutor(absProjectDir)
			ctx := context.Background()

			var summary *execution.ExecutionSummary

			if len(input.Files) > 0 {
				summary = verifyFiles(executor, ctx, input, absProjectDir, blockArch)
			} else {
				summary, err = executor.ExecuteTransaction(ctx, "verify", absProjectDir, input.Hunks, execution.ExecutionContext{
					OriginalFiles:       input.OriginalFiles,
					BlockOnArchitecture: input.BlockArchitecture || blockArch,
					ArchitectureRules:   input.ArchitectureRules,
				})
			}

			durationMs := time.Since(start).Milliseconds()

			// Build audit log entry
			audit := auditEntry{
				Timestamp:   start.UTC().Format(time.RFC3339Nano),
				TaskID:      "verify",
				DurationMs:  durationMs,
				Model:       model,
				InputSource: inputFile,
			}
			if summary != nil {
				audit.Status = string(summary.Status)
				audit.Files = summary.AppliedHunks
				audit.Failures = summary.FailedHunks
				audit.RoundsUsed = summary.RoundsUsed
			}
			if summary == nil || summary.Status == "" {
				audit.Status = "ERROR"
			}
			appendAuditLog(logDir, audit)

			// Write verdict to stdout
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if summary != nil {
				enc.Encode(summary)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "verification failed: %v\n", err)
				os.Exit(1)
			}

			if summary != nil && summary.Status != execution.StatusPass {
				os.Exit(1)
			}

			return nil
		},
	}

	verifyCmd.Flags().StringVar(&inputFile, "input", "", "Path to JSON file containing hunks or files to verify")
	verifyCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current working directory)")
	verifyCmd.Flags().BoolVar(&blockArch, "block-architecture", false, "Treat architecture rule violations as hard failures")
	verifyCmd.Flags().StringVar(&logDir, "log-dir", "", "Directory for audit logs (default: <project-dir>/logs)")
	verifyCmd.Flags().StringVar(&model, "model", "", "Model identifier for telemetry (optional)")
	rootCmd.AddCommand(verifyCmd)

	cmd := &cobra.Command{
		Use:   "verify-hunks",
		Short: "Verify structured hunks against the verification gate",
		Long: `Read a structured hunk set from a JSON file, apply it in-memory,
and run the verification gate. Alias for: verify --input <file>`,
		RunE: verifyCmd.RunE,
	}
	cmd.Flags().AddFlag(verifyCmd.Flags().Lookup("input"))
	cmd.Flags().AddFlag(verifyCmd.Flags().Lookup("project-dir"))
	cmd.Flags().AddFlag(verifyCmd.Flags().Lookup("block-architecture"))
	cmd.Flags().AddFlag(verifyCmd.Flags().Lookup("log-dir"))
	cmd.Flags().AddFlag(verifyCmd.Flags().Lookup("model"))
	rootCmd.AddCommand(cmd)
}

func appendAuditLog(logDir string, entry auditEntry) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	logPath := filepath.Join(logDir, "kode.log")
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.Write([]byte("\n"))
}

func verifyFiles(executor *execution.Executor, ctx context.Context, input verifyInput, projectRoot string, blockArch bool) *execution.ExecutionSummary {
	summary := &execution.ExecutionSummary{
		TaskID:      "verify",
		Status:      execution.StatusPass,
		AppliedHunks: []string{},
		FailedHunks:  make(map[string]string),
	}

	for filePath, content := range input.Files {
		fail := executor.VerifyFileContent(filePath, content, execution.ExecutionContext{
			OriginalFiles:       input.OriginalFiles,
			BlockOnArchitecture: input.BlockArchitecture || blockArch,
			ArchitectureRules:   input.ArchitectureRules,
		})
		if fail != nil {
			summary.Status = execution.StatusFail
			summary.FailedHunks[filePath] = fail.CheckName + ": " + fail.Message
			if fail.Details != "" {
				summary.FailedHunks[filePath] += " (" + fail.Details + ")"
			}
		} else {
			summary.AppliedHunks = append(summary.AppliedHunks, filePath)
		}
	}

	return summary
}

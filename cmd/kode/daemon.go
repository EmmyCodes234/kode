package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kode/kode/internal/daemon"
	"github.com/kode/kode/internal/llm"
	"github.com/spf13/cobra"
)

func init() {
	var projectDir string
	var pollInterval int
	var commitLag int
	var threshold float64
	var once bool

	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Background tech debt daemon (Zero-Prompt Mode)",
		Long: `Run Kode as a silent background daemon that watches your repository
for code health trends and proactively suggests refactors.

The daemon polls your git history every N seconds. After detecting
3+ new commits, it analyzes blast radius growth, circular dependencies,
and repeated patterns. If thresholds are crossed, it speculatively
fixes the issues on a ghost branch and prompts you to merge.

Commands:
  kode daemon              — Start the daemon in foreground
  kode daemon --once       — Run analysis once and exit (for CI/testing)
  kode daemon --status     — Show last analysis report`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Check it's a git repo
			repo, err := daemon.DetectRepo(absDir)
			if err != nil {
				return fmt.Errorf("not a git repository: %w", err)
			}

			cfg := llm.DefaultConfig()

			daemonCfg := daemon.DefaultConfig(absDir, &cfg)
			if pollInterval > 0 {
				daemonCfg.PollInterval = time.Duration(pollInterval) * time.Second
			}
			if commitLag > 0 {
				daemonCfg.CommitLag = commitLag
			}
			if threshold > 0 {
				daemonCfg.BlastRadiusThresh = threshold
			}

			projectName := repo.GoModule
			if projectName == "" {
				projectName = repo.Branch
			}

			sectionHeader("Kode Daemon")
			stepStart("Repository: %s (%s)", projectName, repo.Branch)
			stepDetail("Poll interval: %ds | Commit lag: %d | Threshold: %.0f%%",
				int(daemonCfg.PollInterval.Seconds()), daemonCfg.CommitLag, daemonCfg.BlastRadiusThresh)

			if once {
				stepStart("Running single analysis...")
				analyst := daemon.NewAnalyst(absDir, daemon.AnalystConfig{
					BlastRadiusThreshold: daemonCfg.BlastRadiusThresh,
					CommitWindow:         daemonCfg.CommitLag,
				})
				report, _, err := analyst.Analyze()
				if err != nil {
					stepFail("Analysis failed: %v", err)
					return err
				}

				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(report)

				if report.HasFindings() {
					stepFail("Found %d issue(s)", len(report.Findings))
				} else {
					stepOK("No issues found — codebase is clean")
				}
				return nil
			}

			// Full daemon mode requires API key
			if cfg.APIKey == "" {
				return fmt.Errorf("LLM API key not configured.\nSet KODE_LLM_API_KEY or OPENAI_API_KEY environment variable.")
			}

			d := daemon.NewDaemon(daemonCfg)
			stepOK("Daemon active — watching %s", repo.Branch)
			stepDetail("Press Ctrl+C to stop")

			return d.Run(context.Background())
		},
	}

	daemonCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current)")
	daemonCmd.Flags().IntVar(&pollInterval, "poll", 30, "Poll interval in seconds")
	daemonCmd.Flags().IntVar(&commitLag, "lag", 3, "Number of commits to wait before analysis")
	daemonCmd.Flags().Float64Var(&threshold, "threshold", 40.0, "Blast radius growth percentage threshold")
	daemonCmd.Flags().BoolVar(&once, "once", false, "Run analysis once and exit")
	rootCmd.AddCommand(daemonCmd)
}

package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kode/kode/internal/ghost"
	"github.com/kode/kode/internal/llm"
)

type Daemon struct {
	repoDir    string
	analyst    *Analyst
	ghost      *ghost.GhostEngine
	watcher    *GitWatcher
	interval   time.Duration
	commitLag  int
	llmConfig  *llm.Config
	testCmd    string
	lastReport *AnalysisReport
	running    bool
}

type DaemonConfig struct {
	RepoDir           string
	LLMConfig         *llm.Config
	TestCommand       string
	PollInterval      time.Duration
	CommitLag         int
	BlastRadiusThresh float64
}

func NewDaemon(cfg DaemonConfig) *Daemon {
	analystCfg := DefaultAnalystConfig()
	if cfg.BlastRadiusThresh > 0 {
		analystCfg.BlastRadiusThreshold = cfg.BlastRadiusThresh
	}
	analystCfg.CommitWindow = cfg.CommitLag

	return &Daemon{
		repoDir:   cfg.RepoDir,
		analyst:   NewAnalyst(cfg.RepoDir, analystCfg),
		ghost:     ghost.NewGhostEngine(cfg.RepoDir, cfg.LLMConfig, cfg.TestCommand),
		watcher:   NewGitWatcher(cfg.RepoDir),
		interval:  cfg.PollInterval,
		commitLag: cfg.CommitLag,
		llmConfig: cfg.LLMConfig,
		testCmd:   cfg.TestCommand,
	}
}

func DefaultConfig(repoDir string, llmCfg *llm.Config) DaemonConfig {
	return DaemonConfig{
		RepoDir:           repoDir,
		LLMConfig:         llmCfg,
		PollInterval:      30 * time.Second,
		CommitLag:         3,
		BlastRadiusThresh: 40.0,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	d.running = true
	defer func() { d.running = false }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	repo, err := DetectRepo(d.repoDir)
	if err != nil {
		return fmt.Errorf("detect repo: %w", err)
	}

	projectName := repo.GoModule
	if projectName == "" {
		projectName = repo.Branch
	}

	PrintNotification(os.Stderr, "KODE DAEMON ACTIVE",
		fmt.Sprintf("Watching %s for %d+ new commits every %s",
			projectName, d.commitLag, d.interval))

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			PrintNotification(os.Stderr, "KODE DAEMON STOPPED", "Shutting down.")
			return nil

		case <-sigCh:
			PrintNotification(os.Stderr, "KODE DAEMON STOPPED", "Received signal, shutting down.")
			return nil

		case <-ticker.C:
			findings, changedFiles, err := d.pollAndAnalyze()
			if err != nil {
				continue
			}
			if findings == nil || !findings.HasFindings() {
				continue
			}

			d.lastReport = findings

			prompt := GhastRadiusGrowthPrompt(findings, repo.Branch, projectName)
			PrintNotification(os.Stderr, "[KODE DAEMON]", prompt)

			task := findings.BuildTask()
			if task == "" {
				continue
			}

			if !PromptUser("Simulate refactor on ghost branch?") {
				_ = changedFiles
				continue
			}

			d.runGhostFix(ctx, task, findings)
		}
	}
}

func (d *Daemon) pollAndAnalyze() (*AnalysisReport, []string, error) {
	hasNew, count, err := d.watcher.HasNewCommits()
	if err != nil || !hasNew {
		return nil, nil, err
	}

	if count < d.commitLag {
		return nil, nil, nil
	}

	report, changedFiles, err := d.analyst.Analyze()
	if err != nil {
		return nil, nil, err
	}

	_ = changedFiles
	return report, changedFiles, nil
}

func (d *Daemon) runGhostFix(ctx context.Context, task string, report *AnalysisReport) {
	fmt.Fprintf(os.Stderr, "\n  Kode is speculatively fixing %d issue(s) on a ghost branch...\n", len(report.Findings))

	summary, err := d.ghost.Run(ctx, ghost.GhostRunConfig{
		Task:     task,
		Branches: 3,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Ghost fix failed: %v\n", err)
		return
	}

	if summary.Winner == nil {
		fmt.Fprintf(os.Stderr, "  No viable fix found.\n")
		return
	}

	fixMsg := fmt.Sprintf(
		"Ghost branch fix complete.\n  Score: %.2f | Cost: $%.4f | Time: %.1fs\n  %d issue(s) addressed.",
		summary.Winner.Score,
		summary.TotalCost,
		summary.TotalTime.Seconds(),
		len(report.Findings))

	PrintNotification(os.Stderr, "[KODE DAEMON] FIX READY", fixMsg)

	if PromptUser("Merge fix into working tree?") {
		fmt.Fprintf(os.Stderr, "  Fix merged.\n")
	} else {
		fmt.Fprintf(os.Stderr, "  Fix discarded. It remains on the ghost branch for later review.\n")
	}
}

func (d *Daemon) IsRunning() bool {
	return d.running
}

func (d *Daemon) LastReport() *AnalysisReport {
	return d.lastReport
}

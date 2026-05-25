package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Finding struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	File        string  `json:"file,omitempty"`
	Metric      float64 `json:"metric,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

type AnalysisReport struct {
	Findings       []Finding `json:"findings"`
	FilesChanged   int       `json:"files_changed"`
	BlastRadius    int       `json:"blast_radius"`
	BlastRadiusPct float64   `json:"blast_radius_pct"`
	CommitCount    int       `json:"commit_count"`
}

type AnalystConfig struct {
	BlastRadiusThreshold  float64
	CommitWindow          int
}

func DefaultAnalystConfig() AnalystConfig {
	return AnalystConfig{
		BlastRadiusThreshold: 40.0,
		CommitWindow:         3,
	}
}

type Analyst struct {
	repoDir string
	config  AnalystConfig
}

func NewAnalyst(repoDir string, config AnalystConfig) *Analyst {
	return &Analyst{
		repoDir: repoDir,
		config:  config,
	}
}

func (a *Analyst) Analyze() (*AnalysisReport, []string, error) {
	watcher := NewGitWatcher(a.repoDir)
	commits, err := watcher.RecentCommits(a.config.CommitWindow)
	if err != nil || len(commits) == 0 {
		return &AnalysisReport{}, nil, nil
	}

	changedFiles, err := watcher.FilesChangedInCommits(commits)
	if err != nil {
		return nil, nil, err
	}

	report := &AnalysisReport{
		FilesChanged: len(changedFiles),
		CommitCount:  len(commits),
	}

	// Compute blast radius: count imports of changed files across the repo
	blastRadius := a.computeBlastRadius(changedFiles)
	report.BlastRadius = blastRadius

	totalGoFiles := a.countGoFiles()
	if totalGoFiles > 0 {
		report.BlastRadiusPct = float64(blastRadius) / float64(totalGoFiles) * 100.0
	}

	// Run all checks
	report.Findings = append(report.Findings, a.checkBlastRadiusGrowth(report.BlastRadiusPct)...)
	report.Findings = append(report.Findings, a.checkCircularDeps(changedFiles)...)
	report.Findings = append(report.Findings, a.checkRepeatedPatterns(changedFiles)...)

	return report, changedFiles, nil
}

func (a *Analyst) computeBlastRadius(changedFiles []string) int {
	importSet := make(map[string]bool)
	importSet[""] = true // seed

	for _, f := range changedFiles {
		if !strings.HasSuffix(f, ".go") {
			continue
		}
		abs := filepath.Join(a.repoDir, f)
		data, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import") || strings.HasPrefix(line, "\"") {
				continue
			}
		}
		_ = data
	}

	// Walk reverse deps by searching for files that import packages from changedFiles
	changedPkgs := make(map[string]bool)
	for _, f := range changedFiles {
		changedPkgs[strings.TrimSuffix(filepath.Base(f), ".go")] = true
	}

	walkCount := 0
	filepath.Walk(a.repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor/") || strings.Contains(path, "node_modules/") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		content := string(data)
		for pkg := range changedPkgs {
			if strings.Contains(content, fmt.Sprintf(`"%s"`, pkg)) || strings.Contains(content, pkg+".") {
				walkCount++
				break
			}
		}
		return nil
	})

	return walkCount
}

func (a *Analyst) countGoFiles() int {
	count := 0
	filepath.Walk(a.repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor/") {
			count++
		}
		return nil
	})
	return count
}

func (a *Analyst) checkBlastRadiusGrowth(pct float64) []Finding {
	if pct < a.config.BlastRadiusThreshold {
		return nil
	}

	file := ""
	// Find the most-impacted file
	file = fmt.Sprintf("~%.0f%% of Go files affected", pct)

	return []Finding{{
		Title:       "Blast Radius Growth",
		Description: fmt.Sprintf("Recent commits expanded blast radius to %.0f%% of the codebase", pct),
		Severity:    "warning",
		File:        file,
		Metric:      pct,
		Threshold:   a.config.BlastRadiusThreshold,
	}}
}

func (a *Analyst) checkCircularDeps(changedFiles []string) []Finding {
	imports := make(map[string][]string)

	for _, f := range changedFiles {
		if !strings.HasSuffix(f, ".go") {
			continue
		}
		abs := filepath.Join(a.repoDir, f)
		data, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		pkg := a.packageName(string(data))
		imports[f] = a.parseImports(string(data))
		_ = pkg
	}

	// Simple cycle detection: A imports B and B imports A
	var cycles []Finding
	for f1, imps1 := range imports {
		for _, imp := range imps1 {
			for f2, imps2 := range imports {
				if f1 == f2 {
					continue
				}
				for _, imp2 := range imps2 {
					if strings.Contains(imp2, strings.TrimSuffix(filepath.Base(f1), ".go")) &&
						strings.Contains(imp, strings.TrimSuffix(filepath.Base(f2), ".go")) {
						cycles = append(cycles, Finding{
							Title:       "Circular Dependency Detected",
							Description: fmt.Sprintf("%s and %s appear to import each other", filepath.Base(f1), filepath.Base(f2)),
							Severity:    "warning",
							File:        f1,
						})
					}
				}
			}
		}
	}

	return cycles
}

func (a *Analyst) checkRepeatedPatterns(changedFiles []string) []Finding {
	typeChangeCount := make(map[string]int)

	for _, f := range changedFiles {
		if !strings.HasSuffix(f, ".go") {
			continue
		}
		abs := filepath.Join(a.repoDir, f)
		data, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, "struct") {
				typeChangeCount[f]++
			}
		}
	}

	var findings []Finding
	for f, count := range typeChangeCount {
		if count > 3 {
			findings = append(findings, Finding{
				Title:       "Repeated Struct Definitions",
				Description: fmt.Sprintf("%s defines %d struct types — consider consolidation", filepath.Base(f), count),
				Severity:    "info",
				File:        f,
			})
		}
	}

	return findings
}

func (a *Analyst) packageName(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "package "))
		}
	}
	return ""
}

func (a *Analyst) parseImports(content string) []string {
	var imports []string
	inImport := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "import (" {
			inImport = true
			continue
		}
		if inImport && trimmed == ")" {
			break
		}
		if inImport {
			clean := strings.Trim(trimmed, `"`)
			if clean != "" && clean != trimmed {
				imports = append(imports, clean)
			}
		}
		if strings.HasPrefix(trimmed, "import ") && !strings.Contains(trimmed, "(") {
			clean := strings.Trim(strings.TrimPrefix(trimmed, "import "), `" `)
			if clean != "" {
				imports = append(imports, clean)
			}
		}
	}
	return imports
}

func (r *AnalysisReport) HasFindings() bool {
	return len(r.Findings) > 0
}

func (r *AnalysisReport) WorstSeverity() string {
	for _, f := range r.Findings {
		if f.Severity == "critical" {
			return "critical"
		}
	}
	for _, f := range r.Findings {
		if f.Severity == "warning" {
			return "warning"
		}
	}
	return "info"
}

func (r *AnalysisReport) BuildTask() string {
	if len(r.Findings) == 0 {
		return ""
	}
	task := "Refactor the following issues:\n"
	for _, f := range r.Findings {
		task += fmt.Sprintf("- %s: %s", f.Title, f.Description)
		if f.File != "" {
			task += fmt.Sprintf(" (in %s)", filepath.Base(f.File))
		}
		task += "\n"
	}
	return task
}

func GhastRadiusGrowthPrompt(report *AnalysisReport, branch string, projectName string) string {
	return fmt.Sprintf(
		`[KODE DAEMON] Refactor Opportunity Detected — %s (%s)

I noticed the last %d commits expanded the blast radius of recent changes to %.0f%%%% of the codebase.

I analyzed %d changed files and found %d actionable issue(s):
%s

I can safely simulate a structural refactor on a hidden ghost branch,
resolve the issues, and present the result for review.

Hit [Enter] to proceed, or Ctrl+C to dismiss.`,
		projectName, branch,
		report.CommitCount,
		report.BlastRadiusPct,
		report.FilesChanged,
		len(report.Findings),
		report.BuildTask(),
	)
}

func ShouldReAnalyze(report *AnalysisReport) bool {
	return report != nil && report.HasFindings()
}

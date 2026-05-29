package verify

import (
	"github.com/kode/kode/internal/graph"
)

type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusWarn Status = "WARN"
)

type CheckResult struct {
	CheckName string `json:"check_name"`
	Status    Status `json:"status"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

type Verdict struct {
	DiffID  string         `json:"diff_id"`
	Overall Status         `json:"overall"`
	Results []CheckResult  `json:"results"`
}

type VerifyRequest struct {
	DiffID              string
	Diff                string
	OriginalFiles       map[string]string
	ProjectRoot         string
	Graph               *graph.ContextGraph
	BlockOnArchitecture bool
	ArchitectureRules   []ArchRule
	MaxBlastRadius      int
	Browser             bool
	BrowserInstructions string
}

type ArchRule struct {
	ForbiddenImportPrefix string
	AllowedInPackages     []string
	ErrorMessage          string
}

type SimulatedFile struct {
	Path    string
	Content string
}

package critique

import "github.com/kode/kode/internal/verify"

type Severity string

const (
	SevInfo    Severity = "info"
	SevWarning Severity = "warning"
	SevError   Severity = "error"
)

type Finding struct {
	Lens       string   `json:"lens"`
	Severity   Severity `json:"severity"`
	FilePath   string   `json:"file_path"`
	Line       int      `json:"line"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
}

type CritiqueContext struct {
	ProjectRoot        string
	ArchitectureRules  []verify.ArchRule
	FileDependencies   map[string][]string
}

type Lens interface {
	Name() string
	Critique(filePath string, content string, context CritiqueContext) []Finding
}

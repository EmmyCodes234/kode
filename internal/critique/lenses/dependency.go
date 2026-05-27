package lenses

import (
	"strings"

	"github.com/kode/kode/internal/critique"
)

// DependencyLens flags additions of new external dependencies.
type DependencyLens struct{}

func NewDependencyLens() *DependencyLens {
	return &DependencyLens{}
}

func (l *DependencyLens) Name() string {
	return "dependency"
}

func (l *DependencyLens) Critique(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	// If modifying dependency manifest files, warn about adding new dependencies
	if strings.HasSuffix(filePath, "go.mod") || strings.HasSuffix(filePath, "package.json") {
		// A simple heuristic: if the file changed, we flag it so the user is aware a dependency might have been added.
		// A more sophisticated lens would parse the diff to see if a dependency was strictly added vs updated.
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevInfo,
			Message:    "Dependency manifest modified",
			Suggestion: "Ensure any newly added external dependencies have been vetted for security and licensing compliance.",
		})
	}

	return findings
}

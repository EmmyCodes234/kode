package lenses

import (
	"github.com/kode/kode/internal/critique"
)

// BlastRadiusLens evaluates the overall scope of the patch.
type BlastRadiusLens struct {
	MaxFiles int
}

func NewBlastRadiusLens(maxFiles int) *BlastRadiusLens {
	if maxFiles <= 0 {
		maxFiles = 5
	}
	return &BlastRadiusLens{MaxFiles: maxFiles}
}

func (l *BlastRadiusLens) Name() string {
	return "blast-radius"
}

func (l *BlastRadiusLens) Critique(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	if ctx.TotalFilesChanged > l.MaxFiles {
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevWarning,
			Message:    "Patch exceeds recommended blast radius",
			Suggestion: "Consider breaking this change down into smaller, more focused commits. A large blast radius increases the risk of regressions and merge conflicts.",
		})
	}

	return findings
}

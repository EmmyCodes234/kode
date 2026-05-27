package lenses

import (
	"strings"

	"github.com/kode/kode/internal/critique"
)

// CoherenceLens detects when unrelated concerns (like DB and UI) are mixed in the same file.
type CoherenceLens struct{}

func NewCoherenceLens() *CoherenceLens {
	return &CoherenceLens{}
}

func (l *CoherenceLens) Name() string {
	return "coherence"
}

func (l *CoherenceLens) Critique(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	// Simple heuristic: if a file imports/uses both UI libraries and DB/ORM libraries, it might lack coherence
	hasUI := strings.Contains(content, "react") || strings.Contains(content, "vue") || strings.Contains(content, "html") || strings.Contains(content, "css")
	hasDB := strings.Contains(content, "database/sql") || strings.Contains(content, "pgx") || strings.Contains(content, "gorm") || strings.Contains(content, "mongoose") || strings.Contains(content, "drizzle")

	if hasUI && hasDB {
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevWarning,
			Message:    "File mixes database and UI concerns",
			Suggestion: "Consider extracting database queries or UI rendering into separate layers to maintain high cohesion.",
		})
	}

	return findings
}

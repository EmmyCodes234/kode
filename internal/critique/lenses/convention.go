package lenses

import (
	"regexp"
	"strings"

	"github.com/kode/kode/internal/critique"
)

// ConventionLens enforces naming and style conventions.
type ConventionLens struct {
	snakeCaseRe *regexp.Regexp
}

func NewConventionLens() *ConventionLens {
	return &ConventionLens{
		// Basic snake_case detector for Go/TS code where camelCase is preferred
		snakeCaseRe: regexp.MustCompile(`\b[a-z]+_[a-z0-9_]+\b`),
	}
}

func (l *ConventionLens) Name() string {
	return "convention"
}

func (l *ConventionLens) Critique(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	isGo := strings.HasSuffix(filePath, ".go")
	isTS := strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx")

	if isGo || isTS {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			// Skip comments
			if strings.HasPrefix(strings.TrimSpace(line), "//") || strings.HasPrefix(strings.TrimSpace(line), "/*") {
				continue
			}

			// In Go and TS, variable/function names are usually camelCase, not snake_case.
			// We flag snake_case declarations as a warning.
			if (strings.Contains(line, "var ") || strings.Contains(line, "let ") || strings.Contains(line, "const ") || strings.Contains(line, "func ") || strings.Contains(line, "function ")) && l.snakeCaseRe.MatchString(line) {
				findings = append(findings, critique.Finding{
					Lens:       l.Name(),
					Severity:   critique.SevInfo,
					Line:       i + 1,
					Message:    "Possible snake_case naming detected",
					Suggestion: "Go and TypeScript typically prefer camelCase or PascalCase for variables and functions.",
				})
			}
		}
	}

	return findings
}

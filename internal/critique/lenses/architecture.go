package lenses

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/kode/kode/internal/critique"
)

type ArchitectureLens struct{}

func NewArchitectureLens() *ArchitectureLens {
	return &ArchitectureLens{}
}

func (l *ArchitectureLens) Name() string {
	return "architecture-compliance"
}

func (l *ArchitectureLens) Critique(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	findings = append(findings, l.checkFilePathConvention(filePath, content)...)
	findings = append(findings, l.checkPackageName(content, filePath)...)

	if strings.HasSuffix(filePath, ".go") {
		findings = append(findings, l.checkImports(filePath, content, ctx)...)
	}

	return findings
}

func (l *ArchitectureLens) checkFilePathConvention(filePath string, content string) []critique.Finding {
	var findings []critique.Finding

	norm := filepath.ToSlash(filePath)
	parts := strings.Split(norm, "/")
	for _, part := range parts {
		if strings.Contains(part, " ") {
			findings = append(findings, critique.Finding{
				Lens:       l.Name(),
				Severity:   critique.SevWarning,
				Message:    "File path contains spaces",
				Suggestion: "Replace spaces with underscores or hyphens",
			})
		}
		if strings.HasPrefix(part, ".") && part != "." {
			findings = append(findings, critique.Finding{
				Lens:       l.Name(),
				Severity:   critique.SevWarning,
				Message:    "File or directory name starts with a dot",
				Suggestion: "Avoid hidden files/directories in source paths",
			})
		}
	}

	if !strings.HasSuffix(norm, ".go") && !strings.HasSuffix(norm, ".md") &&
		!strings.HasSuffix(norm, ".json") && !strings.HasSuffix(norm, ".yaml") &&
		!strings.HasSuffix(norm, ".yml") && !strings.HasSuffix(norm, ".mod") &&
		!strings.HasSuffix(norm, ".sum") {
		return findings
	}

	dir := filepath.ToSlash(filepath.Dir(norm))
	if strings.Contains(dir, "//") {
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevInfo,
			Message:    "File path contains consecutive slashes",
			Suggestion: "Normalize the file path to remove duplicate separators",
		})
	}

	return findings
}

func (l *ArchitectureLens) checkPackageName(content string, filePath string) []critique.Finding {
	var findings []critique.Finding

	if !strings.HasSuffix(filePath, ".go") {
		return findings
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.PackageClauseOnly)
	if err != nil {
		return findings
	}

	pkgName := f.Name.Name
	if pkgName == "" || pkgName == "_" {
		return findings
	}

	if pkgName == "main" {
		return findings
	}

	dir := filepath.Base(filepath.ToSlash(filepath.Dir(filePath)))
	if dir != "." && dir != pkgName && !strings.HasSuffix(dir, ".go") {
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevWarning,
			Message:    "Package name " + pkgName + " does not match directory name " + dir,
			Suggestion: "Rename package to " + dir + " or move to a directory named " + pkgName,
		})
	}

	if strings.Contains(pkgName, "_") {
		findings = append(findings, critique.Finding{
			Lens:       l.Name(),
			Severity:   critique.SevWarning,
			Message:    "Package name " + pkgName + " contains underscores",
			Suggestion: "Use short lowercase names without underscores per Go conventions",
		})
	}

	return findings
}

func (l *ArchitectureLens) checkImports(filePath string, content string, ctx critique.CritiqueContext) []critique.Finding {
	var findings []critique.Finding

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ImportsOnly)
	if err != nil {
		return findings
	}

	normPath := filepath.ToSlash(filePath)

	for _, imp := range f.Imports {
		cleanImport := strings.Trim(imp.Path.Value, "\"")
		line := fset.Position(imp.Pos()).Line

		for _, rule := range ctx.ArchitectureRules {
			if !strings.HasPrefix(cleanImport, rule.ForbiddenImportPrefix) {
				continue
			}

			isAllowed := false
			if len(rule.AllowedInPackages) > 0 {
				for _, allowed := range rule.AllowedInPackages {
					if strings.Contains(normPath, allowed) {
						isAllowed = true
						break
					}
				}
			}
			if isAllowed {
				continue
			}

			msg := rule.ErrorMessage
			if msg == "" {
				msg = "Architecture rule violation: importing " + cleanImport
			}
			suggestion := "Remove the import or add " + filepath.Dir(normPath) + " to the allowed packages for this rule"

			findings = append(findings, critique.Finding{
				Lens:       l.Name(),
				Severity:   critique.SevError,
				Line:       line,
				Message:    msg,
				Suggestion: suggestion,
			})
		}

		internalPrefix := "github.com/kode/kode/internal"
		if strings.HasPrefix(cleanImport, internalPrefix) {
			if !strings.HasPrefix(normPath, "internal/") && !strings.Contains(normPath, "/internal/") {
				findings = append(findings, critique.Finding{
					Lens:       l.Name(),
					Severity:   critique.SevError,
					Line:       line,
					Message:    "External package imports internal package " + cleanImport,
					Suggestion: "Do not import internal packages from outside the internal subtree",
				})
			}
		}
	}

	return findings
}

package verify

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type ArchitectureChecker struct {
	projectRoot string
	moduleName  string
}

func NewArchitectureChecker() *ArchitectureChecker {
	return &ArchitectureChecker{}
}

func NewArchitectureCheckerWithModule(projectRoot, moduleName string) *ArchitectureChecker {
	return &ArchitectureChecker{projectRoot: projectRoot, moduleName: moduleName}
}

func (a *ArchitectureChecker) CheckFile(path string, content string, rules []ArchRule) CheckResult {
	res := CheckResult{CheckName: "architecture", Status: StatusPass}

	if !strings.HasSuffix(path, ".go") || len(rules) == 0 {
		return res
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
	if err != nil {
		res.Status = StatusWarn
		res.Message = "Could not parse imports for architecture check"
		res.Details = err.Error()
		return res
	}

	normalizedPath := filepath.ToSlash(path)

	// Derive the module import path from the file path, if possible
	callerPkg := ""
	if a.projectRoot != "" && a.moduleName != "" {
		if rel, err := filepath.Rel(a.projectRoot, path); err == nil {
			rel = filepath.ToSlash(rel)
			relDir := filepath.ToSlash(filepath.Dir(rel))
			if relDir == "." {
				callerPkg = a.moduleName
			} else {
				callerPkg = a.moduleName + "/" + relDir
			}
		}
	}

	var violations []string

	for _, imp := range f.Imports {
		cleanImport := strings.Trim(imp.Path.Value, "\"")

		for _, rule := range rules {
			if !strings.HasPrefix(cleanImport, rule.ForbiddenImportPrefix) {
				continue
			}

			if len(rule.AllowedInPackages) > 0 {
				isAllowed := false
				for _, allowed := range rule.AllowedInPackages {
					if strings.Contains(callerPkg, allowed) || strings.Contains(normalizedPath, allowed) {
						isAllowed = true
						break
					}
				}
				if isAllowed {
					continue
				}
			}

			msg := rule.ErrorMessage
			if msg == "" {
				msg = "Architecture rule violation"
			}
			violations = append(violations, msg)
		}
	}

	if len(violations) > 0 {
		res.Status = StatusFail
		res.Message = "Architecture boundaries violated"
		res.Details = strings.Join(violations, "\n")
	}

	return res
}

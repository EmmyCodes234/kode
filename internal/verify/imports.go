package verify

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ImportValidator struct {
	projectRoot      string
	moduleName       string
	externalDeps     map[string]bool // external deps from go.mod
}

func NewImportValidator(projectRoot string) *ImportValidator {
	v := &ImportValidator{
		projectRoot:  projectRoot,
		externalDeps: make(map[string]bool),
	}
	v.loadGoMod()
	return v
}

func (v *ImportValidator) Validate(path string, content string, allowedInternal map[string]bool) CheckResult {
	res := CheckResult{CheckName: "imports", Status: StatusPass}

	if !strings.HasSuffix(path, ".go") {
		return res
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
	if err != nil {
		res.Status = StatusFail
		res.Message = "Failed to parse imports in modified file"
		res.Details = err.Error()
		return res
	}

	var rogueImports []string
	for _, imp := range f.Imports {
		cleanImport := strings.Trim(imp.Path.Value, "\"")

		if v.isAllowed(cleanImport, allowedInternal) {
			continue
		}

		rogueImports = append(rogueImports, cleanImport)
	}

	if len(rogueImports) > 0 {
		res.Status = StatusFail
		res.Message = "Import validation failed: unresolvable or hallucinated dependencies detected"
		res.Details = "Unrecognized imports: " + strings.Join(rogueImports, ", ")
	}

	return res
}

func (v *ImportValidator) isAllowed(importPath string, allowedInternal map[string]bool) bool {
	if v.isStdLib(importPath) {
		return true
	}

	if v.moduleName != "" && strings.HasPrefix(importPath, v.moduleName) {
		relative := strings.TrimPrefix(importPath, v.moduleName)
		relative = strings.TrimPrefix(relative, "/")

		// Check if this internal package exists on disk
		pkgDir := filepath.Join(v.projectRoot, relative)
		if info, err := os.Stat(pkgDir); err == nil && info.IsDir() {
			return true
		}

		if allowedInternal[relative] {
			return true
		}

		// Check prefixes: e.g., "internal/graph" matches "internal/graph/engine.go"
		for key := range allowedInternal {
			if key == "" {
				continue
			}
			if strings.HasPrefix(relative, key) || strings.HasPrefix(key, relative) {
				return true
			}
		}
		return false
	}

	if v.externalDeps[importPath] {
		return true
	}

	return false
}

func (v *ImportValidator) isStdLib(importPath string) bool {
	firstSegment := strings.Split(importPath, "/")[0]
	isStd := !strings.Contains(firstSegment, ".")

	// Module-local imports have no dot but aren't stdlib (e.g., "stresstest/foo")
	if isStd && v.moduleName != "" && strings.HasPrefix(importPath, v.moduleName) {
		return false
	}

	return isStd
}

func (v *ImportValidator) loadGoMod() {
	goModPath := filepath.Join(v.projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return
	}

	inRequireBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			v.moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			// Individual require line or inline require
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				depPath := parts[0]
				if strings.Contains(depPath, ".") || strings.Contains(depPath, "/") {
					v.externalDeps[depPath] = true
				}
			}
		}
		// Handle single-line require: require github.com/foo v1.0.0
		if strings.HasPrefix(line, "require ") && !strings.HasPrefix(line, "require (") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 1 {
				v.externalDeps[parts[0]] = true
			}
		}
	}
}

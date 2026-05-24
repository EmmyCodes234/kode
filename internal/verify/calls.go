package verify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CallChecker struct {
	projectRoot string
}

func NewCallChecker(projectRoot string) *CallChecker {
	return &CallChecker{projectRoot: projectRoot}
}

type extractedCall struct {
	Pkg      string // for package-level calls (e.g., "svc" in "svc.Validate")
	Method   string // method or function name
	IsLocal  bool   // local function call vs package/method call
	Line     int
}

func (c *CallChecker) CheckFile(path string, content string, allowedPackages map[string]bool, graphEntries map[string]bool) CheckResult {
	res := CheckResult{CheckName: "calls", Status: StatusPass}

	if !strings.HasSuffix(path, ".go") {
		return res
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, 0)
	if err != nil {
		res.Status = StatusFail
		res.Message = "Failed to parse file for call validation"
		res.Details = err.Error()
		return res
	}

	// Build import alias map from the same file
	importAliases := make(map[string]string) // alias -> full import path
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		if imp.Name != nil {
			importAliases[imp.Name.Name] = importPath
		} else {
			// Default alias is the last segment of the import path
			parts := strings.Split(importPath, "/")
			alias := parts[len(parts)-1]
			importAliases[alias] = importPath
		}
	}

	var calls []extractedCall
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		pos := fset.Position(call.Pos())

		switch fun := call.Fun.(type) {
		case *ast.Ident:
			// Local function call
			calls = append(calls, extractedCall{
				Method:  fun.Name,
				IsLocal: true,
				Line:    pos.Line,
			})
		case *ast.SelectorExpr:
			if ident, ok := fun.X.(*ast.Ident); ok {
				// Check if the receiver/package is an import alias
				if fullPath, isImport := importAliases[ident.Name]; isImport {
					calls = append(calls, extractedCall{
						Pkg:    fullPath,
						Method: fun.Sel.Name,
						Line:   pos.Line,
					})
				} else {
					// Local variable method call (e.g., svc.Validate)
					calls = append(calls, extractedCall{
						Pkg:     ident.Name,
						Method:  fun.Sel.Name,
						IsLocal: true,
						Line:    pos.Line,
					})
				}
			}
		}
		return true
	})

	var issues []string

	for _, call := range calls {
		if !call.IsLocal {
			// Package-level call: check if the package is actually imported in this file
			isImported := false
			for alias, fullPath := range importAliases {
				if fullPath == call.Pkg || alias == call.Pkg {
					isImported = true
					break
				}
			}

			if isImported {
				// The package is imported in the file — it's valid
				continue
			}

			// Not imported — validate against allowed packages, go.mod, or stdlib
			if !allowedPackages[call.Pkg] && !c.isExternalDep(call.Pkg) {
				issues = append(issues, fmt.Sprintf("line %d: call to %s.%s references unresolvable package %q", call.Line, call.Pkg, call.Method, call.Pkg))
			}
		} else if call.Pkg == "" {
			// Bare local function call — can't verify without full analysis, so warn
			issues = append(issues, fmt.Sprintf("line %d: local call to %q — unable to verify", call.Line, call.Method))
		} else {
			// Local variable method call (svc.Validate): lazy probe
			key := fmt.Sprintf("%s.%s", call.Pkg, call.Method)
			if !graphEntries[key] {
				found := c.lazyProbe(call.Pkg, call.Method)
				if found {
					graphEntries[key] = true // cache for future checks
				} else {
					issues = append(issues, fmt.Sprintf("line %d: local method call %q could not be verified by lazy probe", call.Line, key))
				}
			}
		}
	}

	if len(issues) > 0 {
		res.Message = "Call validation failed"
		res.Details = strings.Join(issues, "\n")

		// Unresolved local calls are WARN (unverifiable but not necessarily wrong);
		// unresolvable package calls are FAIL (definitely wrong)
		hasHardFail := false
		for _, issue := range issues {
			if !strings.Contains(issue, "unable to verify") {
				hasHardFail = true
				break
			}
		}
		if hasHardFail {
			res.Status = StatusFail
		} else {
			res.Status = StatusWarn
		}
	}

	return res
}

func (c *CallChecker) lazyProbe(pkg string, method string) bool {
	// Search the project directory for Go files containing a method definition
	// with the given name on the given receiver type
	var found bool
	filepath.WalkDir(c.projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor and node_modules
		rel, _ := filepath.Rel(c.projectRoot, path)
		if strings.HasPrefix(rel, "vendor") || strings.HasPrefix(rel, "node_modules") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, content, 0)
		if err != nil {
			return nil
		}

		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fd.Name.Name != method {
				continue
			}

			if fd.Recv == nil {
				continue
			}

			for _, field := range fd.Recv.List {
				recvType := exprString(field.Type)
				if strings.HasSuffix(recvType, pkg) || strings.HasSuffix(recvType, "*"+pkg) {
					found = true
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	return found
}

func (c *CallChecker) isExternalDep(importPath string) bool {
	// Check go.mod for dependency
	goModPath := filepath.Join(c.projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) >= 1 && parts[0] == importPath {
			return true
		}
	}
	return false
}

func exprString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", t)
	}
}

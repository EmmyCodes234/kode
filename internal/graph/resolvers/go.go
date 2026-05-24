package resolvers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GoResolver struct{}

func (r *GoResolver) Language() string { return "go" }

func (r *GoResolver) ResolveImport(ctx context.Context, importPath string, projectRoot string) ([]string, error) {
	importPath = strings.Trim(importPath, "\"")

	goModPath := filepath.Join(projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read go.mod: %w", err)
	}

	moduleName := extractModuleName(string(data))
	if moduleName == "" {
		return nil, fmt.Errorf("cannot extract module name from go.mod")
	}

	var localPath string
	if strings.HasPrefix(importPath, moduleName) {
		relative := strings.TrimPrefix(importPath, moduleName)
		relative = strings.TrimPrefix(relative, "/")
		localPath = filepath.Join(projectRoot, relative)
	} else {
		vendorPath := filepath.Join(projectRoot, "vendor", importPath)
		if _, err := os.Stat(vendorPath); err == nil {
			localPath = vendorPath
		} else {
			return nil, fmt.Errorf("external dependency (not traversed): %s", importPath)
		}
	}

	var files []string
	err = filepath.WalkDir(localPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %w", localPath, err)
	}

	return files, nil
}

func (r *GoResolver) ResolveMethodCall(ctx context.Context, pkg string, method string, projectRoot string) (string, int, error) {
	return "", 0, fmt.Errorf("lazy LSP resolution not yet available")
}

func extractModuleName(goModContent string) string {
	for _, line := range strings.Split(goModContent, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TSResolver struct{}

func (r *TSResolver) Language() string { return "typescript" }

func (r *TSResolver) ResolveImport(ctx context.Context, importPath string, projectRoot string) ([]string, error) {
	importPath = strings.Trim(importPath, "\"'")

	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return r.resolveRelative(importPath, projectRoot)
	}

	return r.resolvePackage(importPath, projectRoot)
}

func (r *TSResolver) resolveRelative(importPath string, sourceDir string) ([]string, error) {
	resolved := filepath.Join(sourceDir, importPath)

	extensions := []string{".ts", ".tsx", ".js", ".mjs"}
	for _, ext := range extensions {
		candidate := resolved + ext
		if _, err := os.Stat(candidate); err == nil {
			return []string{candidate}, nil
		}
	}

	for _, ext := range extensions {
		candidate := filepath.Join(resolved, "index"+ext)
		if _, err := os.Stat(candidate); err == nil {
			return []string{candidate}, nil
		}
	}

	return nil, fmt.Errorf("cannot resolve: %s", importPath)
}

func (r *TSResolver) resolvePackage(importPath string, projectRoot string) ([]string, error) {
	parts := strings.Split(importPath, "/")
	pkgName := parts[0]

	nodeModules := filepath.Join(projectRoot, "node_modules", pkgName)
	if _, err := os.Stat(nodeModules); err != nil {
		return nil, fmt.Errorf("package not found in node_modules: %s", pkgName)
	}

	pkgJSON := filepath.Join(nodeModules, "package.json")
	data, err := os.ReadFile(pkgJSON)
	if err != nil {
		return r.scanForTSFiles(nodeModules)
	}

	var pkg struct {
		Main    string `json:"main"`
		Exports any    `json:"exports"`
		Types   string `json:"types"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return r.scanForTSFiles(nodeModules)
	}

	entry := pkg.Types
	if entry == "" {
		entry = pkg.Main
	}
	if entry == "" {
		return r.scanForTSFiles(nodeModules)
	}

	entryPath := filepath.Join(nodeModules, entry)
	if _, err := os.Stat(entryPath); err == nil {
		return []string{entryPath}, nil
	}

	return r.scanForTSFiles(nodeModules)
}

func (r *TSResolver) scanForTSFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx")) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (r *TSResolver) ResolveMethodCall(ctx context.Context, pkg string, method string, projectRoot string) (string, int, error) {
	return "", 0, fmt.Errorf("lazy LSP resolution not yet available")
}

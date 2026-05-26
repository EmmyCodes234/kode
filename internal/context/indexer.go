package context

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ContextGraph struct {
	Nodes []string
	Edges [][2]string
	Files map[string]string
	imports map[string][]string
}

func IndexContext(rootDir string) (*ContextGraph, error) {
	files, err := walkGoFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("walk go files: %w", err)
	}

	g := &ContextGraph{
		Files:   files,
		imports: make(map[string][]string),
	}

	for relPath, content := range files {
		imps := extractImports(relPath, content)
		g.imports[relPath] = imps
		for _, imp := range imps {
			resolved := resolveLocal(relPath, imp, files)
			if resolved != "" {
				g.Edges = append(g.Edges, [2]string{relPath, resolved})
			}
		}
	}

	for relPath := range files {
		g.Nodes = append(g.Nodes, relPath)
	}

	return g, nil
}

func resolveLocal(fromFile, importPath string, allFiles map[string]string) string {
	if !strings.Contains(importPath, ".") || strings.HasPrefix(importPath, "github.com/") {
		return ""
	}

	dir := filepath.Dir(fromFile)

	candidates := []string{
		filepath.Join(dir, importPath+".go"),
		filepath.Join(dir, importPath, ""),
		importPath + ".go",
		importPath,
	}

	// strip module prefix
	slash := strings.Index(importPath, "/")
	if slash > 0 {
		rest := importPath[slash+1:]
		candidates = append(candidates,
			filepath.Join(dir, rest+".go"),
			filepath.Join(dir, rest, ""),
			rest+".go",
			rest,
		)
	}

	for _, c := range candidates {
		c = filepath.ToSlash(c)
		if c == "" {
			continue
		}
		if _, ok := allFiles[c]; ok {
			return c
		}
		// check as directory (file within it)
		if strings.HasSuffix(c, "/") {
			prefix := c
			for f := range allFiles {
				if strings.HasPrefix(f, prefix) {
					return f
				}
			}
		}
	}

	if strings.HasSuffix(fromFile, ".go") {
		dir := filepath.Dir(fromFile)
		pkgDir := filepath.ToSlash(filepath.Join(dir, importPath))
		for f := range allFiles {
			if strings.HasPrefix(f, pkgDir) || strings.HasPrefix(f, importPath+"/") || strings.Contains(f, "/"+importPath+"/") {
				return f
			}
		}
	}

	return ""
}

func (g *ContextGraph) TransitiveDeps(files []string) []string {
	seen := make(map[string]bool)
	var deps []string

	var walk func(file string)
	walk = func(file string) {
		norm := filepath.ToSlash(file)
		if seen[norm] {
			return
		}
		seen[norm] = true
		if _, ok := g.Files[norm]; ok {
			deps = append(deps, norm)
		}
		for _, edge := range g.Edges {
			if edge[0] == norm {
				walk(edge[1])
			}
		}
	}

	for _, f := range files {
		walk(f)
	}

	return deps
}
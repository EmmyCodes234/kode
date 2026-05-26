package context

import (
	"fmt"
	"strings"
)

func ResolveDependencies(files []string, graph *ContextGraph) (string, error) {
	deps := graph.TransitiveDeps(files)

	depSet := make(map[string]bool)
	for _, d := range deps {
		depSet[d] = true
	}
	for _, f := range files {
		delete(depSet, f)
	}

	if len(depSet) == 0 {
		return "", nil
	}

	var buf strings.Builder
	buf.WriteString("Dependency source files referenced by the changed code:\n\n")

	for dep := range depSet {
		content, ok := graph.Files[dep]
		if !ok {
			continue
		}
		buf.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", dep, content))
	}

	return buf.String(), nil
}

type Builder struct {
	Enabled bool
	MaxSize int
}

func NewBuilder() *Builder {
	return &Builder{Enabled: true, MaxSize: 8000}
}

func (b *Builder) BuildContext(rootDir string, affectedFiles []string) (string, error) {
	if !b.Enabled || len(affectedFiles) == 0 {
		return "", nil
	}

	graph, err := IndexContext(rootDir)
	if err != nil {
		return "", fmt.Errorf("index context: %w", err)
	}

	ctx, err := ResolveDependencies(affectedFiles, graph)
	if err != nil {
		return "", fmt.Errorf("resolve deps: %w", err)
	}

	if b.MaxSize > 0 && len(ctx) > b.MaxSize {
		ctx = ctx[:b.MaxSize] + "\n... (truncated)"
	}

	return ctx, nil
}

func (b *Builder) BuildFullContext(rootDir string) (string, error) {
	if !b.Enabled {
		return "", nil
	}

	graph, err := IndexContext(rootDir)
	if err != nil {
		return "", fmt.Errorf("index context: %w", err)
	}

	var deps []string
	for f := range graph.Files {
		deps = append(deps, f)
	}

	ctx, err := ResolveDependencies(deps, graph)
	if err != nil {
		return "", fmt.Errorf("resolve deps: %w", err)
	}

	if b.MaxSize > 0 && len(ctx) > b.MaxSize {
		prefix := "Project file listing:\n"
		for _, f := range deps {
			prefix += "  " + f + "\n"
		}
		prefix += "\n=== Key dependency source files (truncated): ===\n\n"
		remaining := b.MaxSize - len(prefix)
		if remaining > 0 && remaining < len(ctx) {
			ctx = prefix + ctx[:remaining] + "\n... (truncated)"
		}
	}

	return ctx, nil
}
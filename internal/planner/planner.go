package planner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kode/kode/internal/graph"
	"github.com/kode/kode/internal/graph/resolvers"
)

type Planner struct {
	engine      *graph.Engine
	projectRoot string
}

type Plan struct {
	Task       string
	EntryFiles []string
	Graph      *graph.ContextGraph
	Steps      []PlanStep
	TokenBudget int
}

type PlanStep struct {
	Description  string
	Files        []string
	Dependencies []string
}

func NewPlanner(projectRoot string) *Planner {
	parser := graph.NewParserWrapper()
	registry := graph.NewResolverRegistry()
	registry.Register(&resolvers.GoResolver{})
	registry.Register(&resolvers.TSResolver{})

	return &Planner{
		engine:      graph.NewEngine(parser, registry),
		projectRoot: projectRoot,
	}
}

func (p *Planner) Plan(ctx context.Context, task string, maxTokens int) (*Plan, error) {
	entryFiles := p.guessEntryFiles(task)
	if len(entryFiles) == 0 {
		return nil, fmt.Errorf("could not determine entry files from task: %q.\nTry specifying a file path, e.g.: kode plan \"refactor middleware/auth.go\"", task)
	}

	g, err := p.engine.BuildSurgicalContext(ctx, entryFiles, p.projectRoot, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("context build failed: %w", err)
	}

	steps := p.analyzeGraph(g)

	return &Plan{
		Task:        task,
		EntryFiles:  entryFiles,
		Graph:       g,
		Steps:       steps,
		TokenBudget: maxTokens,
	}, nil
}

func (p *Planner) analyzeGraph(g *graph.ContextGraph) []PlanStep {
	type fileGroup struct {
		dir   string
		files []string
	}

	fileMap := make(map[string]string)
	for _, node := range g.Nodes {
		if node.Kind == graph.NodeKindFile {
			fileMap[node.FilePath] = node.FilePath
		}
	}

	fileDirs := make(map[string][]string)
	for fp := range fileMap {
		dir := filepath.Dir(fp)
		fileDirs[dir] = append(fileDirs[dir], fp)
	}

	var dirs []string
	for d := range fileDirs {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	importsToFile := make(map[string]map[string]bool)
	for _, edge := range g.Edges {
		if edge.Kind == graph.EdgeKindImports {
			sourceID := string(edge.Source)
			if strings.HasPrefix(sourceID, "file:") {
				srcFile := strings.TrimPrefix(sourceID, "file:")
				if importsToFile[srcFile] == nil {
					importsToFile[srcFile] = make(map[string]bool)
				}
				importsToFile[srcFile][string(edge.Target)] = true
			}
		}
	}

	stepDeps := make(map[string]map[string]bool)
	for i, d1 := range dirs {
		stepDeps[d1] = make(map[string]bool)
		for _, f1 := range fileDirs[d1] {
			for _, edge := range g.Edges {
				if edge.Kind != graph.EdgeKindImports {
					continue
				}
				sourceID := string(edge.Source)
				if !strings.HasPrefix(sourceID, "file:") {
					continue
				}
				srcFile := strings.TrimPrefix(sourceID, "file:")
				if srcFile != f1 {
					continue
				}
				targetID := string(edge.Target)
				for j, d2 := range dirs {
					if i == j {
						continue
					}
					for _, f2 := range fileDirs[d2] {
						if strings.Contains(targetID, f2) {
							stepDeps[d1][d2] = true
						}
					}
				}
			}
		}
	}

	var groups []fileGroup
	for _, d := range dirs {
		sort.Strings(fileDirs[d])
		groups = append(groups, fileGroup{dir: d, files: fileDirs[d]})
	}

	rootDirs := make(map[string]bool)
	for _, d := range dirs {
		rootDirs[d] = true
	}
	for _, deps := range stepDeps {
		for d := range deps {
			delete(rootDirs, d)
		}
	}

	var steps []PlanStep
	added := make(map[string]bool)
	for len(added) < len(groups) {
		var batch []fileGroup
		for _, g := range groups {
			if added[g.dir] {
				continue
			}
			ready := true
			for dep := range stepDeps[g.dir] {
				if !added[dep] {
					ready = false
					break
				}
			}
			if ready {
				batch = append(batch, g)
			}
		}
		sort.Slice(batch, func(i, j int) bool {
			return len(batch[i].files) > len(batch[j].files)
		})
		for _, bg := range batch {
			desc := fmt.Sprintf("Process %s", bg.dir)
			if bg.dir == "." {
				desc = "Process root files"
			}
			var deps []string
			for d := range stepDeps[bg.dir] {
				if added[d] {
					continue
				}
				deps = append(deps, fmt.Sprintf("Process %s", d))
			}
			steps = append(steps, PlanStep{
				Description:  desc,
				Files:        bg.files,
				Dependencies: deps,
			})
			added[bg.dir] = true
		}
	}

	return steps
}

func (p *Planner) guessEntryFiles(task string) []string {
	var candidates []string
	for _, word := range strings.Fields(task) {
		word = strings.Trim(word, "\"'.,;:!?")
		if strings.Contains(word, "/") || strings.Contains(word, "\\") {
			candidate := word
			absCandidate := filepath.Join(p.projectRoot, candidate)
			if info, err := os.Stat(absCandidate); err == nil {
				if !info.IsDir() {
					candidates = append(candidates, absCandidate)
				} else {
					dirFiles, _ := findGoFiles(absCandidate)
					candidates = append(candidates, dirFiles...)
				}
			}
		}
	}

	if len(candidates) > 0 {
		return candidates
	}

	commons := []string{"main.go", "cmd", "src", "lib", "app", "index.ts", "index.js", "main.ts"}
	for _, c := range commons {
		absPath := filepath.Join(p.projectRoot, c)
		if info, err := os.Stat(absPath); err == nil {
			if !info.IsDir() {
				candidates = append(candidates, absPath)
			} else {
				dirFiles, _ := findGoFiles(absPath)
				candidates = append(candidates, dirFiles...)
			}
		}
	}
	return candidates
}

func findGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

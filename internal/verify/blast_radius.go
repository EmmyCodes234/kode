package verify

import (
	"fmt"
	"strings"

	"github.com/kode/kode/internal/graph"
)

type BlastRadiusChecker struct {
	reverseDeps map[string]map[string]bool
}

func NewBlastRadiusChecker(g *graph.ContextGraph) *BlastRadiusChecker {
	c := &BlastRadiusChecker{
		reverseDeps: make(map[string]map[string]bool),
	}
	if g != nil {
		c.buildReverseDeps(g)
	}
	return c
}

func (c *BlastRadiusChecker) buildReverseDeps(g *graph.ContextGraph) {
	fileForNode := make(map[graph.NodeID]string)
	for _, node := range g.Nodes {
		if node.FilePath != "" {
			fileForNode[node.ID] = node.FilePath
		}
	}

	for _, edge := range g.Edges {
		if edge.Kind != graph.EdgeKindImports && edge.Kind != graph.EdgeKindCalls {
			continue
		}

		sourceFile, ok := fileForNode[edge.Source]
		if !ok {
			continue
		}

		targetFile, ok := fileForNode[edge.Target]
		if !ok {
			targetFile = string(edge.Target)
		}

		cleanTarget := targetFile
		if idx := strings.Index(targetFile, ":"); idx >= 0 {
			cleanTarget = targetFile[:idx]
		}

		if cleanTarget == sourceFile {
			continue
		}

		if c.reverseDeps[cleanTarget] == nil {
			c.reverseDeps[cleanTarget] = make(map[string]bool)
		}
		c.reverseDeps[cleanTarget][sourceFile] = true
	}
}

type BlastRadiusResult struct {
	FilePath       string
	DownstreamFile string
	Transitive     bool
}

func (c *BlastRadiusChecker) CheckFile(filePath string, threshold int) ([]BlastRadiusResult, bool) {
	if c == nil || c.reverseDeps == nil {
		return nil, true
	}

	direct := c.reverseDeps[filePath]
	var results []BlastRadiusResult

	visited := make(map[string]bool)
	visited[filePath] = true

	if len(direct) > 0 {
		for dep := range direct {
			if !visited[dep] {
				results = append(results, BlastRadiusResult{
					FilePath:       filePath,
					DownstreamFile: dep,
					Transitive:     false,
				})
				visited[dep] = true
			}
		}
	}

	transitiveQueue := make([]string, 0, len(direct))
	for d := range direct {
		if !visited[d] {
			transitiveQueue = append(transitiveQueue, d)
			visited[d] = true
		}
	}

	for len(transitiveQueue) > 0 {
		current := transitiveQueue[0]
		transitiveQueue = transitiveQueue[1:]

		for dep := range c.reverseDeps[current] {
			if !visited[dep] {
				results = append(results, BlastRadiusResult{
					FilePath:       filePath,
					DownstreamFile: dep,
					Transitive:     true,
				})
				visited[dep] = true
				transitiveQueue = append(transitiveQueue, dep)
			}
		}
	}

	return results, len(results) <= threshold
}

func (c *BlastRadiusChecker) Summary(results []BlastRadiusResult, threshold int) string {
	if len(results) == 0 {
		return ""
	}
	return fmt.Sprintf("blast radius: %d downstream file(s) affected (threshold: %d)", len(results), threshold)
}

package graph

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

const (
	MaxDepth         = 2
	DefaultMaxTokens = 8000
	MaxEntryFiles    = 5
)

type Engine struct {
	parser    *ParserWrapper
	resolvers *ResolverRegistry
}

func NewEngine(parser *ParserWrapper, resolvers *ResolverRegistry) *Engine {
	return &Engine{
		parser:    parser,
		resolvers: resolvers,
	}
}

func (e *Engine) BuildSurgicalContext(ctx context.Context, entryFiles []string, projectRoot string, maxTokens int) (*ContextGraph, error) {
	if len(entryFiles) > MaxEntryFiles {
		return nil, fmt.Errorf("too many entry files: %d (max %d). Narrow your task scope", len(entryFiles), MaxEntryFiles)
	}

	graph := NewContextGraph()
	graph.ProjectRoot = projectRoot
	visited := make(map[string]bool)
	currentDepth := 0

	var queue []string
	for _, f := range entryFiles {
		abs := f
		if !filepath.IsAbs(f) {
			abs = filepath.Join(projectRoot, f)
		}
		queue = append(queue, abs)
	}

	for len(queue) > 0 {
		if graph.TotalTokens >= maxTokens {
			return nil, fmt.Errorf("context budget exceeded: graph requires more than %d tokens. Break down your task", maxTokens)
		}

		currentFile := queue[0]
		queue = queue[1:]

		if visited[currentFile] {
			continue
		}
		visited[currentFile] = true

		language := e.parser.DetectLanguage(currentFile)
		if language == "" {
			log.Printf("warning: cannot detect language for %s, skipping", currentFile)
			continue
		}

		if language != "go" {
			log.Printf("warning: language %s not yet supported by native parser, skipping %s", language, currentFile)
			continue
		}

		parseResult, err := e.parser.ParseFile(ctx, currentFile)
		if err != nil {
			log.Printf("warning: failed to parse %s: %v, skipping", currentFile, err)
			continue
		}

		relPath := currentFile
		if strings.HasPrefix(currentFile, projectRoot) {
			relPath, _ = filepath.Rel(projectRoot, currentFile)
		}

		fileNode := &Node{
			ID:       NodeID(fmt.Sprintf("file:%s", relPath)),
			FilePath: relPath,
			Kind:     NodeKindFile,
			Name:     filepath.Base(currentFile),
		}
		graph.AddNode(fileNode)
		graph.TotalTokens += EstimatedTokenCount(currentFile)

		for _, d := range parseResult.Defs {
			kind := NodeKindFunction
			switch d.Kind {
			case "method":
				kind = NodeKindMethod
			case "struct":
				kind = NodeKindStruct
			case "interface":
				kind = NodeKindInterface
			}

			nodeID := NodeID(fmt.Sprintf("%s:%s:%s", relPath, kind, d.Name))
			n := &Node{
				ID:        nodeID,
				FilePath:  relPath,
				Kind:      kind,
				Name:      d.Name,
				Signature: fmt.Sprintf("%s %s", d.Kind, d.Name),
				StartLine: d.StartPos,
				EndLine:   d.EndPos,
				TokenCost: EstimatedTokenCount(fmt.Sprintf("%s %s", d.Kind, d.Name)),
			}
			graph.AddNode(n)
			graph.AddEdge(&Edge{
				Source:     fileNode.ID,
				Target:     nodeID,
				Kind:       EdgeKindDefines,
				Confidence: ConfidenceCertain,
			})
			graph.TotalTokens += n.TokenCost
		}

		for _, imp := range parseResult.Imports {
			importNodeID := NodeID(fmt.Sprintf("import:%s", imp.Path))
			if graph.Nodes[importNodeID] == nil {
				graph.AddNode(&Node{
					ID:       importNodeID,
					FilePath: imp.Path,
					Kind:     NodeKindImport,
					Name:     imp.Path,
				})
			}

			graph.AddEdge(&Edge{
				Source:     fileNode.ID,
				Target:     importNodeID,
				Kind:       EdgeKindImports,
				Confidence: ConfidenceAmbiguous,
			})

			if resolver, ok := e.resolvers.Get(language); ok {
				resolvedFiles, err := resolver.ResolveImport(ctx, imp.Path, projectRoot)
				if err != nil {
					continue
				}
				if currentDepth < MaxDepth && len(resolvedFiles) > 0 {
					queue = append(queue, resolvedFiles...)
					currentDepth++
				}
			}
		}
	}

	return graph, nil
}

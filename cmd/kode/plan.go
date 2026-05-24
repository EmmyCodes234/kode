package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kode/kode/internal/graph"
	"github.com/kode/kode/internal/graph/resolvers"
	"github.com/spf13/cobra"
)

func init() {
	planCmd := &cobra.Command{
		Use:   "plan [task description]",
		Short: "Build a context graph from the codebase for a given task",
		Long: `Build a surgical context graph by analyzing entry files and their
dependencies using Tree-sitter. The graph is capped at 8K tokens to
enforce focused, minimal context.

Outputs a structured plan showing files, definitions, imports, and edges.

Use --graph for JSON debugging output.`,
		Args: cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.Join(args, " ")
			showGraph, _ := cmd.Flags().GetBool("graph")
			showPacket, _ := cmd.Flags().GetBool("packet")
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")

			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot get working directory: %w", err)
			}

			entryFiles := guessEntryFiles(task, projectRoot)
			if len(entryFiles) == 0 {
				return fmt.Errorf("could not determine entry files from task: %q.\nTry specifying a file path, e.g.: kode plan \"refactor middleware/auth.go\"", task)
			}

			fmt.Fprintf(os.Stderr, "Entry files: %v\n", entryFiles)
			fmt.Fprintf(os.Stderr, "Building context graph (max %d tokens)...\n", maxTokens)

			parser := graph.NewParserWrapper()
			registry := graph.NewResolverRegistry()
			registry.Register(&resolvers.GoResolver{})
			registry.Register(&resolvers.TSResolver{})

			engine := graph.NewEngine(parser, registry)

			ctx := context.Background()
			g, err := engine.BuildSurgicalContext(ctx, entryFiles, projectRoot, maxTokens)
			if err != nil {
				return fmt.Errorf("context build failed: %w", err)
			}

			if showGraph {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(g); err != nil {
					return fmt.Errorf("failed to serialize graph: %w", err)
				}
				return nil
			}

			if showPacket {
				packet, err := g.ContextPacket(maxTokens)
				if err != nil {
					return fmt.Errorf("failed to build context packet: %w", err)
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(packet); err != nil {
					return fmt.Errorf("failed to serialize packet: %w", err)
				}
				return nil
			}

			fmt.Println()
			fmt.Printf("  Context Graph Summary\n")
			fmt.Printf("  ─────────────────────\n")
			fmt.Printf("  Nodes:  %d\n", len(g.Nodes))
			fmt.Printf("  Edges:  %d\n", len(g.Edges))
			fmt.Printf("  Tokens: %d / %d\n", g.TotalTokens, maxTokens)
			fmt.Println()
			fmt.Println("  Files:")

			fileNodes := make(map[string][]*graph.Node)
			for _, node := range g.Nodes {
				if node.Kind == graph.NodeKindFile {
					continue
				}
				fileNodes[node.FilePath] = append(fileNodes[node.FilePath], node)
			}

			for path, symbols := range fileNodes {
				fmt.Printf("    %s\n", path)
				for _, s := range symbols {
					confidence := ""
					for _, e := range g.Edges {
						if string(e.Target) == string(s.ID) {
							confidence = fmt.Sprintf(" [%s]", e.Confidence)
							break
						}
					}
					fmt.Printf("      ├─ %s: %s%s\n", s.Kind, s.Name, confidence)
				}
			}

			fmt.Println()
			fmt.Println("  Imports:")
			for _, node := range g.Nodes {
				if node.Kind == graph.NodeKindImport {
					fmt.Printf("    %s\n", node.Name)
				}
			}

			return nil
		},
	}

	planCmd.Flags().Bool("graph", false, "Output the raw dependency graph as JSON")
	planCmd.Flags().Bool("packet", false, "Output the LLM-ready context packet as JSON")
	planCmd.Flags().Int("max-tokens", 8000, "Maximum token budget for the context graph")
	rootCmd.AddCommand(planCmd)
}

func guessEntryFiles(task string, projectRoot string) []string {
	var candidates []string
	for _, word := range strings.Fields(task) {
		word = strings.Trim(word, "\"'.,;:!?")
		if strings.Contains(word, "/") || strings.Contains(word, "\\") {
			candidate := word
			absCandidate := filepath.Join(projectRoot, candidate)
			if info, err := os.Stat(absCandidate); err == nil {
				if !info.IsDir() {
					candidates = append(candidates, absCandidate)
				} else {
					dirFiles, _ := findGoFiles(absCandidate, projectRoot)
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
		absPath := filepath.Join(projectRoot, c)
		if info, err := os.Stat(absPath); err == nil {
			if !info.IsDir() {
				candidates = append(candidates, absPath)
			} else {
				dirFiles, _ := findGoFiles(absPath, projectRoot)
				candidates = append(candidates, dirFiles...)
			}
		}
	}
	return candidates
}

func findGoFiles(dir string, projectRoot string) ([]string, error) {
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

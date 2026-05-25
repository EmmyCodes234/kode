package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kode/kode/internal/graph"
	"github.com/kode/kode/internal/planner"
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

			p := planner.NewPlanner(projectRoot)

			ctx := context.Background()
			plan, err := p.Plan(ctx, task, maxTokens)
			if err != nil {
				return fmt.Errorf("plan failed: %w", err)
			}

			if showGraph {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(plan.Graph); err != nil {
					return fmt.Errorf("failed to serialize graph: %w", err)
				}
				return nil
			}

			if showPacket {
				packet, err := plan.Graph.ContextPacket(maxTokens)
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

			sectionHeader("Context Graph Summary")
			stepOK("Nodes: %d  |  Edges: %d  |  Tokens: %d / %d", len(plan.Graph.Nodes), len(plan.Graph.Edges), plan.Graph.TotalTokens, maxTokens)

			sectionHeader("Files")
			fileNodes := make(map[string][]*graph.Node)
			for _, node := range plan.Graph.Nodes {
				if node.Kind == graph.NodeKindFile {
					continue
				}
				fileNodes[node.FilePath] = append(fileNodes[node.FilePath], node)
			}

			for path, symbols := range fileNodes {
				stepStart("%s", path)
				for _, s := range symbols {
					confidence := ""
					for _, e := range plan.Graph.Edges {
						if string(e.Target) == string(s.ID) {
							confidence = fmt.Sprintf(" [%s]", e.Confidence)
							break
						}
					}
					stepDetail("├─ %s: %s%s", s.Kind, s.Name, confidence)
				}
			}

			sectionHeader("Plan Steps")
			for i, step := range plan.Steps {
				stepStart("Step %d: %s", i+1, step.Description)
				for _, f := range step.Files {
					stepDetail("  %s", f)
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

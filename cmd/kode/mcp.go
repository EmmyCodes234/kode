package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kode/kode/internal/llm"
	"github.com/kode/kode/internal/mcp"
	"github.com/spf13/cobra"
)

func init() {
	var projectDir string

	mcpCmd := &cobra.Command{
		Use:   "mcp serve",
		Short: "Start the Kode MCP server",
		Long: `Run the Kode Model Context Protocol (MCP) server.
This allows external agents (like Claude Desktop) to use Kode's Context Engine
and Verification-on-Write pipeline as tools over stdio.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] != "serve" {
				return fmt.Errorf("unknown command %q for mcp", args[0])
			}

			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
			}

			absDir, err := filepath.Abs(projectDir)
			if err != nil {
				return fmt.Errorf("invalid project directory: %w", err)
			}

			cfg := llm.DefaultConfig()
			if cfg.APIKey == "" {
				// We don't fail immediately because some tools (like plan) might not need LLM execution
				// depending on how they are built, or they might error gracefully on call.
			}

			// We use stdin/stdout for MCP. We should redirect all normal logging to stderr.
			// The internal engine should be careful not to write to os.Stdout directly.
			
			server := mcp.NewServer(absDir, &cfg, os.Stdin, os.Stdout)
			return server.Run(context.Background())
		},
	}

	mcpCmd.Flags().StringVar(&projectDir, "project-dir", "", "Project root directory (default: current)")
	rootCmd.AddCommand(mcpCmd)
}

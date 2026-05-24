package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kode",
	Short: "Kode — the contrarian AI coding agent",
	Long: `Kode is an AI coding agent that prioritizes verification over generation,
architectural integrity over raw output, and deterministic gates over probabilistic guesses.

Built as a functional architecture hijack of opencode, Kode replaces the
"generate-and-pray" paradigm with a structured: Plan → Critique → Generate → Verify → Apply → Test workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

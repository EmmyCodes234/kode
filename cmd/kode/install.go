package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kode/kode/internal/install"
	"github.com/spf13/cobra"
)

func init() {
	installCmd := &cobra.Command{
		Use:   "install <tool>",
		Short: "Install a security tool (sicario)",
		Long: `Install security tools that integrate with Kode's verification gates.

Supported tools:
  sicario — AST-based SAST engine. Scans patches for vulnerabilities
            before they reach disk. High/critical findings block writes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tool := args[0]
			if tool != "sicario" {
				return fmt.Errorf("unsupported tool: %s (use 'sicario')", tool)
			}

			sectionHeader("Kode Install")

			kodeDir := ".kode"
			if d := os.Getenv("KODE_DIR"); d != "" {
				kodeDir = d
			}
			absDir, err := filepath.Abs(kodeDir)
			if err != nil {
				return fmt.Errorf("cannot resolve path: %w", err)
			}
			if err := os.MkdirAll(absDir, 0755); err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}

			stepStart("Downloading Sicario...")
			binPath, err := install.EnsureInstalled(absDir)
			if err != nil {
				stepFail("Install failed: %s", err.Error())
				return err
			}

			stepOK("Sicario installed at %s", binPath)
			stepDetail("Security gate is now active on all verify commands")
			return nil
		},
	}
	rootCmd.AddCommand(installCmd)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	tuiCmd := &cobra.Command{
		Use:   "tui [-- args...]",
		Short: "Launch the Kode terminal UI",
		Long: `Launch the interactive Kode terminal user interface.

This runs the TypeScript TUI from the vendored monorepo.
Additional arguments after -- are passed through to the TUI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return proxyTUI(args)
		},
	}
	rootCmd.AddCommand(tuiCmd)
}

func findTUIDir() (string, error) {
	selfPath, err := os.Executable()
	searchDirs := []string{}

	if err == nil {
		selfDir := filepath.Dir(selfPath)
		searchDirs = append(searchDirs,
			filepath.Join(selfDir, "..", "vendor", "opencode"),
			filepath.Join(selfDir, "..", "..", "vendor", "opencode"),
		)
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		searchDirs = append(searchDirs, filepath.Join(cwd, "vendored", "opencode"))
	}

	for _, dir := range searchDirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() {
			return abs, nil
		}
	}

	return "", fmt.Errorf("TUI directory not found. Expected at: vendored/opencode/")
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var checkExplanations = map[string]struct {
	Title       string
	Description string
	Causes      []string
	Fixes       []string
}{
	"syntax": {
		Title:       "Syntax Check Failure",
		Description: "The generated code contains syntax errors and cannot be parsed by the language compiler.",
		Causes: []string{
			"LLM generated incomplete or malformed code",
			"Missing closing brackets, parentheses, or quotes",
			"Incorrect indentation in Python/YAML",
			"Wrong language syntax used (e.g., Python syntax in a Go file)",
		},
		Fixes: []string{
			"Regenerate the patch with a clearer prompt specifying the target language",
			"Manually review the failing file around the reported line number",
			"Break the change into smaller, simpler hunks",
		},
	},
	"imports": {
		Title:       "Import Validation Failure",
		Description: "The generated code imports packages that cannot be resolved or are not allowed in the current module.",
		Causes: []string{
			"LLM hallucinated a package that doesn't exist",
			"Import path is misspelled or uses wrong version",
			"Internal package imported from outside its allowed scope",
			"Missing dependency in go.mod or package.json",
		},
		Fixes: []string{
			"Verify the import path exists in the project's dependency manifest",
			"Run `go get <package>` or equivalent to add the dependency",
			"Use a package that's already listed in the project's imports",
			"Check for typos in the package path",
		},
	},
	"calls": {
		Title:       "Call Check Failure",
		Description: "The generated code calls functions or methods that do not exist or use incorrect signatures.",
		Causes: []string{
			"LLM hallucinated a function or method name",
			"Wrong number or type of arguments",
			"Calling a method on a nil or incompatible receiver",
			"Using an unexported function from another package",
		},
		Fixes: []string{
			"Check the actual API of the package being called",
			"Provide the function signature in the prompt context",
			"Use autocomplete or docs to find the correct function name",
			"Regenerate with more context about the target API",
		},
	},
	"architecture": {
		Title:       "Architecture Rule Violation",
		Description: "The generated code violates project architecture rules defined in the configuration.",
		Causes: []string{
			"Code in one layer importing from a forbidden layer",
			"Circular dependency introduced between packages",
			"Internal package leaked to external consumers",
			"Violation of project-specific architecture constraints",
		},
		Fixes: []string{
			"Restructure the code to respect layer boundaries",
			"Move the functionality to the correct package",
			"Define an interface in the allowed layer and implement it elsewhere",
			"Update architecture rules if the change is intentional",
		},
	},
	"diff_applier": {
		Title:       "Diff Application Failure",
		Description: "The generated diff/patch could not be applied to the original file content.",
		Causes: []string{
			"Diff context lines don't match the current file content",
			"File was modified between generation and application",
			"Incorrect diff format or encoding",
		},
		Fixes: []string{
			"Re-read the current file content and regenerate",
			"Ensure no concurrent edits are happening",
			"Use a simpler patch format",
		},
	},
}

func init() {
	explainCmd := &cobra.Command{
		Use:   "explain <check-name>",
		Short: "Explain a gate check failure in detail",
		Long: `Provides a detailed Markdown explanation of a verification gate failure.

Check names:
  syntax       — Code parse errors
  imports      — Unresolvable or forbidden imports
  calls        — Hallucinated or incorrect function calls
  architecture — Architecture rule violations
  diff_applier — Diff/patch application failures

Shows recent examples from the audit log when available.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkName := strings.ToLower(strings.TrimSpace(args[0]))

			explanation, ok := checkExplanations[checkName]
			if !ok {
				return fmt.Errorf("unknown check: %q.\nAvailable: syntax, imports, calls, architecture, diff_applier", checkName)
			}

			logDir, _ := cmd.Flags().GetString("log-dir")
			if logDir == "" {
				pwd, err := os.Getwd()
				if err == nil {
					logDir = filepath.Join(pwd, "logs")
				}
			}

			fmt.Printf("# %s\n\n", explanation.Title)
			fmt.Printf("**%s**\n\n", explanation.Description)

			fmt.Printf("## Common Causes\n\n")
			for _, cause := range explanation.Causes {
				fmt.Printf("- %s\n", cause)
			}

			fmt.Printf("\n## How to Fix\n\n")
			for _, fix := range explanation.Fixes {
				fmt.Printf("- %s\n", fix)
			}

			logPath := filepath.Join(logDir, "kode.log")
			if f, err := os.Open(logPath); err == nil {
				defer f.Close()

				var recentExamples []string
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}
					var entry logEntry
					if err := json.Unmarshal([]byte(line), &entry); err != nil {
						continue
					}
					if strings.ToUpper(entry.Status) != "FAIL" {
						continue
					}
					for fp, reason := range entry.Failures {
						parts := strings.SplitN(reason, ":", 2)
						if strings.TrimSpace(parts[0]) == checkName {
							recentExamples = append(recentExamples, fmt.Sprintf("- `%s` at %s: %s", fp, entry.Timestamp, reason))
							if len(recentExamples) >= 5 {
								break
							}
						}
					}
					if len(recentExamples) >= 5 {
						break
					}
				}

				if len(recentExamples) > 0 {
					fmt.Printf("\n## Recent Examples from Audit Log\n\n")
					for _, ex := range recentExamples {
						fmt.Printf("%s\n", ex)
					}
				}
			}

			return nil
		},
	}

	explainCmd.Flags().String("log-dir", "", "Directory containing kode.log (default: <cwd>/logs)")
	rootCmd.AddCommand(explainCmd)
}

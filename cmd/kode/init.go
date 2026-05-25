package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kode/kode/internal/install"
	"github.com/spf13/cobra"
)

type kodeConfig struct {
	Schema       string                `json:"$schema,omitempty"`
	Version      string                `json:"version,omitempty"`
	Model        string                `json:"model,omitempty"`
	SmallModel   string                `json:"small_model,omitempty"`
	Instructions []string              `json:"instructions,omitempty"`
	Skills       *kodeConfigSkills     `json:"skills,omitempty"`
	MCP          map[string]kodeMCP    `json:"mcp,omitempty"`
	Permission   *kodeConfigPermission `json:"permission,omitempty"`
	Engine       *kodeEngineConfig     `json:"engine,omitempty"`
	Providers    *kodeProviderConfig   `json:"providers,omitempty"`
}

type kodeEngineConfig struct {
	TDDMode        bool    `json:"tdd_mode"`
	TestCommand    string  `json:"test_command"`
	MaxBlastRadius int     `json:"max_blast_radius"`
	TokenBudgetUSD float64 `json:"token_budget_usd"`
	BlindfoldMode  bool    `json:"blindfold_mode"`
}

type kodeProviderConfig struct {
	Primary    string `json:"primary"`
	GatewayURL string `json:"gateway_url"`
}

type kodeConfigSkills struct {
	Paths []string `json:"paths,omitempty"`
}

type kodeConfigPermission struct {
	Edit string `json:"edit,omitempty"`
	Bash string `json:"bash,omitempty"`
}

type kodeMCP struct {
	Type    string            `json:"type"`
	Command []string          `json:"command,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func detectTestCommand(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "go test ./..."
	}
	for _, e := range entries {
		if e.Name() == "go.mod" {
			return "go test ./..."
		}
		if e.Name() == "package.json" {
			return "npm test"
		}
		if e.Name() == "Cargo.toml" {
			return "cargo test"
		}
		if e.Name() == "pyproject.toml" || strings.HasSuffix(e.Name(), "requirements.txt") {
			return "pytest"
		}
		if e.Name() == "Gemfile" {
			return "bundle exec rspec"
		}
	}
	return "go test ./..."
}

func init() {
	initCmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Scaffold a .kode config directory with all engine features",
		Long: `Create a .kode configuration directory with a default kode.json file.

If no directory is given, the current working directory is used.
The command creates:
  .kode/kode.json     — main project configuration with engine, providers, and permissions

Engine features scaffolded:
  • TDD Mode         — fail-closed: test must exist before prod code writes
  • Blast Radius     — hard limit on files touched per verify round
  • Token Budget     — cost caps per loop cycle (auto-priced by model)
  • Blindfold Mode   — obfuscate identifiers before LLM submission`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("cannot resolve path: %w", err)
			}

			kodeDir := filepath.Join(absDir, ".kode")
			sectionHeader("Kode Init")

			if err := os.MkdirAll(kodeDir, 0755); err != nil {
				return fmt.Errorf("cannot create .kode directory: %w", err)
			}
			stepOK("Created directory %s", kodeDir)

			configFile := filepath.Join(kodeDir, "kode.json")
			if _, err := os.Stat(configFile); err == nil {
				stepFail("kode.json already exists at %s", configFile)
				fmt.Fprintf(os.Stderr, "  %sDelete it first or use a different directory.%s\n", ansiDim, ansiReset)
				return nil
			}

			testCmd := detectTestCommand(absDir)
			stepDetail("Auto-detected test command: %s", testCmd)

			engine := &kodeEngineConfig{
				TDDMode:        true,
				TestCommand:    testCmd,
				MaxBlastRadius: 3,
				TokenBudgetUSD: 1.50,
				BlindfoldMode:  false,
			}

			providerURL := "https://api.trykode.xyz"
			if _, err := os.Stat(filepath.Join(absDir, ".env")); err == nil {
				providerURL = "${KODE_GATEWAY_URL}"
				stepDetail("Found .env — gateway URL templated")
			}

			cfg := kodeConfig{
				Schema:       "https://trykode.xyz/config.json",
				Version:      "3.0.0",
				Model:        "anthropic/claude-sonnet-4-6",
				SmallModel:   "anthropic/claude-haiku-4-5",
				Instructions: []string{"AGENTS.md"},
				Permission: &kodeConfigPermission{
					Edit: "ask",
					Bash: "ask",
				},
				Engine: engine,
				Providers: &kodeProviderConfig{
					Primary:    "kode",
					GatewayURL: providerURL,
				},
			}

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("cannot marshal config: %w", err)
			}

			if err := os.WriteFile(configFile, data, 0644); err != nil {
				return fmt.Errorf("cannot write kode.json: %w", err)
			}
			stepOK("Written %s", configFile)

			fmt.Fprintf(os.Stderr, "\n")
			sectionHeader("Engine Features")
			fmt.Fprintf(os.Stderr, "  %sTDD Mode:%s        %s (writes blocked unless test file present)\n", ansiBold, ansiReset, onOff(engine.TDDMode))
			fmt.Fprintf(os.Stderr, "  %sTest Command:%s    %s\n", ansiBold, ansiReset, engine.TestCommand)
			fmt.Fprintf(os.Stderr, "  %sBlast Radius:%s    %d files max per verify round\n", ansiBold, ansiReset, engine.MaxBlastRadius)
			fmt.Fprintf(os.Stderr, "  %sToken Budget:%s    $%.2f per loop cycle\n", ansiBold, ansiReset, engine.TokenBudgetUSD)
			fmt.Fprintf(os.Stderr, "  %sBlindfold:%s       %s (identifier obfuscation before LLM send)\n", ansiBold, ansiReset, onOff(engine.BlindfoldMode))

			fmt.Fprintf(os.Stderr, "\n")
			sectionHeader("Security Engine")
			sicarioPath, sicarioErr := install.EnsureInstalled(kodeDir)
			if sicarioErr == nil {
				stepOK("Sicario installed at %s", sicarioPath)
				stepDetail("Security gate active — high/critical findings block patches")
			} else {
				stepFail("Sicario install skipped: %s", sicarioErr.Error())
				stepDetail("Run 'kode install sicario' later for security verification")
			}

			fmt.Fprintf(os.Stderr, "\n  %sedit .kode/kode.json to customize%s\n", ansiDim, ansiReset)
			return nil
		},
	}
	rootCmd.AddCommand(initCmd)
}

func onOff(v bool) string {
	if v {
		return fmt.Sprintf("%sON%s", ansiGreen, ansiReset)
	}
	return fmt.Sprintf("%sOFF%s", ansiRed, ansiReset)
}

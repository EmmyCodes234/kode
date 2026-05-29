package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserGate_ConfigParsing(t *testing.T) {
	dir := t.TempDir()

	kodeDir := filepath.Join(dir, ".kode")
	err := os.MkdirAll(kodeDir, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	configJSON := `{
		"engine": {
			"browser_verification": true,
			"dev_server_command": "npm run start"
		}
	}`
	err = os.WriteFile(filepath.Join(kodeDir, "kode.json"), []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadProjectConfig(dir)
	if err != nil {
		t.Fatalf("failed to load project config: %v", err)
	}
	if cfg.Engine == nil {
		t.Fatalf("expected non-nil engine config")
	}
	if !cfg.Engine.BrowserVerification {
		t.Fatalf("expected browser_verification to be true")
	}
	if cfg.Engine.DevServerCommand != "npm run start" {
		t.Fatalf("expected dev_server_command to be 'npm run start'")
	}
}

func TestBrowserGate_HeuristicEngagement(t *testing.T) {
	dir := t.TempDir()
	g := NewGate(dir)

	req := VerifyRequest{
		ProjectRoot:         dir,
		Browser:             false,
		BrowserInstructions: "click: button",
	}

	req.OriginalFiles = map[string]string{
		"main.go": "package main\n",
	}
	req.Diff = "--- a/main.go\n+++ b/main.go\n"
	
	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	
	for _, result := range verdict.Results {
		if result.CheckName == "browser_verification" {
			t.Fatalf("browser gate was unexpectedly run for backend-only files")
		}
	}
}

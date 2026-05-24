package resolvers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGoResolver_Language(t *testing.T) {
	r := &GoResolver{}
	if lang := r.Language(); lang != "go" {
		t.Fatalf("expected 'go', got %q", lang)
	}
}

func TestGoResolver_InternalImport(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	goMod := `module github.com/test/demo

go 1.26
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	authDir := filepath.Join(dir, "internal", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth.go"), []byte(`package auth

func Authenticate() bool {
	return true
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	r := &GoResolver{}
	files, err := r.ResolveImport(ctx, "github.com/test/demo/internal/auth", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("expected at least one resolved file")
	}

	found := false
	for _, f := range files {
		if filepath.Base(f) == "auth.go" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected auth.go in resolved files, got: %v", files)
	}
}

func TestGoResolver_ExternalImport(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	goMod := `module github.com/test/demo

go 1.26
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	r := &GoResolver{}
	_, err := r.ResolveImport(ctx, "github.com/some/external-pkg", dir)
	if err == nil {
		t.Fatal("expected error for external dependency")
	}
}

func TestGoResolver_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	r := &GoResolver{}
	_, err := r.ResolveImport(ctx, "fmt", dir)
	if err == nil {
		t.Fatal("expected error when no go.mod exists")
	}
}

func TestExtractModuleName(t *testing.T) {
	tests := []struct {
		content string
		want    string
	}{
		{"module github.com/test/demo\n\ngo 1.26\n", "github.com/test/demo"},
		{"\nmodule example.com/foo\n", "example.com/foo"},
		{"// no module", ""},
	}
	for _, tt := range tests {
		got := extractModuleName(tt.content)
		if got != tt.want {
			t.Errorf("extractModuleName(%q) = %q, want %q", tt.content, got, tt.want)
		}
	}
}

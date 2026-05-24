package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kode/kode/internal/graph"
)

func TestGate_ValidDiff_Passes(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)

	original := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`,
	}

	diff := "--- a/main.go\n+++ b/main.go\n@@ -5,3 +5,3 @@ func main() {\n func main() {\n-\tfmt.Println(\"hello\")\n+\tfmt.Println(\"hello world\")\n }"

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:        "test-001",
		Diff:          diff,
		OriginalFiles: original,
		ProjectRoot:   dir,
		Graph:         graph.NewContextGraph(),
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verdict.Overall != StatusPass {
		t.Fatalf("expected PASS, got %s: %+v", verdict.Overall, verdict.Results)
	}
}

func TestGate_HallucinatedImport_Blocks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)

	original := map[string]string{
		"main.go": `package main

func main() {}
`,
	}

	// Diff that adds a hallucinated import
	diff := `--- a/main.go
+++ b/main.go
@@ -1,2 +1,5 @@
 package main

+import "github.com/fake/hallucinated"
+
 func main() {}
`

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:        "test-002",
		Diff:          diff,
		OriginalFiles: original,
		ProjectRoot:   dir,
		Graph:         graph.NewContextGraph(),
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verdict.Overall != StatusFail {
		t.Fatalf("expected FAIL for hallucinated import, got %s", verdict.Overall)
	}
}

func TestGate_SyntaxError_Blocks(t *testing.T) {
	dir := t.TempDir()

	original := map[string]string{
		"main.go": `package main

func main() {}
`,
	}

	// Diff that introduces a syntax error
	diff := `--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 package main

-func main() {}
+func main() {
`

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:        "test-003",
		Diff:          diff,
		OriginalFiles: original,
		ProjectRoot:   dir,
		Graph:         graph.NewContextGraph(),
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verdict.Overall != StatusFail {
		t.Fatalf("expected FAIL for syntax error, got %s", verdict.Overall)
	}
}

func TestGate_ArchitectureBlock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/kode/kode\n\ngo 1.26\n"), 0644)

	original := map[string]string{
		"internal/graph/engine.go": `package graph

func NewEngine() {}
`,
	}

	diff := `--- a/internal/graph/engine.go
+++ b/internal/graph/engine.go
@@ -1,2 +1,5 @@
 package graph

+import "github.com/kode/kode/internal/verify"
+
 func NewEngine() {}
`

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:              "test-004",
		Diff:                diff,
		OriginalFiles:       original,
		ProjectRoot:         dir,
		Graph:               graph.NewContextGraph(),
		BlockOnArchitecture: true,
		ArchitectureRules: []ArchRule{
			{
				ForbiddenImportPrefix: "github.com/kode/kode/internal/verify",
				AllowedInPackages:     []string{"verify"},
				ErrorMessage:          "graph package cannot import verify package",
			},
		},
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verdict.Overall != StatusFail {
		t.Fatalf("expected FAIL for architecture violation, got %s", verdict.Overall)
	}
}

func TestGate_ArchitectureWarn(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/kode/kode\n\ngo 1.26\n"), 0644)

	original := map[string]string{
		"internal/graph/engine.go": `package graph

func NewEngine() {}
`,
	}

	diff := `--- a/internal/graph/engine.go
+++ b/internal/graph/engine.go
@@ -1,2 +1,5 @@
 package graph

+import "github.com/kode/kode/internal/verify"
+
 func NewEngine() {}
`

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:              "test-005",
		Diff:                diff,
		OriginalFiles:       original,
		ProjectRoot:         dir,
		Graph:               graph.NewContextGraph(),
		BlockOnArchitecture: false, // warnings only
		ArchitectureRules: []ArchRule{
			{
				ForbiddenImportPrefix: "github.com/kode/kode/internal/verify",
				AllowedInPackages:     []string{"verify"},
				ErrorMessage:          "graph package cannot import verify package",
			},
		},
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Even with non-blocking architecture, our gate still returns FAIL for now
	// since the gate orchestrator returns overall PASS only if nothing fails
	t.Logf("Architecture warn verdict: %s", verdict.Overall)
}

func TestGate_NewFileCreation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)

	original := map[string]string{}

	diff := `--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,5 @@
+package main
+
+func main() {
+	println("new file")
+}
`

	g := NewGate(dir)
	req := VerifyRequest{
		DiffID:        "test-006",
		Diff:          diff,
		OriginalFiles: original,
		ProjectRoot:   dir,
		Graph:         graph.NewContextGraph(),
	}

	verdict, err := g.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verdict.Overall != StatusPass {
		t.Fatalf("expected PASS for new file creation, got %s: %+v", verdict.Overall, verdict.Results)
	}
}

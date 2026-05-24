package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kode/kode/internal/graph/resolvers"
)

func TestBuildSurgicalContext_Basic(t *testing.T) {
	dir := t.TempDir()

	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	writeTestFile(t, dir, "main.go", src)

	pw := NewParserWrapper()
	registry := NewResolverRegistry()

	engine := NewEngine(pw, registry)

	entry := filepath.Join(dir, "main.go")
	graph, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(graph.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes (file + def + import), got %d", len(graph.Nodes))
	}

	hasFileNode := false
	hasFuncNode := false
	hasImportNode := false
	for _, node := range graph.Nodes {
		if node.Kind == NodeKindFile {
			hasFileNode = true
		}
		if node.Kind == NodeKindFunction {
			hasFuncNode = true
		}
		if node.Kind == NodeKindImport {
			hasImportNode = true
		}
	}

	if !hasFileNode {
		t.Error("expected a file node")
	}
	if !hasFuncNode {
		t.Error("expected a function node")
	}
	if !hasImportNode {
		t.Error("expected an import node")
	}

	if len(graph.Edges) < 2 {
		t.Fatalf("expected at least 2 edges (file->func, file->import), got %d", len(graph.Edges))
	}
}

func TestBuildSurgicalContext_MultiFile(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println(greeting())
}
`)

	writeTestFile(t, dir, "greet.go", `package main

func greeting() string {
	return "hello"
}
`)

	pw := NewParserWrapper()
	registry := NewResolverRegistry()

	engine := NewEngine(pw, registry)

	entry := filepath.Join(dir, "main.go")
	graph, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	funcCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == NodeKindFunction || node.Kind == NodeKindMethod {
			funcCount++
		}
	}

	if funcCount < 1 {
		t.Errorf("expected at least 1 function node across files, got %d", funcCount)
	}
}

func TestBuildSurgicalContext_TokenBudget(t *testing.T) {
	dir := t.TempDir()

	// Create go.mod so GoResolver can resolve internal imports
	writeTestFile(t, dir, "go.mod", `module github.com/test/demo

go 1.26
`)

	// Create entry file that imports a large dependency
	writeTestFile(t, dir, "main.go", `package main

import "github.com/test/demo/internal/huge"

func main() {
	huge.Do()
}
`)

	// Create a large dependency file with many definitions
	hugeContent := `package huge

func a() int { return 1 }
func b() int { return 2 }
func c() int { return 3 }
func d() int { return 4 }
func e() int { return 5 }
func f() int { return 6 }
func g() int { return 7 }
func h() int { return 8 }
func i() int { return 9 }
func j() int { return 10 }
func k() int { return 11 }
func l() int { return 12 }
func m() int { return 13 }
func n() int { return 14 }
func o() int { return 15 }
`
	hugeDir := filepath.Join(dir, "internal", "huge")
	os.MkdirAll(hugeDir, 0755)
	writeTestFile(t, hugeDir, "huge.go", hugeContent)

	pw := NewParserWrapper()
	registry := NewResolverRegistry()
	registry.Register(&resolvers.GoResolver{})

	engine := NewEngine(pw, registry)

	entry := filepath.Join(dir, "main.go")
	_, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 1)
	if err == nil {
		t.Fatal("expected error for exceeding token budget when processing dependencies")
	}
}



func TestBuildSurgicalContext_TooManyEntryFiles(t *testing.T) {
	dir := t.TempDir()

	pw := NewParserWrapper()
	registry := NewResolverRegistry()
	engine := NewEngine(pw, registry)

	var entries []string
	for i := 0; i < 10; i++ {
		entries = append(entries, filepath.Join(dir, "main.go"))
	}

	_, err := engine.BuildSurgicalContext(context.Background(), entries, dir, 8000)
	if err == nil {
		t.Fatal("expected error for too many entry files")
	}
}

func TestBuildSurgicalContext_NonExistentFile(t *testing.T) {
	dir := t.TempDir()

	pw := NewParserWrapper()
	registry := NewResolverRegistry()
	engine := NewEngine(pw, registry)

	entry := filepath.Join(dir, "nonexistent.go")
	graph, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes for nonexistent file, got %d", len(graph.Nodes))
	}
}

func TestContextPacket(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)

	pw := NewParserWrapper()
	registry := NewResolverRegistry()
	engine := NewEngine(pw, registry)

	entry := filepath.Join(dir, "main.go")
	graph, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packet, err := graph.ContextPacket(4000)
	if err != nil {
		t.Fatalf("ContextPacket error: %v", err)
	}

	if len(packet.Files) == 0 {
		t.Fatal("expected at least 1 file in context packet")
	}
	if packet.TotalTokens <= 0 {
		t.Fatalf("expected positive TotalTokens, got %d", packet.TotalTokens)
	}
	if len(packet.Edges) == 0 {
		t.Fatal("expected edges in context packet")
	}
}

func TestBuildSurgicalContext_WithVendorDir(t *testing.T) {
	dir := t.TempDir()

	// Simulate project with go.mod
	writeTestFile(t, dir, "go.mod", `module github.com/test/demo

go 1.26
`)

	writeTestFile(t, dir, "main.go", `package main

import "github.com/test/demo/internal/auth"

func main() {
	auth.Authenticate()
}
`)

	os.MkdirAll(filepath.Join(dir, "internal", "auth"), 0755)
	writeTestFile(t, dir, "internal/auth/auth.go", `package auth

func Authenticate() bool {
	return true
}
`)

	pw := NewParserWrapper()
	registry := NewResolverRegistry()
	engine := NewEngine(pw, registry)

	registry.Register(&resolvers.GoResolver{})

	entry := filepath.Join(dir, "main.go")
	graph, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	authNodeFound := false
	for _, node := range graph.Nodes {
		if node.Kind == NodeKindFunction && node.Name == "Authenticate" {
			authNodeFound = true
			break
		}
	}

	if !authNodeFound {
		t.Error("expected Authenticate function node resolved from internal import")
	}
}

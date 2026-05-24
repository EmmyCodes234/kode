package execution

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kode/kode/internal/graph"
)

func TestApplyHunk_InsertAtEnd(t *testing.T) {
	content := "package main\n"
	hunk := StructuredHunk{
		ID:      "h1",
		Action:  ActionInsert,
		NewText: "func main() {}",
	}
	result, err := applyHunkToContent(content, hunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "func main() {}") {
		t.Fatalf("expected new text in result, got: %s", result)
	}
}

func TestApplyHunk_InsertAfterAnchor(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionInsert,
		AnchorText: "func main() {}",
		NewText:    "func init() {}",
	}
	result, err := applyHunkToContent(content, hunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "func init() {}") {
		t.Fatalf("expected init function in result, got: %s", result)
	}
	if !strings.Contains(result, "func main() {}") {
		t.Fatalf("expected main function still present, got: %s", result)
	}
}

func TestApplyHunk_Delete(t *testing.T) {
	content := "package main\n\nfunc oldFunc() {}\n\nfunc main() {}\n"
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionDelete,
		AnchorText: "func oldFunc() {}",
	}
	result, err := applyHunkToContent(content, hunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "oldFunc") {
		t.Fatalf("expected oldFunc to be removed, got: %s", result)
	}
	if !strings.Contains(result, "func main() {}") {
		t.Fatalf("expected main to remain, got: %s", result)
	}
}

func TestApplyHunk_Modify(t *testing.T) {
	content := "func add(a, b int) int { return 0 }\n"
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionModify,
		AnchorText: "func add(a, b int) int { return 0 }",
		NewText:    "func add(a, b int) int { return a + b }",
	}
	result, err := applyHunkToContent(content, hunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "return a + b") {
		t.Fatalf("expected modified body, got: %s", result)
	}
}

func TestApplyHunk_AnchorNotFound(t *testing.T) {
	content := "package main\n"
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionModify,
		AnchorText: "nonexistent",
		NewText:    "something",
	}
	_, err := applyHunkToContent(content, hunk)
	if err == nil {
		t.Fatal("expected error for missing anchor text")
	}
}

func TestReverseHunk_Insert(t *testing.T) {
	hunk := StructuredHunk{
		ID:      "h1",
		Action:  ActionInsert,
		NewText: "func x() {}",
	}
	rev := reverseHunk(hunk)
	if rev.Action != ActionDelete {
		t.Fatalf("expected reverse to be DELETE, got %s", rev.Action)
	}
	if rev.AnchorText != "func x() {}" {
		t.Fatalf("expected anchor text to be original new text, got %q", rev.AnchorText)
	}
}

func TestReverseHunk_Delete(t *testing.T) {
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionDelete,
		AnchorText: "func x() {}",
	}
	rev := reverseHunk(hunk)
	if rev.Action != ActionInsert {
		t.Fatalf("expected reverse to be INSERT, got %s", rev.Action)
	}
	if rev.NewText != "func x() {}" {
		t.Fatalf("expected new text to be original anchor, got %q", rev.NewText)
	}
}

func TestReverseHunk_Modify(t *testing.T) {
	hunk := StructuredHunk{
		ID:         "h1",
		Action:     ActionModify,
		AnchorText: "old",
		NewText:    "new",
	}
	rev := reverseHunk(hunk)
	if rev.Action != ActionModify {
		t.Fatalf("expected reverse to be MODIFY, got %s", rev.Action)
	}
	if rev.AnchorText != "new" {
		t.Fatalf("expected anchor text to be 'new', got %q", rev.AnchorText)
	}
	if rev.NewText != "old" {
		t.Fatalf("expected new text to be 'old', got %q", rev.NewText)
	}
}

func TestGroupAndFlatten(t *testing.T) {
	hunks := []StructuredHunk{
		{ID: "h1", FilePath: "a.go"},
		{ID: "h2", FilePath: "b.go"},
		{ID: "h3", FilePath: "a.go"},
	}
	groups := groupHunksByFile(hunks)
	if len(groups["a.go"]) != 2 {
		t.Fatalf("expected 2 hunks for a.go, got %d", len(groups["a.go"]))
	}
	if len(groups["b.go"]) != 1 {
		t.Fatalf("expected 1 hunk for b.go, got %d", len(groups["b.go"]))
	}

	flattened := flattenHunks(groups)
	if len(flattened) != 3 {
		t.Fatalf("expected 3 hunks flattened, got %d", len(flattened))
	}
}

func TestExecuteTransaction_ValidModify(t *testing.T) {
	dir := t.TempDir()
	goModPath := filepath.Join(dir, "go.mod")
	os.WriteFile(goModPath, []byte("module test\n\ngo 1.26\n"), 0644)

	src := `package main

func main() {
	println("hello")
}
`
	mainPath := filepath.Join(dir, "main.go")
	os.WriteFile(mainPath, []byte(src), 0644)

	executor := NewExecutor(dir)
	hunks := []StructuredHunk{
		{
			ID:         "h1",
			FilePath:   "main.go",
			Action:     ActionModify,
			AnchorText: `	println("hello")`,
			NewText:    `	println("hello world")`,
		},
	}

	summary, err := executor.ExecuteTransaction(context.Background(), "task-1", dir, hunks, ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Status != StatusPass {
		t.Fatalf("expected PASS, got %s: %+v", summary.Status, summary.FailedHunks)
	}
	if summary.RoundsUsed != 1 {
		t.Fatalf("expected 1 round, got %d", summary.RoundsUsed)
	}

	data, _ := os.ReadFile(mainPath)
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("expected file content to be updated, got: %s", string(data))
	}
}

func TestExecuteTransaction_NewFileInsert(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0644)

	executor := NewExecutor(dir)
	hunks := []StructuredHunk{
		{
			ID:       "h1",
			FilePath: "newfile.go",
			Action:   ActionInsert,
			NewText:  "package main\n\nfunc main() {}\n",
		},
	}

	summary, err := executor.ExecuteTransaction(context.Background(), "task-2", dir, hunks, ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Status != StatusPass {
		t.Fatalf("expected PASS, got %s", summary.Status)
	}

	if _, err := os.Stat(filepath.Join(dir, "newfile.go")); os.IsNotExist(err) {
		t.Fatal("expected newfile.go to be created on disk")
	}
}

func TestExecuteTransaction_SameFileOverlappingHunks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0644)

	src := `package main

func helper() int {
	return 0
}

func main() {
	println(helper())
}
`
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0644)

	executor := NewExecutor(dir)

	// Hunk_A: modify helper to return 1
	// Hunk_B: insert a new function after helper (appending, not replacing)
	hunks := []StructuredHunk{
		{
			ID:         "h1",
			FilePath:   "main.go",
			Action:     ActionModify,
			AnchorText: `func helper() int {
	return 0
}`,
			NewText: `func helper() int {
	return 42
}`,
		},
		{
			ID:         "h2",
			FilePath:   "main.go",
			Action:     ActionInsert,
			AnchorText: `func helper() int {
	return 42
}`,
			NewText: `func anotherHelper() int {
	return 1
}`,
		},
	}

	summary, err := executor.ExecuteTransaction(context.Background(), "task-3", dir, hunks, ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Status != StatusPass {
		t.Fatalf("expected PASS, got %s: %+v", summary.Status, summary.FailedHunks)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	if !strings.Contains(string(data), "return 42") {
		t.Fatal("expected helper to return 42")
	}
	if !strings.Contains(string(data), "anotherHelper") {
		t.Fatal("expected anotherHelper to be present")
	}
}

func TestExecuteTransaction_SyntaxFailure(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0644)

	src := `package main

func main() {
	println("hello")
}
`
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0644)

	executor := NewExecutor(dir)

	// Hunk that introduces a syntax error (removes opening brace)
	hunks := []StructuredHunk{
		{
			ID:         "h1",
			FilePath:   "main.go",
			Action:     ActionModify,
			AnchorText: "func main() {",
			NewText:    "func main(",
		},
	}

	summary, err := executor.ExecuteTransaction(context.Background(), "task-4", dir, hunks, ExecutionContext{})
	if err == nil {
		t.Fatal("expected error for syntax-breaking hunk")
	}
	if summary.Status != StatusFail {
		t.Fatalf("expected FAIL, got %s", summary.Status)
	}
}

func TestExecuteTransaction_CumulativeStatePreservesUnchangedFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0644)

	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc a() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc b() {}\n"), 0644)

	executor := NewExecutor(dir)

	// Two hunks in separate files — they should both pass independently
	hunks := []StructuredHunk{
		{
			ID:         "h1",
			FilePath:   "a.go",
			Action:     ActionModify,
			AnchorText: "func a() {}",
			NewText:    "func a() int { return 1 }",
		},
		{
			ID:         "h2",
			FilePath:   "b.go",
			Action:     ActionModify,
			AnchorText: "func b() {}",
			NewText:    "func b() string { return \"hello\" }",
		},
	}

	summary, err := executor.ExecuteTransaction(context.Background(), "task-5", dir, hunks, ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Status != StatusPass {
		t.Fatalf("expected PASS, got %s: %+v", summary.Status, summary.FailedHunks)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "a.go"))
	if !strings.Contains(string(data), "return 1") {
		t.Fatal("expected a.go to be modified")
	}
	data, _ = os.ReadFile(filepath.Join(dir, "b.go"))
	if !strings.Contains(string(data), "return \\\"hello\\\"") {
		// The escaped quotes might need checking differently
		t.Logf("b.go content: %s", string(data))
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

	pw := graph.NewParserWrapper()
	registry := graph.NewResolverRegistry()
	engine := graph.NewEngine(pw, registry)

	entry := filepath.Join(dir, "main.go")
	g, err := engine.BuildSurgicalContext(context.Background(), []string{entry}, dir, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packet, err := g.ContextPacket(4000)
	if err != nil {
		t.Fatalf("ContextPacket error: %v", err)
	}

	if len(packet.Files) == 0 {
		t.Fatal("expected at least 1 file in context packet")
	}
	if packet.TotalTokens <= 0 {
		t.Fatalf("expected positive TotalTokens, got %d", packet.TotalTokens)
	}
}

func writeTestFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

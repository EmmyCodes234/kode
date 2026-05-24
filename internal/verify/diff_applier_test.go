package verify

import (
	"strings"
	"testing"
)

func TestDiffApplier_AddLine(t *testing.T) {
	da := NewDiffApplier()

	original := map[string]string{
		"main.go": strings.Join([]string{
			`package main`,
			``,
			`import "fmt"`,
			``,
			`func main() {`,
			`	fmt.Println("hello")`,
			`}`,
		}, "\n"),
	}

	// Proper unified diff format: @@ header is a comment, context lines are prefixed with space
	diff := strings.Join([]string{
		`--- a/main.go`,
		`+++ b/main.go`,
		`@@ -5,3 +5,3 @@ func main() {`,
		` func main() {`,
		`-	fmt.Println("hello")`,
		`+	fmt.Println("hello world")`,
		` }`,
	}, "\n")

	result, err := da.ApplyInMemory(diff, original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	modified := result["main.go"]
	if !strings.Contains(modified, "hello world") {
		t.Fatalf("expected modified content to contain 'hello world', got:\n%s", modified)
	}
	if strings.Contains(modified, `fmt.Println("hello")`) {
		t.Fatalf("expected removed line to be gone, but found it:\n%s", modified)
	}
}

func TestDiffApplier_NewFile(t *testing.T) {
	da := NewDiffApplier()

	original := map[string]string{}

	diff := strings.Join([]string{
		`--- /dev/null`,
		`+++ b/newfile.go`,
		`@@ -0,0 +1,5 @@`,
		`+package main`,
		`+`,
		`+func main() {`,
		`+	println("new")`,
		`+}`,
	}, "\n")

	result, err := da.ApplyInMemory(diff, original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, exists := result["newfile.go"]
	if !exists {
		t.Fatal("expected newfile.go in result")
	}
	if !strings.Contains(content, "println") {
		t.Fatalf("expected content to contain 'println', got:\n%s", content)
	}
}

func TestDiffApplier_MultipleFiles(t *testing.T) {
	da := NewDiffApplier()

	original := map[string]string{
		"a.go": "package a\nfunc A() int { return 0 }\n",
		"b.go": "package b\nfunc B() int { return 0 }\n",
	}

	diff := strings.Join([]string{
		`--- a/a.go`,
		`+++ b/a.go`,
		`@@ -1,2 +1,3 @@`,
		` package a`,
		` func A() int { return 0 }`,
		`+func A2() int { return 1 }`,
		`--- a/b.go`,
		`+++ b/b.go`,
		`@@ -1,2 +1,3 @@`,
		` package b`,
		` func B() int { return 0 }`,
		`+func B2() int { return 2 }`,
	}, "\n")

	result, err := da.ApplyInMemory(diff, original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result["a.go"], "A2") {
		t.Error("expected A2 in a.go")
	}
	if !strings.Contains(result["b.go"], "B2") {
		t.Error("expected B2 in b.go")
	}
}

func TestDiffApplier_NoHeaders(t *testing.T) {
	da := NewDiffApplier()

	original := map[string]string{
		"main.go": "package main\n",
	}

	_, err := da.ApplyInMemory("some garbage content", original)
	if err == nil {
		t.Fatal("expected error for diff without headers")
	}
}

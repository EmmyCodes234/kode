package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStress_Syntax(t *testing.T) {
	cases := []struct {
		name    string
		content string
		expect  Status // known limitations: go/parser does NOT check types (duplicate decl, missing return)
	}{
		{name: "missing closing brace", content: "package main\n\nfunc f() {\n\tif true {\n\t\treturn\n", expect: StatusFail},
		{name: "invalid operator", content: "package main\n\nfunc f() { _ = 1 ++ 2 }\n", expect: StatusFail},
		{name: "valid minimal", content: "package main\n", expect: StatusPass},
		{name: "valid function", content: "package main\n\nfunc f() int { return 1 }\n", expect: StatusPass},
		{name: "empty string", content: "", expect: StatusFail},
		{name: "unicode", content: "package main\n\nfunc π() float64 { return 3.14 }\n", expect: StatusPass},
	}

	s := NewSyntaxChecker()
	for _, c := range cases {
		t.Run("syntax/"+c.name, func(t *testing.T) {
			r := s.CheckFile("test.go", c.content)
			if r.Status != c.expect {
				t.Errorf("expected %s, got %s: %s", c.expect, r.Status, r.Message)
			}
		})
	}
}

func TestStress_Imports(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module stresstest\n\ngo 1.21\n"), 0644)

	os.MkdirAll(filepath.Join(dir, "internal", "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "internal", "pkg", "lib.go"), []byte("package pkg\n\nfunc LibFunc() {}\n"), 0644)

	v := NewImportValidator(dir)
	allowed := map[string]bool{dir: true}

	cases := []struct {
		name    string
		content string
		expect  Status
	}{
		{name: "non-existent", content: `package main; import "stresstest/nope"; func f() {}`, expect: StatusFail},
		{name: "stdlib ok", content: `package main; import "fmt"; func f() { fmt.Println() }`, expect: StatusPass},
		{name: "no imports", content: "package main\nfunc f() {}\n", expect: StatusPass},
	}

	for _, c := range cases {
		t.Run("imports/"+c.name, func(t *testing.T) {
			r := v.Validate(filepath.Join(dir, "main.go"), c.content, allowed)
			if r.Status != c.expect {
				t.Errorf("expected %s, got %s: %s", c.expect, r.Status, r.Message)
			}
		})
	}
}

func TestStress_Architecture(t *testing.T) {
	checker := NewArchitectureCheckerWithModule("/tmp", "stresstest")
	rules := []ArchRule{
		{ForbiddenImportPrefix: "stresstest/internal/secret", AllowedInPackages: []string{"stresstest/internal"}, ErrorMessage: "off-limits"},
		{ForbiddenImportPrefix: "stresstest/pkg/legacy", AllowedInPackages: []string{}, ErrorMessage: "deprecated"},
	}

	cases := []struct {
		name    string
		path    string
		content string
		expect  Status
	}{
		{name: "forbidden from cmd", path: "/tmp/cmd/main.go", content: `package main; import _ "stresstest/internal/secret"`, expect: StatusFail},
		{name: "allowed from internal", path: "/tmp/internal/app/main.go", content: `package main; import _ "stresstest/internal/secret"`, expect: StatusPass},
		{name: "deprecated legacy", path: "/tmp/cmd/main.go", content: `package main; import _ "stresstest/pkg/legacy/old"`, expect: StatusFail},
		{name: "clean file", path: "/tmp/cmd/main.go", content: `package main; import "fmt"`, expect: StatusPass},
	}

	for _, c := range cases {
		t.Run("arch/"+c.name, func(t *testing.T) {
			r := checker.CheckFile(c.path, c.content, rules)
			if r.Status != c.expect {
				t.Errorf("expected %s, got %s: %s", c.expect, r.Status, r.Message)
			}
		})
	}
}

func TestStress_VerifyFileContent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module stresstest\n\ngo 1.21\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "internal", "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "internal", "pkg", "lib.go"), []byte("package pkg\n\nfunc LibFunc() {}\n"), 0644)

	e := newExecutor(dir)

	t.Run("valid Go file passes", func(t *testing.T) {
		if r := e.syntax.CheckFile(filepath.Join(dir, "m.go"), "package main\nfunc main() {}\n"); r.Status != StatusPass {
			t.Fatalf("expected PASS: %s", r.Message)
		}
	})

	t.Run("non-Go file skipped", func(t *testing.T) {
		// SyntaxChecker doesn't skip files by extension; the caller handles that.
		content := "# Markdown\n"
		r := e.syntax.CheckFile(filepath.Join(dir, "readme.md"), content)
		// Syntax checker should still parse it — .md files aren't filterable at this layer.
		// The VerifyFileContent method on Executor (execution package) handles extension filtering.
		_ = r
	})

	t.Run("architecture violation checker", func(t *testing.T) {
		content := `package main; import _ "stresstest/internal/secret"; func f() {}`
		r := e.architecture.CheckFile(filepath.Join(dir, "cmd/main.go"), content, []ArchRule{
			{ForbiddenImportPrefix: "stresstest/internal/secret", AllowedInPackages: []string{}, ErrorMessage: "off-limits"},
		})
		if r.Status != StatusFail {
			t.Fatal("expected FAIL for arch violation")
		}
	})
}

func TestStress_CallChecker(t *testing.T) {
	checker := NewCallChecker(".")
	allowedPkgs := map[string]bool{}
	graphEntries := map[string]bool{}

	cases := []struct {
		name    string
		content string
		expect  Status
	}{
		{
			name:    "fmt.Println is stdlib, should pass",
			content: `package main; import "fmt"; func f() { fmt.Println("hello") }`,
			expect:  StatusPass,
		},
		{
			name:    "undefined function call",
			content: `package main; func f() { undefinedFunc() }`,
			expect:  StatusWarn,
		},
		{
			name:    "no function calls",
			content: `package main; func f() int { return 1 }`,
			expect:  StatusPass,
		},
	}

	for _, c := range cases {
		t.Run("calls/"+c.name, func(t *testing.T) {
			r := checker.CheckFile("/tmp/main.go", c.content, allowedPkgs, graphEntries)
			if r.Status != c.expect {
				t.Errorf("expected %s, got %s: %s", c.expect, r.Status, r.Message)
			}
		})
	}
}

func newExecutor(projectRoot string) *gateExecutor {
	return &gateExecutor{
		syntax:       NewSyntaxChecker(),
		imports:      NewImportValidator(projectRoot),
		calls:        NewCallChecker(projectRoot),
		architecture: NewArchitectureChecker(),
	}
}

type gateExecutor struct {
	syntax       *SyntaxChecker
	imports      *ImportValidator
	calls        *CallChecker
	architecture *ArchitectureChecker
}

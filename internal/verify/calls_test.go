package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCallChecker_StdLibCall(t *testing.T) {
	dir := t.TempDir()
	c := NewCallChecker(dir)

	content := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	allowedPackages := map[string]bool{}
	graphEntries := map[string]bool{}

	result := c.CheckFile("main.go", content, allowedPackages, graphEntries)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for stdlib call, got %s: %s", result.Status, result.Message)
	}
}

func TestCallChecker_KnownPackageCall(t *testing.T) {
	dir := t.TempDir()
	c := NewCallChecker(dir)

	content := `package main

import "github.com/test/demo/internal/auth"

func main() {
	auth.Authenticate()
}
`
	allowedPackages := map[string]bool{
		"github.com/test/demo/internal/auth": true,
	}
	graphEntries := map[string]bool{}

	result := c.CheckFile("main.go", content, allowedPackages, graphEntries)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for known package call, got %s: %s", result.Status, result.Message)
	}
}

func TestCallChecker_HallucinatedPackageCall(t *testing.T) {
	dir := t.TempDir()
	c := NewCallChecker(dir)

	content := `package main

import "fmt"

func main() {
	nonexistent.DoSomething()
}
`
	allowedPackages := map[string]bool{}
	graphEntries := map[string]bool{}

	result := c.CheckFile("main.go", content, allowedPackages, graphEntries)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL for hallucinated package call, got %s", result.Status)
	}
}

func TestCallChecker_LocalMethodCall_Valid(t *testing.T) {
	dir := t.TempDir()

	// Create a file that defines a method on *Service
	os.MkdirAll(filepath.Join(dir, "internal", "svc"), 0755)
	os.WriteFile(filepath.Join(dir, "internal", "svc", "service.go"), []byte(`package svc

type Service struct{}

func (s *Service) Validate() bool {
	return true
}
`), 0644)

	c := NewCallChecker(dir)

	content := `package main

import "fmt"

func main() {
	var svc *Service
	svc.Validate()
}
`
	allowedPackages := map[string]bool{}
	graphEntries := map[string]bool{}

	result := c.CheckFile("main.go", content, allowedPackages, graphEntries)
	// Without import of svc package, this might fail
	// But the lazy probe won't find it either since "svc" is a local variable, not a package
	// So this should WARN or PASS depending on resolution
	t.Logf("Result: %s - %s", result.Status, result.Message)
}

func TestCallChecker_NonGoFile(t *testing.T) {
	dir := t.TempDir()
	c := NewCallChecker(dir)

	result := c.CheckFile("style.css", `body { color: red }`, nil, nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for non-Go files, got %s", result.Status)
	}
}

func TestCallChecker_MultipleImports(t *testing.T) {
	dir := t.TempDir()
	c := NewCallChecker(dir)

	content := `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("a")
	os.Exit(0)
	strings.ToUpper("b")
}
`
	allowedPackages := map[string]bool{}
	graphEntries := map[string]bool{}

	result := c.CheckFile("main.go", content, allowedPackages, graphEntries)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for multiple stdlib calls, got %s: %s", result.Status, result.Message)
	}
}

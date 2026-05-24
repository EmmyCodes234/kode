package verify

import (
	"testing"
)

func TestSyntaxChecker_ValidGo(t *testing.T) {
	c := NewSyntaxChecker()
	content := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	result := c.CheckFile("main.go", content)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestSyntaxChecker_InvalidGo(t *testing.T) {
	c := NewSyntaxChecker()
	content := `package main

func main() {
	fmt.Println("hello"
}
`
	result := c.CheckFile("main.go", content)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
}

func TestSyntaxChecker_NonGoFile(t *testing.T) {
	c := NewSyntaxChecker()
	result := c.CheckFile("readme.md", "# hello")
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for non-Go files, got %s", result.Status)
	}
}

func TestSyntaxChecker_EmptyGo(t *testing.T) {
	c := NewSyntaxChecker()
	result := c.CheckFile("empty.go", ``)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL for empty file, got %s", result.Status)
	}
}

package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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

func TestParseFile_GoSyntax(t *testing.T) {
	dir := t.TempDir()
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	path := writeTestFile(t, dir, "main.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasError {
		t.Fatal("expected no parse errors")
	}
	if len(result.Defs) != 1 {
		t.Fatalf("expected 1 def, got %d", len(result.Defs))
	}
	if result.Defs[0].Name != "main" {
		t.Fatalf("expected def name 'main', got %s", result.Defs[0].Name)
	}
	if len(result.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(result.Imports))
	}
	if result.Imports[0].Path != "fmt" {
		t.Fatalf("expected import 'fmt', got %s", result.Imports[0].Path)
	}
}

func TestParseFile_GoSyntaxError(t *testing.T) {
	dir := t.TempDir()
	src := `package main

func main() {
	fmt.Println("hello"
}
`
	path := writeTestFile(t, dir, "broken.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasError {
		t.Fatal("expected HasError=true for broken syntax")
	}
}

func TestParseFile_GoMethod(t *testing.T) {
	dir := t.TempDir()
	src := `package service

type TokenService struct{}

func (s *TokenService) Validate(token string) bool {
	return token != ""
}
`
	path := writeTestFile(t, dir, "service.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, d := range result.Defs {
		if d.Kind == "method" && d.Name == "(*TokenService).Validate" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected method '(*TokenService).Validate', got defs: %+v", result.Defs)
	}
}

func TestParseFile_GoStructAndInterface(t *testing.T) {
	dir := t.TempDir()
	src := `package store

type User struct {
	ID   int
	Name string
}

type Validator interface {
	Validate() error
}
`
	path := writeTestFile(t, dir, "types.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasStruct := false
	hasInterface := false
	for _, d := range result.Defs {
		if d.Kind == "struct" && d.Name == "User" {
			hasStruct = true
		}
		if d.Kind == "interface" && d.Name == "Validator" {
			hasInterface = true
		}
	}
	if !hasStruct {
		t.Fatal("expected struct 'User'")
	}
	if !hasInterface {
		t.Fatal("expected interface 'Validator'")
	}
}

func TestDetectLanguage(t *testing.T) {
	pw := NewParserWrapper()
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "go"},
		{"server.ts", "typescript"},
		{"component.tsx", "typescript"},
		{"app.js", "javascript"},
		{"util.py", "python"},
		{"lib.rs", "rust"},
		{"README.md", ""},
		{"dockerfile", ""},
	}
	for _, tt := range tests {
		got := pw.DetectLanguage(tt.path)
		if got != tt.expected {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestParseFile_ImportsMultiple(t *testing.T) {
	dir := t.TempDir()
	src := `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {}
`
	path := writeTestFile(t, dir, "main.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(result.Imports))
	}
}

func TestParseFile_AliasedImport(t *testing.T) {
	dir := t.TempDir()
	src := `package main

import jwt "github.com/golang-jwt/jwt"

func main() {}
`
	path := writeTestFile(t, dir, "main.go", src)
	pw := NewParserWrapper()

	result, err := pw.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(result.Imports))
	}
	if result.Imports[0].Alias != "jwt" {
		t.Fatalf("expected alias 'jwt', got %q", result.Imports[0].Alias)
	}
}

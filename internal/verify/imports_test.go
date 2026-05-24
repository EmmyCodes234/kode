package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportValidator_StdLib(t *testing.T) {
	dir := t.TempDir()
	v := NewImportValidator(dir)

	content := `package main

import "fmt"
`
	result := v.Validate("main.go", content, nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for stdlib import, got %s: %s", result.Status, result.Message)
	}
}

func TestImportValidator_InternalPackage(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "internal", "auth"), 0755)

	v := NewImportValidator(dir)

	content := `package main

import "github.com/test/demo/internal/auth"
`
	result := v.Validate("main.go", content, nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for valid internal import, got %s: %s", result.Status, result.Message)
	}
}

func TestImportValidator_RogueImport(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)

	v := NewImportValidator(dir)

	content := `package main

import "github.com/hallucinated/pkg"
`
	result := v.Validate("main.go", content, nil)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL for hallucinated import, got %s", result.Status)
	}
}

func TestImportValidator_NonGoFile(t *testing.T) {
	dir := t.TempDir()
	v := NewImportValidator(dir)

	result := v.Validate("style.css", "body { color: red }", nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for non-Go files, got %s", result.Status)
	}
}

func TestImportValidator_AllowedInternalMap(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/demo\n\ngo 1.26\n"), 0644)

	v := NewImportValidator(dir)

	content := `package main

import "github.com/test/demo/internal/known"
`
	allowed := map[string]bool{"internal/known": true}
	result := v.Validate("main.go", content, allowed)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for allowed internal import, got %s: %s", result.Status, result.Message)
	}
}

func TestImportValidator_ExternalDep(t *testing.T) {
	dir := t.TempDir()

	goMod := `module github.com/test/demo

go 1.26

require (
	github.com/gorilla/mux v1.8.0
)
`
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)

	v := NewImportValidator(dir)

	content := `package main

import "github.com/gorilla/mux"
`
	result := v.Validate("main.go", content, nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for known external dep, got %s: %s", result.Status, result.Message)
	}
}

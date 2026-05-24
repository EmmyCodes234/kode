package verify

import (
	"testing"
)

func TestArchitectureChecker_NoRules(t *testing.T) {
	c := NewArchitectureChecker()
	content := `package main
import "fmt"
`
	result := c.CheckFile("main.go", content, nil)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS when no rules defined, got %s", result.Status)
	}
}

func TestArchitectureChecker_NoViolation(t *testing.T) {
	c := NewArchitectureChecker()
	content := `package verify
`
	// Rule says no importing "internal/core" except from "verify"
	rules := []ArchRule{
		{
			ForbiddenImportPrefix: "github.com/kode/kode/internal/core",
			AllowedInPackages:     []string{"verify"},
			ErrorMessage:          "verify package cannot import internal/core",
		},
	}
	result := c.CheckFile("internal/verify/types.go", content, rules)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestArchitectureChecker_Violation(t *testing.T) {
	c := NewArchitectureChecker()
	content := `package graph

import "github.com/kode/kode/internal/verify"
`
	rules := []ArchRule{
		{
			ForbiddenImportPrefix: "github.com/kode/kode/internal/verify",
			AllowedInPackages:     []string{"verify"},
			ErrorMessage:          "graph package cannot import verify",
		},
	}
	result := c.CheckFile("internal/graph/engine.go", content, rules)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL for architecture violation, got %s", result.Status)
	}
}

func TestArchitectureChecker_WildcardViolation(t *testing.T) {
	c := NewArchitectureChecker()
	content := `package handler

import "github.com/kode/kode/internal/db"
`
	rules := []ArchRule{
		{
			ForbiddenImportPrefix: "github.com/kode/kode/internal/",
			AllowedInPackages:     []string{"internal/db", "internal/service"},
			ErrorMessage:          "handler package cannot import internal/ directly",
		},
	}
	result := c.CheckFile("handler/route.go", content, rules)
	if result.Status != StatusFail {
		t.Fatalf("expected FAIL for wildcard prefix violation, got %s", result.Status)
	}
}

func TestArchitectureChecker_NonGoFile(t *testing.T) {
	c := NewArchitectureChecker()
	rules := []ArchRule{
		{ForbiddenImportPrefix: "test", ErrorMessage: "no"},
	}
	result := c.CheckFile("data.json", `{"key": "value"}`, rules)
	if result.Status != StatusPass {
		t.Fatalf("expected PASS for non-Go file, got %s", result.Status)
	}
}

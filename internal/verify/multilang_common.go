package verify

import (
	"strings"
)

type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangJavaScript Language = "javascript"
	LangPython     Language = "python"
	LangRust       Language = "rust"
	LangUnknown    Language = "unknown"
)

// DetectLanguage returns the language for a file path based on extension.
func DetectLanguage(path string) Language {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return LangGo
	case strings.HasSuffix(lower, ".ts"), strings.HasSuffix(lower, ".tsx"):
		return LangTypeScript
	case strings.HasSuffix(lower, ".js"), strings.HasSuffix(lower, ".jsx"),
		strings.HasSuffix(lower, ".mjs"), strings.HasSuffix(lower, ".cjs"):
		return LangJavaScript
	case strings.HasSuffix(lower, ".py"), strings.HasSuffix(lower, ".pyi"):
		return LangPython
	case strings.HasSuffix(lower, ".rs"):
		return LangRust
	default:
		return LangUnknown
	}
}

// LanguageSupported returns true if the language has verification support.
func LanguageSupported(lang Language) bool {
	return lang != LangUnknown
}

// ParsedFile holds extracted structure from a source file.
type ParsedFile struct {
	Language  Language
	Imports   []ImportEntry
	Functions []string
	Classes   []string
	Calls     []CallEntry
}

// ImportEntry represents a single import from a source file.
type ImportEntry struct {
	Path    string // "react", "./utils", "github.com/foo/bar"
	Names   []string // imported identifiers (may be empty for namespace imports)
	IsLocal bool   // true if relative path (./  ../)
}

// CallEntry represents a function/method call.
type CallEntry struct {
	Name   string // "foo" or "pkg.Method"
	Line   int
}

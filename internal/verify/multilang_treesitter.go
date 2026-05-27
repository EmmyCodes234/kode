//go:build cgo

package verify

import (
	"context"
	_ "embed"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

//go:embed queries/typescript.scm
var tsQuery []byte

//go:embed queries/python.scm
var pyQuery []byte

//go:embed queries/rust.scm
var rsQuery []byte

// ParseFile extracts structure from a source file using tree-sitter.
func ParseFile(path, content string) *ParsedFile {
	lang := DetectLanguage(path)
	pf := &ParsedFile{Language: lang}

	var parser *sitter.Parser
	var queryBytes []byte
	var tsLang *sitter.Language

	parser = sitter.NewParser()

	switch lang {
	case LangTypeScript, LangJavaScript:
		tsLang = typescript.GetLanguage()
		parser.SetLanguage(tsLang)
		queryBytes = tsQuery
	case LangPython:
		tsLang = python.GetLanguage()
		parser.SetLanguage(tsLang)
		queryBytes = pyQuery
	case LangRust:
		tsLang = rust.GetLanguage()
		parser.SetLanguage(tsLang)
		queryBytes = rsQuery
	case LangGo, LangUnknown:
		// Not handled by tree-sitter here, could add pure go fallback
		return pf
	default:
		return pf
	}

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil {
		return pf
	}

	query, err := sitter.NewQuery(queryBytes, tsLang)
	if err != nil {
		return pf
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		
		for _, cap := range m.Captures {
			capName := query.CaptureNameForId(cap.Index)
			nodeText := cap.Node.Content([]byte(content))
			
			// Imports
			if capName == "import.source" || capName == "import.name" || capName == "import.path" || capName == "import.require" || capName == "import.dynamic" {
				// Very basic handling: if we capture a string/path, add it
				pathStr := stripQuotes(nodeText)
				if pathStr != "" {
					pf.Imports = append(pf.Imports, ImportEntry{
						Path:    pathStr,
						IsLocal: isLocalPath(pathStr),
					})
				}
			}

			// Functions
			if capName == "function.name" || capName == "method.name" {
				pf.Functions = append(pf.Functions, nodeText)
			}

			// Classes/Structs
			if capName == "class.name" || capName == "struct.name" {
				pf.Classes = append(pf.Classes, nodeText)
			}

			// Calls
			if capName == "call.name" || capName == "call.macro" {
				pf.Calls = append(pf.Calls, CallEntry{
					Name: nodeText,
					Line: int(cap.Node.StartPoint().Row) + 1,
				})
			}
		}
	}

	// Deduplicate imports
	pf.Imports = uniqueImports(pf.Imports)
	pf.Functions = uniqueStrings(pf.Functions)
	pf.Classes = uniqueStrings(pf.Classes)
	pf.Calls = uniqueCalls(pf.Calls)

	return pf
}

func stripQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

func isLocalPath(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] == '.' || s[0] == '/' || (len(s) >= 2 && s[:2] == "@/") || (len(s) >= 5 && s[:5] == "crate")
}

func uniqueImports(in []ImportEntry) []ImportEntry {
	seen := make(map[string]bool)
	var out []ImportEntry
	for _, v := range in {
		if !seen[v.Path] {
			seen[v.Path] = true
			out = append(out, v)
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func uniqueCalls(in []CallEntry) []CallEntry {
	seen := make(map[string]bool)
	var out []CallEntry
	for _, v := range in {
		if !seen[v.Name] {
			seen[v.Name] = true
			out = append(out, v)
		}
	}
	return out
}

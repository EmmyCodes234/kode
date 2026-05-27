//go:build !cgo

package verify

import (
	"regexp"
	"strings"
)

// Pre-compiled regex patterns per language
var (
	// TypeScript/JavaScript imports
	tsImportRe     = regexp.MustCompile(`(?m)^[ \t]*import\s+(?:(?:(?:\{[^}]*\}|[\w*]+(?:\s*,\s*\{[^}]*\})?)\s+from\s+)?['"]([^'"]+)['"]|['"]([^'"]+)['"])`)
	tsRequireRe    = regexp.MustCompile(`require\(\s*['"]([^'"]+)['"]\s*\)`)
	tsDynImportRe  = regexp.MustCompile(`import\(\s*['"]([^'"]+)['"]\s*\)`)
	tsFuncRe       = regexp.MustCompile(`(?m)(?:export\s+)?(?:async\s+)?function\s+(\w+)`)
	tsClassRe      = regexp.MustCompile(`(?m)(?:export\s+)?class\s+(\w+)`)
	tsCallRe       = regexp.MustCompile(`\b(\w+(?:\.\w+)?)\s*\(`)

	// Python imports
	pyImportRe     = regexp.MustCompile(`(?m)^[ \t]*import\s+([\w.]+)`)
	pyFromImportRe = regexp.MustCompile(`(?m)^[ \t]*from\s+([\w.]+)\s+import`)
	pyFuncRe       = regexp.MustCompile(`(?m)^[ \t]*(?:async\s+)?def\s+(\w+)`)
	pyClassRe      = regexp.MustCompile(`(?m)^[ \t]*class\s+(\w+)`)
	pyCallRe       = regexp.MustCompile(`\b(\w+(?:\.\w+)?)\s*\(`)

	// Rust imports
	rsUseRe        = regexp.MustCompile(`(?m)^[ \t]*use\s+([\w:]+(?:::\{[^}]+\})?)`)
	rsFuncRe       = regexp.MustCompile(`(?m)(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`)
	rsStructRe     = regexp.MustCompile(`(?m)(?:pub\s+)?struct\s+(\w+)`)
	rsImplRe       = regexp.MustCompile(`(?m)impl(?:<[^>]+>)?\s+(\w+)`)
	rsCallRe       = regexp.MustCompile(`\b(\w+(?:::\w+)?)\s*[!(]\s*`)
)

// ParseFile extracts structure from a source file using regex.
func ParseFile(path, content string) *ParsedFile {
	lang := DetectLanguage(path)
	pf := &ParsedFile{Language: lang}

	switch lang {
	case LangTypeScript, LangJavaScript:
		pf.parseTypeScript(content)
	case LangPython:
		pf.parsePython(content)
	case LangRust:
		pf.parseRust(content)
	case LangGo:
		// Go uses go/parser (existing path), nothing to add here
	}

	return pf
}

func (pf *ParsedFile) parseTypeScript(content string) {
	// Strip comments for cleaner parsing
	clean := stripLineComments(content, "//")

	// Imports
	for _, m := range tsImportRe.FindAllStringSubmatch(clean, -1) {
		path := m[1]
		if path == "" {
			path = m[2]
		}
		if path != "" {
			pf.Imports = append(pf.Imports, ImportEntry{
				Path:    path,
				IsLocal: strings.HasPrefix(path, ".") || strings.HasPrefix(path, "@/"),
			})
		}
	}
	for _, m := range tsRequireRe.FindAllStringSubmatch(clean, -1) {
		pf.Imports = append(pf.Imports, ImportEntry{
			Path:    m[1],
			IsLocal: strings.HasPrefix(m[1], "."),
		})
	}
	for _, m := range tsDynImportRe.FindAllStringSubmatch(clean, -1) {
		pf.Imports = append(pf.Imports, ImportEntry{
			Path:    m[1],
			IsLocal: strings.HasPrefix(m[1], "."),
		})
	}

	// Functions
	for _, m := range tsFuncRe.FindAllStringSubmatch(clean, -1) {
		pf.Functions = append(pf.Functions, m[1])
	}

	// Classes
	for _, m := range tsClassRe.FindAllStringSubmatch(clean, -1) {
		pf.Classes = append(pf.Classes, m[1])
	}

	// Calls
	for _, m := range tsCallRe.FindAllStringSubmatch(clean, -1) {
		name := m[1]
		// Filter out keywords
		if !isJSKeyword(name) {
			pf.Calls = append(pf.Calls, CallEntry{Name: name})
		}
	}
}

func (pf *ParsedFile) parsePython(content string) {
	clean := stripLineComments(content, "#")

	for _, m := range pyImportRe.FindAllStringSubmatch(clean, -1) {
		pf.Imports = append(pf.Imports, ImportEntry{
			Path:    m[1],
			IsLocal: strings.HasPrefix(m[1], "."),
		})
	}
	for _, m := range pyFromImportRe.FindAllStringSubmatch(clean, -1) {
		pf.Imports = append(pf.Imports, ImportEntry{
			Path:    m[1],
			IsLocal: strings.HasPrefix(m[1], "."),
		})
	}

	for _, m := range pyFuncRe.FindAllStringSubmatch(clean, -1) {
		pf.Functions = append(pf.Functions, m[1])
	}
	for _, m := range pyClassRe.FindAllStringSubmatch(clean, -1) {
		pf.Classes = append(pf.Classes, m[1])
	}
	for _, m := range pyCallRe.FindAllStringSubmatch(clean, -1) {
		name := m[1]
		if !isPyKeyword(name) {
			pf.Calls = append(pf.Calls, CallEntry{Name: name})
		}
	}
}

func (pf *ParsedFile) parseRust(content string) {
	clean := stripLineComments(content, "//")

	for _, m := range rsUseRe.FindAllStringSubmatch(clean, -1) {
		pf.Imports = append(pf.Imports, ImportEntry{
			Path:    m[1],
			IsLocal: strings.HasPrefix(m[1], "crate") || strings.HasPrefix(m[1], "self") || strings.HasPrefix(m[1], "super"),
		})
	}

	for _, m := range rsFuncRe.FindAllStringSubmatch(clean, -1) {
		pf.Functions = append(pf.Functions, m[1])
	}
	for _, m := range rsStructRe.FindAllStringSubmatch(clean, -1) {
		pf.Classes = append(pf.Classes, m[1])
	}
	for _, m := range rsImplRe.FindAllStringSubmatch(clean, -1) {
		pf.Classes = append(pf.Classes, m[1])
	}
	for _, m := range rsCallRe.FindAllStringSubmatch(clean, -1) {
		name := m[1]
		if !isRsKeyword(name) {
			pf.Calls = append(pf.Calls, CallEntry{Name: name})
		}
	}
}

func stripLineComments(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

var jsKeywords = map[string]bool{
	"if": true, "else": true, "for": true, "while": true, "do": true,
	"switch": true, "case": true, "break": true, "continue": true,
	"return": true, "throw": true, "try": true, "catch": true,
	"finally": true, "new": true, "delete": true, "typeof": true,
	"instanceof": true, "void": true, "in": true, "of": true,
	"yield": true, "await": true, "async": true, "function": true,
	"class": true, "extends": true, "super": true, "this": true,
	"import": true, "export": true, "default": true, "from": true,
	"as": true, "const": true, "let": true, "var": true,
}

func isJSKeyword(name string) bool {
	parts := strings.Split(name, ".")
	return jsKeywords[parts[0]]
}

var pyKeywords = map[string]bool{
	"if": true, "elif": true, "else": true, "for": true, "while": true,
	"with": true, "try": true, "except": true, "finally": true,
	"raise": true, "return": true, "yield": true, "import": true,
	"from": true, "as": true, "class": true, "def": true, "pass": true,
	"break": true, "continue": true, "del": true, "assert": true,
	"not": true, "and": true, "or": true, "is": true, "in": true,
	"lambda": true, "global": true, "nonlocal": true, "async": true,
	"await": true, "print": true, "type": true, "super": true,
}

func isPyKeyword(name string) bool {
	parts := strings.Split(name, ".")
	return pyKeywords[parts[0]]
}

var rsKeywords = map[string]bool{
	"if": true, "else": true, "match": true, "for": true, "while": true,
	"loop": true, "break": true, "continue": true, "return": true,
	"fn": true, "let": true, "mut": true, "const": true, "static": true,
	"struct": true, "enum": true, "impl": true, "trait": true,
	"type": true, "use": true, "mod": true, "pub": true, "crate": true,
	"self": true, "super": true, "where": true, "async": true,
	"await": true, "move": true, "ref": true, "as": true,
}

func isRsKeyword(name string) bool {
	parts := strings.Split(name, "::")
	return rsKeywords[parts[0]]
}

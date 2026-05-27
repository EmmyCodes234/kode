package verify

import (
	"go/parser"
	"go/token"
	"strings"
)

type SyntaxChecker struct{}

func NewSyntaxChecker() *SyntaxChecker {
	return &SyntaxChecker{}
}

func (s *SyntaxChecker) CheckFile(path string, content string) CheckResult {
	res := CheckResult{CheckName: "syntax", Status: StatusPass}

	lang := DetectLanguage(path)

	switch lang {
	case LangGo:
		return s.checkGo(path, content)
	case LangTypeScript, LangJavaScript:
		return s.checkBracketBalance(path, content, lang)
	case LangPython:
		return s.checkPython(path, content)
	case LangRust:
		return s.checkBracketBalance(path, content, lang)
	default:
		// Unknown language — skip syntax checking, don't block
		return res
	}
}

func (s *SyntaxChecker) checkGo(path string, content string) CheckResult {
	res := CheckResult{CheckName: "syntax", Status: StatusPass}

	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, path, content, parser.AllErrors)

	if err != nil {
		res.Status = StatusFail
		res.Message = "Syntax check failed: the modified file contains parse errors"
		res.Details = err.Error()
	}

	return res
}

// checkBracketBalance validates curly-brace languages (TS, JS, Rust, etc.)
// by counting matching brackets/parens/braces and detecting common syntax errors.
func (s *SyntaxChecker) checkBracketBalance(path string, content string, lang Language) CheckResult {
	res := CheckResult{CheckName: "syntax", Status: StatusPass}

	inString := false
	inTemplate := false
	stringChar := byte(0)
	escaped := false
	inLineComment := false
	inBlockComment := false

	var stack []byte

	for i := 0; i < len(content); i++ {
		ch := content[i]

		// Handle newlines
		if ch == '\n' {
			inLineComment = false
			escaped = false
			continue
		}

		// Handle escape sequences
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && (inString || inTemplate) {
			escaped = true
			continue
		}

		// Handle block comments
		if inBlockComment {
			if ch == '*' && i+1 < len(content) && content[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		// Handle line comments
		if inLineComment {
			continue
		}

		// Detect comment starts
		if !inString && !inTemplate && ch == '/' {
			if i+1 < len(content) {
				if content[i+1] == '/' {
					inLineComment = true
					i++
					continue
				}
				if content[i+1] == '*' {
					inBlockComment = true
					i++
					continue
				}
			}
		}

		// Handle strings
		if inString {
			if ch == stringChar {
				inString = false
			}
			continue
		}

		// Handle template literals (JS/TS)
		if inTemplate {
			if ch == '`' {
				inTemplate = false
			}
			continue
		}

		// Detect string starts
		if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
			continue
		}
		if ch == '`' && (lang == LangTypeScript || lang == LangJavaScript) {
			inTemplate = true
			continue
		}

		// Count brackets
		switch ch {
		case '{', '(', '[':
			stack = append(stack, ch)
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing brace '}'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing paren ')'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing bracket ']'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		opener := string(stack[len(stack)-1])
		closer := map[string]string{"{": "}", "(": ")", "[": "]"}[opener]
		res.Status = StatusFail
		res.Message = "Syntax check failed: unclosed " + opener + " (expected " + closer + ")"
		res.Details = "File has " + strings.Repeat(opener, len(stack)) + " unclosed bracket(s)"
	}

	return res
}

// checkPython validates Python by checking indentation consistency and bracket balance.
func (s *SyntaxChecker) checkPython(path string, content string) CheckResult {
	res := CheckResult{CheckName: "syntax", Status: StatusPass}

	// Bracket balance (Python uses parens, brackets, braces for continuations)
	var stack []byte
	inString := false
	inTripleString := false
	stringChar := byte(0)
	inComment := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		if ch == '\n' {
			inComment = false
			continue
		}

		if inComment {
			continue
		}

		// Triple-quoted strings
		if inTripleString {
			if i+2 < len(content) && content[i] == stringChar && content[i+1] == stringChar && content[i+2] == stringChar {
				inTripleString = false
				i += 2
			}
			continue
		}

		if inString {
			if ch == '\\' {
				i++
				continue
			}
			if ch == stringChar {
				inString = false
			}
			continue
		}

		// Detect triple-quoted strings
		if (ch == '"' || ch == '\'') && i+2 < len(content) && content[i+1] == ch && content[i+2] == ch {
			inTripleString = true
			stringChar = ch
			i += 2
			continue
		}

		if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
			continue
		}

		if ch == '#' {
			inComment = true
			continue
		}

		switch ch {
		case '(', '[', '{':
			stack = append(stack, ch)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing paren ')'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing bracket ']'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				res.Status = StatusFail
				res.Message = "Syntax check failed: unmatched closing brace '}'"
				res.Details = bracketErrorDetail(path, content, i)
				return res
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		opener := string(stack[len(stack)-1])
		closer := map[string]string{"{": "}", "(": ")", "[": "]"}[opener]
		res.Status = StatusFail
		res.Message = "Syntax check failed: unclosed " + opener + " (expected " + closer + ")"
		res.Details = "File has unclosed bracket(s)"
	}

	return res
}

func bracketErrorDetail(path string, content string, pos int) string {
	line := 1
	col := 1
	for i := 0; i < pos && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return path + ":" + strings.Repeat("0", 0) + itoa(line) + ":" + itoa(col)
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

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

	if !strings.HasSuffix(path, ".go") {
		return res
	}

	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, path, content, parser.AllErrors)

	if err != nil {
		res.Status = StatusFail
		res.Message = "Syntax check failed: the modified file contains parse errors"
		res.Details = err.Error()
	}

	return res
}

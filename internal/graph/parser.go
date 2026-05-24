package graph

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ParserWrapper struct{}

func NewParserWrapper() *ParserWrapper {
	return &ParserWrapper{}
}

func (pw *ParserWrapper) DetectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	default:
		return ""
	}
}

type ParseResult struct {
	Language string
	FilePath string
	Imports  []ImportInfo
	Defs     []DefInfo
	Calls    []CallInfo
	HasError bool
}

type ImportInfo struct {
	Path  string
	Alias string
}

type DefInfo struct {
	Kind     string // "function", "method", "struct", "interface"
	Name     string
	Receiver string // for methods
	StartPos int
	EndPos   int
}

type CallInfo struct {
	Type     string // "local", "package", "method"
	FuncName string
	Pkg      string
	StartPos int
}

func (pw *ParserWrapper) ParseFile(ctx context.Context, filePath string) (*ParseResult, error) {
	language := pw.DetectLanguage(filePath)
	if language == "" {
		return nil, fmt.Errorf("cannot detect language: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	switch language {
	case "go":
		return pw.parseGo(ctx, filePath, content)
	default:
		return nil, fmt.Errorf("language not yet supported by native parser: %s", language)
	}
}

func (pw *ParserWrapper) parseGo(ctx context.Context, filePath string, content []byte) (*ParseResult, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return &ParseResult{
			Language: "go",
			FilePath: filePath,
			HasError: true,
		}, nil
	}

	result := &ParseResult{
		Language: "go",
		FilePath: filePath,
	}

	for _, imp := range f.Imports {
		info := ImportInfo{
			Path: strings.Trim(imp.Path.Value, "\""),
		}
		if imp.Name != nil {
			info.Alias = imp.Name.Name
		}
		result.Imports = append(result.Imports, info)
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			info := DefInfo{
				Name:     d.Name.Name,
				StartPos: fset.Position(d.Pos()).Line,
				EndPos:   fset.Position(d.End()).Line,
			}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				info.Kind = "method"
				recvType := exprToString(d.Recv.List[0].Type)
				info.Receiver = recvType
				info.Name = fmt.Sprintf("(%s).%s", recvType, d.Name.Name)
			} else {
				info.Kind = "function"
			}
			result.Defs = append(result.Defs, info)

			ast.Inspect(d.Body, func(n ast.Node) bool {
				switch call := n.(type) {
				case *ast.CallExpr:
					info := CallInfo{
						StartPos: fset.Position(call.Pos()).Line,
					}
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						info.Type = "local"
						info.FuncName = fun.Name
					case *ast.SelectorExpr:
						if ident, ok := fun.X.(*ast.Ident); ok {
							info.Type = "package"
							info.Pkg = ident.Name
							info.FuncName = fun.Sel.Name
						}
					}
					result.Calls = append(result.Calls, info)
				}
				return true
			})

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						info := DefInfo{
							Name:     ts.Name.Name,
							StartPos: fset.Position(ts.Pos()).Line,
							EndPos:   fset.Position(ts.End()).Line,
						}
						switch ts.Type.(type) {
						case *ast.StructType:
							info.Kind = "struct"
						case *ast.InterfaceType:
							info.Kind = "interface"
						default:
							info.Kind = "type"
						}
						result.Defs = append(result.Defs, info)
					}
				}
			}
		}
	}

	return result, nil
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

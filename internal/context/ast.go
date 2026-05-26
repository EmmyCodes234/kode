package context

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func extractImports(path string, content string) []string {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	var imports []string
	for _, imp := range f.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, "\""))
	}
	return imports
}

func walkGoFiles(rootDir string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			skip := strings.HasPrefix(d.Name(), ".") || d.Name() == "node_modules" || d.Name() == "vendored"
			if skip && path != rootDir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		files[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	return files, err
}
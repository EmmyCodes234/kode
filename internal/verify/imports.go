package verify

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ImportValidator struct {
	projectRoot      string
	moduleName       string
	externalDeps     map[string]bool // external deps from go.mod
}

func NewImportValidator(projectRoot string) *ImportValidator {
	v := &ImportValidator{
		projectRoot:  projectRoot,
		externalDeps: make(map[string]bool),
	}
	v.loadGoMod()
	return v
}

func (v *ImportValidator) Validate(path string, content string, allowedInternal map[string]bool) CheckResult {
	res := CheckResult{CheckName: "imports", Status: StatusPass}

	lang := DetectLanguage(path)

	switch lang {
	case LangGo:
		return v.validateGo(path, content, allowedInternal)
	case LangTypeScript, LangJavaScript:
		return v.validateTSJS(path, content)
	case LangPython:
		return v.validatePython(path, content)
	case LangRust:
		// Rust imports are validated by cargo — skip for now
		return res
	default:
		return res
	}
}

func (v *ImportValidator) validateGo(path string, content string, allowedInternal map[string]bool) CheckResult {
	res := CheckResult{CheckName: "imports", Status: StatusPass}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
	if err != nil {
		res.Status = StatusFail
		res.Message = "Failed to parse imports in modified file"
		res.Details = err.Error()
		return res
	}

	var rogueImports []string
	for _, imp := range f.Imports {
		cleanImport := strings.Trim(imp.Path.Value, "\"")

		if v.isAllowed(cleanImport, allowedInternal) {
			continue
		}

		rogueImports = append(rogueImports, cleanImport)
	}

	if len(rogueImports) > 0 {
		res.Status = StatusFail
		res.Message = "Import validation failed: unresolvable or hallucinated dependencies detected"
		res.Details = "Unrecognized imports: " + strings.Join(rogueImports, ", ")
	}

	return res
}

func (v *ImportValidator) validateTSJS(path string, content string) CheckResult {
	res := CheckResult{CheckName: "imports", Status: StatusPass}

	parsed := ParseFile(path, content)
	var rogueImports []string

	for _, imp := range parsed.Imports {
		if imp.IsLocal {
			// Local imports (./foo, ../bar, @/utils) — check relative file existence
			// We skip this check since we can't reliably resolve TS path aliases without tsconfig
			continue
		}

		// External package imports — check if they could be valid npm packages
		pkgName := imp.Path
		// Scoped packages like @scope/name
		if strings.HasPrefix(pkgName, "@") {
			parts := strings.SplitN(pkgName, "/", 3)
			if len(parts) < 2 {
				rogueImports = append(rogueImports, pkgName)
				continue
			}
			// @scope/name is valid
			continue
		}

		// Bare specifiers like "react", "lodash/get" — these are valid npm conventions
		// Only flag clearly invalid imports (empty, starting with numbers, etc.)
		if pkgName == "" || (pkgName[0] >= '0' && pkgName[0] <= '9') {
			rogueImports = append(rogueImports, pkgName)
		}

		// Check if the package exists in node_modules on disk
		nodeModulesPath := filepath.Join(v.projectRoot, "node_modules", pkgName)
		if info, err := os.Stat(nodeModulesPath); err == nil && info.IsDir() {
			continue // Found in node_modules — valid
		}

		// Check common monorepo workspace patterns
		pkgJsonPath := filepath.Join(v.projectRoot, "node_modules", pkgName, "package.json")
		if _, err := os.Stat(pkgJsonPath); err == nil {
			continue // Found package.json — valid
		}

		// Node built-ins
		if isNodeBuiltin(pkgName) {
			continue
		}
	}

	if len(rogueImports) > 0 {
		res.Status = StatusWarn
		res.Message = "Import validation warning: potentially unresolvable dependencies"
		res.Details = "Unrecognized imports: " + strings.Join(rogueImports, ", ")
	}

	return res
}

func (v *ImportValidator) validatePython(path string, content string) CheckResult {
	res := CheckResult{CheckName: "imports", Status: StatusPass}

	parsed := ParseFile(path, content)
	var rogueImports []string

	for _, imp := range parsed.Imports {
		if imp.IsLocal {
			continue // Relative imports in Python are fine
		}

		// Check if it's a Python stdlib module
		topLevel := strings.Split(imp.Path, ".")[0]
		if isPythonStdlib(topLevel) {
			continue
		}

		// Check if the package exists as a directory in the project
		pkgDir := filepath.Join(v.projectRoot, topLevel)
		if info, err := os.Stat(pkgDir); err == nil && info.IsDir() {
			continue // Local package directory
		}

		// Check if there's a .py file for the import
		pkgFile := filepath.Join(v.projectRoot, topLevel+".py")
		if _, err := os.Stat(pkgFile); err == nil {
			continue // Local Python file
		}

		// Don't flag — we can't reliably validate all third-party Python packages
		// without a venv/requirements.txt parser. Just note it.
	}

	if len(rogueImports) > 0 {
		res.Status = StatusWarn
		res.Message = "Import validation warning: potentially unresolvable Python imports"
		res.Details = "Unrecognized imports: " + strings.Join(rogueImports, ", ")
	}

	return res
}

func (v *ImportValidator) isAllowed(importPath string, allowedInternal map[string]bool) bool {
	if v.isStdLib(importPath) {
		return true
	}

	if v.moduleName != "" && strings.HasPrefix(importPath, v.moduleName) {
		relative := strings.TrimPrefix(importPath, v.moduleName)
		relative = strings.TrimPrefix(relative, "/")

		// Check if this internal package exists on disk
		pkgDir := filepath.Join(v.projectRoot, relative)
		if info, err := os.Stat(pkgDir); err == nil && info.IsDir() {
			return true
		}

		if allowedInternal[relative] {
			return true
		}

		// Check prefixes: e.g., "internal/graph" matches "internal/graph/engine.go"
		for key := range allowedInternal {
			if key == "" {
				continue
			}
			if strings.HasPrefix(relative, key) || strings.HasPrefix(key, relative) {
				return true
			}
		}
		return false
	}

	if v.externalDeps[importPath] {
		return true
	}

	return false
}

func (v *ImportValidator) isStdLib(importPath string) bool {
	firstSegment := strings.Split(importPath, "/")[0]
	isStd := !strings.Contains(firstSegment, ".")

	// Module-local imports have no dot but aren't stdlib (e.g., "stresstest/foo")
	if isStd && v.moduleName != "" && strings.HasPrefix(importPath, v.moduleName) {
		return false
	}

	return isStd
}

func (v *ImportValidator) loadGoMod() {
	goModPath := filepath.Join(v.projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return
	}

	inRequireBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			v.moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			// Individual require line or inline require
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				depPath := parts[0]
				if strings.Contains(depPath, ".") || strings.Contains(depPath, "/") {
					v.externalDeps[depPath] = true
				}
			}
		}
		// Handle single-line require: require github.com/foo v1.0.0
		if strings.HasPrefix(line, "require ") && !strings.HasPrefix(line, "require (") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 1 {
				v.externalDeps[parts[0]] = true
			}
		}
	}
}

var nodeBuiltins = map[string]bool{
	"assert": true, "buffer": true, "child_process": true, "cluster": true,
	"console": true, "constants": true, "crypto": true, "dgram": true,
	"dns": true, "domain": true, "events": true, "fs": true, "http": true,
	"https": true, "module": true, "net": true, "os": true, "path": true,
	"perf_hooks": true, "process": true, "punycode": true, "querystring": true,
	"readline": true, "repl": true, "stream": true, "string_decoder": true,
	"sys": true, "timers": true, "tls": true, "tty": true, "url": true,
	"util": true, "v8": true, "vm": true, "wasi": true, "worker_threads": true,
	"zlib": true, "async_hooks": true, "diagnostics_channel": true,
	"inspector": true, "test": true, "trace_events": true,
}

func isNodeBuiltin(name string) bool {
	clean := strings.TrimPrefix(name, "node:")
	return nodeBuiltins[clean]
}

var pythonStdlib = map[string]bool{
	"abc": true, "aifc": true, "argparse": true, "array": true, "ast": true,
	"asyncio": true, "atexit": true, "base64": true, "bdb": true, "binascii": true,
	"binhex": true, "bisect": true, "builtins": true, "bz2": true, "calendar": true,
	"cgi": true, "cgitb": true, "chunk": true, "cmath": true, "cmd": true,
	"code": true, "codecs": true, "codeop": true, "collections": true,
	"colorsys": true, "compileall": true, "concurrent": true, "configparser": true,
	"contextlib": true, "contextvars": true, "copy": true, "copyreg": true,
	"cProfile": true, "crypt": true, "csv": true, "ctypes": true, "curses": true,
	"dataclasses": true, "datetime": true, "dbm": true, "decimal": true,
	"difflib": true, "dis": true, "distutils": true, "doctest": true, "email": true,
	"encodings": true, "enum": true, "errno": true, "faulthandler": true,
	"fcntl": true, "filecmp": true, "fileinput": true, "fnmatch": true,
	"fractions": true, "ftplib": true, "functools": true, "gc": true, "getopt": true,
	"getpass": true, "gettext": true, "glob": true, "grp": true, "gzip": true,
	"hashlib": true, "heapq": true, "hmac": true, "html": true, "http": true,
	"idlelib": true, "imaplib": true, "imghdr": true, "imp": true, "importlib": true,
	"inspect": true, "io": true, "ipaddress": true, "itertools": true, "json": true,
	"keyword": true, "lib2to3": true, "linecache": true, "locale": true,
	"logging": true, "lzma": true, "mailbox": true, "mailcap": true, "marshal": true,
	"math": true, "mimetypes": true, "mmap": true, "modulefinder": true,
	"multiprocessing": true, "netrc": true, "nis": true, "nntplib": true,
	"numbers": true, "operator": true, "optparse": true, "os": true, "ossaudiodev": true,
	"parser": true, "pathlib": true, "pdb": true, "pickle": true, "pickletools": true,
	"pipes": true, "pkgutil": true, "platform": true, "plistlib": true,
	"poplib": true, "posix": true, "posixpath": true, "pprint": true, "profile": true,
	"pstats": true, "pty": true, "pwd": true, "py_compile": true, "pyclbr": true,
	"pydoc": true, "queue": true, "quopri": true, "random": true, "re": true,
	"readline": true, "reprlib": true, "resource": true, "rlcompleter": true,
	"runpy": true, "sched": true, "secrets": true, "select": true, "selectors": true,
	"shelve": true, "shlex": true, "shutil": true, "signal": true, "site": true,
	"smtpd": true, "smtplib": true, "sndhdr": true, "socket": true,
	"socketserver": true, "sqlite3": true, "ssl": true, "stat": true,
	"statistics": true, "string": true, "stringprep": true, "struct": true,
	"subprocess": true, "sunau": true, "symtable": true, "sys": true,
	"sysconfig": true, "syslog": true, "tabnanny": true, "tarfile": true,
	"telnetlib": true, "tempfile": true, "termios": true, "test": true,
	"textwrap": true, "threading": true, "time": true, "timeit": true,
	"tkinter": true, "token": true, "tokenize": true, "tomllib": true,
	"trace": true, "traceback": true, "tracemalloc": true, "tty": true,
	"turtle": true, "turtledemo": true, "types": true, "typing": true,
	"unicodedata": true, "unittest": true, "urllib": true, "uu": true,
	"uuid": true, "venv": true, "warnings": true, "wave": true,
	"weakref": true, "webbrowser": true, "winreg": true, "winsound": true,
	"wsgiref": true, "xdrlib": true, "xml": true, "xmlrpc": true,
	"zipapp": true, "zipfile": true, "zipimport": true, "zlib": true,
	"_thread": true, "__future__": true,
}

func isPythonStdlib(name string) bool {
	return pythonStdlib[name]
}

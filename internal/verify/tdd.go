package verify

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var testFileSuffixes = []string{"_test.go", ".test.ts", ".test.tsx", ".test.js", ".spec.ts", ".spec.tsx", ".spec.js", "_test.py", "test_"}

type TDDEnforcer struct {
	projectRoot  string
	testCommand  string
	testPassed   bool
	testAttempted bool
}

func NewTDDEnforcer(projectRoot string) *TDDEnforcer {
	return &TDDEnforcer{
		projectRoot: projectRoot,
	}
}

func (e *TDDEnforcer) WithTestCommand(cmd string) *TDDEnforcer {
	e.testCommand = cmd
	return e
}

func IsTestFile(filePath string) bool {
	name := filepath.Base(filePath)
	for _, suffix := range testFileSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

type TDDResult struct {
	Blocked   bool
	Message   string
	HasTests  bool
	TestRan   bool
	TestPassed bool
}

func (e *TDDEnforcer) Check(modifiedFiles []string, testCmd string) *TDDResult {
	result := &TDDResult{}

	hasTestFiles := false
	hasProdFiles := false
	for _, f := range modifiedFiles {
		if IsTestFile(f) {
			hasTestFiles = true
		} else {
			hasProdFiles = true
		}
	}
	result.HasTests = hasTestFiles

	if !hasTestFiles && hasProdFiles {
		result.Blocked = true
		result.Message = "TDD mode: write or modify a test file before modifying production code"
		return result
	}

	if hasTestFiles {
		cmd := testCmd
		if cmd == "" {
			cmd = e.detectTestCommand()
		}

		e.testAttempted = true
		parts := strings.Fields(cmd)
		if len(parts) > 0 {
			c := exec.Command(parts[0], parts[1:]...)
			c.Dir = e.projectRoot
			err := c.Run()
			e.testPassed = err == nil
		}
		result.TestRan = true
		result.TestPassed = e.testPassed
	}

	if hasTestFiles && hasProdFiles && result.TestRan && result.TestPassed {
		result.Blocked = false
		result.Message = "TDD mode: tests pass, production code accepted"
		return result
	}

	if hasTestFiles && !hasProdFiles {
		result.Blocked = hasTestFiles && result.TestRan && result.TestPassed
		if result.TestRan && result.TestPassed {
			result.Message = "TDD mode: tests pass, you may now write production code"
		} else if result.TestRan {
			result.Message = "TDD mode: tests fail as expected, good. Production code unlocked."
			result.Blocked = false
		} else {
			result.Message = "TDD mode: test files detected, production code blocked until test run confirms failure"
			result.Blocked = true
		}
		return result
	}

	result.Blocked = true
	result.Message = "TDD mode: test-first enforcement active"
	return result
}

func (e *TDDEnforcer) detectTestCommand() string {
	detectCmds := []struct {
		pattern string
		cmd     string
	}{
		{"go.mod", "go test ./..."},
		{"package.json", "npm test"},
		{"Cargo.toml", "cargo test"},
		{"pyproject.toml", "pytest"},
		{"requirements.txt", "python -m pytest"},
	}

	var deepSearch func(dir string) (string, string)
	deepSearch = func(dir string) (string, string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", ""
		}
		for _, entry := range entries {
			for _, dc := range detectCmds {
				if entry.Name() == dc.pattern {
					return entry.Name(), dc.cmd
				}
			}
			if entry.IsDir() && entry.Name() != "node_modules" && entry.Name() != ".git" && entry.Name() != "vendor" {
				if f, c := deepSearch(filepath.Join(dir, entry.Name())); f != "" {
					return f, c
				}
			}
		}
		return "", ""
	}

	_, cmd := deepSearch(e.projectRoot)
	return cmd
}

func TDDSummary(result *TDDResult) string {
	if result == nil {
		return ""
	}
	parts := []string{"tdd: " + result.Message}
	if result.TestRan {
		if result.TestPassed {
			parts = append(parts, "tests passed")
		} else {
			parts = append(parts, "tests failed (expected)")
		}
	}
	return strings.Join(parts, "; ")
}

func (e *TDDEnforcer) TestPassed() bool {
	return e.testPassed
}

func (e *TDDEnforcer) TestAttempted() bool {
	return e.testAttempted
}

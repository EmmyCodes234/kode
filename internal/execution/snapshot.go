package execution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParseCommand(cmd string) []string {
	var parts []string
	current := ""
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current += string(c)
			}
		} else if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func NormalizeTestCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "go test ./..."
	}
	return cmd
}

type Snapshot map[string]string

func CreateSnapshot(projectRoot string, files []string) (Snapshot, error) {
	snap := make(Snapshot)
	for _, f := range files {
		absPath := filepath.Join(projectRoot, f)
		data, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				snap[f] = ""
				continue
			}
			return nil, fmt.Errorf("snapshot read %s: %w", f, err)
		}
		snap[f] = string(data)
	}
	return snap, nil
}

func (s Snapshot) Restore(projectRoot string) error {
	var lastErr error
	for filePath, content := range s {
		absPath := filepath.Join(projectRoot, filePath)
		if content == "" {
			if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
				lastErr = fmt.Errorf("restore remove %s: %w", filePath, err)
			}
		} else {
			if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
				lastErr = fmt.Errorf("restore write %s: %w", filePath, err)
			}
		}
	}
	return lastErr
}

func DetectTestCommand(projectRoot string) string {
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
		return "go test ./..."
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "package.json")); err == nil {
		return "npm test"
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "Cargo.toml")); err == nil {
		return "cargo test"
	}
	return "go test ./..."
}

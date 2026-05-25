package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SicarioFinding struct {
	RuleID    string  `json:"ruleId"`
	FilePath  string  `json:"filePath"`
	Line      int     `json:"line"`
	Column    int     `json:"column"`
	Snippet   string  `json:"snippet"`
	Severity  string  `json:"severity"`
	CWEID     *string `json:"cweId"`
	ScanType  string  `json:"scanType"`
	Suppressed bool   `json:"suppressed"`
}

type SecurityChecker struct {
	sicarioPath string
	enabled     bool
}

func NewSecurityChecker() *SecurityChecker {
	path, _ := findSicario()
	return &SecurityChecker{
		sicarioPath: path,
		enabled:     path != "",
	}
}

func (s *SecurityChecker) CheckFile(path string, content string) CheckResult {
	res := CheckResult{CheckName: "security", Status: StatusPass}

	if !s.enabled {
		res.Status = StatusWarn
		res.Message = "Security check skipped — Sicario not found (install with 'kode install sicario')"
		return res
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("kode-sicario-%d", os.Getpid()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		res.Status = StatusWarn
		res.Message = "Security check skipped — could not create temp dir"
		return res
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, filepath.Base(path))
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		res.Status = StatusWarn
		res.Message = "Security check skipped — could not write temp file"
		res.Details = err.Error()
		return res
	}

	cmd := exec.Command(s.sicarioPath, "scan", tmpFile, "--format", "json", "--quiet")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			res.Status = StatusWarn
			res.Message = "Security check skipped — Sicario scan failed"
			res.Details = string(exitErr.Stderr)
			return res
		}
		res.Status = StatusWarn
		res.Message = "Security check skipped — Sicario scan error"
		res.Details = err.Error()
		return res
	}

	var findings []SicarioFinding
	if err := json.Unmarshal(output, &findings); err != nil || len(findings) == 0 {
		return res
	}

	var highCrit []string
	var lowMed []string
	for _, f := range findings {
		if f.Suppressed {
			continue
		}
		sev := strings.ToLower(f.Severity)
		loc := fmt.Sprintf("%s:%d", f.RuleID, f.Line)
		cwe := ""
		if f.CWEID != nil {
			cwe = fmt.Sprintf(" [%s]", *f.CWEID)
		}
		entry := fmt.Sprintf("  %s %s%s — %s", f.Severity, loc, cwe, truncateSnippet(f.Snippet))
		if sev == "high" || sev == "critical" {
			highCrit = append(highCrit, entry)
		} else {
			lowMed = append(lowMed, entry)
		}
	}

	if len(highCrit) > 0 {
		res.Status = StatusFail
		res.Message = fmt.Sprintf("Blocked by %d security finding(s)", len(highCrit))
		res.Details = strings.Join(highCrit, "\n")
		return res
	}

	if len(lowMed) > 0 {
		res.Status = StatusWarn
		res.Message = fmt.Sprintf("%d low/medium security finding(s) found", len(lowMed))
		res.Details = strings.Join(lowMed, "\n")
		return res
	}

	return res
}

func (s *SecurityChecker) Available() bool {
	return s.enabled
}

func (s *SecurityChecker) BinaryPath() string {
	return s.sicarioPath
}

func findSicario() (string, error) {
	if p := os.Getenv("SICARIO_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	paths := []string{"sicario", "sicario.exe"}
	for _, name := range paths {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	kodeBin := filepath.Join(".kode", "bin")
	if kodeDir, err := os.Getwd(); err == nil {
		local := filepath.Join(kodeDir, kodeBin, "sicario")
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
		localExe := local + ".exe"
		if _, err := os.Stat(localExe); err == nil {
			return localExe, nil
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		local := filepath.Join(home, ".local", "bin", "sicario")
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
	}

	return "", fmt.Errorf("sicario not found")
}

func truncateSnippet(s string) string {
	cleaned := strings.ReplaceAll(s, "\n", "\\n")
	if len(cleaned) > 120 {
		return cleaned[:117] + "..."
	}
	return cleaned
}

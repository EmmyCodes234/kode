package daemon

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitWatcher struct {
	repoDir   string
	lastKnown string
}

func NewGitWatcher(repoDir string) *GitWatcher {
	return &GitWatcher{
		repoDir:   repoDir,
		lastKnown: "",
	}
}

func (w *GitWatcher) CurrentHEAD() (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (w *GitWatcher) HasNewCommits() (bool, int, error) {
	head, err := w.CurrentHEAD()
	if err != nil {
		return false, 0, err
	}

	if w.lastKnown == "" {
		w.lastKnown = head
		return false, 0, nil
	}

	if head == w.lastKnown {
		return false, 0, nil
	}

	// Count commits since last known
	out, err := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..HEAD", w.lastKnown)).Output()
	if err != nil {
		// HEAD may have moved non-linearly; reset and skip
		w.lastKnown = head
		return false, 0, nil
	}

	count := 0
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count)
	w.lastKnown = head

	return count > 0, count, nil
}

func (w *GitWatcher) RecentCommits(n int) ([]string, error) {
	out, err := exec.Command("git", "log", "--oneline", fmt.Sprintf("-%d", n), "--format=%H").Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

func (w *GitWatcher) FilesChangedInCommits(hashes []string) ([]string, error) {
	args := append([]string{"diff", "--name-only"}, hashes[len(hashes)-1:]...)
	args = append(args, "^"+hashes[0])
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		// Fallback: get files from last commit range
		args2 := []string{"diff", "--name-only", fmt.Sprintf("%s..HEAD", hashes[0])}
		out2, err2 := exec.Command("git", args2...).Output()
		if err2 != nil {
			return nil, fmt.Errorf("git diff: %w", err2)
		}
		return splitLines(string(out2)), nil
	}
	return splitLines(string(out)), nil
}

func (w *GitWatcher) DiffStat(hashes []string) (string, error) {
	args := []string{"diff", "--stat", fmt.Sprintf("%s..HEAD", hashes[0])}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func splitLines(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

type RepoInfo struct {
	ProjectRoot string
	Branch      string
	GoModule    string
}

func DetectRepo(dir string) (*RepoInfo, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(abs, ".git")); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", abs)
	}

	out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	branch := strings.TrimSpace(string(out))

	moduleName := ""
	if data, err := os.ReadFile(filepath.Join(abs, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	return &RepoInfo{
		ProjectRoot: abs,
		Branch:      branch,
		GoModule:    moduleName,
	}, nil
}

type GhostRunner interface {
	RunRefactor(task string) error
}

func PrintNotification(w io.Writer, title, body string) {
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  ┌─ %s ─────────────────────────────────────┐\n", title)
	for _, line := range strings.Split(body, "\n") {
		fmt.Fprintf(w, "  │ %-60s │\n", line)
	}
	fmt.Fprintf(w, "  └──────────────────────────────────────────────┘\n")
}

func PromptUser(msg string) bool {
	fmt.Fprintf(os.Stderr, "\n  %s [Y/n]: ", msg)
	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

package ghost

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WorktreeManager struct {
	repoDir string
	ghostDir string
}

func NewWorktreeManager(repoDir string) *WorktreeManager {
	return &WorktreeManager{
		repoDir:  repoDir,
		ghostDir: filepath.Join(repoDir, ".kode", "ghost"),
	}
}

func (w *WorktreeManager) Create(spec BranchSpec) (string, error) {
	branchName := fmt.Sprintf("ghost/%s", spec.ID)
	worktreePath := filepath.Join(w.ghostDir, string(spec.ID))

	if err := os.MkdirAll(w.ghostDir, 0755); err != nil {
		return "", fmt.Errorf("create ghost dir: %w", err)
	}

	if _, err := os.Stat(worktreePath); err == nil {
		w.Remove(spec.ID)
	}

	// Create orphan branch for this ghost
	orphan := exec.Command("git", "switch", "--orphan", branchName)
	orphan.Dir = w.repoDir
	orphan.Run()
	_ = exec.Command("git", "rm", "-rf", ".").Run()

	// Reset to current HEAD first
	reset := exec.Command("git", "checkout", "-b", branchName)
	reset.Dir = w.repoDir
	reset.Run()

	// Add the worktree pointing at this branch
	add := exec.Command("git", "worktree", "add", worktreePath, "HEAD")
	add.Dir = w.repoDir
	if out, err := add.CombinedOutput(); err != nil {
		// Fallback: checkout branch first
		exec.Command("git", "checkout", "-b", branchName).Dir = w.repoDir
		add = exec.Command("git", "worktree", "add", worktreePath, branchName)
		add.Dir = w.repoDir
		if out2, err2 := add.CombinedOutput(); err2 != nil {
			return "", fmt.Errorf("create worktree: %s: %w", string(out2), err2)
		}
		_ = string(out)
	}

	return worktreePath, nil
}

func (w *WorktreeManager) Remove(id BranchID) error {
	worktreePath := filepath.Join(w.ghostDir, string(id))
	remove := exec.Command("git", "worktree", "remove", worktreePath)
	remove.Dir = w.repoDir
	remove.Run()

	branchName := fmt.Sprintf("ghost/%s", id)
	branchDel := exec.Command("git", "branch", "-D", branchName)
	branchDel.Dir = w.repoDir
	branchDel.Run()

	os.RemoveAll(worktreePath)
	return nil
}

func (w *WorktreeManager) MergeWinner(id BranchID) error {
	branchName := fmt.Sprintf("ghost/%s", id)
	merge := exec.Command("git", "merge", "--squash", branchName)
	merge.Dir = w.repoDir
	if out, err := merge.CombinedOutput(); err != nil {
		// Try with --allow-unrelated-histories
		merge2 := exec.Command("git", "merge", "--squash", "--allow-unrelated-histories", branchName)
		merge2.Dir = w.repoDir
		if out2, err2 := merge2.CombinedOutput(); err2 != nil {
			return fmt.Errorf("merge failed: %s: %w", string(out2), err2)
		}
		_ = string(out)
	}
	return nil
}

func (w *WorktreeManager) SwitchBack() error {
	checkout := exec.Command("git", "checkout", "-")
	checkout.Dir = w.repoDir
	return checkout.Run()
}

func (w *WorktreeManager) GhostDir() string {
	return w.ghostDir
}

func (w *WorktreeManager) CurrentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(out))
}

func (w *WorktreeManager) Cleanup() {
	entries, err := os.ReadDir(w.ghostDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := BranchID(e.Name())
		w.Remove(id)
	}
}

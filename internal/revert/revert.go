package revert

import (
	"fmt"
	"sync"
)

type SnapshotEntry struct {
	FilePath   string
	HunkID     string
	StartLine  int
	EndLine    int
	OldContent []string
}

type Store struct {
	mu      sync.Mutex
	entries []SnapshotEntry
}

var globalStore = &Store{}

func Record(hunkID, filePath string, startLine, endLine int, oldContent []string) {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	for i, e := range globalStore.entries {
		if e.HunkID == hunkID && e.FilePath == filePath {
			globalStore.entries[i] = SnapshotEntry{
				FilePath: filePath, HunkID: hunkID,
				StartLine: startLine, EndLine: endLine,
				OldContent: oldContent,
			}
			return
		}
	}
	globalStore.entries = append(globalStore.entries, SnapshotEntry{
		FilePath: filePath, HunkID: hunkID,
		StartLine: startLine, EndLine: endLine,
		OldContent: oldContent,
	})
}

func Revert(hunkID, filePath string) error {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	idx := -1
	for i, e := range globalStore.entries {
		if e.HunkID == hunkID && e.FilePath == filePath {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("no snapshot found for hunk %s in %s", hunkID, filePath)
	}
	entry := globalStore.entries[idx]
	if entry.StartLine == 0 || len(entry.OldContent) == 0 {
		return fmt.Errorf("invalid snapshot for hunk %s", hunkID)
	}
	lines, err := readLines(entry.FilePath)
	if err != nil {
		return fmt.Errorf("read file for revert: %w", err)
	}
	replaceLen := entry.EndLine - entry.StartLine + 1
	var newLines []string
	newLines = append(newLines, lines[:entry.StartLine-1]...)
	newLines = append(newLines, entry.OldContent...)
	newLines = append(newLines, lines[entry.StartLine-1+replaceLen:]...)
	if err := writeLines(entry.FilePath, newLines); err != nil {
		return fmt.Errorf("write file for revert: %w", err)
	}
	globalStore.entries = append(globalStore.entries[:idx], globalStore.entries[idx+1:]...)
	return nil
}

func List() []SnapshotEntry {
	globalStore.mu.Lock()
	defer globalStore.mu.Unlock()
	out := make([]SnapshotEntry, len(globalStore.entries))
	copy(out, globalStore.entries)
	return out
}

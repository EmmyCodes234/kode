package revert

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordAndRevert(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	original := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(filePath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	Record("hunk-1", filePath, 2, 3, []string{"line2", "line3"})

	modified := "line1\nreplaced2\nreplaced3\nline4\nline5\n"
	if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Revert("hunk-1", filePath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("expected %q, got %q", original, string(data))
	}
}

func TestRevertNotFound(t *testing.T) {
	err := Revert("nonexistent", "/fake/file.go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestList(t *testing.T) {
	globalStore.entries = nil
	Record("h1", "/a.go", 1, 1, []string{"old"})
	entries := List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].HunkID != "h1" {
		t.Fatalf("expected h1, got %s", entries[0].HunkID)
	}
}

func TestRecordDuplicate(t *testing.T) {
	globalStore.entries = nil
	Record("h1", "/a.go", 1, 1, []string{"old"})
	Record("h1", "/a.go", 2, 2, []string{"newer"})
	entries := List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after duplicate, got %d", len(entries))
	}
	if entries[0].StartLine != 2 {
		t.Fatalf("expected StartLine 2, got %d", entries[0].StartLine)
	}
}

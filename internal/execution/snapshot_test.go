package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSnapshot_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	os.WriteFile(path, []byte("package main"), 0644)

	snap, err := CreateSnapshot(dir, []string{"test.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap["test.go"] != "package main" {
		t.Fatalf("expected 'package main', got %q", snap["test.go"])
	}
}

func TestCreateSnapshot_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	snap, err := CreateSnapshot(dir, []string{"nonexistent.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap["nonexistent.go"] != "" {
		t.Fatalf("expected empty string for nonexistent file, got %q", snap["nonexistent.go"])
	}
}

func TestRestoreSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	os.WriteFile(path, []byte("original content"), 0644)

	snap, _ := CreateSnapshot(dir, []string{"test.go"})

	// Modify the file
	os.WriteFile(path, []byte("modified content"), 0644)

	// Restore
	if err := snap.Restore(dir); err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "original content" {
		t.Fatalf("expected 'original content' after restore, got %q", string(data))
	}
}

func TestRestoreSnapshot_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.go")

	// Snapshot when file doesn't exist yet (empty content)
	snap := Snapshot{"new.go": ""}

	// Create the file (simulating apply)
	os.WriteFile(path, []byte("package main"), 0644)

	// Restore should remove it
	if err := snap.Restore(dir); err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected file to be removed after restore")
	}
}

func TestDetectTestCommand_Go(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	cmd := DetectTestCommand(dir)
	if cmd != "go test ./..." {
		t.Fatalf("expected 'go test ./...', got %q", cmd)
	}
}

func TestDetectTestCommand_Node(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	cmd := DetectTestCommand(dir)
	if cmd != "npm test" {
		t.Fatalf("expected 'npm test', got %q", cmd)
	}
}

func TestDetectTestCommand_Cargo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644)

	cmd := DetectTestCommand(dir)
	if !strings.HasPrefix(cmd, "cargo") {
		t.Fatalf("expected 'cargo test', got %q", cmd)
	}
}

func TestDetectTestCommand_Fallback(t *testing.T) {
	dir := t.TempDir()

	cmd := DetectTestCommand(dir)
	if cmd != "go test ./..." {
		t.Fatalf("expected fallback 'go test ./...', got %q", cmd)
	}
}

func TestParseCommand_Simple(t *testing.T) {
	parts := ParseCommand("go test ./...")
	if len(parts) != 3 || parts[0] != "go" || parts[1] != "test" || parts[2] != "./..." {
		t.Fatalf("unexpected parts: %v", parts)
	}
}

func TestParseCommand_Quoted(t *testing.T) {
	parts := ParseCommand(`echo "hello world" foo`)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %v", len(parts), parts)
	}
	if parts[1] != "hello world" {
		t.Fatalf("expected 'hello world', got %q", parts[1])
	}
}

func TestParseCommand_Empty(t *testing.T) {
	parts := ParseCommand("")
	if len(parts) != 0 {
		t.Fatalf("expected empty, got %v", parts)
	}
}

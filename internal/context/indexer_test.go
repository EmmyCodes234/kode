package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContextBuilder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kode_context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder()
	_, err = builder.BuildFullContext(tmpDir)
	if err != nil {
		t.Fatalf("BuildFullContext failed: %v", err)
	}
}

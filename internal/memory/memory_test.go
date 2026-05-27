package memory

import (
	"os"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kode_memory_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test writing
	err = store.RecordVerificationFailure("task123", "missing import", "file.go")
	if err != nil {
		t.Errorf("RecordVerificationFailure error: %v", err)
	}

	// Test reading - no get function exists, so we just check the internal slice
	if len(store.data.Failures) != 1 {
		t.Errorf("Expected 1 history item, got %d", len(store.data.Failures))
	} else {
		if store.data.Failures[0].ErrorMessage != "missing import" {
			t.Errorf("Expected missing import, got %s", store.data.Failures[0].ErrorMessage)
		}
	}
}

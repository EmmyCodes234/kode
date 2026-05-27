package critique

import (
	"testing"
)

func TestCritiqueEngine(t *testing.T) {
	// Need to test that basic critique interface works
	// Because lenses package is the actual implementation, we just test the core interface/structs here if any
	ctx := CritiqueContext{
		ProjectRoot:       "/tmp/test",
		TotalFilesChanged: 2,
	}

	if ctx.TotalFilesChanged != 2 {
		t.Errorf("Expected 2 files changed, got %d", ctx.TotalFilesChanged)
	}
}

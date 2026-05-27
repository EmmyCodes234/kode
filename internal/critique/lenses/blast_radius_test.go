package lenses

import (
	"testing"

	"github.com/kode/kode/internal/critique"
)

func TestBlastRadiusLens(t *testing.T) {
	lens := NewBlastRadiusLens(10)
	
	ctxFail := critique.CritiqueContext{
		ProjectRoot:       "/tmp/test",
		TotalFilesChanged: 15,
	}
	
	findings := lens.Critique("file.go", "code", ctxFail)
	if len(findings) == 0 {
		t.Error("Expected findings for blast radius > 10")
	}

	ctxPass := critique.CritiqueContext{
		ProjectRoot:       "/tmp/test",
		TotalFilesChanged: 5,
	}

	findingsPass := lens.Critique("file.go", "code", ctxPass)
	if len(findingsPass) > 0 {
		t.Error("Expected no findings for blast radius <= 10")
	}
}

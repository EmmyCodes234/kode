package lenses

import (
	"github.com/kode/kode/internal/critique"
)

// DefaultEngine creates a critique engine with all the standard lenses registered.
func DefaultEngine() *critique.Engine {
	engine := critique.NewEngine()

	// Register all standard lenses
	engine.RegisterLens(NewArchitectureLens())
	engine.RegisterLens(NewBlastRadiusLens(5))
	engine.RegisterLens(NewCoherenceLens())
	engine.RegisterLens(NewConventionLens())
	engine.RegisterLens(NewDependencyLens())

	return engine
}

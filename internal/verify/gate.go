package verify

import (
	"context"
	"fmt"
	"strings"
)

type Gate struct {
	diffApplier *DiffApplier
	syntax      *SyntaxChecker
	imports     *ImportValidator
	calls       *CallChecker
	architecture *ArchitectureChecker
}

func NewGate(projectRoot string) *Gate {
	return &Gate{
		diffApplier:  NewDiffApplier(),
		syntax:       NewSyntaxChecker(),
		imports:      NewImportValidator(projectRoot),
		calls:        NewCallChecker(projectRoot),
		architecture: NewArchitectureChecker(),
	}
}

func (g *Gate) Verify(ctx context.Context, req VerifyRequest) (*Verdict, error) {
	modifiedFiles, err := g.diffApplier.ApplyInMemory(req.Diff, req.OriginalFiles)
	if err != nil {
		return &Verdict{
			DiffID:  req.DiffID,
			Overall: StatusFail,
			Results: []CheckResult{{
				CheckName: "diff_applier",
				Status:    StatusFail,
				Message:   "Failed to apply diff to original files in memory",
				Details:   err.Error(),
			}},
		}, nil
	}

	verdict := &Verdict{DiffID: req.DiffID, Overall: StatusPass}

	// Build allowed package manifest from ContextGraph
	allowedPackages := make(map[string]bool)
	graphEntries := make(map[string]bool)
	if req.Graph != nil {
		for _, node := range req.Graph.Nodes {
			if node.Kind == "import" {
				allowedPackages[node.FilePath] = true
			}
		}

		// Build graph entries for call validation: "pkg.Method" -> true
		for _, edge := range req.Graph.Edges {
			if edge.Kind == "calls" || edge.Kind == "defines" {
				sourceFile := string(edge.Source)
				targetID := string(edge.Target)
				graphEntries[fmt.Sprintf("%s.%s", sourceFile, targetID)] = true
			}
		}
	}

	// Build allowed internal packages from modified files
	allowedInternal := make(map[string]bool)
	for path := range modifiedFiles {
		dir := path
		if idx := strings.LastIndex(dir, "/"); idx >= 0 {
			dir = dir[:idx]
		} else if idx := strings.LastIndex(dir, "\\"); idx >= 0 {
			dir = dir[:idx]
		}
		allowedInternal[dir] = true
		allowedInternal[path] = true

		// Add every node name as allowed internal
		if req.Graph != nil {
			for _, node := range req.Graph.Nodes {
				if node.Kind == "import" {
					allowedInternal[node.FilePath] = true
				}
				if node.Kind == "file" {
					allowedInternal[node.Name] = true
				}
			}
		}
	}

	// Check 1: Syntax — hard block on parse errors
	for path, content := range modifiedFiles {
		result := g.syntax.CheckFile(path, content)
		verdict.Results = append(verdict.Results, result)
		if result.Status == StatusFail {
			verdict.Overall = StatusFail
			return verdict, nil
		}
	}

	// Check 2: Imports — hard block on unresolvable imports
	for path, content := range modifiedFiles {
		result := g.imports.Validate(path, content, allowedInternal)
		verdict.Results = append(verdict.Results, result)
		if result.Status == StatusFail {
			verdict.Overall = StatusFail
			return verdict, nil
		}
	}

	// Check 3: Calls — hard block on hallucinated package calls, warn on unresolvable local calls
	for path, content := range modifiedFiles {
		result := g.calls.CheckFile(path, content, allowedPackages, graphEntries)
		verdict.Results = append(verdict.Results, result)
		if result.Status == StatusFail {
			verdict.Overall = StatusFail
			return verdict, nil
		}
	}

	// Check 4: Architecture — configurable: block or warn
	if len(req.ArchitectureRules) > 0 {
		for path, content := range modifiedFiles {
			result := g.architecture.CheckFile(path, content, req.ArchitectureRules)
			verdict.Results = append(verdict.Results, result)
			if result.Status == StatusFail {
				if req.BlockOnArchitecture {
					verdict.Overall = StatusFail
					return verdict, nil
				}
				// If not blocking, downgrade to WARN
				verdict.Overall = StatusPass
				for i, r := range verdict.Results {
					if r.CheckName == "architecture" && r.Status == StatusFail {
						verdict.Results[i].Status = StatusWarn
					}
				}
			}
		}
	}

	return verdict, nil
}

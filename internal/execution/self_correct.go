package execution

import (
	"fmt"
	"strings"
)

type SelfCorrectionEngine struct{}

func NewSelfCorrectionEngine() *SelfCorrectionEngine {
	return &SelfCorrectionEngine{}
}

func (sc *SelfCorrectionEngine) BuildRepairPrompt(failedHunk StructuredHunk, failureMessage string, passingHunks []StructuredHunk) string {
	var builder strings.Builder

	builder.WriteString("### [KODE SELF-CORRECTION LOOP]\n")
	builder.WriteString("An isolated block of your generated patch failed deterministic verification.\n\n")

	builder.WriteString("### [IMMUTABLE CONSTRAINTS — DO NOT MODIFY]\n")
	builder.WriteString("The following hunks passed all verification params and are locked:\n\n")
	for _, pass := range passingHunks {
		builder.WriteString(fmt.Sprintf("- Hunk %s on %s (%s)\n", pass.ID, pass.FilePath, pass.Action))
	}
	builder.WriteString("\n")

	builder.WriteString("### [FAILING HUNK — REGENERATE ONLY THIS]\n")
	builder.WriteString(fmt.Sprintf("- Hunk ID: %s\n", failedHunk.ID))
	builder.WriteString(fmt.Sprintf("- File: %s\n", failedHunk.FilePath))
	builder.WriteString(fmt.Sprintf("- Action: %s\n", failedHunk.Action))
	if failedHunk.TargetSymbol != "" {
		builder.WriteString(fmt.Sprintf("- Target: %s\n", failedHunk.TargetSymbol))
	}
	builder.WriteString(fmt.Sprintf("- Failure: %s\n\n", failureMessage))

	builder.WriteString("### [OUTPUT]\n")
	builder.WriteString("Emit a single JSON hunk correcting the failure. Respect immutable constraints above.\n")

	return builder.String()
}

package execution

type HunkAction string

const (
	ActionInsert HunkAction = "INSERT"
	ActionDelete HunkAction = "DELETE"
	ActionModify HunkAction = "MODIFY"
)

type StructuredHunk struct {
	ID           string     `json:"id"`
	FilePath     string     `json:"file_path"`
	Action       HunkAction `json:"action"`
	TargetSymbol string     `json:"target_symbol,omitempty"`
	AnchorText   string     `json:"anchor_text,omitempty"`
	NewText      string     `json:"new_text,omitempty"`
	Explanation  string     `json:"explanation,omitempty"`
}

type PatchIntent struct {
	TaskID string           `json:"task_id"`
	Hunks  []StructuredHunk `json:"hunks"`
}

type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
)

type ExecutionSummary struct {
	TaskID       string            `json:"task_id"`
	Status       Status            `json:"status"`
	RoundsUsed   int               `json:"rounds_used"`
	AppliedHunks []string          `json:"applied_hunks"`
	FailedHunks  map[string]string `json:"failed_hunks"`
}

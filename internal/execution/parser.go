package execution

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HunkParser struct{}

func NewHunkParser() *HunkParser {
	return &HunkParser{}
}

func (hp *HunkParser) ParseLLMResponse(rawInput string) ([]StructuredHunk, error) {
	cleaned := strings.TrimSpace(rawInput)

	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "[]" {
		return nil, fmt.Errorf("empty generation packet returned from LLM execution pipeline")
	}

	var hunks []StructuredHunk
	if err := json.Unmarshal([]byte(cleaned), &hunks); err != nil {
		return nil, fmt.Errorf("failed to decode execution structural layout: %w", err)
	}

	for i, hunk := range hunks {
		if hunk.FilePath == "" {
			return nil, fmt.Errorf("validation error at hunk idx %d: missing file target route", i)
		}
		if hunk.Action != ActionInsert && hunk.Action != ActionDelete && hunk.Action != ActionModify {
			return nil, fmt.Errorf("validation error at hunk ID %s: invalid operation signature (%s)", hunk.ID, hunk.Action)
		}
		if hunk.Action == ActionModify && hunk.AnchorText == "" {
			return nil, fmt.Errorf("validation error at hunk ID %s: MODIFY requires anchor text", hunk.ID)
		}
		if hunk.Action == ActionDelete && hunk.AnchorText == "" {
			return nil, fmt.Errorf("validation error at hunk ID %s: DELETE requires anchor text", hunk.ID)
		}
		if hunk.Action == ActionInsert && hunk.NewText == "" {
			return nil, fmt.Errorf("validation error at hunk ID %s: INSERT requires new text", hunk.ID)
		}
	}

	return hunks, nil
}

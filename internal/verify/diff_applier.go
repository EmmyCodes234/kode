package verify

import (
	"fmt"
	"strconv"
	"strings"
)

type DiffApplier struct{}

func NewDiffApplier() *DiffApplier {
	return &DiffApplier{}
}

type HunkLineType int

const (
	HunkContext  HunkLineType = 0
	HunkAddition HunkLineType = 1
	HunkDeletion HunkLineType = 2
)

type HunkLine struct {
	Type HunkLineType
	Text string
}

type HunkHeader struct {
	OrigStart int
	OrigLines int
	NewStart  int
	NewLines  int
}

type Hunk struct {
	Header HunkHeader
	Lines  []HunkLine
}

func (da *DiffApplier) ApplyInMemory(diff string, originalFiles map[string]string) (map[string]string, error) {
	mutated := make(map[string]string)
	for k, v := range originalFiles {
		mutated[k] = v
	}

	lines := strings.Split(diff, "\n")
	var currentFile string
	var hunks []Hunk
	var currentHunk Hunk
	inHunk := false
	seenHeader := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "--- ") {
			continue
		}

		if strings.HasPrefix(line, "+++ ") {
			if inHunk && len(currentHunk.Lines) > 0 {
				hunks = append(hunks, currentHunk)
				currentHunk = Hunk{}
			}
			if len(hunks) > 0 && currentFile != "" {
				applied, err := applyHunks(mutated[currentFile], hunks)
				if err != nil {
					return nil, fmt.Errorf("failed to apply hunks to %s: %w", currentFile, err)
				}
				mutated[currentFile] = applied
				hunks = nil
			}

			target := strings.TrimSpace(line[4:])
			if strings.HasPrefix(target, "b/") {
				target = target[2:]
			}
			currentFile = target
			if _, exists := mutated[currentFile]; !exists {
				mutated[currentFile] = ""
			}
			inHunk = false
			seenHeader = true
			currentHunk = Hunk{}
			continue
		}

		if strings.HasPrefix(line, "@@") {
			if inHunk && len(currentHunk.Lines) > 0 {
				hunks = append(hunks, currentHunk)
			}
			currentHunk = parseHunkHeader(line)
			inHunk = true
			continue
		}

		if inHunk {
			if len(line) == 0 {
				currentHunk.Lines = append(currentHunk.Lines, HunkLine{Type: HunkContext, Text: ""})
			} else {
				prefix := line[0]
				text := line[1:]
				switch prefix {
				case ' ':
					currentHunk.Lines = append(currentHunk.Lines, HunkLine{Type: HunkContext, Text: text})
				case '+':
					currentHunk.Lines = append(currentHunk.Lines, HunkLine{Type: HunkAddition, Text: text})
				case '-':
					currentHunk.Lines = append(currentHunk.Lines, HunkLine{Type: HunkDeletion, Text: text})
				}
			}
		}
	}

	if inHunk && len(currentHunk.Lines) > 0 {
		hunks = append(hunks, currentHunk)
	}
	if len(hunks) > 0 && currentFile != "" {
		applied, err := applyHunks(mutated[currentFile], hunks)
		if err != nil {
			return nil, fmt.Errorf("failed to apply hunks to %s: %w", currentFile, err)
		}
		mutated[currentFile] = applied
	}

	if !seenHeader && diff != "" {
		return nil, fmt.Errorf("no valid diff headers found (expected '+++ b/...' lines)")
	}

	return mutated, nil
}

func parseHunkHeader(line string) Hunk {
	h := Hunk{}
	line = strings.TrimPrefix(line, "@@ ")
	idx := strings.LastIndex(line, " @@")
	if idx >= 0 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)

	parts := strings.Split(line, " ")
	if len(parts) >= 2 {
		origPart := parts[0]
		newPart := parts[1]

		origStart, origLines := parseHunkRange(origPart)
		newStart, newLines := parseHunkRange(newPart)
		h.Header = HunkHeader{
			OrigStart: origStart,
			OrigLines: origLines,
			NewStart:  newStart,
			NewLines:  newLines,
		}
	}
	return h
}

func parseHunkRange(s string) (int, int) {
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "+")
	parts := strings.Split(s, ",")
	start, _ := strconv.Atoi(parts[0])
	lines := 1
	if len(parts) > 1 {
		lines, _ = strconv.Atoi(parts[1])
	}
	return start, lines
}

func applyHunks(original string, hunks []Hunk) (string, error) {
	lines := strings.Split(original, "\n")
	if len(original) == 0 {
		lines = []string{}
	}

	currentLine := 0
	var result []string

	for _, hunk := range hunks {
		origStart := hunk.Header.OrigStart
		if origStart > 0 {
			origStart--
		}

		for currentLine < origStart && currentLine < len(lines) {
			result = append(result, lines[currentLine])
			currentLine++
		}

		for _, hl := range hunk.Lines {
			switch hl.Type {
			case HunkContext:
				if currentLine < len(lines) {
					result = append(result, lines[currentLine])
					currentLine++
				}
			case HunkDeletion:
				if currentLine < len(lines) {
					currentLine++
				}
			case HunkAddition:
				result = append(result, hl.Text)
			}
		}
	}

	for currentLine < len(lines) {
		result = append(result, lines[currentLine])
		currentLine++
	}

	return strings.Join(result, "\n"), nil
}

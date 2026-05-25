package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
)

type stepSpinner struct {
	frames []string
	idx    int
	label  string
}

func newSpinner(label string) *stepSpinner {
	return &stepSpinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		label:  label,
	}
}

func (s *stepSpinner) tick() {
	s.idx = (s.idx + 1) % len(s.frames)
	fmt.Fprintf(os.Stderr, "\r%s %s%s%s ", ansiCyan, s.frames[s.idx], ansiReset, s.label)
}

func stepStart(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %s•%s %s\n", ansiBlue, ansiReset, fmt.Sprintf(format, args...))
}

func stepOK(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %s✓%s %s\n", ansiGreen, ansiReset, fmt.Sprintf(format, args...))
}

func stepFail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %s✗%s %s\n", ansiRed, ansiReset, fmt.Sprintf(format, args...))
}

func stepDetail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "    %s%s%s\n", ansiDim, fmt.Sprintf(format, args...), ansiReset)
}

func sectionHeader(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "\n%s%s%s\n", ansiBold, line, ansiReset)
	fmt.Fprintf(os.Stderr, "%s%s%s\n", ansiDim, strings.Repeat("─", len(line)), ansiReset)
}

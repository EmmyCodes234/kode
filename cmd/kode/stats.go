package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type logEntry struct {
	Timestamp  string            `json:"timestamp"`
	Status     string            `json:"status"`
	Files      []string          `json:"files,omitempty"`
	Failures   map[string]string `json:"failures,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Model      string            `json:"model,omitempty"`
}

func init() {
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show gatekeeper verification statistics from logs/kode.log",
		Long: `Parse logs/kode.log and display aggregate verification metrics,
including pass/fail rates, most frequently failing files, and
failure type breakdown.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logDir, _ := cmd.Flags().GetString("log-dir")
			top, _ := cmd.Flags().GetInt("top")

			if logDir == "" {
				pwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine project directory: %w", err)
				}
				logDir = filepath.Join(pwd, "logs")
			}

			logPath := filepath.Join(logDir, "kode.log")
			f, err := os.Open(logPath)
			if err != nil {
				return fmt.Errorf("cannot open log file %s: %w", logPath, err)
			}
			defer f.Close()

			type fileFail struct {
				path   string
				count  int
			}
			type typeFail struct {
				name  string
				count int
			}

			var entries []logEntry
			failByFile := make(map[string]int)
			failByType := make(map[string]int)

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				var entry logEntry
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					continue
				}
				entries = append(entries, entry)

				if strings.ToUpper(entry.Status) == "FAIL" {
					for fp, reason := range entry.Failures {
						failByFile[fp]++
						parts := strings.SplitN(reason, ":", 2)
						failByType[strings.TrimSpace(parts[0])]++
					}
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading log: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("No verification records found in", logPath)
				return nil
			}

			total := len(entries)
			passed := 0
			failed := 0
			var totalDuration int64
			for _, e := range entries {
				totalDuration += e.DurationMs
				if strings.ToUpper(e.Status) == "PASS" {
					passed++
				} else {
					failed++
				}
			}
			passRate := float64(passed) / float64(total) * 100
			failRate := float64(failed) / float64(total) * 100
			avgDuration := float64(totalDuration) / float64(total)

			w := func(format string, args ...interface{}) {
				fmt.Printf("  │  "+format+"\n", args...)
			}

			fmt.Println("  ┌─ Kode Gatekeeper Statistics ──────────────────────────────────────────┐")
			w("")
			w("Total Verifications:  %d", total)
			w("Passed:               %d  (%.1f%%)", passed, passRate)
			w("Failed:               %d  (%.1f%%)", failed, failRate)
			w("Avg Duration:         %.1fms", avgDuration)
			w("")

			if failed > 0 {
				w("Failure Breakdown (%d total):", failed)
				var sortedTypes []typeFail
				for n, c := range failByType {
					sortedTypes = append(sortedTypes, typeFail{n, c})
				}
				sort.Slice(sortedTypes, func(i, j int) bool {
					return sortedTypes[i].count > sortedTypes[j].count
				})
				for _, ft := range sortedTypes {
					pct := float64(ft.count) / float64(failed) * 100
					w("  %-12s %d  (%.1f%%)", ft.name+":", ft.count, pct)
				}
				w("")

				w("Most Failed Files (top %d):", top)
				var sortedFiles []fileFail
				for p, c := range failByFile {
					sortedFiles = append(sortedFiles, fileFail{p, c})
				}
				sort.Slice(sortedFiles, func(i, j int) bool {
					return sortedFiles[i].count > sortedFiles[j].count
				})
				limit := top
				if len(sortedFiles) < limit {
					limit = len(sortedFiles)
				}
				for _, ff := range sortedFiles[:limit] {
					w("  %-45s %d failures", truncate(ff.path, 45), ff.count)
				}
				w("")
			}

			w("Recent Trend (last 20):")
			tStart := 0
			if len(entries) > 20 {
				tStart = len(entries) - 20
			}
			var trend []string
			for _, e := range entries[tStart:] {
				if strings.ToUpper(e.Status) == "PASS" {
					trend = append(trend, "✓")
				} else {
					trend = append(trend, "✗")
				}
			}
			w("  %s", strings.Join(trend, " "))
			w("")

			fmt.Println("  └────────────────────────────────────────────────────────────────────────┘")
			return nil
		},
	}

	statsCmd.Flags().String("log-dir", "", "Directory containing kode.log (default: <cwd>/logs)")
	statsCmd.Flags().Int("top", 10, "Number of top-failing files to show")
	rootCmd.AddCommand(statsCmd)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

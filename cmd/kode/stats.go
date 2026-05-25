package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type logEntry struct {
	Timestamp  string            `json:"timestamp"`
	TaskID     string            `json:"task_id,omitempty"`
	Status     string            `json:"status"`
	Files      []string          `json:"files,omitempty"`
	Failures   map[string]string `json:"failures,omitempty"`
	RoundsUsed int               `json:"rounds_used,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Model      string            `json:"model,omitempty"`
}

type fileFail struct {
	path  string
	count int
}

type typeFail struct {
	name  string
	count int
}

type dayBucket struct {
	date   string
	passed int
	failed int
}

func init() {
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show gatekeeper verification statistics from logs/kode.log",
		Long: `Parse logs/kode.log and display aggregate verification metrics,
including pass/fail rates, trend analysis, most frequently failing
files, and failure type breakdown.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logDir, _ := cmd.Flags().GetString("log-dir")
			top, _ := cmd.Flags().GetInt("top")
			days, _ := cmd.Flags().GetInt("days")

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

			var entries []logEntry
			failByFile := make(map[string]int)
			failByType := make(map[string]int)
			modelCount := make(map[string]int)
			dailyTrend := make(map[string]*dayBucket)

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

				if entry.Model != "" {
					modelCount[entry.Model]++
				}

				dayKey := ""
				if ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
					dayKey = ts.Format("2006-01-02")
				}
				if dayKey != "" {
					if dailyTrend[dayKey] == nil {
						dailyTrend[dayKey] = &dayBucket{date: dayKey}
					}
				}

				if strings.ToUpper(entry.Status) == "PASS" {
					if dayKey != "" && dailyTrend[dayKey] != nil {
						dailyTrend[dayKey].passed++
					}
				} else {
					if dayKey != "" && dailyTrend[dayKey] != nil {
						dailyTrend[dayKey].failed++
					}
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
			w("Avg Duration:         %.1fms  (total: %.1fs)", avgDuration, float64(totalDuration)/1000)
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

			if len(modelCount) > 0 {
				w("Models Used:")
				var sortedModels []struct {
					name  string
					count int
				}
				for n, c := range modelCount {
					sortedModels = append(sortedModels, struct {
						name  string
						count int
					}{n, c})
				}
				sort.Slice(sortedModels, func(i, j int) bool {
					return sortedModels[i].count > sortedModels[j].count
				})
				for _, m := range sortedModels {
					pct := float64(m.count) / float64(total) * 100
					w("  %-30s %d  (%.1f%%)", m.name+":", m.count, pct)
				}
				w("")
			}

			w("Daily Trend (last %d days):", days)
			var sortedDays []*dayBucket
			for _, b := range dailyTrend {
				sortedDays = append(sortedDays, b)
			}
			sort.Slice(sortedDays, func(i, j int) bool {
				return sortedDays[i].date < sortedDays[j].date
			})
			start := 0
			if len(sortedDays) > days {
				start = len(sortedDays) - days
			}
			for _, d := range sortedDays[start:] {
				totalDay := d.passed + d.failed
				rate := float64(d.passed) / float64(totalDay) * 100
				bar := ""
				barLen := (d.passed * 20) / totalDay
				for i := 0; i < 20; i++ {
					if i < barLen {
						bar += "█"
					} else {
						bar += "░"
					}
				}
				w("  %s %s %d/%d (%.0f%%)", d.date, bar, d.passed, totalDay, rate)
			}
			w("")

			w("Recent Verdicts (last 20):")
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
	statsCmd.Flags().Int("days", 14, "Number of days for daily trend view")
	rootCmd.AddCommand(statsCmd)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

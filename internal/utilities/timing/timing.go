// Package timing provides utilities for measuring and reporting execution time
// of various code sections across the codebase.
package timing

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// Enabled controls whether timing is active (set via TIMING=1 environment variable)
var Enabled = os.Getenv("TIMING") == "1"

// Collector aggregates timing measurements from multiple code sections
type Collector struct {
	mu       sync.Mutex
	timings  map[string][]time.Duration
	order    []string // preserve insertion order
	category string
}

// NewCollector creates a new timing collector with an optional category name
func NewCollector(category string) *Collector {
	return &Collector{
		timings:  make(map[string][]time.Duration),
		order:    make([]string, 0),
		category: category,
	}
}

// globalCollector is used when no specific collector is provided
var globalCollector = NewCollector("Global")
var globalMu sync.Mutex

// Track measures the execution time of a code block.
// Usage:
//
//	defer timing.Track("MyFunction")()
//
// Or with a collector:
//
//	defer collector.Track("MyFunction")()
func Track(name string) func() {
	if !Enabled {
		return func() {}
	}
	start := time.Now()
	return func() {
		globalMu.Lock()
		defer globalMu.Unlock()
		globalCollector.record(name, time.Since(start))
	}
}

// Track measures the execution time and records it to this collector
func (c *Collector) Track(name string) func() {
	if !Enabled {
		return func() {}
	}
	start := time.Now()
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.record(name, time.Since(start))
	}
}

// record adds a timing measurement (must be called with lock held)
func (c *Collector) record(name string, d time.Duration) {
	if _, exists := c.timings[name]; !exists {
		c.order = append(c.order, name)
	}
	c.timings[name] = append(c.timings[name], d)
}

// TimingStats holds statistics for a single named timing
type TimingStats struct {
	Name    string
	Count   int
	Total   time.Duration
	Average time.Duration
}

// Stats returns statistics for all recorded timings
func (c *Collector) Stats() []TimingStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := make([]TimingStats, 0, len(c.timings))
	for _, name := range c.order {
		durations := c.timings[name]
		if len(durations) == 0 {
			continue
		}

		var total time.Duration
		for _, d := range durations {
			total += d
		}

		stats = append(stats, TimingStats{
			Name:    name,
			Count:   len(durations),
			Total:   total,
			Average: total / time.Duration(len(durations)),
		})
	}

	return stats
}

// Reset clears all recorded timings
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timings = make(map[string][]time.Duration)
	c.order = make([]string, 0)
}

// Summary returns a formatted summary string
func (c *Collector) Summary() string {
	stats := c.Stats()
	if len(stats) == 0 {
		return "No timing data collected"
	}

	var total time.Duration
	for _, s := range stats {
		total += s.Total
	}

	var sb strings.Builder
	width := 80

	// Header
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("═", width) + "\n")
	title := fmt.Sprintf("  TIMING SUMMARY: %s  ", c.category)
	padding := (width - len(title)) / 2
	sb.WriteString(strings.Repeat(" ", padding) + title + "\n")
	sb.WriteString(strings.Repeat("═", width) + "\n")

	// Column headers
	sb.WriteString(fmt.Sprintf("%-30s │ %8s │ %12s │ %12s │ %6s\n",
		"Name", "Count", "Average", "Total", "%"))
	sb.WriteString(strings.Repeat("─", width) + "\n")

	// Data rows
	for _, s := range stats {
		pct := float64(s.Total) / float64(total) * 100
		sb.WriteString(fmt.Sprintf("%-30s │ %8d │ %12v │ %12v │ %5.1f%%\n",
			truncate(s.Name, 30), s.Count, formatDuration(s.Average), formatDuration(s.Total), pct))
	}

	// Footer
	sb.WriteString(strings.Repeat("─", width) + "\n")
	sb.WriteString(fmt.Sprintf("%-30s │ %8s │ %12s │ %12v │ %5.1f%%\n",
		"TOTAL", "", "", formatDuration(total), 100.0))
	sb.WriteString(strings.Repeat("═", width) + "\n")

	return sb.String()
}

// PrintSummary prints the timing summary to stdout
func (c *Collector) PrintSummary() {
	if !Enabled {
		return
	}
	fmt.Print(c.Summary())
}

// GlobalSummary returns the global collector's summary
func GlobalSummary() string {
	return globalCollector.Summary()
}

// PrintGlobalSummary prints the global timing summary
func PrintGlobalSummary() {
	if !Enabled {
		return
	}
	globalCollector.PrintSummary()
}

// ResetGlobal clears the global collector
func ResetGlobal() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalCollector.Reset()
}

// SortByTotal sorts stats by total time (descending)
func SortByTotal(stats []TimingStats) {
	slices.SortFunc(stats, func(a, b TimingStats) int {
		switch {
		case a.Total > b.Total:
			return -1
		case a.Total < b.Total:
			return 1
		default:
			return 0
		}
	})
}

// Helper functions

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d)/float64(time.Microsecond))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.2fs", float64(d)/float64(time.Second))
}

// BenchmarkStats holds statistical data for benchmark runs
type BenchmarkStats struct {
	Min    time.Duration
	Max    time.Duration
	Mean   time.Duration
	StdDev time.Duration
	P50    time.Duration // Median
	P75    time.Duration
	P90    time.Duration
	P99    time.Duration
}

// CalculateBenchmarkStats calculates statistics from multiple durations
func CalculateBenchmarkStats(durations []time.Duration) BenchmarkStats {
	if len(durations) == 0 {
		return BenchmarkStats{}
	}

	// Sort for percentile calculation
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	slices.Sort(sorted)

	var sum time.Duration
	for _, d := range durations {
		sum += d
	}

	mean := sum / time.Duration(len(durations))

	// Calculate variance and std dev
	var variance float64
	for _, d := range durations {
		diff := float64(d - mean)
		variance += diff * diff
	}
	variance /= float64(len(durations))
	stdDev := time.Duration(math.Sqrt(variance))

	// Calculate percentiles
	percentile := func(p float64) time.Duration {
		idx := int(float64(len(sorted)) * p)
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		return sorted[idx]
	}

	return BenchmarkStats{
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Mean:   mean,
		StdDev: stdDev,
		P50:    percentile(0.50),
		P75:    percentile(0.75),
		P90:    percentile(0.90),
		P99:    percentile(0.99),
	}
}

// BenchmarkSummary holds aggregated statistics for all custom timings
type BenchmarkSummary struct {
	Timings map[string]BenchmarkStats
}

// CalculateBenchmarkSummary calculates statistics from multiple collector snapshots
// Each element in allRunsStats should be the Stats() result from a single run
func CalculateBenchmarkSummary(allRunsStats [][]TimingStats) BenchmarkSummary {
	summary := BenchmarkSummary{
		Timings: make(map[string]BenchmarkStats),
	}

	// Collect all timing names
	timingNames := make(map[string]bool)
	for _, runStats := range allRunsStats {
		for _, stat := range runStats {
			timingNames[stat.Name] = true
		}
	}

	// Calculate stats for each timing
	for name := range timingNames {
		var durations []time.Duration
		for _, runStats := range allRunsStats {
			for _, stat := range runStats {
				if stat.Name == name {
					// Use average from this run
					durations = append(durations, stat.Average)
					break
				}
			}
		}
		if len(durations) > 0 {
			summary.Timings[name] = CalculateBenchmarkStats(durations)
		}
	}

	return summary
}

// FormatBenchmarkTable formats benchmark statistics as a table
func (bs BenchmarkSummary) FormatBenchmarkTable(runCount int) string {
	if len(bs.Timings) == 0 {
		return "No benchmark data available"
	}

	// Sort timings by mean time (descending)
	type timingStat struct {
		name  string
		stats BenchmarkStats
	}

	timings := make([]timingStat, 0, len(bs.Timings))
	for name, stats := range bs.Timings {
		timings = append(timings, timingStat{name: name, stats: stats})
	}

	slices.SortFunc(timings, func(a, b timingStat) int {
		if a.stats.Mean > b.stats.Mean {
			return -1
		} else if a.stats.Mean < b.stats.Mean {
			return 1
		}
		return 0
	})

	// Find max for bar scaling
	maxMean := time.Duration(0)
	for _, t := range timings {
		if t.stats.Mean > maxMean {
			maxMean = t.stats.Mean
		}
	}

	var sb strings.Builder
	width := 90

	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", width))
	fmt.Fprintf(&sb, "  CUSTOM TIMING BENCHMARK STATISTICS (%d runs)\n", runCount)
	fmt.Fprintf(&sb, "%s\n\n", strings.Repeat("=", width))

	fmt.Fprintf(&sb, "  Name                      │    Mean    │     P50     │     P90     │     P99     │    StdDev   │  Graph\n")
	fmt.Fprintf(&sb, "  %s\n", strings.Repeat("-", width-2))

	for _, t := range timings {
		bar := createTimingBar(t.stats.Mean, maxMean, 25)
		fmt.Fprintf(&sb, "  %-25s │ %10v │ %11v │ %11v │ %11v │ %11v │ %s\n",
			truncate(t.name, 25),
			formatBenchmarkDuration(t.stats.Mean),
			formatBenchmarkDuration(t.stats.P50),
			formatBenchmarkDuration(t.stats.P90),
			formatBenchmarkDuration(t.stats.P99),
			formatBenchmarkDuration(t.stats.StdDev),
			bar)
	}

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", width))

	return sb.String()
}

func createTimingBar(value, max time.Duration, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int((float64(value) / float64(max)) * float64(width))
	filled = min(filled, width)
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func formatBenchmarkDuration(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)
	if ms < 1 {
		return fmt.Sprintf("%.2fμs", ms*1000)
	} else if ms < 1000 {
		return fmt.Sprintf("%.2fms", ms)
	} else {
		return fmt.Sprintf("%.2fs", ms/1000)
	}
}

// PrintBenchmarkSummary prints benchmark statistics
func PrintBenchmarkSummary(allRunsStats [][]TimingStats, runCount int) {
	summary := CalculateBenchmarkSummary(allRunsStats)
	fmt.Println(summary.FormatBenchmarkTable(runCount))
}

// GetGlobalStats returns a snapshot of the global collector's stats
func GetGlobalStats() []TimingStats {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalCollector.Stats()
}

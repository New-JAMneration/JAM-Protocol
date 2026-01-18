// Package timing provides utilities for measuring and reporting execution time
// of various code sections across the codebase.
package timing

import (
	"fmt"
	"os"
	"sort"
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
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Total > stats[j].Total
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

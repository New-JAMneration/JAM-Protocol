package fuzz

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/timing"
)

var importBlockTimings []stf.STFTiming

// ResetImportBlockTimings clears collected STF timings for a new fuzz server session.
func ResetImportBlockTimings() {
	importBlockTimings = nil
	timing.ResetGlobal()
}

// RecordImportBlockTiming appends one ImportBlock STF timing sample.
func RecordImportBlockTiming(t stf.STFTiming) {
	importBlockTimings = append(importBlockTimings, t)
}

// PrintImportBlockTimingSummary prints aggregated STF timing when TIMING=1.
func PrintImportBlockTimingSummary() {
	if !timing.Enabled || len(importBlockTimings) == 0 {
		return
	}
	stf.PrintTimingSummary(importBlockTimings)
}

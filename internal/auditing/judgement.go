package auditing

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

// DefaultBundleFetcher — swap with real impl once CE bundle request is ready.
var DefaultBundleFetcher BundleFetcher = &StubBundleFetcher{}

// GetJudgement implements GP §17.16–17.17: fetch bundle → re-execute Ξ(p,c) →
// compare against claimed report. Returns true if match, false otherwise.
func GetJudgement(auditReport types.AuditReport) bool {
	report := auditReport.Report
	coreIndex := report.CoreIndex

	// Step A: Fetch original bundle from guarantor or erasure reconstruction.
	bundleBytes, err := DefaultBundleFetcher.FetchBundle(report)
	if err != nil {
		return false
	}

	// Step B: Re-execute Ξ(p,c) — decode → ΨI → ΨR per item → assemble report.
	controller := work_package.NewSharedController(bundleBytes, coreIndex)
	recomputed, err := controller.Process()
	if err != nil {
		return false
	}

	// Step C: Compare recomputed vs original.
	return workReportsEqual(recomputed, report)
}

// workReportsEqual compares two WorkReports via canonical serialization (GP §C.27).
func workReportsEqual(a, b types.WorkReport) bool {
	encodedA := utilities.WorkReportSerialization(a)
	encodedB := utilities.WorkReportSerialization(b)
	return bytes.Equal(encodedA, encodedB)
}

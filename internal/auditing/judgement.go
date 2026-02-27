package auditing

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

// DefaultBundleFetcher is the bundle fetcher used by GetJudgement.
// Node layer should replace this with a real implementation that
// fetches bundles via CE protocol from guarantors or via erasure-coding
// reconstruction from validators.
var DefaultBundleFetcher BundleFetcher = &StubBundleFetcher{}

// GetJudgement implements GP §17.16–17.17: the auditor's evaluation of a
// work-report. It re-executes Ξ(p,c) on the original work-package bundle
// and compares the result against the claimed work-report.
//
// GP §17.16:
//
//	∀(c, w) ∈ aₙ :
//	  eₙ(c) ⟺ w = Ξ(p, c)  if ∃p ∈ P : E(p) = F(r)
//	            ⊥             otherwise
//
// Returns true if the recomputed report matches the original (valid),
// false otherwise (invalid, including fetch/decode/execution failures).
func GetJudgement(auditReport types.AuditReport) bool {
	report := auditReport.Report
	coreIndex := report.CoreIndex

	// Step A: Fetch the original work-package bundle.
	// The auditor needs the raw bundle (work-package + extrinsic data +
	// import segments) to re-execute. This is obtained from a guarantor
	// (fast path) or reconstructed from erasure-coded chunks (slow path).
	bundleBytes, err := DefaultBundleFetcher.FetchBundle(report)
	if err != nil {
		// GP §17.16: failure to decode implies an invalid work-report.
		return false
	}

	// Step B: Re-execute Ξ(p, c) using WorkPackageController.
	// NewSharedController follows the same "shared guarantor" path:
	// decode bundle → extract inputs → ΨI → ΨR per item → assemble WorkReport.
	// All PVM infrastructure (ΨI, ΨR, host calls 0-13) is already in place.
	controller := work_package.NewSharedController(bundleBytes, coreIndex)
	recomputed, err := controller.Process()
	if err != nil {
		return false
	}

	// Step C: Compare recomputed report against original.
	return workReportsEqual(recomputed, report)
}

// workReportsEqual compares two WorkReports by their canonical serialization.
// This follows GP §17.16: eₙ(c) ⟺ w = Ξ(p, c), where equality is over
// the full encoded work-report.
//
// We use WorkReportSerialization (GP §C.27) which is the canonical encoding
// already used throughout the codebase for hashing and signing work-reports.
func workReportsEqual(a, b types.WorkReport) bool {
	encodedA := utilities.WorkReportSerialization(a)
	encodedB := utilities.WorkReportSerialization(b)
	return bytes.Equal(encodedA, encodedB)
}

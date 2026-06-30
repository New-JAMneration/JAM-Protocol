package work_package

import "github.com/New-JAMneration/JAM-Protocol/internal/pvmcost"

// WorkPackageTelemetryCost is the observability sidecar for one work-package
// evaluation (#974 Phase 1). It carries the PVM cost summaries for JIP-3
// events 95 (Is-Authorized) and 101 (Refine) parallel to the consensus
// types.WorkReport — cost must never ride on WorkReport / WorkResult (the
// CI guard in internal/pvmcost enforces this). Values are zero-filled until
// Phase 2a instrumentation lands.
//
// Failure paths discard cost by design: when WorkReportCompute errors
// (Psi_I not ok / oversize output), no sidecar is produced. JIP-3 maps
// those failures to event 92, which carries a reason string and no cost;
// events 95/101 fire on success only.
type WorkPackageTelemetryCost struct {
	IsAuthorized pvmcost.IsAuthorizedCost
	// Refine is index-aligned with WorkPackage.Items (and therefore with
	// WorkReport.Results) — event 101's payload is one RefineCost per work
	// item, so per-item granularity is kept here. Failed items keep their
	// slot so alignment holds.
	Refine []pvmcost.RefineCost
}

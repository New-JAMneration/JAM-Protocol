package accumulation

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (12.1) ξ ∈ ⟦{H}⟧_E: store.Xi

// (12.2) ©ξ ≡ ⋃x∈ξ
func GetAccumulatedHashes() (output []types.WorkPackageHash) {
	xi := store.GetInstance().GetPriorStates().GetXi()
	for _, history := range xi {
		for _, hash := range history {
			output = append(output, hash)
		}
	}
	return output
}

// (12.3) ϑ ∈ ⟦⟦(W, {H})⟧⟧_E available work reports: store.theta

// (12.4) W! ≡ [w S w <− W, S(wx)pS = 0 ∧ wl = {}]

func UpdateImmediatelyAccumulateWorkReports(reports []types.WorkReport) {
	var accumulatable []types.WorkReport
	for _, report := range reports {
		if len(report.Context.Prerequisites) == 0 && len(report.SegmentRootLookup) == 0 {
			accumulatable = append(accumulatable, report)
		}
	}
	store.GetInstance().GetAccumulatedWorkReportsPointer().SetAccumulatedWorkReports(accumulatable)
}

// (12.5) WQ ≡ E([D(w) S w <− W, S(wx)pS > 0 ∨ wl ≠ {}], ©ξ )
func UpdateQueuedWorkReports(reports []types.WorkReport) {
	var reports_with_dependency types.ReadyQueueItem
	for _, report := range reports {
		if len(report.Context.Prerequisites) != 0 || len(report.SegmentRootLookup) != 0 {
			reports_with_dependency = append(reports_with_dependency, GetDependencyFromWorkReport(report))
		}
	}
	store.GetInstance().GetQueuedWorkReportsPointer().SetQueuedWorkReports(QueueEditingFunction(reports_with_dependency, GetAccumulatedHashes()))
}

// (12.6) D(w) ≡ (w, {(wx)p} ∪ K(wl))

func GetDependencyFromWorkReport(report types.WorkReport) (output types.ReadyRecord) {
	output.Report = report
	if len(report.Context.Prerequisites) > 0 {
		for _, hash := range report.Context.Prerequisites {
			output.Dependencies = append(output.Dependencies, types.WorkPackageHash(hash))
		}
	}
	if len(report.SegmentRootLookup) > 0 {
		for _, segment := range report.SegmentRootLookup {
			output.Dependencies = append(output.Dependencies, types.WorkPackageHash(segment.WorkPackageHash))
		}
	}
	return output
}

// (12.7)
// the queue-editing function E,
// which is essentially a mutator function for items such as those of ϑ,
// parameterized by sets of now-accumulated work-package hashes (those in ξ)
// It is used to update queues  of work-reports when some of them are accumulated.
// Functionally, it removes all entries whose work-report's hash is in the set provided as a parameter,
// and removes any dependencies which apper in said set.
func QueueEditingFunction(queue types.ReadyQueueItem, workPackageHashes []types.WorkPackageHash) (newQueue types.ReadyQueueItem) {
	doneSet := make(map[types.WorkPackageHash]bool)
	for _, h := range workPackageHashes {
		doneSet[h] = true
	}
	for _, item := range queue {
		if doneSet[item.Report.PackageSpec.Hash] {
			continue
		}
		var remainingDeps []types.WorkPackageHash
		for _, dep := range item.Dependencies {
			if !doneSet[dep] {
				remainingDeps = append(remainingDeps, dep)
			}
		}
		item.Dependencies = remainingDeps
		newQueue = append(newQueue, item)
	}
	return newQueue
}

// (12.8) Q get accumulatable work reports

func AccumulationPriorityQueue(queue types.ReadyQueueItem) (output []types.WorkReport) {
	var g []types.WorkReport // items_without dependency
	for _, item := range queue {
		if len(item.Dependencies) == 0 {
			g = append(g, item.Report)
		}
	}
	if len(g) == 0 {
		return output
	}
	hashes := ExtractWorkReportHashes(g)
	extra_avaliable_reports := AccumulationPriorityQueue(QueueEditingFunction(queue, hashes))
	output = append(output, extra_avaliable_reports...)
	return output
}

// (12.9) P
// the mapping function P which extracts the corresponding work-package hashes from a set of work-reports

func ExtractWorkReportHashes(workReports []types.WorkReport) (output []types.WorkPackageHash) {
	for _, workReport := range workReports {
		output = append(output, workReport.PackageSpec.Hash)
	}
	return output
}

// (12.10)
// (12.11) W∗ ≡ W! ⌢ Q(q)
// (12.12) q = E(Ïϑm... ⌢ Ïϑ...m ⌢ WQ, P (W!))

func UpdateAccumulatableWorkReports() {
	// Get prior state
	slot_index := store.GetInstance().GetPriorStates().GetTau() % types.TimeSlot(types.EpochLength)
	theta := store.GetInstance().GetPriorStates().GetTheta()

	var candidate_queue types.ReadyQueue
	candidate_queue = append(candidate_queue, theta[slot_index:]...)
	candidate_queue = append(candidate_queue, theta[:slot_index]...)
	candidate_queue = append(candidate_queue, store.GetInstance().GetQueuedWorkReportsPointer().GetQueuedWorkReports())
	var candidate_reports types.ReadyQueueItem
	for _, queue := range candidate_queue {
		for _, item := range queue {
			candidate_reports = append(candidate_reports, item)
		}
	}
	accumulated_hashes := ExtractWorkReportHashes(store.GetInstance().GetAccumulatedWorkReportsPointer().GetAccumulatedWorkReports())
	q := QueueEditingFunction(candidate_reports, accumulated_hashes)
	new_accumulatable_workreports := store.GetInstance().GetAccumulatedWorkReportsPointer().GetAccumulatedWorkReports()
	new_accumulatable_workreports = append(new_accumulatable_workreports, AccumulationPriorityQueue(q)...)
	store.GetInstance().GetAccumulatableWorkReportsPointer().SetAccumulatableWorkReports(new_accumulatable_workreports)

}

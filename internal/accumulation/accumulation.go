package accumulation

import (
	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
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
	store := store.GetInstance()
	m := store.GetIntermediateHeader().Slot
	theta := store.GetPriorStates().GetTheta()

	var candidate_queue types.ReadyQueue
	candidate_queue = append(candidate_queue, theta[m:]...)
	candidate_queue = append(candidate_queue, theta[:m]...)
	WQ := store.GetQueuedWorkReportsPointer().GetQueuedWorkReports()
	candidate_queue = append(candidate_queue, WQ)
	var candidate_reports types.ReadyQueueItem
	for _, queue := range candidate_queue {
		candidate_reports = append(candidate_reports, queue...)
	}
	accumulated_hashes := ExtractWorkReportHashes(store.GetAccumulatedWorkReportsPointer().GetAccumulatedWorkReports())
	q := QueueEditingFunction(candidate_reports, accumulated_hashes)
	new_accumulatable := store.GetAccumulatedWorkReportsPointer().GetAccumulatedWorkReports()
	new_accumulatable = append(new_accumulatable, AccumulationPriorityQueue(q)...)
	store.GetAccumulatableWorkReportsPointer().SetAccumulatableWorkReports(new_accumulatable)
}

//(12.13) U ≡ (d ∈ D⟨NS → A⟩ , i ∈ ⟦K⟧V , q ∈ C⟦H⟧QHC ,
//				x ∈ (NS , NS , NS , D⟨NS → NG⟩)) type.types.PartialStateSet
/*
type PartialStateSet struct {
	ServiceAccounts ServiceAccountState
	ValidatorKeys   ValidatorsData
	Authorizers     AuthQueues
	Privileges      Privileges
}
*/

// (12.14) T ≡ (s ∈ NS , d ∈ NS , a ∈ NB , m ∈ YWT , g ∈ NG) types.DeferredTransfer
/*
type DeferredTransfer struct {
	SenderID   ServiceId `json:"senderid"`
	ReceiverID ServiceId `json:"receiverid"`
	Balance    U64       `json:"balance"`
	Memo       [128]byte `json:"memo"`
	GasLimit   Gas       `json:"gas"`
}
*/

// (12.15) U types.ServiceGasUsedList
// 		   B types.ServiceHashSet
/*
// (12.15) U
type ServiceGasUsedList []ServiceGasUsed

type ServiceGasUsed struct {
	ServiceId ServiceId
	Gas       Gas
}

type ServiceHash struct {
	ServiceId ServiceId
	Hash      OpaqueHash // AccumulationOutput
}

// (12.15) B
// FIXME: Naming issue
type ServiceHashSet map[ServiceHash]struct{}
*/

// (12.18)
/* types.Operand
type Operand struct {
	Hash           WorkPackageHash
	ExportsRoot    ExportsRoot
	AuthorizerHash OpaqueHash
	AuthOutput     ByteSequence
	PayloadHash    OpaqueHash
	Result         WorkExecResult
}
*/

// (12.16) ∆+ outer accumulation function
// (NG, ⟦W⟧, U, D⟨NS → NG⟩) → (N, U, ⟦T⟧, B, U )
// (g, w, o, f )↦ (0, o, [], {}) if i = 0
//
//	(i + j, o′, t∗⌢ t, b∗ ∪ b, u∗⌢ u) o/w
//
// where i = max(NSwS+1) ∶   ∑   ∑     (rg ) ≤ g
//
//							w∈w...i  r∈wr
//	 and (u∗, o∗, t∗, b∗) = ∆∗(o, w...i, f )
//
// and (j, o′, t, b, u) = ∆+(g − ∑u, wi..., o∗, {})
//
//	(s,u)∈u∗
func OuterAccumulation(input OuterAccumulationInput) (output OuterAccumulationOutput) {
	gas_sum := 0
	i := 0
	for idx, report := range input.WorkReports {
		for _, result := range report.Results {
			gas_sum += int(result.AccumulateGas)
		}
		if gas_sum <= int(input.GasLimit) {
			i = idx + 1
		} else {
			break
		}
	}
	if i == 0 {
		output.NumberOfWorkResultsAccumulated = 0
		output.PartialStateSet = input.InitPartialStateSet
		return output
	}
	var parallel_input ParallelizedAccumulationInput
	parallel_input.WorkReports = input.WorkReports[:i]
	parallel_input.PartialStateSet = input.InitPartialStateSet
	parallel_input.AlwaysAccumulateMap = input.ServicesWithFreeAccumulation

	parallel_result := ParallelizedAccumulation(parallel_input)
	remain_gas := input.GasLimit
	for _, gas_use := range parallel_result.ServiceGasUsedList {
		remain_gas -= gas_use.Gas
	}
	var recursive_outer_input OuterAccumulationInput
	recursive_outer_input.GasLimit = remain_gas
	recursive_outer_input.WorkReports = input.WorkReports[i:]
	recursive_outer_input.InitPartialStateSet = parallel_result.PartialStateSet

	recursive_outer_output := OuterAccumulation(recursive_outer_input)
	output.NumberOfWorkResultsAccumulated = types.U64(i) + recursive_outer_output.NumberOfWorkResultsAccumulated
	output.PartialStateSet = recursive_outer_output.PartialStateSet
	output.DeferredTransfers = append(parallel_result.DeferredTransfers, recursive_outer_output.DeferredTransfers...)
	output.ServiceGasUsedList = append(parallel_result.ServiceGasUsedList, recursive_outer_output.ServiceGasUsedList...)
	output.ServiceHashSet = parallel_result.ServiceHashSet

	for key, value := range recursive_outer_output.ServiceHashSet {
		output.ServiceHashSet[key] = value
	}

	return output
}

// (12.17) ∆∗ parallelized accumulation function
// (U, ⟦W⟧, D⟨NS → NG⟩) → (U, ⟦T⟧, B, U )
// (o, w, f ) ↦ (((d ∪ n) ∖ m, i′, q′, x′),Ìt, b, u)
// where:
// s = {rs S w ∈ w, r ∈ wr} ∪ K(f )
// u = [(s, ∆1(o, w, f , s)u) S s <− s]
// b = {(s, b) S s ∈ s, b = ∆1(o, w, f , s)b, b ≠ ∅}
// t = [∆1(o, w, f , s)t S s <− s]
//
//	(d, i, q, (m, a, v, z)) = o
//
// x′ = (∆1(o, w, f , m)o)x
// i′ = (∆1(o, w, f , v)o)i
// q′ = (∆1(o, w, f , a)o)q
// n = ⋃ ({(∆1(o, w, f , s)o)d ∖ K(d ∖ {s})})
//
//	s∈s
//
// m = ⋃ (K(d) ∖ K((∆1(o, w, f , s)o)d))
//
//	s∈s
//
// This formula still have some parts needs conclusion
func ParallelizedAccumulation(input ParallelizedAccumulationInput) (output ParallelizedAccumulationOutput) {
	var service_set map[types.ServiceId]bool
	for _, report := range input.WorkReports {
		for _, reuslt := range report.Results {
			service_set[reuslt.ServiceId] = true
		}
	}
	for serivce_id := range input.AlwaysAccumulateMap {
		service_set[serivce_id] = true
	}
	current_partial_state := input.PartialStateSet
	for service_id := range service_set {
		var single_input SingleServiceAccumulationInput
		single_input.ServiceId = service_id
		single_input.PartialStateSet = current_partial_state
		single_input.WorkReports = input.WorkReports
		single_input.AlwaysAccumulateMap = input.AlwaysAccumulateMap
		single_output := SingleServiceAccumulation(single_input)
		var gas_used types.ServiceGasUsed
		gas_used.ServiceId = service_id
		gas_used.Gas = single_output.GasUsed
		output.ServiceGasUsedList = append(output.ServiceGasUsedList, gas_used)
		if single_output.AccumulationOutput != nil {
			var service_hash types.ServiceHash
			service_hash.ServiceId = service_id
			service_hash.Hash = *single_output.AccumulationOutput
			output.ServiceHashSet[service_hash] = struct{}{}
		}
		for _, defer_transfer := range single_output.DeferredTransfers {
			output.DeferredTransfers = append(output.DeferredTransfers, defer_transfer)
		}
		new_partial_state := single_output.PartialStateSet
		new_partial_state.ServiceAccounts = current_partial_state.ServiceAccounts
		for key, value := range single_output.PartialStateSet.ServiceAccounts {
			if _, ok := new_partial_state.ServiceAccounts[key]; ok {
				delete(new_partial_state.ServiceAccounts, key)
			} else {
				new_partial_state.ServiceAccounts[key] = value
			}
		}
		current_partial_state = new_partial_state
	}
	output.PartialStateSet = current_partial_state
	return output
}

// (12.19) ∆1 single-service accumulation function

// ∆1∶
// (U, ⟦W⟧, D⟨NS → NG⟩, NS ) → o ∈ U , t ∈ ⟦T⟧ ,
//
//					  b ∈ H? , u ∈ NG
//	(o, w, f , s) ↦ ΨA(o, τ ′, s, g, p)
//
// where:
//
//	g = U(fs, 0) + ∑(rg )
//				w∈w,r∈wr,rs=s
//
// p d: rd, e: (ws)e, o:wo,    w <− w, r <− wr, rs = s
//
//	y: ry ,h: (ws)h, a:wa
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (output SingleServiceAccumulationOutput) {
	var operands []types.Operand
	g := max(input.AlwaysAccumulateMap[input.ServiceId], 0)
	for _, report := range input.WorkReports {
		for _, item := range report.Results {
			if item.ServiceId == input.ServiceId {
				g += item.AccumulateGas
				operand := types.Operand{
					Hash:           report.PackageSpec.Hash,
					ExportsRoot:    report.PackageSpec.ExportsRoot,
					AuthorizerHash: report.AuthorizerHash,
					PayloadHash:    item.PayloadHash,
					AuthOutput:     report.AuthOutput,
					Result:         item.Result,
				}
				operands = append(operands, operand)
			}
		}
	}
	pvm_result := PolkaVM.Psi_A(input.PartialStateSet, store.GetInstance().GetPriorStates().GetTau()%types.TimeSlot(types.EpochLength), input.ServiceId, g, operands)
	output.AccumulationOutput = pvm_result.Result
	output.DeferredTransfers = pvm_result.DeferredTransfers
	output.GasUsed = pvm_result.Gas
	output.PartialStateSet = pvm_result.PartialStateSet
	return output
}

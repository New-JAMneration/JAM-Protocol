package accumulation

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (12.1) ξ ∈ ⟦{H}⟧_E: store.Xi

// (12.2) ©ξ ≡ ⋃x∈ξ
// This function extracts all known (past) accumulated WorkPackageHashes.
func GetAccumulatedHashes() (output []types.WorkPackageHash) {
	xi := store.GetInstance().GetPriorStates().GetXi() // Retrieve ξ
	for _, history := range xi {
		output = append(output, history...) // Form ©ξ ≡ union over ξ
	}
	return output
}

// (12.3) ϑ ∈ ⟦⟦(W, {H})⟧⟧_E available work reports: store.theta

// (12.4) W! ≡ [w S w <− W, S(wx)pS = 0 ∧ wl = {}]
// This function identifies and stores work reports that are immediately
// eligible for accumulation, i.e.:
//   - (wx)p == 0: the report has no unresolved prerequisites
//   - wl == {}  : there is no segment root lookup required
//
// These work reports are independent and can be accumulated without waiting.
func UpdateImmediatelyAccumulateWorkReports() {
	intermediateState := store.GetInstance().GetIntermediateStates()
	availableReports := intermediateState.GetAvailableWorkReports()

	var accumulatable_reports []types.WorkReport
	for _, report := range availableReports {
		// Check for no prerequisites and no segment root lookup dependencies
		if len(report.Context.Prerequisites) == 0 && len(report.SegmentRootLookup) == 0 {
			accumulatable_reports = append(accumulatable_reports, report)
		}
	}
	// Store W! — immediately accumulatable work reports
	intermediateState.SetAccumulatedWorkReports(accumulatable_reports)
}

// (12.5) WQ ≡ E([D(w) S w <− W, S(wx)pS > 0 ∨ wl ≠ {}], ©ξ )
// Get all workreport with dependency, and store in QueuedWorkReports
func UpdateQueuedWorkReports() {
	intermediateState := store.GetInstance().GetIntermediateStates()
	availableReports := intermediateState.GetAvailableWorkReports()
	var reports_with_dependency types.ReadyQueueItem
	for _, report := range availableReports {
		if len(report.Context.Prerequisites) != 0 || len(report.SegmentRootLookup) != 0 {
			// D(w): extract the dependency structure from report
			reports_with_dependency = append(reports_with_dependency, GetDependencyFromWorkReport(report))
		}
	}
	// E(..., ©ξ): perform dependency resolution and ordering
	work_reports_queue := QueueEditingFunction(reports_with_dependency, GetAccumulatedHashes())
	// Store WQ — queued reports awaiting prerequisite satisfaction
	intermediateState.SetQueuedWorkReports(work_reports_queue)
}

// (12.6) D(w) ≡ (w, {(wx)p} ∪ K(wl))
// Extract all dependencies from single work report
func GetDependencyFromWorkReport(report types.WorkReport) (output types.ReadyRecord) {
	output.Report = report
	// Add all explicit prerequisites (wx)p to the dependency list
	for _, hash := range report.Context.Prerequisites {
		output.Dependencies = append(output.Dependencies, types.WorkPackageHash(hash))
	}

	// Add all work package hashes found in the segment root lookup (i.e., K(wl))
	for _, segment := range report.SegmentRootLookup {
		output.Dependencies = append(output.Dependencies, types.WorkPackageHash(segment.WorkPackageHash))
	}
	return output
}

// (12.7)
//
//	(⟦(W, {H})⟧, {H}) → ⟦(W, {H})⟧
//
// E∶   (r, x) ↦ (w, d ∖ x) W     (w, d) <− r,
//
//	(ws)h ~∈ x
//	  - For each work report (w) with dependency set (d)
//	  - If w’s own hash (ws)h is *not* in x (i.e. not yet accumulated),
//	  - Remove from d any dependencies already present in x (i.e., prune known satisfied deps)
//	  - Return the pruned ReadyQueueItem (w, d \ x)
func QueueEditingFunction(r types.ReadyQueueItem, x []types.WorkPackageHash) (newQueue types.ReadyQueueItem) {
	finished_report_hashes := make(map[types.WorkPackageHash]bool)
	for _, h := range x {
		finished_report_hashes[h] = true
	}
	for _, item := range r {
		// If the report itself is already accumulated, skip it, remove from queue
		if exist, _ := finished_report_hashes[item.Report.PackageSpec.Hash]; exist {
			continue
		}
		// Otherwise, filter its dependencies: keep only those NOT in the finished set
		var remainingDeps []types.WorkPackageHash
		for _, dep := range item.Dependencies {
			if exist, _ := finished_report_hashes[dep]; !exist {
				remainingDeps = append(remainingDeps, dep)
			}
		}

		// Update the item with pruned dependencies and keep it in the new queue
		item.Dependencies = remainingDeps
		newQueue = append(newQueue, item)
	}
	return newQueue
}

// (12.8) Q get accumulatable work reports

func AccumulationPriorityQueue(r types.ReadyQueueItem) (output []types.WorkReport) {
	var g []types.WorkReport // g := items with no remaining dependencies

	// Collect all reports that are ready for accumulation (i.e., dependencies resolved)
	for _, item := range r {
		if len(item.Dependencies) == 0 {
			g = append(g, item.Report)
		}
	}

	// If no items are currently eligible, return empty result
	if len(g) == 0 {
		return output
	}

	output = g
	// Recursively prune the queue and resolve additional eligible reports
	hashes := ExtractWorkReportHashes(g)
	recursivelyReadyReports := AccumulationPriorityQueue(QueueEditingFunction(r, hashes))
	output = append(output, recursivelyReadyReports...)
	return output
}

// (12.9) P
// the mapping function P which extracts the corresponding work-package hashes from a set of work-reports

func ExtractWorkReportHashes(w []types.WorkReport) (output []types.WorkPackageHash) {
	for _, workReport := range w {
		output = append(output, workReport.PackageSpec.Hash)
	}
	return output
}

// (12.10) let m = Ht mod E(12.10)
// (12.11) W∗ ≡ W! ⌢ Q(q)
// (12.12) q = E(ϑm... ⌢ ϑ...m ⌢ WQ, P (W!))
func UpdateAccumulatableWorkReports() {
	store := store.GetInstance()

	// (12.10) Get current slot index 'm'
	slot := store.GetLatestBlock().Header.Slot
	E := types.EpochLength
	m := int(slot) % E

	// Get θ: the available work reports (ϑ)
	theta := store.GetPriorStates().GetTheta()
	WQ := store.GetIntermediateStates().GetQueuedWorkReports()
	Wbang := store.GetIntermediateStates().GetAccumulatedWorkReports()

	// E(ϑm... ⌢ ϑ...m ⌢ WQ)
	var composedQueue types.ReadyQueueItem
	for _, record := range theta[m:] {
		composedQueue = append(composedQueue, record...)
	}

	for _, record := range theta[:m] {
		composedQueue = append(composedQueue, record...)
	}

	composedQueue = append(composedQueue, WQ...)

	accumulatedHashes := ExtractWorkReportHashes(Wbang)

	// (12.12) Compute q = E(..., P(W!))
	// Use accumulated hashes from W! to prune dependencies
	q := QueueEditingFunction(composedQueue, accumulatedHashes)

	// (12.11) W* ≡ W! ⌢ Q(q)
	// Reconstruct W* by appending newly-resolved reports to previously accumulated W!
	Wstar := []types.WorkReport{}
	Wstar = append(Wstar, Wbang...)
	Wstar = append(Wstar, AccumulationPriorityQueue(q)...)

	// Update W*
	store.GetIntermediateStates().SetAccumulatableWorkReports(Wstar)
}

// (12.16) ∆+ outer accumulation function
// (NG, ⟦W⟧, U, D⟨NS → NG⟩) → (N, U, ⟦T⟧, B, U )
// (g, w, o, f )↦ (0, o, [], {}, []) if i = 0
//
//	(i + j, o′, t∗⌢ t, b∗ ∪ b, u∗⌢ u) o/w
//
// where i = max(NSwS+1) ∶   ∑   ∑     (rg ) ≤ g
//
//							w∈w...i  r∈wr
//	 and (o∗, t∗, b∗, u∗) = ∆∗(o, w...i, f )
//
// and (j, o′, t, b, u) = ∆+(g − ∑u, wi..., o∗, {})
//
//	(s,u)∈u∗
func OuterAccumulation(input OuterAccumulationInput) (output OuterAccumulationOutput, err error) {
	gas_sum := 0
	i := 0

	// Determine the maximal prefix of reports that fits within the gas limit
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
		return output, nil
	}

	// Accumulate the first i reports in parallel across services (∆)
	var parallel_input ParallelizedAccumulationInput
	parallel_input.WorkReports = input.WorkReports[:i]
	parallel_input.PartialStateSet = input.InitPartialStateSet
	parallel_input.AlwaysAccumulateMap = input.ServicesWithFreeAccumulation

	parallel_result, err := ParallelizedAccumulation(parallel_input)
	if err != nil {
		return output, fmt.Errorf("parallel accumulation failed: %w", err)
	}

	// Recurse on the remaining reports with the remaining gas
	remain_gas := input.GasLimit
	for _, gas_use := range parallel_result.ServiceGasUsedList {
		remain_gas -= gas_use.Gas
	}
	var recursive_outer_input OuterAccumulationInput
	recursive_outer_input.GasLimit = remain_gas
	recursive_outer_input.WorkReports = input.WorkReports[i:]
	recursive_outer_input.InitPartialStateSet = parallel_result.PartialStateSet

	recursive_outer_output, err := OuterAccumulation(recursive_outer_input)
	if err != nil {
		return output, fmt.Errorf("recursive accumulation failed: %w", err)
	}
	// Combine results from this batch and the recursive tail
	output.NumberOfWorkResultsAccumulated = types.U64(i) + recursive_outer_output.NumberOfWorkResultsAccumulated
	output.PartialStateSet = recursive_outer_output.PartialStateSet
	output.DeferredTransfers = append(parallel_result.DeferredTransfers, recursive_outer_output.DeferredTransfers...)
	output.ServiceGasUsedList = append(parallel_result.ServiceGasUsedList, recursive_outer_output.ServiceGasUsedList...)
	output.AccumulatedServiceOutput = parallel_result.AccumulatedServiceOutput

	for key, value := range recursive_outer_output.AccumulatedServiceOutput {
		output.AccumulatedServiceOutput[key] = value
	}

	return output, nil
}

// Helper function to compute the set s for(12.17)
// s = {s S s ∈ (rs S w ∈ w, r ∈ wr)} ∪ K(f)
func set_s(W []types.WorkReport, M types.AlwaysAccumulateMap) map[types.ServiceId]bool {
	s := make(map[types.ServiceId]bool)
	// {rs S w ∈ w, r ∈ wr}
	for _, w := range W {
		for _, r := range w.Results {
			s[r.ServiceId] = true
		}
	}
	// K(f)
	for service_id := range M {
		s[service_id] = true
	}
	return s
}

// Merge three maps, d, n, m
func merge(d, n, m types.ServiceAccountState) types.ServiceAccountState {
	result := make(types.ServiceAccountState)
	// Copy d
	for key, value := range d {
		result[key] = value
	}
	// Merge n, overwriting any keys from d
	for key, value := range n {
		result[key] = value
	}
	// Remove keys present in m
	for key := range m {
		delete(result, key)
	}
	return result
}

// (12.17) ∆∗ parallelized accumulation function
func ParallelizedAccumulation(input ParallelizedAccumulationInput) (output ParallelizedAccumulationOutput, err error) {
	// Initialize output maps
	output.AccumulatedServiceOutput = make(map[types.AccumulatedServiceHash]bool)

	// s = {rs S w ∈ w, r ∈ wd} ∪ K(f)
	s := set_s(input.WorkReports, input.AlwaysAccumulateMap)

	// Needed notations from partial state set
	d := input.PartialStateSet.ServiceAccounts
	var t []types.DeferredTransfer
	n := make(types.ServiceAccountState)
	m := make(types.ServiceAccountState)

	var single_input SingleServiceAccumulationInput
	single_input.PartialStateSet = input.PartialStateSet         // e
	single_input.WorkReports = input.WorkReports                 // w
	single_input.AlwaysAccumulateMap = input.AlwaysAccumulateMap // f

	serviceResultCache := make(map[types.ServiceId]SingleServiceAccumulationOutput)
	// Helper to run single service accumulation for a given service ID
	runSingleReplaceService := func(serviceId types.ServiceId) (SingleServiceAccumulationOutput, error) {
		if result, exists := serviceResultCache[serviceId]; exists {
			return result, nil
		}
		// Replace service ID in input
		single_input.ServiceId = serviceId
		single_output, err := SingleServiceAccumulation(single_input)
		if err != nil {
			serviceResultCache[serviceId] = single_output
		}
		return single_output, err
	}

	// collective service blobs output
	p := types.ServiceBlobs{}

	// ∀s ∈ s ∶ run ∆1(e, w, f, s)
	for service_id := range s {
		single_output, err := runSingleReplaceService(service_id)
		if err != nil {
			return output, fmt.Errorf("single replace service failed: %w", err)
		}

		// u = [(s, ∆1(e, w, f, s)u) S s <− s]
		var u types.ServiceGasUsed
		u.ServiceId = service_id
		u.Gas = single_output.GasUsed
		output.ServiceGasUsedList = append(output.ServiceGasUsedList, u)

		// b = {(s, b) S s ∈ s, b = ∆1(e, w, f , s)b, b ≠ ∅}
		if single_output.AccumulationOutput != nil {
			var b types.AccumulatedServiceHash
			b.ServiceId = service_id
			b.Hash = *single_output.AccumulationOutput
			output.AccumulatedServiceOutput[b] = true
		}

		// t = [∆1(e, w, f, s)t S s <− s]
		t = append(t, single_output.DeferredTransfers...)

		single_outout_d := single_output.PartialStateSet.ServiceAccounts

		// n = ⋃ ((∆(s)e)d ∖ K(d ∖ { s }))
		// n = union of (d_prime without keys in d except service_id)
		for key, value := range single_outout_d {
			if key == service_id {
				n[key] = value
			} else if _, exists := d[key]; !exists {
				n[key] = value
			} else {
				// exclude part: ∖ K(d ∖ {s})
			}
		}

		// m = ⋃ (K(d) ∖ K((∆(s)e)d))
		// m = union of (keys in d but missing in d_prime)
		d_exclude_single_output_d := make(types.ServiceAccountState)
		for key := range d {
			if _, exists := single_outout_d[key]; !exists {
				d_exclude_single_output_d[key] = d[key]
			} else {
				// exclude part: ∖ K((∆(s)e)d)
			}
		}

		// add to m
		for key, value := range d_exclude_single_output_d {
			m[key] = value
		}

		// p =  ⋃∆1(e, w, f, s)p
		p = append(p, single_output.ServiceBlobs...)
	}

	// x′ = (∆1(e, w, f, m)o)x
	// (m′, a∗, v∗, z′) = (∆1(e, w, f, m)o)(m,a,v,z)
	single_output, err := runSingleReplaceService(input.PartialStateSet.Bless)
	if err != nil {
		return output, fmt.Errorf("single replace service failed: %w", err)
	}
	m_prime := single_output.PartialStateSet.Bless
	a_star := single_output.PartialStateSet.Assign
	v_star := single_output.PartialStateSet.Designate
	z_prime := single_output.PartialStateSet.AlwaysAccum
	a_prime := make(types.ServiceIdList, types.CoresCount)

	// ∀c ∈ NC ∶ a′c = ((∆1(o, w, f, a∗c )o)a)c
	if len(a_star) != types.CoresCount {
		return output, fmt.Errorf("service assign length mismatch: expected %d, got %d", types.CoresCount, len(a_star))
	}
	for c := range types.CoresCount {
		single_output, err := runSingleReplaceService(a_star[c])
		if err != nil {
			return output, fmt.Errorf("single replace service failed: %w", err)
		}
		a_prime[c] = single_output.PartialStateSet.Assign[c]
	}

	// v′ = (∆1(o, w, f , v∗)o)v
	single_output, err = runSingleReplaceService(v_star)
	if err != nil {
		return output, fmt.Errorf("single replace service failed: %w", err)
	}
	v_prime := single_output.PartialStateSet.Designate

	// i′ = (∆1(o, w, f, v)o)i
	var i_prime types.ValidatorsData
	{
		single_output, err := runSingleReplaceService(input.PartialStateSet.Designate)
		if err != nil {
			return output, fmt.Errorf("single replace service failed: %w", err)
		}
		i_prime = single_output.PartialStateSet.ValidatorKeys
	}

	// ∀c ∈ NC ∶ q′c = (∆1(o, w, f , ac)o)q
	var q_prime types.AuthQueues
	{
		q_prime = make(types.AuthQueues, types.CoresCount)
		for c, service_id := range input.PartialStateSet.Assign {
			single_output, err := runSingleReplaceService(service_id)
			if err != nil {
				return output, fmt.Errorf("single replace service failed: %w", err)
			}
			q_prime[c] = single_output.PartialStateSet.Authorizers[c]
		}
	}

	// d′ = P ((d ∪ n) ∖ m, ⋃ ∆(s)p)
	//	    		         s∈s
	// d_prime, err = Provide(merge(d, n, m), p)
	d_prime, err := Provide(merge(d, n, m), p)
	if err != nil {
		return output, fmt.Errorf("failed to provide service accounts: %w", err)
	}

	// Set posterior state
	{
		store := store.GetInstance()
		store.GetPosteriorStates().SetChi(types.Privileges{
			Bless:       m_prime,
			Assign:      a_prime,
			Designate:   v_prime,
			AlwaysAccum: z_prime,
		})
		store.GetPosteriorStates().SetVarphi(q_prime)
		store.GetPosteriorStates().SetIota(i_prime)
	}

	// new partial state set: (d′, i′, q′, m′, a′, v′, z′)
	var new_partial_state types.PartialStateSet
	{
		new_partial_state.ServiceAccounts = d_prime
		new_partial_state.ValidatorKeys = i_prime
		new_partial_state.Authorizers = q_prime
		new_partial_state.Bless = m_prime
		new_partial_state.Assign = a_prime
		new_partial_state.Designate = v_prime
		new_partial_state.AlwaysAccum = z_prime
	}
	output.PartialStateSet = new_partial_state
	output.DeferredTransfers = t
	return output, nil
}

// (12.20) ∆1 single-service accumulation function

// ∆1∶
// (U, ⟦W⟧, D⟨NS → NG⟩, NS ) → o ∈ U , t ∈ ⟦T⟧ ,
//
//					  b ∈ H? , u ∈ NG p ∈ {NS , Y}
//	(o, w, f, s) ↦ ΨA(o, τ ′, s, g, i)
//
// where:
//
//	g = U(fs, 0) + ∑(rg )
//				w∈w,r∈wr,rs=s
//
// i d: rd, e: (ws)e, o:wo,    w <− w, r <− wr, rs = s
//
//	y: ry ,h: (ws)h, a:wa
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (output SingleServiceAccumulationOutput, err error) {
	var operands []types.Operand // all operand inputs for Ψₐ
	// U(fs, 0)
	g := types.Gas(0)
	if preset, ok := input.AlwaysAccumulateMap[input.ServiceId]; ok {
		g = preset
	}
	for _, report := range input.WorkReports {
		for _, item := range report.Results {
			if item.ServiceId == input.ServiceId {
				//    ∑(rg )
				// w∈w,r∈wr,rs=s
				g += item.AccumulateGas

				// p d: rd, e: (ws)e, o:wo
				//	y: ry ,h: (ws)h, a:wa
				operand := types.Operand{
					Hash:           report.PackageSpec.Hash,        // h: (ws)h — work package hash，
					ExportsRoot:    report.PackageSpec.ExportsRoot, // e: (ws)e — exports root
					AuthorizerHash: report.AuthorizerHash,          // a: wa — authorizer hash
					PayloadHash:    item.PayloadHash,               // y: ry — result payload hash
					AuthOutput:     report.AuthOutput,              // o: wo
					Result:         item.Result,                    // d: rd
					GasLimit:       item.AccumulateGas,             // g: rg
				}
				operands = append(operands, operand)
			}
		}
	}

	// τ′: Posterior validator state used by Ψₐ
	tau_prime := store.GetInstance().GetPosteriorStates().GetTau()
	eta0 := store.GetInstance().GetPosteriorStates().GetState().Eta[0]
	pvm_result := PVM.Psi_A(input.PartialStateSet, tau_prime, input.ServiceId, g, operands, eta0)
	output.AccumulationOutput = pvm_result.Result
	output.DeferredTransfers = pvm_result.DeferredTransfers
	output.GasUsed = pvm_result.Gas
	output.PartialStateSet = pvm_result.PartialStateSet
	output.ServiceBlobs = pvm_result.ServiceBlobs
	return output, nil
}

func ProcessAccumulation() error {
	// Compute W!
	UpdateImmediatelyAccumulateWorkReports()

	// Compute WQ
	UpdateQueuedWorkReports()

	// Compute W*
	UpdateAccumulatableWorkReports()
	return nil
}

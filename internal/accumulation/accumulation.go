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

	var accumulatableReports []types.WorkReport
	for _, report := range availableReports {
		// Check for no prerequisites and no segment root lookup dependencies
		if len(report.Context.Prerequisites) == 0 && len(report.SegmentRootLookup) == 0 {
			accumulatableReports = append(accumulatableReports, report)
		}
	}
	// Store W! — immediately accumulatable work reports
	intermediateState.SetAccumulatedWorkReports(accumulatableReports)
}

// (12.5) WQ ≡ E([D(w) S w <− W, S(wx)pS > 0 ∨ wl ≠ {}], ©ξ )
// Get all workreport with dependency, and store in QueuedWorkReports
func UpdateQueuedWorkReports() {
	intermediateState := store.GetInstance().GetIntermediateStates()
	availableReports := intermediateState.GetAvailableWorkReports()
	var reportsWithDependency types.ReadyQueueItem
	for _, report := range availableReports {
		if len(report.Context.Prerequisites) != 0 || len(report.SegmentRootLookup) != 0 {
			// D(w): extract the dependency structure from report
			reportsWithDependency = append(reportsWithDependency, GetDependencyFromWorkReport(report))
		}
	}
	// E(..., ©ξ): perform dependency resolution and ordering
	workReportsQueue := QueueEditingFunction(reportsWithDependency, GetAccumulatedHashes())
	// Store WQ — queued reports awaiting prerequisite satisfaction
	intermediateState.SetQueuedWorkReports(workReportsQueue)
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
	finishedReportHashes := make(map[types.WorkPackageHash]bool)
	for _, h := range x {
		finishedReportHashes[h] = true
	}
	for _, item := range r {
		// If the report itself is already accumulated, skip it, remove from queue
		if _, exist := finishedReportHashes[item.Report.PackageSpec.Hash]; exist {
			continue
		}
		// Otherwise, filter its dependencies: keep only those NOT in the finished set
		var remainingDeps []types.WorkPackageHash
		for _, dep := range item.Dependencies {
			if _, exist := finishedReportHashes[dep]; !exist {
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
func OuterAccumulation(input OuterAccumulationInput) (output OuterAccumulationOutput, err error) {
	// input parameters
	g := input.GasLimit
	w := input.WorkReports
	e := input.InitPartialStateSet
	f := input.ServicesWithFreeAccumulation

	gasSum := 0
	i := 0
	t := input.DeferredTransfers

	// Determine the maximal prefix of reports that fits within the gas limit
	for idx, report := range w {
		for _, result := range report.Results {
			gasSum += int(result.AccumulateGas)
		}
		if gasSum <= int(g) {
			i = idx + 1
		} else {
			break
		}
	}
	// n = |t| + i + |f|
	n := len(t) + i + len(f)
	if n == 0 {
		output.NumberOfWorkResultsAccumulated = 0
		output.PartialStateSet = e
		output.AccumulatedServiceOutput = make(map[types.AccumulatedServiceHash]bool)
		output.ServiceGasUsedList = []types.ServiceGasUsed{}
		return output, nil
	}

	// Accumulate the first i reports in parallel across services (∆)
	//(e∗, t∗, b∗, u∗) = ∆∗(e, t, r...i, f)
	var parallel_input ParallelizedAccumulationInput
	parallel_input.PartialStateSet = e
	parallel_input.DeferredTransfers = t
	parallel_input.WorkReports = w[:i]
	parallel_input.AlwaysAccumulateMap = f

	parallel_result, err := ParallelizedAccumulation(parallel_input)
	e_star := parallel_result.PartialStateSet
	t_star := parallel_result.DeferredTransfers
	b_star := parallel_result.AccumulatedServiceOutput
	u_star := parallel_result.ServiceGasUsedList
	if err != nil {
		return output, fmt.Errorf("parallel accumulation failed: %w", err)
	}

	// Recurse on the remaining reports with the remaining gas
	// (j, e′, b, u) = ∆+(g∗ − ∑(s,u)∈u∗(u), t∗, ri..., e∗, {})
	g_star := input.GasLimit
	for _, DeferredTransfer := range t {
		g_star += DeferredTransfer.GasLimit
	}

	gas_limit_for_recursion := g_star
	for _, u := range u_star {
		gas_limit_for_recursion -= u.Gas
	}
	var recursive_outer_input OuterAccumulationInput
	recursive_outer_input.GasLimit = gas_limit_for_recursion
	recursive_outer_input.DeferredTransfers = t_star
	recursive_outer_input.WorkReports = w[i:]
	recursive_outer_input.InitPartialStateSet = e_star
	recursive_outer_input.ServicesWithFreeAccumulation = make(map[types.ServiceId]types.Gas) // {}
	recursive_outer_output, err := OuterAccumulation(recursive_outer_input)
	if err != nil {
		return output, fmt.Errorf("recursive accumulation failed: %w", err)
	}
	j := recursive_outer_output.NumberOfWorkResultsAccumulated
	e_prime := recursive_outer_output.PartialStateSet
	b := recursive_outer_output.AccumulatedServiceOutput
	u := recursive_outer_output.ServiceGasUsedList
	// Combine results from this batch and the recursive tail
	// (i + j, e′, b∗ ∪ b, u∗⌢ u)
	{
		output.NumberOfWorkResultsAccumulated = types.U64(i) + j
		output.PartialStateSet = e_prime // need to set post state?
		// merge b_star and b
		mergedAccumulatedServiceOutput := make(map[types.AccumulatedServiceHash]bool)
		for key, value := range b_star {
			mergedAccumulatedServiceOutput[key] = value
		}
		for key, value := range b {
			mergedAccumulatedServiceOutput[key] = value
		}
		output.AccumulatedServiceOutput = mergedAccumulatedServiceOutput
		output.ServiceGasUsedList = append(u_star, u...)
	}

	return output, nil
}

// (12.20)
func R[T comparable](o, a, b T) T {
	if a == o {
		return b
	} else {
		return a
	}
}

// Helper function to compute the set s for(12.17)
// s = {s S s ∈ (rs S r ∈ r, d ∈ rd)} ∪ K(f) ∪ {td S t ∈ t}
func set_s(r []types.WorkReport, f types.AlwaysAccumulateMap, t []types.DeferredTransfer) map[types.ServiceId]bool {
	s := make(map[types.ServiceId]bool)
	// {rs S r ∈ r, d ∈ rd}
	for _, w := range r {
		for _, r := range w.Results {
			s[r.ServiceId] = true // rd
		}
	}

	// K(f)
	for service_id := range f {
		s[service_id] = true
	}

	// td S t ∈ t
	for _, deferred_transfer := range t {
		s[deferred_transfer.ReceiverID] = true // td
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

// Parallelize parts and partial state modification needs confirm what is the correct way to process
func ParallelizedAccumulation(input ParallelizedAccumulationInput) (output ParallelizedAccumulationOutput, err error) {

	// s = {s S s ∈ (rs S w ∈ w, r ∈ wr)} ∪ K(f) ∪ {td S t ∈ t}
	s := set_s(input.WorkReports, input.AlwaysAccumulateMap, input.DeferredTransfers)
	b := make(map[types.AccumulatedServiceHash]bool)
	u := make(types.ServiceGasUsedList, 0)

	// Needed notations from partial state set
	e := input.PartialStateSet
	t := input.DeferredTransfers
	r := input.WorkReports
	f := input.AlwaysAccumulateMap
	// d, a from input partial state set
	d := input.PartialStateSet.ServiceAccounts
	a := input.PartialStateSet.Assign

	var t_prime []types.DeferredTransfer

	// maps for collecting service account state changes and update d_prime
	n := make(types.ServiceAccountState)
	m := make(types.ServiceAccountState)

	var single_input SingleServiceAccumulationInput
	single_input.PartialStateSet = e
	single_input.DeferredTransfers = t
	single_input.WorkReports = r
	single_input.AlwaysAccumulateMap = f

	// Helper to run single service accumulation for a given service ID
	// ∆(s) ≡ ∆1(e, t, r, f, s)
	runSingleReplaceService := func(s types.ServiceId) (SingleServiceAccumulationOutput, error) {
		single_input.ServiceId = s
		single_output, err := SingleServiceAccumulation(single_input)
		return single_output, err
	}

	// p: output service blobs collection
	var p types.ServiceBlobs
	for service_id := range s {
		single_output, err := runSingleReplaceService(service_id)
		if err != nil {
			fmt.Println("SingleServiceAccumulation failed:", err)
		}
		// u = [(s, ∆(s)u) S s <− s]
		var gas_use types.ServiceGasUsed
		gas_use.ServiceId = service_id
		gas_use.Gas = single_output.GasUsed
		u = append(u, gas_use)

		// b = {(s, b) S s ∈ s, b = ∆(s)y, b ≠ ∅}
		if single_output.AccumulationOutput != nil {
			var service_hash types.AccumulatedServiceHash
			service_hash.ServiceId = service_id
			service_hash.Hash = *single_output.AccumulationOutput
			b[service_hash] = true
		}

		// t = [∆(s)t S s <− s]
		for _, deferred_transfer := range single_output.DeferredTransfers {
			t_prime = append(t_prime, deferred_transfer)
		}

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
		// collect blobs updates
		p = append(p, single_output.ServiceBlobs...)
	}

	single_output, err := runSingleReplaceService(input.PartialStateSet.Bless)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for bless failed: %w", err)
	}
	// e∗ = ∆(m)e
	e_star := single_output.PartialStateSet
	// m′, z′ = e∗(m, z)
	m_prime := e_star.Bless
	z_prime := e_star.AlwaysAccum

	// ∀c ∈ NC ∶ a′c = R(ac, (e∗a)c, ((∆(ac)e)a)c)
	a_prime := make(types.ServiceIdList, types.CoresCount)
	if len(a) != types.CoresCount {
		return output, fmt.Errorf("input.PartialStateSet.Assign length does not match types.CoresCount")
	}
	for c := range types.CoresCount {
		single_output, err := runSingleReplaceService(a[c])
		if err != nil {
			return output, fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
		}
		a_prime[c] = R(a[c], e_star.Assign[c], single_output.PartialStateSet.Assign[c])
	}

	// v' = R(v, e∗v , (∆(v)e)v )
	var v_prime, r_prime types.ServiceId
	single_output, err = runSingleReplaceService(input.PartialStateSet.Designate)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for designate failed: %w", err)
	}
	v_prime = R(input.PartialStateSet.Designate, e_star.Designate, single_output.PartialStateSet.Designate)

	// r′ = R(r, e∗r , (∆(r)e)r)
	single_output, err = runSingleReplaceService(input.PartialStateSet.CreateAcct)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for createacct failed: %w", err)
	}
	r_prime = R(input.PartialStateSet.CreateAcct, e_star.CreateAcct, single_output.PartialStateSet.CreateAcct)

	// i′ = (∆(v)e)i
	var i_prime types.ValidatorsData
	{
		single_output, err := runSingleReplaceService(input.PartialStateSet.Designate)
		if err != nil {
			return output, fmt.Errorf("single service accumulation for designate failed: %w", err)

		}
		i_prime = single_output.PartialStateSet.ValidatorKeys
	}

	// ∀c ∈ NC ∶ q′c = ((∆(ac)e)q)c
	var q_prime types.AuthQueues
	{

		q_prime = make(types.AuthQueues, types.CoresCount)
		if len(input.PartialStateSet.Assign) != types.CoresCount {
			fmt.Println("Warning: input.PartialStateSet.Assign length does not match types.CoresCount")
		}
		for c, service_id := range input.PartialStateSet.Assign {
			single_output, err := runSingleReplaceService(service_id)
			if err != nil {
				return output, fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
			}
			q_prime[c] = single_output.PartialStateSet.Authorizers[c]
		}

	}

	// (d ∪ n) ∖ m
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
			CreateAcct:  r_prime,
			AlwaysAccum: z_prime,
		})
		store.GetPosteriorStates().SetVarphi(q_prime)
		store.GetPosteriorStates().SetIota(i_prime)
		store.GetPosteriorStates().SetDelta(d_prime)
		// Do we need t_prime in posterior state?
	}
	// new partial state set: (d′, i′, q′, m′, a′, v′, r′, z′)
	// Set output ((d′, i′, q′, m′, a′, v′, r′, z′), t′, b, u)
	var new_partial_state types.PartialStateSet
	{
		new_partial_state.ServiceAccounts = d_prime
		new_partial_state.ValidatorKeys = i_prime
		new_partial_state.Authorizers = q_prime
		new_partial_state.Bless = m_prime
		new_partial_state.Assign = a_prime
		new_partial_state.Designate = v_prime
		new_partial_state.CreateAcct = r_prime
		new_partial_state.AlwaysAccum = z_prime
	}
	output.PartialStateSet = new_partial_state
	output.DeferredTransfers = t_prime
	output.AccumulatedServiceOutput = b
	output.ServiceGasUsedList = u
	return output, nil
}

// (12.20) ∆1 single-service accumulation function
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (output SingleServiceAccumulationOutput, err error) {
	e := input.PartialStateSet       // e: PartialStateSet
	t := input.DeferredTransfers     // t: DeferredTransfers
	r := input.WorkReports           // r: WorkReports
	f := input.AlwaysAccumulateMap   // f: AlwaysAccumulateMap
	s := input.ServiceId             // s: ServiceId
	var i_T []types.Operand          // all operand inputs for Ψₐ
	var i_U []types.DeferredTransfer // all deferred transfers for Ψₐ

	// U(fs, 0)
	g := types.Gas(0)
	if preset, ok := f[input.ServiceId]; ok {
		g = preset
	}

	// iT: all accumulate work result operands for service s
	for _, r := range r {
		for _, d := range r.Results {
			if d.ServiceId == s {
				//    ∑(rg )
				// w∈w,r∈wr,rs=s
				g += d.AccumulateGas
				// Construct operand
				operand := types.Operand{
					PayloadHash:    d.PayloadHash,             // l: dl
					GasLimit:       d.AccumulateGas,           // g: dg
					Result:         d.Result,                  // y: dy
					AuthOutput:     r.AuthOutput,              // t: rt
					Hash:           r.PackageSpec.Hash,        // h: (rs)p — work package hash，
					ExportsRoot:    r.PackageSpec.ExportsRoot, // e: (rs)e — exports root
					AuthorizerHash: r.AuthorizerHash,          // a: ra — authorizer hash
				}
				i_T = append(i_T, operand)
			}
		}
	}

	// iU: all deferred transfers for service s
	for _, deferred_transfer := range t {
		if deferred_transfer.ReceiverID == input.ServiceId {
			i_U = append(i_U, deferred_transfer)
			g += deferred_transfer.GasLimit
		}
	}

	//  iT ⌢ iU
	var pvm_items []types.OperandOrDeferredTransfer
	for _, operand := range i_T {
		pvm_items = append(pvm_items, types.OperandOrDeferredTransfer{Operand: &operand, DeferredTransfer: nil})
	}
	for _, deferred_transfer := range i_U {
		pvm_items = append(pvm_items, types.OperandOrDeferredTransfer{Operand: nil, DeferredTransfer: &deferred_transfer})
	}

	// τ′: Posterior validator state used by Ψₐ
	tau_prime := store.GetInstance().GetPosteriorStates().GetTau()

	// η0: entropy used by Ψₐ
	eta0 := store.GetInstance().GetPosteriorStates().GetState().Eta[0]

	// ΨA(e, τ′, s, g, iT ⌢ iU )
	pvm_result := PVM.Psi_A(e, tau_prime, s, g, pvm_items, eta0)

	// Collect PVM results as output
	{
		output.AccumulationOutput = pvm_result.Result
		output.DeferredTransfers = pvm_result.DeferredTransfers
		output.GasUsed = pvm_result.Gas
		output.PartialStateSet = pvm_result.PartialStateSet
		output.ServiceBlobs = pvm_result.ServiceBlobs
	}
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

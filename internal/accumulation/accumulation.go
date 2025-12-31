package accumulation

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"golang.org/x/sync/singleflight"
)

// (12.1) ξ ∈ ⟦{H}⟧_E: store.Xi

// (12.2) ©ξ ≡ ⋃x∈ξ
// This function extracts all known (past) accumulated WorkPackageHashes.
func GetAccumulatedHashes() (output []types.WorkPackageHash) {
	xi := store.GetInstance().GetPriorStates().GetXi() // Retrieve ξ

	// Pre-calculate total size to avoid multiple memory reallocations
	totalSize := 0
	for _, history := range xi {
		totalSize += len(history)
	}

	output = make([]types.WorkPackageHash, 0, totalSize)

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
	// Pre-allocate with exact capacity to avoid memory reallocations
	output = make([]types.WorkPackageHash, 0, len(w))
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
	// Pre-calculate total capacity for composedQueue to avoid multiple reallocations
	composedQueueCapacity := len(WQ)
	for _, record := range theta[m:] {
		composedQueueCapacity += len(record)
	}
	for _, record := range theta[:m] {
		composedQueueCapacity += len(record)
	}
	composedQueue := make(types.ReadyQueueItem, 0, composedQueueCapacity)

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
	WStar := []types.WorkReport{}
	WStar = append(WStar, Wbang...)
	WStar = append(WStar, AccumulationPriorityQueue(q)...)

	// Update W*
	store.GetIntermediateStates().SetAccumulatableWorkReports(WStar)
}

// (12.16) ∆+ outer accumulation function
func OuterAccumulation(input OuterAccumulationInput) (output OuterAccumulationOutput, err error) {
	// input parameters
	g := input.GasLimit
	t := input.DeferredTransfers
	r := input.WorkReports
	e := input.InitPartialStateSet
	f := input.ServicesWithFreeAccumulation

	gasSum := 0
	i := 0

	// Determine the maximal prefix of reports that fits within the gas limit
	for idx, report := range r {
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
	var parallelInput ParallelizedAccumulationInput
	parallelInput.PartialStateSet = e
	parallelInput.DeferredTransfers = t
	parallelInput.WorkReports = r[:i]
	parallelInput.AlwaysAccumulateMap = f

	parallelResult, err := ParallelizedAccumulation(parallelInput)
	eStar := parallelResult.PartialStateSet
	tStar := parallelResult.DeferredTransfers
	bStar := parallelResult.AccumulatedServiceOutput
	uStar := parallelResult.ServiceGasUsedList
	if err != nil {
		return output, fmt.Errorf("parallel accumulation failed: %w", err)
	}

	// Recurse on the remaining reports with the remaining gas
	// (j, e′, b, u) = ∆+(g∗ − ∑(s,u)∈u∗(u), t∗, ri..., e∗, {})
	gStar := input.GasLimit
	for _, DeferredTransfer := range t {
		gStar += DeferredTransfer.GasLimit
	}

	gasLimitForRecursion := gStar
	for _, u := range uStar {
		gasLimitForRecursion -= u.Gas
	}
	var recursiveOuterInput OuterAccumulationInput
	recursiveOuterInput.GasLimit = gasLimitForRecursion
	recursiveOuterInput.DeferredTransfers = tStar
	recursiveOuterInput.WorkReports = r[i:]
	recursiveOuterInput.InitPartialStateSet = eStar
	recursiveOuterInput.ServicesWithFreeAccumulation = make(map[types.ServiceId]types.Gas) // {}
	recursiveOuterOutput, err := OuterAccumulation(recursiveOuterInput)
	if err != nil {
		return output, fmt.Errorf("recursive accumulation failed: %w", err)
	}
	j := recursiveOuterOutput.NumberOfWorkResultsAccumulated
	ePrime := recursiveOuterOutput.PartialStateSet
	b := recursiveOuterOutput.AccumulatedServiceOutput
	u := recursiveOuterOutput.ServiceGasUsedList
	// Combine results from this batch and the recursive tail
	// (i + j, e′, b∗ ∪ b, u∗⌢ u)
	{
		output.NumberOfWorkResultsAccumulated = types.U64(i) + j
		output.PartialStateSet = ePrime // need to set post state?
		// merge b_star and b
		bUnion := make(map[types.AccumulatedServiceHash]bool)
		for key, value := range bStar {
			bUnion[key] = value
		}
		for key, value := range b {
			bUnion[key] = value
		}
		output.AccumulatedServiceOutput = bUnion
		output.ServiceGasUsedList = append(uStar, u...)
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
	for serviceId := range f {
		s[serviceId] = true
	}

	// td S t ∈ t
	for _, deferredTransfer := range t {
		s[deferredTransfer.ReceiverID] = true // td
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

type singleResult struct {
	out SingleServiceAccumulationOutput
	err error
}

// Deep copy single service accumulation input for goroutine parallelization
func (in SingleServiceAccumulationInput) CloneForService(s types.ServiceId) SingleServiceAccumulationInput {
	out := in
	out.ServiceId = s
	out.PartialStateSet = in.PartialStateSet.DeepCopy()
	out.DeferredTransfers = slices.Clone(in.DeferredTransfers)
	out.WorkReports = slices.Clone(in.WorkReports)
	out.AlwaysAccumulateMap = maps.Clone(in.AlwaysAccumulateMap)
	out.UnmatchedKeyVals = store.GetInstance().GetPostStateUnmatchedKeyVals()
	return out
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

	var tPrime []types.DeferredTransfer

	// maps for collecting service account state changes and update d_prime
	n := make(types.ServiceAccountState)
	m := make(types.ServiceAccountState)

	var singleInput SingleServiceAccumulationInput
	singleInput.PartialStateSet = e
	singleInput.DeferredTransfers = t
	singleInput.WorkReports = r
	singleInput.AlwaysAccumulateMap = f

	// Helper to run single service accumulation for a given service ID
	// ∆(s) ≡ ∆1(e, t, r, f, s)
	var (
		sf    singleflight.Group
		mu    sync.RWMutex
		cache = make(map[types.ServiceId]SingleServiceAccumulationOutput)
	)
	runSingleReplaceService := func(s types.ServiceId, singleParam SingleServiceAccumulationInput) (SingleServiceAccumulationOutput, error) {
		mu.RLock()
		if out, ok := cache[s]; ok {
			mu.RUnlock()
			return out, nil
		}
		localParam := singleParam.CloneForService(s)
		mu.RUnlock()
		// Use singleflight to deduplicate SingleServiceAccumulation per service.
		// The key(string) is used as identifier deduplicate calls.

		identifier := fmt.Sprintf("%d", s)
		v, err, _ := sf.Do(identifier, func() (any, error) {
			out, err := SingleServiceAccumulation(localParam)
			return out, err
		})
		if err != nil {
			return SingleServiceAccumulationOutput{}, err
		}
		out := v.(SingleServiceAccumulationOutput)

		mu.Lock()
		cache[s] = out
		store.GetInstance().SetPostStateUnmatchedKeyVals(out.UnmatchedKeyVals)
		mu.Unlock()

		return out, nil
	}

	// p: output service blobs collection
	var p types.ServiceBlobs
	// For each service in s, run single service accumulation in parallel
	{
		var wg sync.WaitGroup
		errCh := make(chan error)
		for serviceId := range s {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := runSingleReplaceService(serviceId, singleInput)
				if err != nil {
					errCh <- fmt.Errorf("single service accumulation for service %d failed: %w", serviceId, err)
					return
				}
			}()
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			return output, err
		}
	}
	// Process results from each service
	for service_id := range s {
		singleOutput, ok := cache[service_id]
		if !ok {
			singleOutput, err = runSingleReplaceService(service_id, singleInput)
			if err != nil {
				return output, fmt.Errorf("single service accumulation for service %d failed: %w", service_id, err)
			}
		}
		// u = [(s, ∆(s)u) S s <− s]
		var gasUse types.ServiceGasUsed
		gasUse.ServiceId = service_id
		gasUse.Gas = singleOutput.GasUsed
		u = append(u, gasUse)

		// b = {(s, b) S s ∈ s, b = ∆(s)y, b ≠ ∅}
		if singleOutput.AccumulationOutput != nil {
			var service_hash types.AccumulatedServiceHash
			service_hash.ServiceId = service_id
			service_hash.Hash = *singleOutput.AccumulationOutput
			b[service_hash] = true
		}

		// t = [∆(s)t S s <− s]
		tPrime = append(tPrime, singleOutput.DeferredTransfers...)

		singleOutputD := singleOutput.PartialStateSet.ServiceAccounts

		// n = ⋃ ((∆(s)e)d ∖ K(d ∖ { s }))
		// n = union of (d_prime without keys in d except service_id)
		for key, value := range singleOutputD {
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
		dExcludeSingleOutput := make(types.ServiceAccountState)
		for key := range d {
			if _, exists := singleOutputD[key]; !exists {
				dExcludeSingleOutput[key] = d[key]
			} else {
				// exclude part: ∖ K((∆(s)e)d)
			}
		}
		// add to m
		for key, value := range dExcludeSingleOutput {
			m[key] = value
		}
		// collect blobs updates
		p = append(p, singleOutput.ServiceBlobs...)
	}

	singleOutput, err := runSingleReplaceService(input.PartialStateSet.Bless, singleInput)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for bless failed: %w", err)
	}
	// e∗ = ∆(m)e
	eStar := singleOutput.PartialStateSet
	// m′, z′ = e∗(m, z)
	mPrime := eStar.Bless
	zPrime := eStar.AlwaysAccum

	// ∀c ∈ NC ∶ a′c = R(ac, (e∗a)c, ((∆(ac)e)a)c)
	aPrime := make(types.ServiceIdList, types.CoresCount)
	if len(a) != types.CoresCount {
		return output, fmt.Errorf("input.PartialStateSet.Assign length does not match types.CoresCount")
	}
	// For each core c, parallelize compute a′c
	{
		var wg sync.WaitGroup
		errCh := make(chan error)
		for c := range types.CoresCount {
			c := c
			serviceId := a[c]
			wg.Add(1)
			go func() {
				defer wg.Done()
				singleOutput, err := runSingleReplaceService(serviceId, singleInput)
				if err != nil {
					errCh <- fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
					return
				}
				aPrime[c] = R(serviceId, eStar.Assign[c], singleOutput.PartialStateSet.Assign[c])
			}()
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			return output, err
		}
	}

	// v' = R(v, e∗v , (∆(v)e)v )
	var vPrime, rPrime types.ServiceId
	singleOutput, err = runSingleReplaceService(input.PartialStateSet.Designate, singleInput)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for designate failed: %w", err)
	}
	vPrime = R(input.PartialStateSet.Designate, eStar.Designate, singleOutput.PartialStateSet.Designate)

	// r′ = R(r, e∗r , (∆(r)e)r)
	singleOutput, err = runSingleReplaceService(input.PartialStateSet.CreateAcct, singleInput)
	if err != nil {
		return output, fmt.Errorf("single service accumulation for createacct failed: %w", err)
	}
	rPrime = R(input.PartialStateSet.CreateAcct, eStar.CreateAcct, singleOutput.PartialStateSet.CreateAcct)

	// i′ = (∆(v)e)i
	var iPrime types.ValidatorsData
	{
		singleOutput, err := runSingleReplaceService(input.PartialStateSet.Designate, singleInput)
		if err != nil {
			return output, fmt.Errorf("single service accumulation for designate failed: %w", err)
		}
		iPrime = singleOutput.PartialStateSet.ValidatorKeys
	}

	// ∀c ∈ NC ∶ q′c = ((∆(ac)e)q)c
	var qPrime types.AuthQueues
	{

		qPrime = make(types.AuthQueues, types.CoresCount)
		if len(input.PartialStateSet.Assign) != types.CoresCount {
			logger.Warnf("input.PartialStateSet.Assign length does not match types.CoresCount")
		}
		// For each core c, parallelize compute q′c
		{
			var wg sync.WaitGroup
			errCh := make(chan error)
			for c, serviceId := range input.PartialStateSet.Assign {
				c := c
				serviceId := serviceId

				wg.Add(1)
				go func() {
					defer wg.Done()

					singleOutput, err := runSingleReplaceService(serviceId, singleInput)
					if err != nil {
						errCh <- fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
						return
					}
					qPrime[c] = singleOutput.PartialStateSet.Authorizers[c]
				}()
			}
			wg.Wait()
			close(errCh)

			for err := range errCh {
				return output, err
			}
		}

	}

	// (d ∪ n) ∖ m
	// d′ = P ((d ∪ n) ∖ m, ⋃ ∆(s)p)
	//	    		         s∈s
	dPrime, err := Provide(merge(d, n, m), p)
	if err != nil {
		return output, fmt.Errorf("failed to provide service accounts: %w", err)
	}

	// Set posterior state
	{
		store := store.GetInstance()
		store.GetPosteriorStates().SetChi(types.Privileges{
			Bless:       mPrime,
			Assign:      aPrime,
			Designate:   vPrime,
			CreateAcct:  rPrime,
			AlwaysAccum: zPrime,
		})
		store.GetPosteriorStates().SetVarphi(qPrime)
		store.GetPosteriorStates().SetIota(iPrime)
		store.GetPosteriorStates().SetDelta(dPrime)
	}
	// new partial state set: (d′, i′, q′, m′, a′, v′, r′, z′)
	// Set output ((d′, i′, q′, m′, a′, v′, r′, z′), t′, b, u)
	var newPartialState types.PartialStateSet
	{
		newPartialState.ServiceAccounts = dPrime
		newPartialState.ValidatorKeys = iPrime
		newPartialState.Authorizers = qPrime
		newPartialState.Bless = mPrime
		newPartialState.Assign = aPrime
		newPartialState.Designate = vPrime
		newPartialState.CreateAcct = rPrime
		newPartialState.AlwaysAccum = zPrime
	}
	output.PartialStateSet = newPartialState
	output.DeferredTransfers = tPrime
	output.AccumulatedServiceOutput = b
	output.ServiceGasUsedList = u
	return output, nil
}

// (12.20) ∆1 single-service accumulation function
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (output SingleServiceAccumulationOutput, err error) {
	e := input.PartialStateSet      // e: PartialStateSet
	t := input.DeferredTransfers    // t: DeferredTransfers
	r := input.WorkReports          // r: WorkReports
	f := input.AlwaysAccumulateMap  // f: AlwaysAccumulateMap
	s := input.ServiceId            // s: ServiceId
	var iU []types.Operand          // all operand inputs for Ψₐ
	var iT []types.DeferredTransfer // all deferred transfers for Ψₐ

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
				iU = append(iU, operand)
			}
		}
	}

	// iU: all deferred transfers for service s
	for _, deferredTransfer := range t {
		if deferredTransfer.ReceiverID == input.ServiceId {
			iT = append(iT, deferredTransfer)
			g += deferredTransfer.GasLimit
		}
	}

	sort.Slice(iT, func(i, j int) bool {
		return iT[i].SenderID < iT[j].SenderID
	})

	//  iT ⌢ iU
	var pvmItems []types.OperandOrDeferredTransfer
	for _, deferredTransfer := range iT {
		pvmItems = append(pvmItems, types.OperandOrDeferredTransfer{Operand: nil, DeferredTransfer: &deferredTransfer})
	}
	for _, operand := range iU {
		pvmItems = append(pvmItems, types.OperandOrDeferredTransfer{Operand: &operand, DeferredTransfer: nil})
	}
	// τ′: Posterior validator state used by Ψₐ
	tauPrime := store.GetInstance().GetPosteriorStates().GetTau()

	// η0: entropy used by Ψₐ
	eta0 := store.GetInstance().GetPosteriorStates().GetState().Eta[0]

	// (e, w, f , s)↦ ΨA(e, τ′, s, g, iT ⌢ iU )
	storageKeyVal := input.UnmatchedKeyVals
	pvmResult := PVM.Psi_A(e, tauPrime, s, g, pvmItems, eta0, storageKeyVal)
	store.GetInstance().SetPostStateUnmatchedKeyVals(pvmResult.StorageKeyVal)

	// Collect PVM results as output
	{
		output.AccumulationOutput = pvmResult.Result
		output.DeferredTransfers = pvmResult.DeferredTransfers
		output.GasUsed = pvmResult.Gas
		output.PartialStateSet = pvmResult.PartialStateSet
		output.ServiceBlobs = pvmResult.ServiceBlobs
		output.UnmatchedKeyVals = pvmResult.StorageKeyVal
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

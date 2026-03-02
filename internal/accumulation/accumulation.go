package accumulation

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/timing"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// (12.1) ξ ∈ ⟦{H}⟧_E: blockchain.Xi

// (12.2) ©ξ ≡ ⋃x∈ξ
// This function extracts all known (past) accumulated WorkPackageHashes.
func GetAccumulatedHashes() (output []types.WorkPackageHash) {
	xi := blockchain.GetInstance().GetPriorStates().GetXi() // Retrieve ξ

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

// (12.3) ϑ ∈ ⟦⟦(W, {H})⟧⟧_E available work reports: blockchain.Vartheta

// (12.4) W! ≡ [w S w <− W, S(wx)pS = 0 ∧ wl = {}]
// This function identifies and stores work reports that are immediately
// eligible for accumulation, i.e.:
//   - (wx)p == 0: the report has no unresolved prerequisites
//   - wl == {}  : there is no segment root lookup required
//
// These work reports are independent and can be accumulated without waiting.
func UpdateImmediatelyAccumulateWorkReports() {
	intermediateState := blockchain.GetInstance().GetIntermediateStates()
	availableReports := intermediateState.GetAvailableWorkReports()

	// Pre-allocate capacity: estimate that about half of reports are immediately accumulatable
	accumulatableReports := make([]types.WorkReport, 0, len(availableReports)/2)
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
// Get all workreport with dependency, and cs in QueuedWorkReports
func UpdateQueuedWorkReports() {
	intermediateState := blockchain.GetInstance().GetIntermediateStates()
	availableReports := intermediateState.GetAvailableWorkReports()
	// Pre-allocate capacity: estimate that about half of reports have dependencies
	reportsWithDependency := make(types.ReadyQueueItem, 0, len(availableReports)/2)
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
	// Pre-allocate capacity: total dependencies = prerequisites + segment root lookups
	totalDeps := len(report.Context.Prerequisites) + len(report.SegmentRootLookup)
	output.Dependencies = make([]types.WorkPackageHash, 0, totalDeps)
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
	finishedReportHashes := make(map[types.WorkPackageHash]bool, len(x))
	for _, h := range x {
		finishedReportHashes[h] = true
	}
	for _, item := range r {
		// If the report itself is already accumulated, skip it, remove from queue
		if _, exist := finishedReportHashes[item.Report.PackageSpec.Hash]; exist {
			continue
		}
		// Otherwise, filter its dependencies: keep only those NOT in the finished set
		// Pre-allocate capacity: worst case is all dependencies remain
		remainingDeps := make([]types.WorkPackageHash, 0, len(item.Dependencies))
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
	// Pre-allocate capacity: estimate that most items have no dependencies
	g := make([]types.WorkReport, 0, len(r))

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
	cs := blockchain.GetInstance()

	// (12.10) Get current slot index 'm'
	slot := cs.GetLatestBlock().Header.Slot
	E := types.EpochLength
	m := int(slot) % E

	// Get θ: the available work reports (ϑ)
	vartheta := cs.GetPriorStates().GetVartheta()
	WQ := cs.GetIntermediateStates().GetQueuedWorkReports()
	Wbang := cs.GetIntermediateStates().GetAccumulatedWorkReports()

	// E(ϑm... ⌢ ϑ...m ⌢ WQ)
	// Pre-calculate total capacity for composedQueue to avoid multiple reallocations
	composedQueueCapacity := len(WQ)
	for _, record := range vartheta[m:] {
		composedQueueCapacity += len(record)
	}
	for _, record := range vartheta[:m] {
		composedQueueCapacity += len(record)
	}
	composedQueue := make(types.ReadyQueueItem, 0, composedQueueCapacity)

	for _, record := range vartheta[m:] {
		composedQueue = append(composedQueue, record...)
	}

	for _, record := range vartheta[:m] {
		composedQueue = append(composedQueue, record...)
	}

	composedQueue = append(composedQueue, WQ...)

	accumulatedHashes := ExtractWorkReportHashes(Wbang)

	// (12.12) Compute q = E(..., P(W!))
	// Use accumulated hashes from W! to prune dependencies
	q := QueueEditingFunction(composedQueue, accumulatedHashes)

	// (12.11) W* ≡ W! ⌢ Q(q)
	// Reconstruct W* by appending newly-resolved reports to previously accumulated W!
	qResult := AccumulationPriorityQueue(q)
	WStar := make([]types.WorkReport, 0, len(Wbang)+len(qResult))
	WStar = append(WStar, Wbang...)
	WStar = append(WStar, qResult...)

	// Update W*
	cs.GetIntermediateStates().SetAccumulatableWorkReports(WStar)
}

// (12.16) ∆+ outer accumulation function
func OuterAccumulation(input OuterAccumulationInput) (output OuterAccumulationOutput, err error) {
	defer timing.Track("accumulation.OuterAccumulation")()

	// input parameters
	g := input.GasLimit
	t := input.DeferredTransfers
	r := input.WorkReports
	e := input.InitPartialStateSet
	f := input.ServicesWithFreeAccumulation

	gasSum := types.Gas(0)
	i := 0

	// Determine the maximal prefix of reports that fits within the gas limit
	for idx, report := range r {
		for _, result := range report.Results {
			gasSum += result.AccumulateGas
		}
		if gasSum <= g {
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
	//(e∗, t∗, b∗, u∗) = ∆∗(e, t, r...i, f)
	parallelInput := ParallelizedAccumulationInput{
		PartialStateSet:     e,
		DeferredTransfers:   t,
		WorkReports:         r[:i],
		AlwaysAccumulateMap: f,
	}

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
	recursiveOuterInput := OuterAccumulationInput{
		GasLimit:                     gasLimitForRecursion,
		DeferredTransfers:            tStar,
		WorkReports:                  r[i:],
		InitPartialStateSet:          eStar,
		ServicesWithFreeAccumulation: make(map[types.ServiceID]types.Gas), // {}
	}
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
		bUnion := make(map[types.AccumulatedServiceHash]bool, len(bStar)+len(b))
		maps.Copy(bUnion, bStar)
		maps.Copy(bUnion, b)
		output.AccumulatedServiceOutput = bUnion
		// Pre-allocate capacity for combined list
		combinedGasUsed := make(types.ServiceGasUsedList, 0, len(uStar)+len(u))
		combinedGasUsed = append(combinedGasUsed, uStar...)
		combinedGasUsed = append(combinedGasUsed, u...)
		output.ServiceGasUsedList = combinedGasUsed
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
func set_s(r []types.WorkReport, f types.AlwaysAccumulateMap, t []types.DeferredTransfer) map[types.ServiceID]bool {
	s := make(map[types.ServiceID]bool, len(r)*2+len(f)+len(t))
	// {rs S r ∈ r, d ∈ rd}
	for _, w := range r {
		for _, r := range w.Results {
			s[r.ServiceID] = true // rd
		}
	}

	// K(f)
	for serviceID := range f {
		s[serviceID] = true
	}

	// td S t ∈ t
	for _, deferredTransfer := range t {
		s[deferredTransfer.ReceiverID] = true // td
	}
	return s
}

// Merge three maps, d, n, m
func merge(d, n, m types.ServiceAccountState) types.ServiceAccountState {
	result := make(types.ServiceAccountState, len(d)+len(n)-len(m))
	// Copy d
	maps.Copy(result, d)
	// Merge n, overwriting any keys from d
	maps.Copy(result, n)
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
func (in SingleServiceAccumulationInput) CloneForService(s types.ServiceID) SingleServiceAccumulationInput {
	out := in
	out.ServiceID = s
	out.PartialStateSet = in.PartialStateSet.DeepCopy()
	out.DeferredTransfers = slices.Clone(in.DeferredTransfers)
	out.WorkReports = slices.Clone(in.WorkReports)
	out.AlwaysAccumulateMap = maps.Clone(in.AlwaysAccumulateMap)
	out.UnmatchedKeyVals = blockchain.GetInstance().GetPostStateUnmatchedKeyVals()
	return out
}

// (12.17) ∆∗ parallelized accumulation function

// Parallelize parts and partial state modification needs confirm what is the correct way to process
func ParallelizedAccumulation(input ParallelizedAccumulationInput) (output ParallelizedAccumulationOutput, err error) {
	defer timing.Track("accumulation.ParallelizedAccumulation")()

	// s = {s S s ∈ (rs S w ∈ w, r ∈ wr)} ∪ K(f) ∪ {td S t ∈ t}
	s := set_s(input.WorkReports, input.AlwaysAccumulateMap, input.DeferredTransfers)
	b := make(map[types.AccumulatedServiceHash]bool, len(s))
	u := make(types.ServiceGasUsedList, 0, len(s))

	// Needed notations from partial state set
	e := input.PartialStateSet
	t := input.DeferredTransfers
	r := input.WorkReports
	f := input.AlwaysAccumulateMap
	// d, a from input partial state set
	d := input.PartialStateSet.ServiceAccounts
	a := input.PartialStateSet.Assign

	tPrime := make([]types.DeferredTransfer, 0, len(t))

	// maps for collecting service account state changes and update d_prime
	n := make(types.ServiceAccountState, len(d))
	m := make(types.ServiceAccountState, len(d))

	singleInput := SingleServiceAccumulationInput{
		PartialStateSet:     e,
		DeferredTransfers:   t,
		WorkReports:         r,
		AlwaysAccumulateMap: f,
	}

	// Helper to run single service accumulation for a given service ID
	// ∆(s) ≡ ∆1(e, t, r, f, s)
	var (
		sf    singleflight.Group
		mu    sync.RWMutex
		cache = make(map[types.ServiceID]SingleServiceAccumulationOutput, len(s))
	)
	runSingleReplaceService := func(s types.ServiceID, singleParam SingleServiceAccumulationInput) (SingleServiceAccumulationOutput, error) {
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
		// Do not update global store here during parallel execution
		// UnmatchedKeyVals will be merged after all services complete
		mu.Unlock()

		return out, nil
	}

	// p: output service blobs collection
	p := make(types.ServiceBlobs, 0, len(s))
	// For each service in s, run single service accumulation in parallel
	{
		g := new(errgroup.Group)
		g.SetLimit(types.MaxWorkers)
		for serviceID := range s {
			serviceID := serviceID
			g.Go(func() error {
				_, err := runSingleReplaceService(serviceID, singleInput)
				if err != nil {
					return fmt.Errorf("single service accumulation for service %d failed: %w", serviceID, err)
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return output, err
		}
	}
	// Process results from each service accumulation
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
		gasUse.ServiceID = service_id
		gasUse.Gas = singleOutput.GasUsed
		u = append(u, gasUse)

		// b = {(s, b) S s ∈ s, b = ∆(s)y, b ≠ ∅}
		if singleOutput.AccumulationOutput != nil {
			var service_hash types.AccumulatedServiceHash
			service_hash.ServiceID = service_id
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
		maps.Copy(m, dExcludeSingleOutput)
		// collect blobs updates
		p = append(p, singleOutput.ServiceBlobs...)
	}

	// Merge UnmatchedKeyVals from all services using intersection
	// Each service remove its own keys from UnmatchedKeyVals
	// Keys removed by other services remain in each service's output
	// Use intersection to merge UnmatchedKeyVals which not removed by any service
	{
		if len(s) > 0 {
			keyCountMap := make(map[[31]byte]int, len(s))               // key -> count of how many services have this key
			keyValueMap := make(map[[31]byte]types.StateKeyVal, len(s)) // key -> value of the key
			serviceCount := 0

			// Count occurrences of each key across all service outputs
			for service_id := range s {
				singleOutput, ok := cache[service_id]
				if !ok {
					continue
				}
				serviceCount++
				// Deduplicate keys within each service output first
				serviceKeySet := make(map[[31]byte]bool, len(singleOutput.UnmatchedKeyVals))
				for _, kv := range singleOutput.UnmatchedKeyVals {
					if !serviceKeySet[kv.Key] {
						serviceKeySet[kv.Key] = true
						keyCountMap[kv.Key]++
						keyValueMap[kv.Key] = kv
					}
				}
			}

			// Only keep keys that exist in ALL service outputs (intersection)
			mergedUnmatchedKeyVals := make(types.StateKeyVals, 0, len(keyCountMap))
			for key, count := range keyCountMap {
				if count == serviceCount {
					// This key exists in all service outputs, keep it
					mergedUnmatchedKeyVals = append(mergedUnmatchedKeyVals, keyValueMap[key])
				}
			}

			// Update the global store with merged result
			blockchain.GetInstance().SetPostStateUnmatchedKeyVals(mergedUnmatchedKeyVals)
		}
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
	aPrime := make(types.ServiceIDList, types.CoresCount)
	if len(a) != types.CoresCount {
		return output, fmt.Errorf("input.PartialStateSet.Assign length does not match types.CoresCount")
	}
	// For each core c, parallelize compute a′c
	{
		g := new(errgroup.Group)
		g.SetLimit(types.MaxWorkers)

		for c := range types.CoresCount {
			c := c
			serviceID := a[c]
			g.Go(func() error {
				singleOutput, err := runSingleReplaceService(serviceID, singleInput)
				if err != nil {
					return fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
				}
				aPrime[c] = R(serviceID, eStar.Assign[c], singleOutput.PartialStateSet.Assign[c])
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return output, err
		}
	}
	// v' = R(v, e∗v , (∆(v)e)v )
	var vPrime, rPrime types.ServiceID
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
			g := new(errgroup.Group)
			g.SetLimit(types.MaxWorkers)
			for c, serviceID := range input.PartialStateSet.Assign {
				c := c
				serviceID := serviceID

				g.Go(func() error {

					singleOutput, err := runSingleReplaceService(serviceID, singleInput)
					if err != nil {
						return fmt.Errorf("single service accumulation for assign[%d] failed: %w", c, err)
					}
					qPrime[c] = singleOutput.PartialStateSet.Authorizers[c]
					return nil
				})
			}
			if err := g.Wait(); err != nil {
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

	// Filter tPrime: remove DeferredTransfers where SenderID or ReceiverID is deleted (not in dPrime)
	// A DeferredTransfer can only be valid if both sender and receiver services exist in dPrime
	tPrimeFiltered := make([]types.DeferredTransfer, 0, len(tPrime))
	for _, transfer := range tPrime {
		if _, senderExists := dPrime[transfer.SenderID]; !senderExists {
			continue
		}
		if _, receiverExists := dPrime[transfer.ReceiverID]; !receiverExists {
			continue
		}
		// Both sender and receiver exist in dPrime, keep this transfer
		tPrimeFiltered = append(tPrimeFiltered, transfer)
	}
	tPrime = tPrimeFiltered

	// Set posterior state
	{
		cs := blockchain.GetInstance()
		cs.GetPosteriorStates().SetChi(types.Privileges{
			Bless:       mPrime,
			Assign:      aPrime,
			Designate:   vPrime,
			CreateAcct:  rPrime,
			AlwaysAccum: zPrime,
		})
		cs.GetPosteriorStates().SetVarphi(qPrime)
		cs.GetPosteriorStates().SetIota(iPrime)
		cs.GetPosteriorStates().SetDelta(dPrime)
	}
	// new partial state set: (d′, i′, q′, m′, a′, v′, r′, z′)
	// Set output ((d′, i′, q′, m′, a′, v′, r′, z′), t′, b, u)
	output.PartialStateSet = types.PartialStateSet{
		ServiceAccounts: dPrime,
		ValidatorKeys:   iPrime,
		Authorizers:     qPrime,
		Bless:           mPrime,
		Assign:          aPrime,
		Designate:       vPrime,
		CreateAcct:      rPrime,
		AlwaysAccum:     zPrime,
	}
	output.DeferredTransfers = tPrime
	output.AccumulatedServiceOutput = b
	output.ServiceGasUsedList = u
	return output, nil
}

// (12.20) ∆1 single-service accumulation function
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (output SingleServiceAccumulationOutput, err error) {
	defer timing.Track("accumulation.SingleServiceAccumulation")()

	e := input.PartialStateSet     // e: PartialStateSet
	t := input.DeferredTransfers   // t: DeferredTransfers
	r := input.WorkReports         // r: WorkReports
	f := input.AlwaysAccumulateMap // f: AlwaysAccumulateMap
	s := input.ServiceID           // s: ServiceID
	// Estimate capacity: each work report may have multiple results matching this service
	// Use conservative estimate of 2 results per work report on average
	iU := make([]types.Operand, 0, len(r)*2)        // all operand inputs for Ψₐ
	iT := make([]types.DeferredTransfer, 0, len(t)) // all deferred transfers for Ψₐ

	// U(fs, 0)
	g := types.Gas(0)
	if preset, ok := f[input.ServiceID]; ok {
		g = preset
	}

	// iU: all accumulate work result operands for service s
	for _, r := range r {
		for _, d := range r.Results {
			if d.ServiceID == s {
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

	// iT: all deferred transfers for service s
	for _, deferredTransfer := range t {
		if deferredTransfer.ReceiverID == input.ServiceID {
			iT = append(iT, deferredTransfer)
			g += deferredTransfer.GasLimit
		}
	}

	sort.Slice(iT, func(i, j int) bool {
		return iT[i].SenderID < iT[j].SenderID
	})

	//  iT ⌢ iU
	pvmItems := make([]types.OperandOrDeferredTransfer, 0, len(iT)+len(iU))
	for _, deferredTransfer := range iT {
		pvmItems = append(pvmItems, types.OperandOrDeferredTransfer{Operand: nil, DeferredTransfer: &deferredTransfer})
	}
	for _, operand := range iU {
		pvmItems = append(pvmItems, types.OperandOrDeferredTransfer{Operand: &operand, DeferredTransfer: nil})
	}
	// τ′: Posterior validator state used by Ψₐ
	tauPrime := blockchain.GetInstance().GetPosteriorStates().GetTau()

	// η0: entropy used by Ψₐ
	eta0 := blockchain.GetInstance().GetPosteriorStates().GetState().Eta[0]

	// (e, w, f , s)↦ ΨA(e, τ′, s, g, iT ⌢ iU )
	storageKeyVal := input.UnmatchedKeyVals
	var pvmResult PVM.Psi_A_ReturnType
	func() {
		defer timing.Track("PVM.Psi_A")()
		pvmResult = PVM.Psi_A(e, tauPrime, s, g, pvmItems, eta0, storageKeyVal)
	}()

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
	defer timing.Track("accumulation.ProcessAccumulation")()

	// Compute W!
	UpdateImmediatelyAccumulateWorkReports()

	// Compute WQ
	UpdateQueuedWorkReports()

	// Compute W*
	UpdateAccumulatableWorkReports()
	return nil
}

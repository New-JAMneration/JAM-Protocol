package extrinsic

import (
	"crypto/ed25519"
	"fmt"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/mmr"
)

// GuaranteeController is a struct that contains a slice of ReportGuarantee (for controller logic)
type GuaranteeController struct {
	Guarantees []types.ReportGuarantee
}

// NewGuaranteeController creates a new GuaranteeController (Constructor)
func NewGuaranteeController() *GuaranteeController {
	return &GuaranteeController{
		Guarantees: make([]types.ReportGuarantee, 0),
	}
}

// Validate Guarantee extrinsic | Eq. 11.23
func (g *GuaranteeController) Validate() error {
	if len(g.Guarantees) != types.CoresCount {
		return fmt.Errorf("GuaranteeController.Validate failed: bad_guarantee_count")
	}
	for _, guarantee := range g.Guarantees {
		if err := guarantee.Validate(); err != nil {
			return fmt.Errorf("GuaranteeController.Validate failed: %w", err)
		}
	}
	return nil
}

// Sort Guarantee extrinsic | Eq. 11.24-11.25
func (g *GuaranteeController) Sort() error {
	sort.Slice(g.Guarantees, func(i, j int) bool {
		return g.Guarantees[i].Report.CoreIndex < g.Guarantees[j].Report.CoreIndex
	})
	for i := 1; i < len(g.Guarantees); i++ {
		if g.Guarantees[i].Report.CoreIndex < g.Guarantees[i-1].Report.CoreIndex {
			return fmt.Errorf("GuaranteeController.Sort failed: coreIndex not sorted")
		}
		SortUniqueSignatures(g.Guarantees[i].Signatures)
	}
	return nil
}

func SortUniqueSignatures(signatures []types.ValidatorSignature) {
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].ValidatorIndex < signatures[j].ValidatorIndex
	})
	uniqueSignatures := signatures[:0]
	for i, sig := range signatures {
		if i == 0 || sig.ValidatorIndex != signatures[i-1].ValidatorIndex {
			uniqueSignatures = append(uniqueSignatures, sig)
		}
	}
	copy(signatures, uniqueSignatures)
}

// ValidateSignatures | Eq. 11.26
func (g *GuaranteeController) ValidateSignatures() error {
	state := store.GetInstance().GetPosteriorStates().GetState()

	var guranatorAssignments GuranatorAssignments
	for _, guarantee := range g.Guarantees {
		if (int(state.Tau))/R == int(state.Tau)/int(guarantee.Slot) {
			guranatorAssignments = GFunc()
		} else {
			guranatorAssignments = GStarFunc()
		}

		if !((int(state.Tau)/R-1)*R <= int(guarantee.Slot) && int(guarantee.Slot) <= int(state.Tau)) {
			return fmt.Errorf("invalid_slot")
		}
		message := []byte(jam_types.JamGuarantee)
		hashed := hash.Blake2bHash(utilities.WorkReportSerialization(guarantee.Report))
		message = append(message, hashed[:]...)
		for _, sig := range guarantee.Signatures {
			if guranatorAssignments.CoreAssignments[sig.ValidatorIndex] != guarantee.Report.CoreIndex {
				return fmt.Errorf("invalid_core_index")
			}
			publicKey := guranatorAssignments.PublicKeys[sig.ValidatorIndex][:]
			if !ed25519.Verify(publicKey, message, sig.Signature[:]) {
				return fmt.Errorf("invalid_signature")
			}
		}
	}
	return nil
}

// WorkReportSet | Eq. 11.28
func (g *GuaranteeController) WorkReportSet() []types.WorkReport {
	workReports := make([]types.WorkReport, 0)
	for _, guarantee := range g.Guarantees {
		workReports = append(workReports, guarantee.Report)
	}
	return workReports
}

// ValidateWorkReports | Eq. 11.29-11.30
func (g *GuaranteeController) ValidateWorkReports() error {
	workReports := g.WorkReportSet()
	alpha := store.GetInstance().GetPriorStates().GetAlpha()
	delta := store.GetInstance().GetPosteriorStates().GetDelta()
	rhoDoubleDagger := store.GetInstance().GetIntermediateStates().GetRhoDoubleDagger()
	for _, workReport := range workReports {
		if rhoDoubleDagger[workReport.CoreIndex] != nil {
			return fmt.Errorf("invalid_core_index")
		}
		authPool := alpha[workReport.CoreIndex]
		if !isAuthPoolContains(authPool, workReport.AuthorizerHash) {
			return fmt.Errorf("invalid_authorizer")
		}
		totalGas := types.U64(0)
		for _, workResult := range workReport.Results {
			totalGas += types.U64(workResult.AccumulateGas)
			if workResult.AccumulateGas < delta[workResult.ServiceId].MinItemGas {
				return fmt.Errorf("invalid_gas")
			}
		}
		if totalGas > types.U64(types.GasLimit) {
			return fmt.Errorf("invalid_gas")
		}
	}
	return nil
}

func isAuthPoolContains(authPool []types.AuthorizerHash, authorizerHash types.OpaqueHash) bool {
	for _, auth := range authPool {
		if types.OpaqueHash(auth) == authorizerHash {
			return true
		}
	}
	return false
}

// ContextSet | Eq. 11.31
func (g *GuaranteeController) ContextSet() []types.RefineContext {
	contexts := make([]types.RefineContext, 0)
	for _, guarantee := range g.Guarantees {
		contexts = append(contexts, guarantee.Report.Context)
	}
	return contexts
}

// WorkPackageHashSet | Eq. 11.31
func (g *GuaranteeController) WorkPackageHashSet() []types.WorkPackageHash {
	workPackageHashes := make([]types.WorkPackageHash, 0)
	for _, guarantee := range g.Guarantees {
		workPackageHashes = append(workPackageHashes, guarantee.Report.PackageSpec.Hash)
	}
	return workPackageHashes
}

// CardinalityCheck | Eq. 11.32
func (g *GuaranteeController) CardinalityCheck() error {
	contexts := g.ContextSet()
	workPackageHashes := g.WorkPackageHashSet()
	if len(contexts) != len(workPackageHashes) {
		return fmt.Errorf("invalid_cardinality")
	}
	return nil
}

// ValidateContexts | Eq. 11.33-11.35
func (g *GuaranteeController) ValidateContexts() error {
	contexts := g.ContextSet()
	betaDagger := store.GetInstance().GetIntermediateStates().GetBetaDagger()
	headerTimeSlot := store.GetInstance().GetBlock().Header.Slot
	ancestorHeaders := store.GetInstance().GetAncestorHeaders()
	for _, context := range contexts {
		foundMatch := false
		for _, blockInfo := range betaDagger {
			m := mmr.NewMMRFromPeaks(blockInfo.Mmr.Peaks, hash.Blake2bHash).SuperPeak(blockInfo.Mmr.Peaks)
			if context.Anchor == blockInfo.HeaderHash && context.StateRoot == blockInfo.StateRoot && context.BeefyRoot == types.BeefyRoot(*m) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return fmt.Errorf("invalid_context")
		}
		if context.LookupAnchorSlot < (headerTimeSlot - types.TimeSlot(types.MaxLookupAge)) {
			return fmt.Errorf("invalid_context")
		}
	}
	for _, context := range contexts {
		foundMatch := false
		for _, ancestorHeader := range ancestorHeaders {
			if context.LookupAnchorSlot == ancestorHeader.Slot && hash.Blake2bHash(utilities.HeaderSerialization(ancestorHeader)) == types.OpaqueHash(context.LookupAnchor) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return fmt.Errorf("invalid_context")
		}
	}
	return nil
}

// ValidateWorkPackageHashes | Eq. 11.36-11.38
func (g *GuaranteeController) ValidateWorkPackageHashes() error {
	workPackageHashes := g.WorkPackageHashSet()
	theta := store.GetInstance().GetPriorStates().GetTheta()
	rho := store.GetInstance().GetPriorStates().GetRho()
	xi := store.GetInstance().GetPriorStates().GetXi()
	beta := store.GetInstance().GetPriorStates().GetBeta()
	qMap := make(map[types.WorkPackageHash]bool)
	for _, v := range theta {
		qMap[v.WorkReport.PackageSpec.Hash] = true
	}
	aMap := make(map[types.WorkPackageHash]bool)
	for _, v := range rho {
		aMap[v.Report.PackageSpec.Hash] = true
	}
	xiMap := make(map[types.WorkPackageHash]bool)
	for _, v := range xi {
		for _, w := range v {
			xiMap[w] = true
		}
	}
	betaMap := make(map[types.WorkPackageHash]bool)
	for _, v := range beta {
		for _, w := range v.Reported {
			betaMap[types.WorkPackageHash(w.Hash)] = true
		}
	}
	for _, workPackageHash := range workPackageHashes {
		if qMap[workPackageHash] || aMap[workPackageHash] || xiMap[workPackageHash] || betaMap[workPackageHash] {
			return fmt.Errorf("invalid_work_package_hash")
		}
	}
	return nil
}

// CheckExtrinsicOrRecentHistory | Eq. 11.39
func (g *GuaranteeController) CheckExtrinsicOrRecentHistory() error {
	w := g.WorkReportSet()
	beta := store.GetInstance().GetPriorStates().GetBeta()
	packageSet := make(map[types.OpaqueHash]bool)
	for _, v := range w {
		for _, w := range v.Context.Prerequisites {
			packageSet[types.OpaqueHash(w)] = true
		}
		for _, w := range v.SegmentRootLookup {
			packageSet[types.OpaqueHash(w.WorkPackageHash)] = true
		}
	}
	p := g.WorkPackageHashSet()
	checkPackageSet := make(map[types.OpaqueHash]bool)
	for _, v := range p {
		checkPackageSet[types.OpaqueHash(v)] = true
	}
	for _, v := range beta {
		for _, w := range v.Reported {
			checkPackageSet[types.OpaqueHash(w.Hash)] = true
		}
	}
	for k, _ := range packageSet {
		if !checkPackageSet[k] {
			return fmt.Errorf("invalid_work_package_hash")
		}
	}
	return nil
}

// CheckSegmentRootLookup | Eq. 11.40-11.41
func (g *GuaranteeController) CheckSegmentRootLookup() error {
	blockDicSet := make(map[types.WorkPackageHash]types.OpaqueHash)
	for _, guarantee := range g.Guarantees {
		for _, segmentRootLookup := range guarantee.Report.SegmentRootLookup {
			blockDicSet[segmentRootLookup.WorkPackageHash] = segmentRootLookup.SegmentTreeRoot
		}
	}
	beta := store.GetInstance().GetPriorStates().GetBeta()
	for _, v := range beta {
		for _, w := range v.Reported {
			blockDicSet[types.WorkPackageHash(w.Hash)] = types.OpaqueHash(w.ExportsRoot)
		}
	}
	w := g.WorkReportSet()
	for _, v := range w {
		for _, w := range v.SegmentRootLookup {
			if blockDicSet[w.WorkPackageHash] != w.SegmentTreeRoot {
				return fmt.Errorf("invalid_segment_root_lookup")
			}
		}
	}
	return nil
}

// CheckWorkResult | Eq. 11.42
func (g *GuaranteeController) CheckWorkResult() error {
	w := g.WorkReportSet()
	delta := store.GetInstance().GetPosteriorStates().GetDelta()
	for _, v := range w {
		for _, w := range v.Results {
			if w.CodeHash != delta[w.ServiceId].CodeHash {
				return fmt.Errorf("invalid_work_result")
			}
		}
	}
	return nil
}

// Transitioning for work reports | Eq. 11.43
func (g *GuaranteeController) TransitionWorkReport() {
	rhoDoubleDagger := store.GetInstance().GetIntermediateStates().GetRhoDoubleDagger()
	posteriorTau := store.GetInstance().GetPosteriorStates().GetState().Tau
	coreIndexMap := make(map[types.CoreIndex]bool)
	for _, guarantee := range g.Guarantees {
		coreIndexMap[guarantee.Report.CoreIndex] = true
	}
	for i := range rhoDoubleDagger {
		if coreIndexMap[types.CoreIndex(i)] {
			rhoDoubleDagger[i].Report = g.Guarantees[i].Report
			rhoDoubleDagger[i].Timeout = posteriorTau
		}
	}
	store.GetInstance().GetPosteriorStates().SetRho(rhoDoubleDagger)
}

// Set sets the ReportGuarantee slice
func (g *GuaranteeController) Set(gToSet []types.ReportGuarantee) {
	g.Guarantees = gToSet
}

// Len returns the length of the slice
func (r *GuaranteeController) Len() int {
	return len(r.Guarantees)
}

// Less returns true if the index i is less than the index j
func (r *GuaranteeController) Less(i, j int) bool {
	return r.Guarantees[i].Report.CoreIndex < r.Guarantees[j].Report.CoreIndex
}

// Swap swaps the index i with the index j
func (r *GuaranteeController) Swap(i, j int) {
	r.Guarantees[i], r.Guarantees[j] = r.Guarantees[j], r.Guarantees[i]
}

// Sort sorts the slice
func (r *GuaranteeController) SortSlice() {
	sort.Slice(r.Guarantees, func(i, j int) bool {
		return r.Less(i, j)
	})
}

// Add adds a new Guarantee to the ReportGuarantee slice.
func (r *GuaranteeController) Add(newReportGuarantee types.ReportGuarantee) {
	r.Guarantees = append(r.Guarantees, newReportGuarantee)
	r.Sort()
}

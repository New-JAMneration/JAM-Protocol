package extrinsic

import (
	"crypto/ed25519"
	"log"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	ReportsErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
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
	for _, guarantee := range g.Guarantees {
		if guarantee.Report.CoreIndex >= types.CoreIndex(types.CoresCount) {
			err := ReportsErrorCode.BadCoreIndex
			return &err
		}
	}

	if len(g.Guarantees) > types.CoresCount {
		err := ReportsErrorCode.BadCoreIndex
		return &err
	}
	for _, guarantee := range g.Guarantees {
		if err := guarantee.Validate(); err != nil {
			errCode := ReportsErrorCode.ReportsErrorMap[err.Error()]
			return &errCode
		}
	}

	return nil
}

// Sort Guarantee extrinsic | Eq. 11.24-11.25
func (g *GuaranteeController) Sort() error {
	/*
		sort.Slice(g.Guarantees, func(i, j int) bool {
			return g.Guarantees[i].Report.CoreIndex < g.Guarantees[j].Report.CoreIndex
		})
	*/
	if len(g.Guarantees) == 0 {
		return nil
	}

	err := SortUniqueSignatures(g.Guarantees[0].Signatures)
	if err != nil {
		// not_sorted_guarantors
		return err
	}
	for i := 1; i < len(g.Guarantees); i++ {
		if g.Guarantees[i-1].Report.CoreIndex >= g.Guarantees[i].Report.CoreIndex {
			err := ReportsErrorCode.OutOfOrderGuarantee
			return &err
		}
		err := SortUniqueSignatures(g.Guarantees[i].Signatures)
		if err != nil {
			// "not_sorted_guarantors"
			return err
		}
	}
	return nil
}

func SortUniqueSignatures(signatures []types.ValidatorSignature) error {
	/*
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
	*/
	if len(signatures) == 0 {
		return nil
	}

	for i := 0; i < len(signatures)-1; i++ {
		if signatures[i].ValidatorIndex >= signatures[i+1].ValidatorIndex {
			// return errors.New("not_sorted_or_unique_guarantors")
			err := ReportsErrorCode.NotSortedOrUniqueGuarantors
			return &err
		}
	}

	return nil
}

// ValidateSignatures | Eq. 11.26
func (g *GuaranteeController) ValidateSignatures() error {
	tau := store.GetInstance().GetPosteriorStates().GetTau()
	var guranatorAssignments GuranatorAssignments

	for _, guarantee := range g.Guarantees {
		if (int(tau))/R == int(guarantee.Slot)/R {
			guranatorAssignments = GFunc()
		} else {
			guranatorAssignments = GStarFunc()
		}
		if !((int(tau)/R-1)*R <= int(guarantee.Slot)) {
			err := ReportsErrorCode.ReportEpochBeforeLast
			return &err
		}

		if !(int(guarantee.Slot) <= int(tau)) {
			err := ReportsErrorCode.FutureReportSlot
			return &err
		}

		message := []byte(jam_types.JamGuarantee)
		hashed := hash.Blake2bHash(utilities.WorkReportSerialization(guarantee.Report))
		message = append(message, hashed[:]...)
		for _, sig := range guarantee.Signatures {
			if guranatorAssignments.CoreAssignments[sig.ValidatorIndex] != guarantee.Report.CoreIndex {
				err := ReportsErrorCode.WrongAssignment
				return &err
			}
			publicKey := guranatorAssignments.PublicKeys[sig.ValidatorIndex][:]
			if !ed25519.Verify(publicKey, message, sig.Signature[:]) {
				err := ReportsErrorCode.BadSignature
				return &err
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
			err := ReportsErrorCode.CoreEngaged
			return &err
		}
		authPool := alpha[workReport.CoreIndex]
		if !isAuthPoolContains(authPool, workReport.AuthorizerHash) {
			err := ReportsErrorCode.CoreUnauthorized
			return &err
		}
		totalGas := types.U64(0)
		for _, workResult := range workReport.Results {
			totalGas += types.U64(workResult.AccumulateGas)
			if _, serviceExists := delta[workResult.ServiceId]; !serviceExists {
				err := ReportsErrorCode.BadServiceId
				return &err
			}
			if workResult.AccumulateGas < delta[workResult.ServiceId].ServiceInfo.MinItemGas {
				err := ReportsErrorCode.ServiceItemGasTooLow
				return &err
			}
		}
		if totalGas > types.U64(types.MaxAccumulateGas) {
			err := ReportsErrorCode.WorkReportGasTooHigh
			return &err
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
	workPackageMap := make(map[types.WorkPackageHash]bool)
	// filter duplicate WorkPackageHash
	for _, guarantee := range g.Guarantees {
		workPackageMap[guarantee.Report.PackageSpec.Hash] = true
	}

	for workPackageHash := range workPackageMap {
		workPackageHashes = append(workPackageHashes, workPackageHash)
	}
	return workPackageHashes
}

// CardinalityCheck | Eq. 11.32
func (g *GuaranteeController) CardinalityCheck() error {
	workReports := g.WorkReportSet()
	workPackageHashes := g.WorkPackageHashSet()

	if len(workReports) != len(workPackageHashes) {
		err := ReportsErrorCode.DuplicatePackage
		return &err
	}

	return nil
}

// ValidateContexts | Eq. 11.33-11.35
func (g *GuaranteeController) ValidateContexts() error {
	contexts := g.ContextSet()
	betaDagger := store.GetInstance().GetIntermediateStates().GetBetaHDagger()
	headerTimeSlot := store.GetInstance().GetBlock().Header.Slot

	for _, context := range contexts {
		recentAnchorMatch := false
		stateRootMatch := false
		beefyRootMatch := false
		for _, blockInfo := range betaDagger {
			if context.Anchor == blockInfo.HeaderHash {
				recentAnchorMatch = true
				stateRootMatch = (context.StateRoot == blockInfo.StateRoot)
				beefyRootMatch = context.BeefyRoot == types.BeefyRoot(blockInfo.MmrPeak)
				break
			}
		}
		if !recentAnchorMatch {
			err := ReportsErrorCode.AnchorNotRecent
			return &err
		}
		if !stateRootMatch {
			err := ReportsErrorCode.BadStateRoot
			return &err
		}
		if !beefyRootMatch {
			err := ReportsErrorCode.BadBeefyMmrRoot
			return &err
		}

		if int(context.LookupAnchorSlot) < int(headerTimeSlot)-types.MaxLookupAge {
			// return errors.New("report_before_last_rotation")
			err := ReportsErrorCode.ReportEpochBeforeLast
			return &err
		}
	}
	// 11.35
	ancestorHeaders := store.GetInstance().GetAncestorHeaders()
	for _, context := range contexts {
		foundMatch := false
		for _, ancestorHeader := range ancestorHeaders {
			serializedHeader, err := utilities.HeaderSerialization(ancestorHeader)
			if err != nil {
				return err
			}
			if context.LookupAnchorSlot == ancestorHeader.Slot && hash.Blake2bHash(serializedHeader) == types.OpaqueHash(context.LookupAnchor) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			// return errors.New("invalid_context")
			log.Println("invalid_context")
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
	// q
	for _, v := range theta {
		for _, w := range v {
			qMap[w.Report.PackageSpec.Hash] = true
		}
	}

	aMap := make(map[types.WorkPackageHash]bool)
	for _, v := range rho {
		if v != nil {
			aMap[v.Report.PackageSpec.Hash] = true
		}
	}
	xiMap := make(map[types.WorkPackageHash]bool)
	for _, v := range xi {
		for _, w := range v {
			xiMap[w] = true
		}
	}
	betaMap := make(map[types.WorkPackageHash]bool)
	for _, v := range beta.History {
		for _, w := range v.Reported {
			betaMap[types.WorkPackageHash(w.Hash)] = true
		}
	}

	for _, workPackageHash := range workPackageHashes {
		if qMap[workPackageHash] || aMap[workPackageHash] || xiMap[workPackageHash] || betaMap[workPackageHash] {
			err := ReportsErrorCode.DuplicatePackage
			return &err
		}
	}
	return nil
}

// CheckExtrinsicOrRecentHistory | Eq. 11.39
func (g *GuaranteeController) CheckExtrinsicOrRecentHistory() error {
	w := g.WorkReportSet()
	beta := store.GetInstance().GetPriorStates().GetBeta()
	dependencySet := make(map[types.OpaqueHash]bool)
	segmentRootSet := make(map[types.OpaqueHash]bool)
	for _, v := range w {
		for _, w := range v.Context.Prerequisites {
			dependencySet[types.OpaqueHash(w)] = true
		}
		for _, w := range v.SegmentRootLookup {
			segmentRootSet[types.OpaqueHash(w.WorkPackageHash)] = true
		}
	}
	p := g.WorkPackageHashSet()
	checkPackageSet := make(map[types.OpaqueHash]bool)
	for _, v := range p {
		checkPackageSet[types.OpaqueHash(v)] = true
	}
	for _, v := range beta.History {
		for _, w := range v.Reported {
			checkPackageSet[types.OpaqueHash(w.Hash)] = true
		}
	}
	for k := range dependencySet {
		if !checkPackageSet[k] {
			err := ReportsErrorCode.DependencyMissing
			return &err
		}
	}
	for k := range segmentRootSet {
		if !checkPackageSet[k] {
			err := ReportsErrorCode.SegmentRootLookupInvalid
			return &err
		}
	}

	return nil
}

// CheckSegmentRootLookup | Eq. 11.40-11.41
func (g *GuaranteeController) CheckSegmentRootLookup() error {
	pSet := make(map[types.WorkPackageHash]types.ExportsRoot)
	for _, guarantee := range g.Guarantees {
		pSet[guarantee.Report.PackageSpec.Hash] = guarantee.Report.PackageSpec.ExportsRoot
	}
	beta := store.GetInstance().GetPriorStates().GetBeta()
	for _, v := range beta.History {
		for _, w := range v.Reported {
			pSet[types.WorkPackageHash(w.Hash)] = w.ExportsRoot
		}
	}
	w := g.WorkReportSet()
	for _, v := range w {
		for _, w := range v.SegmentRootLookup {
			// Segments tree root lookup item not found in recent blocks history.
			segmentRootLookup, segmentRootExists := pSet[w.WorkPackageHash]
			if !segmentRootExists {
				err := ReportsErrorCode.SegmentRootLookupInvalid
				return &err
			}
			// Segments tree root lookup item found in recent blocks history but with an unexpected value.
			if segmentRootLookup != types.ExportsRoot(w.SegmentTreeRoot) {
				err := ReportsErrorCode.SegmentRootLookupInvalid
				return &err
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
			if w.CodeHash != delta[w.ServiceId].ServiceInfo.CodeHash {
				err := ReportsErrorCode.BadCodeHash
				return &err
			}
		}
	}
	return nil
}

// Transitioning for work reports | Eq. 11.43
func (g *GuaranteeController) TransitionWorkReport() {
	rhoDoubleDagger := store.GetInstance().GetIntermediateStates().GetRhoDoubleDagger()
	posteriorTau := store.GetInstance().GetPosteriorStates().GetTau()
	coreIndexMap := make(map[types.CoreIndex]bool)
	for _, guarantee := range g.Guarantees {
		coreIndexMap[guarantee.Report.CoreIndex] = true
	}
	for i := range rhoDoubleDagger {
		if coreIndexMap[types.CoreIndex(i)] {
			rhoDoubleDagger[i] = &types.AvailabilityAssignment{
				Report:  g.Guarantees[i].Report,
				Timeout: posteriorTau,
			}
		}
	}

	store.GetInstance().GetPosteriorStates().SetRho(rhoDoubleDagger)

	// Save the work reports to the store
	workReports := g.WorkReportSet()
	store.GetInstance().GetIntermediateStates().SetPresentWorkReports(workReports)
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

func GetGuarantors(guarantee types.ReportGuarantee) []types.Ed25519Public {
	tau := store.GetInstance().GetPosteriorStates().GetTau()
	var guranatorAssignments GuranatorAssignments
	guarantors := make([]types.Ed25519Public, 0)
	if (int(tau))/R == int(guarantee.Slot)/R {
		guranatorAssignments = GFunc()
	} else {
		guranatorAssignments = GStarFunc()
	}

	for _, sig := range guarantee.Signatures {
		guarantors = append(guarantors, guranatorAssignments.PublicKeys[sig.ValidatorIndex])
	}

	return guarantors
}

/*
	for _, sig := range guarantee.Signatures {
		if guranatorAssignments.CoreAssignments[sig.ValidatorIndex] != guarantee.Report.CoreIndex {
			err := ReportsErrorCode.WrongAssignment
			return &err
		}
		publicKey := guranatorAssignments.PublicKeys[sig.ValidatorIndex][:]
		if !ed25519.Verify(publicKey, message, sig.Signature[:]) {
			err := ReportsErrorCode.BadSignature
			return &err
		}
	}
*/

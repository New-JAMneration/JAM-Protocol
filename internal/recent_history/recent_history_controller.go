package recent_history

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	merkle "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/mmr"
)

// RecentHistoryController is a controller for the recent history.
// This controller is used to manage the recent history.
type RecentHistoryController struct {
	Betas types.RecentBlocks
}

// NewRecentHistoryController creates a new RecentHistoryController.
func NewRecentHistoryController() *RecentHistoryController {
	return &RecentHistoryController{
		Betas: types.RecentBlocks{},
	}
}

var maxBlocksHistory = types.MaxBlocksHistory

// Remove duplicated blocks by BlockHash
func (rhc *RecentHistoryController) CheckDuplicate(headerhash types.HeaderHash) bool {
	// Check if headerhash is already in Recent History Controller
	for _, beta := range rhc.Betas.History {
		if beta.HeaderHash == headerhash {
			return true
		}
	}
	return false
}

// Beta_RecentHistory^dagger (7.5) GP 0.6.7
func (rhc *RecentHistoryController) RecentHistory2Dagger(parentStateRoot types.StateRoot) {
	s := store.GetInstance()
	// Get recent beta_H^dagger from store
	betaDagger := s.GetIntermediateStates().GetBetaHDagger()

	if len(rhc.Betas.History) > 0 {
		// Append first avoid empty slice
		// Duplicate beta_H into beta_H^dagger
		betaDagger = append(betaDagger, rhc.Betas.History...)
		// Except for the stateroot need to be updated
		betaDagger[len(rhc.Betas.History)-1].StateRoot = parentStateRoot
	}

	// Check beta_H^dagger is not longer than maxBlocksHistory
	if len(betaDagger) > maxBlocksHistory {
		// Remove old elements to retain maxBlocksHistory
		betaDagger = betaDagger[len(betaDagger)-maxBlocksHistory:]
	}

	// Set beta_H^dagger to intermediate state in store
	s.GetIntermediateStates().SetBetaHDagger(betaDagger)
}

// -----(7.3)-----

// // Accumulation-result tree root $r$
// func r(c types.AccumulatedServiceOutput) (accumulationResultTreeRoot types.OpaqueHash) {
// 	// Empty struct
// 	pairs := make([]types.AccumulatedServiceHash, len(c))

// 	for commitment, exist := range c {
// 		if exist {
// 			pairs = append(pairs, types.AccumulatedServiceHash{
// 				ServiceId: commitment.ServiceId,
// 				Hash:      commitment.Hash,
// 			})
// 		}
// 	}

// 	// Sort by serviceid $s$
// 	sort.Slice(pairs, func(i, j int) bool {
// 		return pairs[i].ServiceId < pairs[j].ServiceId
// 	})

// 	// Serialization
// 	var dataSerialized types.ByteSequence
// 	for _, pair := range pairs {
// 		serviceidSerialized := utils.SerializeFixedLength(types.U32(pair.ServiceId), 4)
// 		dataSerialized = append(dataSerialized, serviceidSerialized...)

// 		hashSerialized := utils.OpaqueHashWrapper{Value: pair.Hash}.Serialize()
// 		dataSerialized = append(dataSerialized, hashSerialized...)
// 	}

// 	// Merklization
// 	accumulationResultTreeRoot = merkle.Mb([]types.ByteSequence{dataSerialized}, hash.KeccakHash)
// 	return accumulationResultTreeRoot
// }

// (7.6) \mathbf{s} GP 0.6.7
// TODO: remove mock theta and read from store(posterior LastAccOut)
func s() (output types.ByteSequence) {
	newEncoder := types.NewEncoder()
	mockTheta := []types.AccumulatedServiceHash{}
	for _, pair := range mockTheta {
		encodedServiceId, err := newEncoder.EncodeUintWithLength(uint64(pair.ServiceId), 4)
		if err != nil {
			return nil
		}
		output = append(output, encodedServiceId...)
		encodedHash, err := newEncoder.Encode(pair.Hash)
		if err != nil {
			return nil
		}
		output = append(output, encodedHash...)
	}
	return output
}

// Merkle Mountain Range $b$
// (7.7) GP 0.6.7
func (rhc *RecentHistoryController) b() types.OpaqueHash {
	mmb := s()
	wrappedMmr := mmr.MmrWrapper(&rhc.Betas.Mmr, hash.KeccakHash)
	accumulationResultTreeRoot := merkle.Mb([]types.ByteSequence{mmb}, hash.KeccakHash)
	// MMR append func $\mathcal{A}$
	beefybeltPrime := wrappedMmr.AppendOne(types.MmrPeak(&accumulationResultTreeRoot))
	return wrappedMmr.SuperPeak(beefybeltPrime)
}

// Work Report map $\mathbf{p}$
func p(eg types.GuaranteesExtrinsic) []types.ReportedWorkPackage {
	var reports []types.ReportedWorkPackage
	// Create a map from eg.Report.PackageSpec.Hash to eg.Report.PackageSpec.ExportsRoot
	for _, eg := range eg {
		report := types.ReportedWorkPackage{
			// Golang cannot compare different struct, so transfer first
			Hash:        types.WorkReportHash(eg.Report.PackageSpec.Hash),
			ExportsRoot: eg.Report.PackageSpec.ExportsRoot,
		}
		reports = append(reports, report)
	}
	return reports
}

// item $n$ = (header hash $h$, accumulation-result mmr $b$, state root $s$, WorkReportHash $\mathbf{p}$)
// (7.8) GP 0.6.7
func (rhc *RecentHistoryController) N(headerHash types.HeaderHash, eg types.GuaranteesExtrinsic) (items types.BlockInfo) {
	accumulationResultMmr := rhc.b()
	workReportHash := p(eg)
	zeroHash := types.StateRoot{}

	items = types.BlockInfo{
		HeaderHash: headerHash,
		BeefyRoot:  accumulationResultMmr,
		StateRoot:  zeroHash,
		Reported:   workReportHash,
	}
	return items
}

// -----(7.3)-----

// Update beta^dagger to beta^prime (7.4)
func (rhc *RecentHistoryController) AddToBetaPrime(items types.BlockInfo) {
	s := store.GetInstance()
	// Get recent beta^dagger from store
	historyDagger := s.GetIntermediateStates().GetBetaHDagger()

	historyDagger = append(historyDagger, items)

	// Ensure beta^prime's length not exceed maxBlocksHistory
	if len(historyDagger) >= maxBlocksHistory {
		// Remove old states, with length is maxBlocksHistory
		historyDagger = historyDagger[(len(historyDagger) - maxBlocksHistory):]
	}

	// Set beta^dagger to beta^prime in store
	s.GetPosteriorStates().SetBetaH(historyDagger)
}

// // STF β† ≺ (H, β) (4.6)
// STF β†_H ≺ (H, β_H) (4.6)
func STFBeta2BetaDagger() {
	var (
		s               = store.GetInstance()
		rhc             = NewRecentHistoryController()
		betas           = s.GetPriorStates().GetBeta()
		block           = s.GetLatestBlock()
		parentStateRoot = block.Header.ParentStateRoot
	)
	rhc.Betas = betas
	rhc.RecentHistory2Dagger(parentStateRoot)
}

// func STFBeta2BetaDagger() {
// 	var (
// 		s               = store.GetInstance()
// 		rhc             = NewRecentHistoryController()
// 		betas           = s.GetPriorStates().GetBeta()
// 		block           = s.GetProcessingBlockPointer().GetBlock()
// 		parentStateRoot = block.Header.ParentStateRoot
// 	)
// 	rhc.Betas = betas
// 	rhc.AddToBetaDagger(parentStateRoot)
// }

// // STF β′ ≺ (H, EG, β†, C) (4.7)
// func STFBetaDagger2BetaPrime() {
// 	var (
// 		s          = store.GetInstance()
// 		rhc        = NewRecentHistoryController()
// 		betas      = s.GetIntermediateStates().GetBetaDagger()
// 		block      = s.GetProcessingBlockPointer().GetBlock()
// 		betaB      = s.GetPriorStates().GetLastAccOut()
// 		headerHash = block.Header.Parent
// 		eg         = block.Extrinsic.Guarantees
// 	)
// 	rhc.Betas = betas
// 	accumulationResultTreeRoot := r(betaB)
// 	items := rhc.N(headerHash, eg, accumulationResultTreeRoot)
// 	rhc.AddToBetaPrime(items)
// }

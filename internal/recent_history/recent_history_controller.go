package recent_history

import (
	"sort"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	merkle "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	mmr "github.com/New-JAMneration/JAM-Protocol/internal/utilities/mmr"
)

// RecentHistoryController is a controller for the recent history.
// This controller is used to manage the recent history.
type RecentHistoryController struct {
	Betas types.BlocksHistory
}

// NewRecentHistoryController creates a new RecentHistoryController.
func NewRecentHistoryController() *RecentHistoryController {
	return &RecentHistoryController{
		Betas: types.BlocksHistory{},
	}
}

// \mathbf{C} in GP from type B (12.15)
type BeefyCommitmentOutput []AccumulationOutput // TODO: How to check unique

// Instant-used struct
type AccumulationOutput struct {
	serviceid  types.ServiceId
	commitment types.OpaqueHash
}

var maxBlocksHistory = types.MaxBlocksHistory

// Remove duplicated blocks by BlockHash
func (rhc *RecentHistoryController) CheckDuplicate(headerhash types.HeaderHash) bool {
	// Check if headerhash is already in Recent History Controller
	for _, beta := range rhc.Betas {
		if beta.HeaderHash == headerhash {
			return true
		}
	}
	return false
}

// Beta^dagger (7.2) and STF (4.6)
func (rhc *RecentHistoryController) AddToBetaDagger(header types.Header) {
	// Get recent beta^dagger from store
	betaDagger := store.GetInstance().GetIntermediateStates().GetState().Beta

	if len(rhc.Betas) > 0 {
		// Append first aviod empty slice
		// Duplicate beta into beta^dagger
		betaDagger = append(betaDagger, rhc.Betas...)
		// Except for the stateroot need to be updated
		betaDagger[len(rhc.Betas)-1].StateRoot = header.ParentStateRoot
	}

	// Check beta^dagger is not longer than maxBlocksHistory
	if len(betaDagger) > maxBlocksHistory {
		// Remove old elements to retain maxBlocksHistory
		betaDagger = betaDagger[len(betaDagger)-maxBlocksHistory:]
	}

	// Set beta^dagger to intermediate state in store
	store.GetInstance().GetIntermediateStates().SetBeta(betaDagger)
}

// -----(7.3)-----

// Accumulation-result tree root $r$
func r(c BeefyCommitmentOutput) (accumulationResultTreeRoot types.OpaqueHash) {
	// Empty struct
	pairs := make([]AccumulationOutput, len(c))

	// Extract slices from c
	for i, output := range c {
		pairs[i] = AccumulationOutput{
			serviceid:  output.serviceid,
			commitment: output.commitment,
		}
	}

	// Sort by serviceid $s$
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].serviceid < pairs[j].serviceid
	})

	// Serialization
	var dataSerialized types.ByteSequence
	for _, pair := range pairs {
		serviceidSerialized := utils.SerializeFixedLength(types.U32(pair.serviceid), 4)
		dataSerialized = append(dataSerialized, serviceidSerialized...)

		commitmentSerialized := utils.OpaqueHashWrapper{Value: pair.commitment}.Serialize()
		dataSerialized = append(dataSerialized, commitmentSerialized...)
	}

	// Merklization
	merkle := merkle.Mb([]types.ByteSequence{dataSerialized}, hash.KeccakHash)
	accumulationResultTreeRoot = merkle
	return accumulationResultTreeRoot
}

// Merkle Mountain Range $\mathbf{b}$
func (rhc *RecentHistoryController) b(accumulationResultTreeRoot types.OpaqueHash) (NewMmr []types.MmrPeak) {
	// Only genesis block goto else
	if len(rhc.Betas) != 0 {
		// Else extract each slice of State.beta and then use the latest slice as input of m.AppendOne
		wrappedMmr := mmr.MmrWrapper(&rhc.Betas[len(rhc.Betas)-1].Mmr, hash.KeccakHash)
		// MMR append func $\mathcal{A}$
		NewMmr := wrappedMmr.AppendOne(types.MmrPeak(&accumulationResultTreeRoot))
		return NewMmr
	} else {
		// If State.Beta is empty -> create a new empty MMR
		m := mmr.NewMMR(hash.KeccakHash)
		// MMR append func $\mathcal{A}$
		NewMmr := m.AppendOne(&accumulationResultTreeRoot)
		return NewMmr
	}
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

// item $n$ = (header hash $h$, accumulation-result mmr $\mathbf{b}$, state root $s$, WorkReportHash $\mathbf{p}$)
func (rhc *RecentHistoryController) n(header types.Header, eg types.GuaranteesExtrinsic, c BeefyCommitmentOutput) (items types.BlockInfo) {
	headerHash := header.Parent
	accumulationResultTreeRoot := r(c)
	accumulationResultMmr := rhc.b(accumulationResultTreeRoot)
	workReportHash := p(eg)
	zeroHash := types.StateRoot(types.OpaqueHash{})

	items = types.BlockInfo{
		HeaderHash: headerHash,
		Mmr:        types.Mmr{Peaks: accumulationResultMmr},
		StateRoot:  zeroHash,
		Reported:   workReportHash,
	}
	return items
}

// -----(7.3)-----

// Update beta^dagger to beta^prime (7.4)
func (rhc *RecentHistoryController) AddToBetaPrime(items types.BlockInfo) {
	// Get recent beta^dagger from store
	betaDagger := store.GetInstance().GetIntermediateStates().GetState().Beta

	betaDagger = append(betaDagger, items)

	// Ensure beta^prime's length not exceed maxBlocksHistory
	if len(betaDagger) >= maxBlocksHistory {
		// Remove old states, with length is maxBlocksHistory
		betaDagger = betaDagger[len(betaDagger)-maxBlocksHistory:]
	}

	// Set beta^dagger to beta^prime in store
	store.GetInstance().GetPosteriorStates().SetBeta(betaDagger)
}

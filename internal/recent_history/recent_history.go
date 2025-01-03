package recent_history

import (
	"sort"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	u "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	merkle "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	mmr "github.com/New-JAMneration/JAM-Protocol/internal/utilities/mmr"
)

type State struct {
	Beta       []types.BlockInfo // prior state
	BetaDagger []types.BlockInfo // intermediate state
	BetaPrime  []types.BlockInfo // posterior state
}

// \mathbf{C} in GP from type B (12.15)
type BeefyCommitmentOutput []AccumulationOutput // TODO: How to check unique

// instant-used struct
type AccumulationOutput struct {
	serviceid  types.ServiceId
	commitment types.OpaqueHash
}

var maxBlocksHistory = types.MaxBlocksHistory

// remove duplicated blocks by BlockHash
func (s *State) RemoveDuplicate(headerhash types.HeaderHash) bool {
	for _, block := range s.Beta {
		if block.HeaderHash == headerhash {
			return true
		}
	}
	return false
}

// beta^dagger (7.2) and STF (4.6)
func (s *State) AddToBetaDagger(h types.Header) {
	if len(s.Beta) > 0 {
		if len(s.Beta) != maxBlocksHistory {
			// except
			s.BetaDagger[len(s.Beta)-1].StateRoot = h.ParentStateRoot
		} else {
			// duplicate beta into beta^dagger
			s.BetaDagger = append(s.BetaDagger, s.Beta...)
		}
	}

	// check BetaDagger is not longer than maxBlocksHistory
	if len(s.BetaDagger) > maxBlocksHistory {
		// remove oldest elements to retain maxBlocksHistory
		s.BetaDagger = s.BetaDagger[len(s.BetaDagger)-maxBlocksHistory:]
	}
}

// -----(7.3)--------

// accumulation-result tree root $r$
func r(c BeefyCommitmentOutput) (accumulationResultTreeRoot types.OpaqueHash) {
	// empty struct
	pairs := make([]AccumulationOutput, len(c))

	// extract slices from c
	for i, output := range c {
		pairs[i] = AccumulationOutput{
			serviceid:  output.serviceid,
			commitment: output.commitment,
		}
	}

	// sort by serviceid $s$
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].serviceid < pairs[j].serviceid
	})

	// serialization
	var dataSerialized types.ByteSequence
	for _, pair := range pairs {
		serviceidSerialized := u.SerializeFixedLength(types.U32(pair.serviceid), 4)
		dataSerialized = append(dataSerialized, serviceidSerialized...)

		commitmentSerialized := u.OpaqueHashWrapper{Value: pair.commitment}.Serialize()
		dataSerialized = append(dataSerialized, commitmentSerialized...)
	}

	// Merklization
	merkle := merkle.Mb([]types.ByteSequence{dataSerialized}, hash.KeccakHash)
	accumulationResultTreeRoot = merkle
	return accumulationResultTreeRoot
}

// Merkle Mountain Range $\mathbf{b}$
func (s *State) b(accumulationResultTreeRoot types.OpaqueHash) (NewMmr []types.MmrPeak) {
	// only genesis block goto else
	if len(s.Beta) != 0 {
		// else extract each slice of State.beta and then use the latest slice as input of m.AppendOne
		wrappedMmr := mmr.MmrWrapper(&s.Beta[len(s.Beta)-1].Mmr, hash.KeccakHash)
		// MMR append func $\mathcal{A}$
		NewMmr := wrappedMmr.AppendOne(types.MmrPeak(&accumulationResultTreeRoot))
		return NewMmr
	} else {
		// if State.Beta is empty -> create a new empty MMR
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
func (s *State) n(h types.Header, eg types.GuaranteesExtrinsic, c BeefyCommitmentOutput) (items types.BlockInfo) {
	headerHash := h.Parent // hash.Blake2bHash(u.HeaderSerialization(h))
	accumulationResultTreeRoot := r(c)
	accumulationResultMmr := s.b(accumulationResultTreeRoot)
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

// -----(7.3)--------

// Update BetaDagger to BetaPrime (7.4)
func (s *State) AddToBetaPrime(items types.BlockInfo) {
	// Ensure BetaPrime's length not exceed maxBlocksHistory
	if len(s.BetaPrime) >= maxBlocksHistory {
		// remove old states, with length is maxBlocksHistory
		s.BetaPrime = s.BetaPrime[1:]
	}

	s.BetaPrime = append(s.BetaPrime, items)
}

package recent_history

import (
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	merkle "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/mmr"
)

var maxBlocksHistory = types.MaxBlocksHistory

// Remove duplicated blocks by BlockHash
func CheckDuplicate(blocksHistory types.BlocksHistory, headerhash types.HeaderHash) bool {
	// Check if headerhash is already in Recent History Controller
	for _, blockInfo := range blocksHistory {
		if blockInfo.HeaderHash == headerhash {
			return true
		}
	}
	return false
}

// Beta_H^dagger (7.5) GP 0.6.7
/*
	β†_H ≡ β_H except β†_H [|β_H| − 1]s = H_r
*/
func History2HistoryDagger(history types.BlocksHistory, parentStateRoot types.StateRoot) types.BlocksHistory {
	// Duplicate beta_H into beta_H^dagger
	historyDagger := history

	if len(history) != 0 {
		// Except for the stateroot need to be updated
		historyDagger[len(history)-1].StateRoot = parentStateRoot
	}

	return historyDagger
}

// \mathbf{s} (7.6) GP 0.6.7
/*
	s = [ E_4(s) ⌢ E(h) | (s, h) <− θ′ ]
*/
func serLastAccOut(lastAccOut types.AccumulatedServiceOutput) (types.ByteSequence, error) {
	newEncoder := types.NewEncoder()
	var output types.ByteSequence
	for pair, exist := range lastAccOut {
		if exist {
			data, err := newEncoder.Encode(&pair)
			if err != nil {
				return nil, fmt.Errorf("failed to encode pair: %v", err)
			}
			output = append(output, data...)
		}
	}
	return output, nil
}

// Merkle root from serializedLastAccOut (s) part of (7.7) GP 0.6.7
/*
	MB ( s, HK )
*/
func lastAccOutRoot(serializedLastAccOut types.ByteSequence) types.OpaqueHash {
	return merkle.Mb([]types.ByteSequence{serializedLastAccOut}, hash.KeccakHash)
}

// Append lastAccOutRoot to mmr and form commitment (7.7) GP 0.6.7
/*
	β′_B ≡ A( β_B , MB ( s, HK ), HK )

	b: MR(β′_B)
*/
func appendAndCommitMmr(beefyBelt types.Mmr, merkleRoot types.OpaqueHash) (types.Mmr, types.OpaqueHash) {
	var Mmr *mmr.MMR
	if len(beefyBelt.Peaks) == 0 {
		Mmr = mmr.NewMMR(hash.KeccakHash)
	} else {
		Mmr = mmr.NewMMRFromPeaks(beefyBelt.Peaks, hash.KeccakHash)
	}
	log.Printf("mmr struct before append: %+v", Mmr)
	beefybeltPrime := Mmr.AppendOne(types.MmrPeak(&merkleRoot))
	log.Printf("mmr struct before superpeak: %+v", Mmr)
	return types.Mmr{Peaks: beefybeltPrime}, Mmr.SuperPeak(beefybeltPrime)
}

// The set of work reports $\mathbf{p}$ (7.8) GP 0.6.7
/*
	p = { ((g_w)s)h ↦ ((g_w)s)e | g ∈ EG }
*/
func mapWorkReportFromEg(eg types.GuaranteesExtrinsic) []types.ReportedWorkPackage {
	var reports []types.ReportedWorkPackage
	// Create a map from eg.Report.PackageSpec.Hash to eg.Report.PackageSpec.ExportsRoot
	for _, eg := range eg {
		report := types.ReportedWorkPackage{
			Hash:        types.WorkReportHash(eg.Report.PackageSpec.Hash),
			ExportsRoot: eg.Report.PackageSpec.ExportsRoot,
		}
		reports = append(reports, report)
	}
	return reports
}

// pack item $n$ (7.8) GP 0.6.7
/*
	item $n$ = (header hash $h$, accumulation-result mmr $b$, state root $s$, WorkReportHash $\mathbf{p}$)
*/
func newItem(headerHash types.HeaderHash, workReportHash []types.ReportedWorkPackage, accumulationResultMmr types.OpaqueHash) (item types.BlockInfo) {
	zeroHash := types.StateRoot{}

	item = types.BlockInfo{
		HeaderHash: headerHash,
		BeefyRoot:  accumulationResultMmr,
		StateRoot:  zeroHash,
		Reported:   workReportHash,
	}
	return item
}

// Update beta^dagger to beta^prime (7.8) GP 0.6.7
/*
	β′_H ≡ β†_H cat. ( p, h = H(H), b = MR(β′_B ), s = H^0 )
*/
func AddItem2BetaHPrime(historyDagger types.BlocksHistory, item types.BlockInfo) types.BlocksHistory {
	historyPrime := append(historyDagger, item)

	// Ensure beta^prime's length not exceed maxBlocksHistory
	if historyPrime.Validate() != nil {
		// Remove old states, with length is maxBlocksHistory
		historyPrime = historyPrime[(len(historyPrime) - maxBlocksHistory):]
	}

	return historyPrime
}

// STF β†_H ≺ (H, β_H) (4.6)
func STFBetaH2BetaHDagger() {
	var (
		s     = store.GetInstance()
		beta  = s.GetPriorStates().GetBeta()
		block = s.GetLatestBlock()
	)
	if beta.History.Validate() != nil {
		log.Fatalf("beta.History.Validate() failed: %v", beta.History.Validate())
	}
	betaDagger := History2HistoryDagger(beta.History, block.Header.ParentStateRoot)

	s.GetIntermediateStates().SetBetaHDagger(betaDagger)
}

// STF β′_H ≺ (H, EG, β†_H, C) (4.7)
func STFBetaHDagger2BetaHPrime() error {
	var (
		store         = store.GetInstance()
		historyDagger = store.GetIntermediateStates().GetBetaHDagger()
		beefyBelt     = store.GetPriorStates().GetBeta().Mmr
		lastAccOut    = store.GetPosteriorStates().GetLastAccOut()
		block         = store.GetLatestBlock()
	)
	serializedLastAccOut, err := serLastAccOut(lastAccOut)
	if err != nil {
		return err
	}
	merkleRoot := lastAccOutRoot(serializedLastAccOut)
	beefyBeltPrime, commitment := appendAndCommitMmr(beefyBelt, merkleRoot)
	workReportHash := mapWorkReportFromEg(block.Extrinsic.Guarantees)
	item := newItem(block.Header.Parent, workReportHash, commitment)
	historyPrime := AddItem2BetaHPrime(historyDagger, item)

	// Set beta_B^prime and beta_H^prime to store
	store.GetPosteriorStates().SetBetaB(beefyBeltPrime)
	store.GetPosteriorStates().SetBetaH(historyPrime)
	return nil
}

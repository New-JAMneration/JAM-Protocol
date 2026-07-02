package statistics

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestValidatorStats_AssurancesBeforeEpochRollover covers the GP v0.8.0
// activity-spec restructure: assurances are credited to pi_V-dagger BEFORE the
// epoch rollover, so on an epoch-boundary block they land in pi_L' (and the
// fresh pi_V' stays at zero); on a same-epoch block they land in pi_V' as
// before.
func TestValidatorStats_AssurancesBeforeEpochRollover(t *testing.T) {
	const assurer = 1

	setup := func(priorTau, postTau types.TimeSlot) *blockchain.ChainState {
		blockchain.ResetInstance()
		cs := blockchain.GetInstance()

		cs.GetPriorStates().SetTau(priorTau)
		cs.GetPosteriorStates().SetTau(postTau)
		cs.GetPriorStates().SetPi(types.Statistics{
			ValsCurr: make(types.ValidatorsStatistics, types.ValidatorsCount),
			ValsLast: make(types.ValidatorsStatistics, types.ValidatorsCount),
		})

		cs.AddBlock(types.Block{
			Header: types.Header{
				Slot:        postTau,
				AuthorIndex: 0,
			},
			Extrinsic: types.Extrinsic{
				Assurances: types.AssurancesExtrinsic{
					{
						ValidatorIndex: assurer,
						// CalculatePopularity indexes Bitfield per core.
						Bitfield: make(types.Bitfield, types.CoresCount),
					},
				},
			},
		})
		return cs
	}

	epochLen := types.TimeSlot(types.EpochLength)

	t.Run("epoch boundary: assurance rolls into pi_L'", func(t *testing.T) {
		cs := setup(epochLen-1, epochLen) // epoch 0 -> epoch 1

		UpdateValidatorActivityStatistics()

		pi := cs.GetPosteriorStates().GetPi()
		if got := pi.ValsLast[assurer].Assurances; got != 1 {
			t.Errorf("ValsLast[%d].Assurances = %d, want 1 (pi_V-dagger rolls into pi_L')", assurer, got)
		}
		if got := pi.ValsCurr[assurer].Assurances; got != 0 {
			t.Errorf("ValsCurr[%d].Assurances = %d, want 0 (fresh accumulator)", assurer, got)
		}
	})

	t.Run("same epoch: assurance lands in pi_V'", func(t *testing.T) {
		cs := setup(epochLen, epochLen+1) // both in epoch 1

		UpdateValidatorActivityStatistics()

		pi := cs.GetPosteriorStates().GetPi()
		if got := pi.ValsCurr[assurer].Assurances; got != 1 {
			t.Errorf("ValsCurr[%d].Assurances = %d, want 1", assurer, got)
		}
		if got := pi.ValsLast[assurer].Assurances; got != 0 {
			t.Errorf("ValsLast[%d].Assurances = %d, want 0", assurer, got)
		}
	})
}

package safrole

import (
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// 6.6 The Markers (graypaper 0.5.4)

func getCurrentTau() types.TimeSlot {
	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	return tauPrime
}

func CreateEpochMarker() {
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorState()

	// Get previous time slot index
	tau := priorState.Tau

	// Get current time slot
	tauPrime := getCurrentTau()

	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)

	if ePrime > e {
		// New epoch, create epoch marker
		// Get eta_0, eta_1
		eta_0 := priorState.Eta[0]
		eta_1 := priorState.Eta[1]

		// Get gamma_k from posterior state
		gamma_k := s.GetPosteriorState().Gamma.GammaK

		// Get bandersnatch key from gamma_k
		bandersnatchKeys := []types.BandersnatchPublic{}
		for _, validator := range gamma_k {
			bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
		}

		epochMarker := &types.EpochMark{
			Entropy:        eta_0,
			TicketsEntropy: eta_1,
			Validators:     bandersnatchKeys,
		}

		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
	} else {
		// The epoch is the same
		var epochMarker *types.EpochMark = nil
		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
	}
}

func CreateWinningTickets() {
	// TODO: Implement create winning tickets
}

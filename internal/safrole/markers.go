// 6.6 The Markers (graypaper 0.5.4)
package safrole

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CreateEpochMarker creates the epoch marker
// (6.27)
func CreateEpochMarker() {
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorState()

	// Get posterior state
	posteriorState := s.GetPosteriorState()

	// Get previous time slot index
	tau := priorState.Tau

	// Get current time slot index
	tauPrime := posteriorState.Tau

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

// CreateWinningTickets creates the winning tickets
// (6.28)
func CreateWinningTickets() {
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorState()

	// Get posterior state
	posteriorState := s.GetPosteriorState()

	// Get previous time slot index
	tau := priorState.Tau

	// Get current time slot index
	tauPrime := posteriorState.Tau

	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)

	m := GetSlotIndex(tau)
	mPrime := GetSlotIndex(tauPrime)

	gamma_a := priorState.Gamma.GammaA

	condition1 := ePrime == e
	condition2 := m < types.TimeSlot(types.Y) && mPrime >= types.TimeSlot(types.Y)
	condition3 := len(gamma_a) == types.EpochLength

	if condition1 && condition2 && condition3 {
		// Z(gamma_a)
		ticketsMark := types.TicketsMark(OutsideInSequencer(&gamma_a))
		s.GetIntermediateHeaderPointer().SetTicketsMark(&ticketsMark)
	} else {
		// The epoch is the same
		var ticketsMark *types.TicketsMark = nil
		s.GetIntermediateHeaderPointer().SetTicketsMark(ticketsMark)
	}
}

// // 6.6 The Markers (graypaper 0.5.4)
package safrole

// import (
// 	"github.com/New-JAMneration/JAM-Protocol/internal/store"
// 	"github.com/New-JAMneration/JAM-Protocol/internal/types"
// 	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
// )

// // CreateEpochMarker creates the epoch marker
// // (6.27)
// func CreateEpochMarker() *types.ErrorCode {
// 	s := store.GetInstance()

// 	// Get previous time slot index
// 	tau := s.GetPriorStates().GetTau()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	// prior time slot must be less than posterior time slot
// 	if tau >= tauPrime {
// 		err := SafroleErrorCode.BadSlot
// 		return &err
// 	}

// 	if ePrime > e {
// 		// New epoch, create epoch marker
// 		// Get eta_0, eta_1
// 		eta := s.GetPriorStates().GetEta()

// 		// Get gamma_k from posterior state
// 		gammaK := s.GetPosteriorStates().GetGammaK()

// 		// Get bandersnatch key from gamma_k
// 		bandersnatchKeys := []types.BandersnatchPublic{}
// 		for _, validator := range gammaK {
// 			bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
// 		}

// 		epochMarker := &types.EpochMark{
// 			Entropy:        eta[0],
// 			TicketsEntropy: eta[1],
// 			Validators:     bandersnatchKeys,
// 		}

// 		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
// 	} else {
// 		// The epoch is the same
// 		var epochMarker *types.EpochMark = nil
// 		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
// 	}

// 	return nil
// }

// // CreateWinningTickets creates the winning tickets
// // (6.28)
// func CreateWinningTickets() {
// 	s := store.GetInstance()

// 	// Get previous time slot index
// 	tau := s.GetPriorStates().GetTau()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	m := GetSlotIndex(tau)
// 	mPrime := GetSlotIndex(tauPrime)

// 	gammaA := s.GetPriorStates().GetGammaA()

// 	condition1 := ePrime == e
// 	condition2 := m < types.TimeSlot(types.SlotSubmissionEnd) && mPrime >= types.TimeSlot(types.SlotSubmissionEnd)
// 	condition3 := len(gammaA) == types.EpochLength

// 	if condition1 && condition2 && condition3 {
// 		// Z(gamma_a)
// 		ticketsMark := types.TicketsMark(OutsideInSequencer(&gammaA))
// 		s.GetIntermediateHeaderPointer().SetTicketsMark(&ticketsMark)
// 	} else {
// 		// The epoch is the same
// 		var ticketsMark *types.TicketsMark = nil
// 		s.GetIntermediateHeaderPointer().SetTicketsMark(ticketsMark)
// 	}
// }

package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func UpdateSafrole() error {

	// --- STEP 1 Get Epoch --- //

	s := store.GetInstance()
	tau := s.GetPriorStates().GetTau()
	tauPrime := s.GetPosteriorStates().GetTau()
	e := safrole.GetEpochIndex(tau)
	ePrime := safrole.GetEpochIndex(tauPrime)

	// --- STEP 2 Update Entropy123 --- //
	// (GP 6.23)
	safrole.UpdateEntropy()

	// --- STEP 3 safrole.go (GP 6.2, 6.13, 6.14) --- //
	// (6.2, 6.13, 6.14)
	// This function will update GammaK, GammaZ, Lambda, Kappa
	/*
		KeyRotate()
	*/

	// Get prior state
	priorState := s.GetPriorStates()

	// Get previous time slot index
	// tau := priorState.GetTau()

	// Get current time slot
	// now := time.Now().UTC()
	// timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	// tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))
	// tauPrime := testCase.Input.Slot
	// s.GetPosteriorStates().SetTau(tauPrime)
	// Execute key rotation
	// newSafroleState := keyRotation(tau, tauPrime, priorState.GetState())
	// e := GetEpochIndex(tau)
	// ePrime := GetEpochIndex(tauPrime)
	// t.Log("tau", tau, "-> tauPrime", tauPrime)
	// t.Log("e: ", e, "-> ePrime: ", ePrime)

	if ePrime > e {
		// Update state to posterior state
		s.GetPosteriorStates().SetGammaK(safrole.ReplaceOffenderKeys(priorState.GetIota()))
		s.GetPosteriorStates().SetKappa(priorState.GetGammaK())
		s.GetPosteriorStates().SetLambda(priorState.GetKappa())
		z, zErr := safrole.UpdateBandersnatchKeyRoot(s.GetPosteriorStates().GetGammaK())
		if zErr != nil {
			log.Printf("Error updating Bandersnatch key root: %v\n", zErr)
		}
		s.GetPosteriorStates().SetGammaZ(z)
	} else {
		s.GetPosteriorStates().SetGammaK(priorState.GetGammaK())
		s.GetPosteriorStates().SetKappa(priorState.GetKappa())
		s.GetPosteriorStates().SetLambda(priorState.GetLambda())
		s.GetPosteriorStates().SetGammaZ(priorState.GetGammaZ())
	}

	// (GP 6.22) This is done in Dump()
	// UpdateEtaPrime0()

	// (GP 6.17)
	// UpdateHeaderEntropy()

	// --- slot_key_sequence.go (GP 6.25, 6.26) --- //
	safrole.UpdateSlotKeySequence()

	// --- STEP 4 Check TicketExtrinsic --- //
	// --- extrinsic_tickets.go (GP 6.30~6.34) --- //
	ourEtErr := safrole.CreateNewTicketAccumulator()
	if ourEtErr != nil {
		return ourEtErr
	}
	// (GP 6.28)
	safrole.CreateWinningTickets()

	// --- sealing.go (GP 6.15~6.24) --- //
	safrole.SealingByBandersnatchs()
	safrole.SealingByTickets()

	// --- markers.go (GP 6.27, 6.28) --- //
	// (GP 6.27)
	ourEpochMarkErr := safrole.CreateEpochMarker()
	if ourEpochMarkErr != nil {
		return ourEpochMarkErr
	}
	return nil
}

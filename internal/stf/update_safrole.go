package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
)

func UpdateSafrole() error {

	// Key rotation
	// This function will update GammaK, GammaZ, Lambda, Kappa
	safrole.KeyRotate()

	ourEtErr := safrole.CreateNewTicketAccumulator()
	if ourEtErr != nil {
		// s.GetIntermediateStates().SetTauInput(s.GetPriorStates().GetTau())
		return ourEtErr
	}

	// Sealing
	safrole.SealingByTickets()
	safrole.SealingByBandersnatchs()

	// Update Eta'0
	safrole.UpdateEtaPrime0()

	// Entropy update
	// This function will update Eta
	safrole.UpdateEntropy()

	// Update Header Entropy
	safrole.UpdateHeaderEntropy()

	// Update GammaS
	safrole.UpdateSlotKeySequence()

	err := safrole.CreateEpochMarker()
	if err != nil {
		return err
	}
	safrole.CreateWinningTickets()

	return nil
}

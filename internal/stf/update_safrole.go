package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
)

func UpdateSafrole() error {

	// Key rotation
	// This function will update GammaK, GammaZ, Lambda, Kappa
	safrole.KeyRotate()

	// Create new ticket accumulator
	// This function will update GammaA

	// will got the VRF error
	// Failed to create verifier: invalid input
	// safrole.CreateNewTicketAccumulator()

	// Update GammaS
	// safrole.UpdateSlotKeySequence()

	// Entropy update
	// This function will update Eta
	// safrole.UpdateEntropy()

	// Update Eta'0
	// safrole.UpdateEtaPrime0()

	// Update Header Entropy
	// safrole.UpdateHeaderEntropy()

	return nil
}

package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
)

func UpdateHistory() error {
	log.Println("Update History")

	// Start test STFBeta2BetaDagger (4.6)
	recent_history.STFBetaH2BetaHDagger()

	// Start test STFBetaDagger2BetaPrime (4.7)
	if err := recent_history.STFBetaHDagger2BetaHPrime_ForTestVector(); err != nil {
		return err
	}

	return nil
}

package stf

import "github.com/New-JAMneration/JAM-Protocol/internal/recent_history"

func UpdateHistory() error {
	// log.Println("Update History")

	// Start test STFBeta2BetaDagger (4.6)
	// We update (4.6) at the beginning of STF
	// recent_history.STFBetaH2BetaHDagger()

	// Start test STFBetaDagger2BetaPrime (4.7)
	/*
		// for stf test-vector
		if err := recent_history.STFBetaHDagger2BetaHPrime_ForTestVector(); err != nil {
			return err
		}
	*/
	// for traces
	if err := recent_history.STFBetaHDagger2BetaHPrime(); err != nil {
		return err
	}
	return nil
}

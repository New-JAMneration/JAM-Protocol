package stf

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	AssurancesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/assurances"
	DisputesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/disputes"
	PreimagesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/preimages"
	ReportsErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
	SafroleErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

// measureTime measures execution time of a function and logs it
func measureTime(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	if err != nil {
		fmt.Printf("⏱️  %-40s took: %12v (ERROR: %v)\n", operation, duration, err)
	} else {
		fmt.Printf("⏱️  %-40s took: %12v\n", operation, duration)
	}
	return err
}

// measureTimeNoError measures execution time of a function that doesn't return error
func measureTimeNoError(operation string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	fmt.Printf("⏱️  %-40s took: %12v\n", operation, duration)
}

// TODO: Implement the following functions to handle state transitions
// Each function should update the corresponding state in the data store
// The functions should validate inputs and handle errors appropriately
// Consider adding proper logging and metrics collection
func isProtocolError(err error) bool {
	if _, ok := err.(*types.ErrorCode); ok {
		// This is a protocol-level error → block invalid
		return true
	}

	// Runtime error → unexpected bug
	return false
}

func RunSTF() (bool, error) {
	totalStart := time.Now()
	defer func() {
		fmt.Printf("\n⏱️  %-40s Total STF took: %12v\n", "RunSTF", time.Since(totalStart))
	}()

	st := store.GetInstance()

	// Update timeslot
	st.GetPosteriorStates().SetTau(st.GetLatestBlock().Header.Slot)

	// update BetaH, GP 0.6.7 formula 4.6
	measureTimeNoError("STFBetaH2BetaHDagger", func() {
		recent_history.STFBetaH2BetaHDagger()
	})

	// Update Disputes
	err := measureTime("UpdateDisputes", func() error {
		return UpdateDisputes()
	})
	if err != nil {
		errorMessage := DisputesErrorCodes.DisputesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Safrole
	err = measureTime("UpdateSafrole", func() error {
		return UpdateSafrole()
	})
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	priorState := st.GetPriorStates().GetState()

	// Validate Non-VRF Header(H_E, H_W, H_O)
	err = measureTime("ValidateNonVRFHeader", func() error {
		return ValidateNonVRFHeader(st.GetLatestBlock().Header, &priorState)
	})
	if err != nil {
		return isProtocolError(err), fmt.Errorf("header validate error: %v", err)
	}

	// Update Assurances
	err = measureTime("UpdateAssurances", func() error {
		return UpdateAssurances()
	})
	if err != nil {
		errorMessage := AssurancesErrorCodes.AssurancesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Reports
	err = measureTime("UpdateReports", func() error {
		return UpdateReports()
	})
	if err != nil {
		errorMessage := ReportsErrorCodes.ReportsErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Accumlate
	err = measureTime("UpdateAccumlate", func() error {
		return UpdateAccumlate()
	})
	if err != nil {
		return isProtocolError(err), fmt.Errorf("accumulate error: %v", err)
	}

	// Update History (beta^dagger -> beta^prime)
	// err = UpdateHistory()
	err = measureTime("STFBetaHDagger2BetaHPrime", func() error {
		return recent_history.STFBetaHDagger2BetaHPrime()
	})
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update history error: %v", err)
	}

	// Update Preimages
	err = measureTime("UpdatePreimages", func() error {
		return UpdatePreimages()
	})
	if err != nil {
		errorMessage := PreimagesErrorCodes.PreimagesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Authorization
	err = measureTime("UpdateAuthorizations", func() error {
		return UpdateAuthorizations()
	})
	if err != nil {
		return isProtocolError(err), fmt.Errorf("authorization error: %v", err)
	}

	// Update Statistics
	err = measureTime("UpdateStatistics", func() error {
		return UpdateStatistics()
	})
	if err != nil {
		return isProtocolError(err), fmt.Errorf("statistics error: %v", err)
	}

	return false, nil
}

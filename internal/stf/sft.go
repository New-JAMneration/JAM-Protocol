package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	AssurancesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/assurances"
	DisputesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/disputes"
	PreimagesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/preimages"
	ReportsErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
	SafroleErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

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
	var (
		err        error
		st         = store.GetInstance()
		priorState = st.GetPriorStates().GetState()
		header     = st.GetLatestBlock().Header
	)

	// Update timeslot
	st.GetPosteriorStates().SetTau(header.Slot)

	// Validate Non-VRF Header(H_E, H_W, H_O, H_I)
	err = ValidateNonVRFHeader(header, &priorState)
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// update BetaH, GP 0.6.7 formula 4.6
	recent_history.STFBetaH2BetaHDagger()

	// Update Disputes
	err = UpdateDisputes()
	if err != nil {
		errorMessage := DisputesErrorCodes.DisputesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}
	postState := st.GetPosteriorStates().GetState()

	// After keyRotate
	err = ValidateHeaderVrf(header, &postState)
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		errorMessage := AssurancesErrorCodes.AssurancesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Reports
	err = UpdateReports()
	if err != nil {
		errorMessage := ReportsErrorCodes.ReportsErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("accumulate error: %v", err)
	}

	// Update History (beta^dagger -> beta^prime)
	err = recent_history.STFBetaHDagger2BetaHPrime()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update history error: %v", err)
	}

	// Update Preimages
	err = UpdatePreimages()
	if err != nil {
		errorMessage := PreimagesErrorCodes.PreimagesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Authorization
	err = UpdateAuthorizations()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("authorization error: %v", err)
	}

	// Update Statistics
	err = UpdateStatistics()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("statistics error: %v", err)
	}

	return false, nil
}

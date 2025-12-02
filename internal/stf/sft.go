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
	var err error
	st := store.GetInstance()

	// Update timeslot
	st.GetPosteriorStates().SetTau(st.GetLatestBlock().Header.Slot)

	// update BetaH, GP 0.6.7 formula 4.6
	recent_history.STFBetaH2BetaHDagger()
	priorState := st.GetPriorStates().GetState()

	block := st.GetLatestBlock()
	header := block.Header

	err = ValidateHeader(header, &priorState)
	if err != nil {
		// INFO: Now, we use Safrole error codes for all header validation errors.
		// If needed, we should manage the error codes for header validation separately.

		// const errPrefix = "block header verification failure: "
		// errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)

		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	err = ValidateBlock(block)
	if err != nil {
		// const errPrefix = "block verification failure: "
		// TODO: Implement proper block error codes
		errorMessage := ""
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Disputes
	err = UpdateDisputes()
	if err != nil {
		// const errPrefix = "disputes error: "
		// errorMessage := DisputesErrorCodes.DisputesErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)
		errorMessage := DisputesErrorCodes.DisputesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		// const errPrefix = "safrole error: "
		// errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Validate Header Vrf(seal, entropy)
	stateForHeaderValidate := st.GetPosteriorStates().GetState()
	err = ValidateHeaderVrf(header, &stateForHeaderValidate)
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Validate Header (other fields)
	err = ValidateHeader(header, &priorState)
	if err != nil {
		// INFO: Now, we use Safrole error codes for all header validation errors.
		// If needed, we should manage the error codes for header validation separately.

		// const errPrefix = "block header verification failure: "
		// errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)

		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		// const errPrefix = "assurances error: "
		// errorMessage := AssurancesErrorCodes.AssurancesErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)
		errorMessage := AssurancesErrorCodes.AssurancesErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Reports
	err = UpdateReports()
	if err != nil {
		// const errPrefix = "reports error: "
		// errorMessage := ReportsErrorCodes.ReportsErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)
		errorMessage := ReportsErrorCodes.ReportsErrorCodeMessages[*err.(*types.ErrorCode)]
		return isProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("accumulate error: %v", err)
	}

	// Update History (beta^dagger -> beta^prime)
	// err = UpdateHistory()
	if err = recent_history.STFBetaHDagger2BetaHPrime(); err != nil {
		return isProtocolError(err), fmt.Errorf("update history error: %v", err)
	}

	// Update Preimages
	err = UpdatePreimages()
	if err != nil {
		// const errPrefix = "preimages error: "
		// errorMessage := PreimagesErrorCodes.PreimagesErrorCodeMessages[*err.(*types.ErrorCode)]
		// return isProtocolError(err), fmt.Errorf("%s%v", errPrefix, errorMessage)
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

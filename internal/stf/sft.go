package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	AssurancesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/assurances"
	DisputesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/disputes"
	PreimagesErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/preimages"
	ReportsErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
	SafroleErrorCodes "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

// IsProtocolError checks if an error is a protocol-level error (defined ErrorCode)
// Returns:
//   - true:  Protocol error → block is invalid, but node should continue processing other blocks
//   - false: Runtime error → unexpected bug, node should terminate
func IsProtocolError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*types.ErrorCode); ok {
		// This is a protocol-level error → block invalid
		return true
	}

	// Runtime error → unexpected bug
	return false
}

// RunSTF executes the State Transition Function
// Returns:
//   - (true, error):  Protocol error - block is invalid but node should continue
//   - (false, error): Runtime error - unexpected bug, node should terminate
//   - (false, nil):   Success - block processed successfully
func RunSTF() (bool, error) {
	var (
		err              error
		st               = store.GetInstance()
		priorState       = st.GetPriorStates().GetState()
		header           = st.GetLatestBlock().Header
		extrinsic        = st.GetLatestBlock().Extrinsic
		unmatchedKeyVals = st.GetPriorStateUnmatchedKeyVals()
	)

	// Update timeslot
	st.GetPosteriorStates().SetTau(header.Slot)

	// Validate Non-VRF Header(H_E, H_W, H_O, H_I)
	err = ValidateNonVRFHeader(header, &priorState, extrinsic)
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// update BetaH, GP 0.6.7 formula 4.6
	recent_history.STFBetaH2BetaHDagger()

	// Update Disputes
	err = UpdateDisputes()
	if err != nil {
		errorMessage := DisputesErrorCodes.DisputesErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}
	postState := st.GetPosteriorStates().GetState()

	// After keyRotate
	err = ValidateHeaderVrf(header, &postState)
	if err != nil {
		errorMessage := SafroleErrorCodes.SafroleErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Validate extrinsic
	err = ValidateExtrinsic(extrinsic, &priorState, unmatchedKeyVals)
	if err != nil {
		var errorMessage string
		// Determine the error message based on the type of error
		if ec, ok := err.(*types.ErrorCode); ok {
			if msg, exists := PreimagesErrorCodes.PreimagesErrorCodeMessages[*ec]; exists {
				errorMessage = msg
			} else if msg, exists := AssurancesErrorCodes.AssurancesErrorCodeMessages[*ec]; exists {
				errorMessage = msg
			} else if msg, exists := ReportsErrorCodes.ReportsErrorCodeMessages[*ec]; exists {
				errorMessage = msg
			} else if msg, exists := DisputesErrorCodes.DisputesErrorCodeMessages[*ec]; exists {
				errorMessage = msg
			} else if msg, exists := SafroleErrorCodes.SafroleErrorCodeMessages[*ec]; exists {
				errorMessage = msg
			} else {
				errorMessage = fmt.Sprintf("runtime error: %v", err)
			}
		}
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		errorMessage := AssurancesErrorCodes.AssurancesErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Reports
	err = UpdateReports()
	if err != nil {
		errorMessage := ReportsErrorCodes.ReportsErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return IsProtocolError(err), fmt.Errorf("%v", err)
	}

	// Update History (beta^dagger -> beta^prime)
	err = recent_history.STFBetaHDagger2BetaHPrime()
	if err != nil {
		return IsProtocolError(err), fmt.Errorf("update history error: %v", err)
	}

	// Update Preimages
	err = accumulation.ProcessPreimageExtrinsics()
	if err != nil {
		errorMessage := PreimagesErrorCodes.PreimagesErrorCodeMessages[*err.(*types.ErrorCode)]
		return IsProtocolError(err), fmt.Errorf("%v", errorMessage)
	}

	// Update Authorization
	err = UpdateAuthorizations()
	if err != nil {
		return IsProtocolError(err), fmt.Errorf("%v", err)
	}

	// Update Statistics
	err = UpdateStatistics()
	if err != nil {
		return IsProtocolError(err), fmt.Errorf("%v", err)
	}

	return false, nil
}

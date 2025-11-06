package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

	st := store.GetInstance()
	{
		// Validate Header Seal
		priorState := st.GetPriorStates().GetState()
		header := st.GetLatestBlock().Header
		err := safrole.ValidateHeaderSeal(header, &priorState)
		if err != nil {
			return true, fmt.Errorf("validate header seal error: %v", err)
		}
	}
	// Update timeslot
	st.GetPosteriorStates().SetTau(st.GetLatestBlock().Header.Slot)
	// update BetaH, GP 0.6.7 formula 4.6
	recent_history.STFBetaH2BetaHDagger()
	// Update Disputes
	err := UpdateDisputes()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update disputes error: %v", err)
	}

	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update safrole error: %v", err)
	}

	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update assurances error: %v", err)
	}
	// Update Reports
	err = UpdateReports()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update reports error: %v", err)
	}

	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update accumulate error: %v", err)
	}

	// Update History (beta^dagger -> beta^prime)
	err = UpdateHistory()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update histroy error: %v", err)
	}

	// Update Preimages
	err = UpdatePreimages()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update preimages error: %v", err)
	}

	// Update Authorization
	err = UpdateAuthorizations()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update authorization error: %v", err)
	}

	// Update Statistics
	err = UpdateStatistics()
	if err != nil {
		return isProtocolError(err), fmt.Errorf("update statistics error: %v", err)
	}

	return false, nil
}

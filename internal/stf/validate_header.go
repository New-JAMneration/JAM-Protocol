package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// TODO: Align the official errorCode
func ValidateNonVRFHeader(header types.Header, priorState *types.State, extrinsic types.Extrinsic) error {
	if err := safrole.ValidateHeaderTicketsMark(header, priorState); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderOffenderMarker(header, priorState); err != nil {
		return err
	}

	if err := validateExtrinsicHash(header, extrinsic); err != nil {
		return err
	}

	// H_R
	unmatchedKeyVals := blockchain.GetInstance().GetPriorStateUnmatchedKeyVals()
	serializedState, _ := m.StateEncoder(*priorState)
	fullStateKeyVals := append(serializedState, unmatchedKeyVals...)
	priorStateRoot := m.MerklizationSerializedState(fullStateKeyVals)
	if header.ParentStateRoot != priorStateRoot {
		errCode := SafroleErrorCode.InvalidParentStateRoot
		return &errCode
	}

	// Validate author_index out of range.
	// NOTE: There is currently no official error code defined for this case.
	// We may need to update this once the spec updates.
	if header.AuthorIndex >= types.ValidatorIndex(len(priorState.Kappa)) {
		errCode := SafroleErrorCode.AuthorIndexOutOfRange
		return &errCode
	}
	return nil
}

func ValidateHeaderVrf(header types.Header, priorState *types.State, posteriorState *types.State) error {
	if err := safrole.ValidateHeaderEpochMark(header, priorState, posteriorState); err != nil {
		return err
	}
	if err := safrole.ValidateHeaderSeal(header, posteriorState); err != nil {
		return err
	}
	if err := safrole.ValidateHeaderEntropy(header, posteriorState); err != nil {
		return err
	}
	return nil
}

func validateExtrinsicHash(header types.Header, extrinsic types.Extrinsic) error {
	actualExtrinsicHash, err := utilities.CreateExtrinsicHash(extrinsic)
	if err != nil {
		return fmt.Errorf("failed to create extrinsic hash: %v", err)
	}

	expectedExtrinsicHash := header.ExtrinsicHash

	if actualExtrinsicHash != expectedExtrinsicHash {
		errCode := SafroleErrorCode.InvalidExtrinsicHash
		return &errCode
	}

	return nil
}

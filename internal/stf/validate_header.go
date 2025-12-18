package stf

import (
	"fmt"

	HeaderController "github.com/New-JAMneration/JAM-Protocol/internal/header"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

// TODO: Align the official errorCode
func ValidateNonVRFHeader(header types.Header, state *types.State, extrinsic types.Extrinsic) error {
	if err := safrole.ValidateHeaderEpochMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderTicketsMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderOffenderMarker(header, state); err != nil {
		return err
	}

	if err := validateExtrinsicHash(header, extrinsic); err != nil {
		return err
	}

	// Validate author_index out of range.
	// NOTE: There is currently no official error code defined for this case.
	// We may need to update this once the spec updates.
	if header.AuthorIndex >= types.ValidatorIndex(len(state.Kappa)) {
		errCode := SafroleErrorCode.AuthorIndexOutOfRange
		return &errCode
	}
	return nil
}

func ValidateHeaderVrf(header types.Header, state *types.State) error {
	if err := safrole.ValidateHeaderSeal(header, state); err != nil {
		return err
	}
	if err := safrole.ValidateHeaderEntropy(header, state); err != nil {
		return err
	}
	return nil
}

func validateExtrinsicHash(header types.Header, extrinsics types.Extrinsic) *types.ErrorCode {
	headerController := HeaderController.NewHeaderController()
	err := headerController.CreateExtrinsicHash(extrinsics)
	if err != nil {
		fmt.Println("Error creating extrinsic hash:", err)
	}

	extrinsicHash := store.GetInstance().GetProcessingBlockPointer().GetExtrinsicHash()

	if extrinsicHash != header.ExtrinsicHash {
		errCode := SafroleErrorCode.InvalidExtrinsicHash
		return &errCode
	}

	return nil
}

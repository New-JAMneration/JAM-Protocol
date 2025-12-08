package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

// TODO: Align the official errorCode
func ValidateHeader(header types.Header, state *types.State) error {
	err := safrole.ValidateHeaderEpochMark(header, state)
	if err != nil {
		return err
	}

	if err := safrole.ValidateHeaderTicketsMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderOffenderMarker(header, state); err != nil {
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
	err := safrole.ValidateHeaderSeal(header, state)
	if err != nil {
		return err
	}
	err = safrole.ValidateHeaderEntropy(header, state)
	if err != nil {
		return err
	}
	return nil
}

package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TODO: Align the official errorCode
// H_E,H_W, H_O
func ValidateNonVRFHeader(header types.Header, state *types.State) error {
	err := safrole.ValidateHeaderEpochMark(header, state)
	if err != nil {
		return err
	}

	if err := safrole.ValidateHeaderTicketsMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderOffenderMarker(header); err != nil {
		return err
	}

	return nil
}

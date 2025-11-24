package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TODO: Align the official errorCode
func ValidateHeader(header types.Header, state *types.State) error {
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

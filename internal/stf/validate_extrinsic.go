package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ValidateExtrinsic checks the structural validity of the block's extrinsic
// data against the prior state.
func ValidateExtrinsic(extrinsic types.Extrinsic, state *types.State) error {
	eps := extrinsic.Preimages
	delta := state.Delta

	// Preimage
	err := accumulation.ValidatePreimageExtrinsics(eps, delta)
	if err != nil {
		return err
	}

	return nil
}

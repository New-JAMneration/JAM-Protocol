package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func ValidateExtrinsic(extrinsic types.Extrinsic, state *types.State, unmatchedKeyVals types.StateKeyVals) error {
	eps := extrinsic.Preimages
	delta := state.Delta

	// Preimage
	err := accumulation.ValidatePreimageExtrinsics(eps, delta, &unmatchedKeyVals)
	if err != nil {
		return err
	}

	return nil
}

package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ValidateExtrinsic checks the structural validity of the block's extrinsic
// data against the prior state.
//
// The legacy unmatchedKeyVals parameter is retained for source compatibility
// during the globalKV transition but is no longer consulted (Method A loads
// the full state into globalKV at deserialization, so no fallback pool
// lookup is needed). Step 7.5 will remove the parameter entirely.
func ValidateExtrinsic(extrinsic types.Extrinsic, state *types.State, unmatchedKeyVals types.StateKeyVals) error {
	_ = unmatchedKeyVals
	eps := extrinsic.Preimages
	delta := state.Delta

	// Preimage
	err := accumulation.ValidatePreimageExtrinsics(eps, delta)
	if err != nil {
		return err
	}

	return nil
}

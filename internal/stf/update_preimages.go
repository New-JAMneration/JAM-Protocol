package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
)

// UpdatePreimages is only used by jam-test-vectors. For production, please
// use stf.go instead.
//
// Note: the legacy fallback pool (cs.GetPriorStateUnmatchedKeyVals()) is no
// longer consulted under Method A; it will be removed entirely in Step 7.5.
func UpdatePreimages() error {
	cs := blockchain.GetInstance()
	delta := cs.GetPriorStates().GetDelta()
	eps := cs.GetLatestBlock().Extrinsic.Preimages
	// Preimage
	err := accumulation.ValidatePreimageExtrinsics(eps, delta)
	if err != nil {
		return err
	}
	err = accumulation.ProcessPreimageExtrinsics()
	if err != nil {
		return err
	}
	return nil
}

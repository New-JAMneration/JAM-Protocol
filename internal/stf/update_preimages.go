package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
)

// This only used by jam-test-vectors
// For production, please use stf.go instead
func UpdatePreimages() error {
	cs := blockchain.GetInstance()
	delta := cs.GetPriorStates().GetDelta()
	unmatchedKeyVals := cs.GetPriorStateUnmatchedKeyVals()
	eps := cs.GetLatestBlock().Extrinsic.Preimages
	// Preimage
	err := accumulation.ValidatePreimageExtrinsics(eps, delta, &unmatchedKeyVals)
	if err != nil {
		return err
	}
	err = accumulation.ProcessPreimageExtrinsics()
	if err != nil {
		return err
	}
	return nil
}

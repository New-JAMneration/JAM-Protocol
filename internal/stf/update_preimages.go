package stf

import "github.com/New-JAMneration/JAM-Protocol/internal/accumulation"

func UpdatePreimages() error {
	err := accumulation.ProcessPreimageExtrinsics()
	if err != nil {
		return err
	}
	return nil
}

package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
)

func UpdateAccumlate() error {
	// logger.Debug("Update Accumlate")

	// 12.1, 12.2
	err := accumulation.ProcessAccumulation()
	if err != nil {
		return err
	}
	// 12.3
	err = accumulation.DeferredTransfers()
	if err != nil {
		return err
	}
	return nil
}

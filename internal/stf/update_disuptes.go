package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
)

func UpdateDisputes() error {
	// logger.Debug("Update Disputes")

	_, err := extrinsic.Disputes()
	if err != nil {
		return err
	}

	return nil
}

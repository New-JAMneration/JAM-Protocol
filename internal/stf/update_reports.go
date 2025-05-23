package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
)

func UpdateReports() error {
	err := extrinsic.Guarantee()
	if err != nil {
		return err
	}

	return nil
}

package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
)

func UpdateAssurances() error {
	err := extrinsic.Assurance()
	if err != nil {
		return err
	}
	return nil
}

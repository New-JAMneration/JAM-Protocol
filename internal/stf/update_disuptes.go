package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
)

func UpdateDisputes() error {
	log.Println("Update Disputes")

	_, err := extrinsic.Disputes()

	if err != nil {
		return err
	}

	return nil
}

package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
)

func UpdateAssurances() error {
	log.Println("Update assurances")

	return extrinsic.Assurance()
}

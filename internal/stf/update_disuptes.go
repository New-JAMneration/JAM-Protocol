package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func UpdateDisputes() error {
	log.Println("Update Disputes")

	s := store.GetInstance()
	disputeExtrinsic := s.GetProcessingBlockPointer().GetDisputesExtrinsic()
	_, err := extrinsic.Disputes(disputeExtrinsic)

	if err != nil {
		return err
	}

	return nil
}

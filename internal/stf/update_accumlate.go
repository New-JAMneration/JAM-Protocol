package stf

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func UpdateAccumlate() error {
	log.Println("Update Accumlate")

	s := store.GetInstance()
	// 12.1~2
	accumulation.GetAccumulatedHashes()

	// Those two functions should be modified to get W from store
	accumulation.UpdateImmediatelyAccumulateWorkReports(s.GetIntermediateStates().GetAccumulatableWorkReports())
	accumulation.UpdateQueuedWorkReports(s.GetIntermediateStates().GetAccumulatableWorkReports())
	accumulation.UpdateAccumulatableWorkReports()

	// 12.3
	err := accumulation.DeferredTransfers()
	if err != nil {
		return err
	}
	return nil
}

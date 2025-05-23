package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func UpdateHistory() error {
	s := store.GetInstance()
	rhc := recent_history.NewRecentHistoryController()
	if rhc == nil {
		return fmt.Errorf("controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		return fmt.Errorf("expected controller to have no states initially, got %d", len(rhc.Betas))
	}
	rhc.Betas = s.GetPriorStates().GetBeta()

	// Test AddToBetaDagger
	// Start test STFBeta2BetaDagger (4.6)
	recent_history.STFBeta2BetaDagger()

	// Start test STFBetaDagger2BetaPrime (4.7)
	// STFBetaDagger2BetaPrime()
	var (
		betas                  = s.GetIntermediateStates().GetBetaDagger()
		block                  = s.GetProcessingBlockPointer().GetBlock()
		headerHash             = block.Header.Parent
		eg                     = block.Extrinsic.Guarantees
		inputAccumulateRootMap = s.GetIntermediateStates().GetBeefyCommitmentOutput()
	)
	var inputAccumulateRoot types.OpaqueHash
	for serviceHash, exist := range inputAccumulateRootMap {
		if exist {
			inputAccumulateRoot = serviceHash.Hash
		}
	}

	rhc.Betas = betas
	items := rhc.N(headerHash, eg, inputAccumulateRoot)
	rhc.AddToBetaPrime(items)
	return nil
}

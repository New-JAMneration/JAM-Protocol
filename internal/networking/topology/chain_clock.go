package topology

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ChainTimeslots exposes the local chain head and finalized timeslots used by topology.
// BestHead follows SyncManager / CE128 import (#567); Finalized drives epoch-transition rules.
type ChainTimeslots struct {
	BestHead  types.TimeSlot
	Finalized types.TimeSlot
}

func readChainTimeslots(chain *blockchain.ChainState) ChainTimeslots {
	if chain == nil {
		return ChainTimeslots{}
	}
	ts := ChainTimeslots{
		Finalized: chain.GetLatestFinalizedBlock().Header.Slot,
	}
	if head, err := chain.GetCurrentHead(); err == nil {
		ts.BestHead = head.Header.Slot
	}
	return ts
}

func finalizedBlock(chain *blockchain.ChainState) types.Block {
	if chain == nil {
		return types.Block{}
	}
	return chain.GetLatestFinalizedBlock()
}

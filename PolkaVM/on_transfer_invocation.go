package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type OnTransferInput struct {
	ServiceAccounts   types.ServiceAccountState
	Timeslot          types.TimeSlot
	ServiceID         types.ServiceId
	DeferredTransfers []types.DeferredTransfer
}

func OnTransferInvoke(OnTransferInput) (types.ServiceAccount, types.Gas) {
	return types.ServiceAccount{}, 0
}

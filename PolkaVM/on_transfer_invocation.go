package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type OnTransferInput struct {
	ServiceAccounts  types.ServiceAccountState
	Timeslot         types.TimeSlot
	ServiceID        types.ServiceId
	DeferredTransfer types.DeferredTransfer
}

func OnTransferInvoke(OnTransferInput) types.ServiceAccount {
	return types.ServiceAccount{}
}

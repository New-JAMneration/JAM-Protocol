package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type OnTransferInput struct {
	ServiceAccounts   types.ServiceAccountState
	Timeslot          types.TimeSlot
	ServiceID         types.ServiceId
	DeferredTransfers []types.DeferredTransfer
}

// Psi_T
func OnTransferInvoke(input OnTransferInput) (types.ServiceAccount, types.Gas) {
	emptyArray := types.OpaqueHash{}

	if account, accountExists := input.ServiceAccounts[input.ServiceID]; accountExists {
		if account.ServiceInfo.CodeHash == emptyArray || len(input.DeferredTransfers) == 0 {
			return input.ServiceAccounts[input.ServiceID], 0
		}
		// arguments
		serialized := []byte{}
		gasLimits := Gas(0)
		balances := types.U64(0)
		for _, deferredTransfer := range input.DeferredTransfers {
			gasLimits += deferredTransfer.GasLimit
			balances += deferredTransfer.Balance
		}

		encoder := types.Encoder{}
		s := input.ServiceAccounts[input.ServiceID]
		s.ServiceInfo.Balance += balances
		// F := hostCallFunctions()

		// timeslot(t) bytes
		timeslotSer, err := encoder.Encode(input.Timeslot)
		if err != nil {
			log.Println("OnTransferInvoke encode TimeSlot error")
		}
		// serviceID(s) bytes
		serviceIDSer, err := encoder.Encode(input.ServiceID)

		serialized = append(timeslotSer, serviceIDSer...)

		// !warning: should input service count blob, not codehash, how to get code blob ?
		result := Psi_M(account.ServiceInfo.CodeHash[:], 10, gasLimits, serialized, F, nil, s)

		// Psi_M
	} else {
		log.Println("OnTransferInvoke account not exists")
	}

	return types.ServiceAccount{}, 0
}

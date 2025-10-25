package PVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type OnTransferInput struct {
	ServiceAccounts   types.ServiceAccountState
	Timeslot          types.TimeSlot
	ServiceID         types.ServiceId
	DeferredTransfers []types.DeferredTransfer
	StorageKeyVal     types.StateKeyVals
}

// Psi_T
func OnTransferInvoke(input OnTransferInput) (types.ServiceAccount, types.Gas, types.StateKeyVals) {
	account, accountExists := input.ServiceAccounts[input.ServiceID]
	if !accountExists {
		log.Fatalf("OnTransferInvoke serviceAccount : %d not exists", input.ServiceID)
		return account, 0, input.StorageKeyVal
	}
	/*
	   codeHash := s.ServiceInfo.CodeHash
	   	_, code, err := service_account.FetchCodeByHash(s, codeHash)
	*/
	_, code, err := service_account.FetchCodeByHash(account, account.ServiceInfo.CodeHash)
	if err != nil {
		logger.Debug("no code to execute in Psi_T")
		return account, 0, input.StorageKeyVal
	}
	// programCode, programCodeExists := account.PreimageLookup[codeHash]
	if len(code) > types.MaxServiceCodeSize || len(input.DeferredTransfers) == 0 {
		logger.Debug(len(code), len(input.DeferredTransfers))
		return account, 0, input.StorageKeyVal
	}
	encoder := types.NewEncoder()
	// Psi_M arguments
	gasLimits := types.Gas(0)
	balances := types.U64(0)

	F := Omegas{}

	var serialized []byte
	// timeslot(t) bytes
	encoded, err := encoder.EncodeUint(uint64(input.Timeslot))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	logger.Debug("serialized: ", serialized)
	// serviceID(s) bytes
	encoded, err = encoder.EncodeUint(uint64(input.ServiceID))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	// deferredTransfer(bold{t})
	encoded, err = encoder.EncodeUint(uint64(len(input.DeferredTransfers)))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	logger.Debug("serialized: ", serialized)

	logger.Debug("serialized: ", serialized)

	for _, deferredTransfer := range input.DeferredTransfers {
		gasLimits += deferredTransfer.GasLimit
		balances += deferredTransfer.Balance
	}
	account.ServiceInfo.Balance += balances
	input.ServiceAccounts[input.ServiceID] = account
	store := store.GetInstance()
	eta0 := store.GetPriorStates().GetEta()[0]

	// omegas
	F[LookupOp] = HostCallFunctions[LookupOp]
	F[FetchOp] = HostCallFunctions[FetchOp] // added 0.6.6
	F[ReadOp] = HostCallFunctions[ReadOp]
	F[WriteOp] = HostCallFunctions[WriteOp]
	F[GasOp] = HostCallFunctions[GasOp]
	F[InfoOp] = HostCallFunctions[InfoOp]
	F[100] = logHostCall

	serviceAccounts := input.ServiceAccounts
	serviceAccount := serviceAccounts[input.ServiceID]
	// addition, on-transfer only uses GeneralArgs
	addition := HostCallArgs{
		GeneralArgs: GeneralArgs{
			ServiceAccount:      &serviceAccount,
			ServiceId:           &input.ServiceID,
			ServiceAccountState: &serviceAccounts,
			StorageKeyVal:       &input.StorageKeyVal,
		},
		AccumulateArgs: AccumulateArgs{
			Eta: eta0,
		},
		OnTransferArgs: OnTransferArgs{
			DeferredTransfer: input.DeferredTransfers,
		},
	}
	result := Psi_M(StandardCodeFormat(code), 10, gasLimits, serialized, F, addition)

	return *result.Addition.ServiceAccount, types.Gas(result.Gas), *result.Addition.StorageKeyVal
}

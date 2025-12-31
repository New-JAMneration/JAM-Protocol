package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (B.8) Ψ_A
func Psi_A(
	partialState types.PartialStateSet, // e
	timeslot types.TimeSlot, // t
	serviceId types.ServiceId, // s
	gas types.Gas, // g
	operandOrDeferTransfers []types.OperandOrDeferredTransfer, // i
	eta types.Entropy,
	storageKeyVal types.StateKeyVals,
) (
	psi_result Psi_A_ReturnType,
) {
	s, ok := partialState.ServiceAccounts[serviceId]
	if !ok {
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: []types.DeferredTransfer{},
			Result:            nil,
			Gas:               0,
			ServiceBlobs:      []types.ServiceBlob{},
			StorageKeyVal:     storageKeyVal,
		}
	}

	// s = e
	var balances uint64
	for _, v := range operandOrDeferTransfers {
		if v.DeferredTransfer != nil {
			balances += uint64(v.DeferredTransfer.Balance)
		}
	}
	s.ServiceInfo.Balance += types.U64(balances)
	partialState.ServiceAccounts[serviceId] = s

	// (9.4) E(↕m, c) = ap[ac]
	// Get actual code (c)
	codeHash := s.ServiceInfo.CodeHash
	_, code, err := service_account.FetchCodeByHash(s, codeHash)
	if err != nil {
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: []types.DeferredTransfer{},
			Result:            nil,
			Gas:               0,
			ServiceBlobs:      []types.ServiceBlob{},
			StorageKeyVal:     storageKeyVal,
		}
	}

	// if c = ∅ or |c| > W_C
	if !ok || len(code) == 0 || len(code) > types.MaxServiceCodeSize {
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: []types.DeferredTransfer{},
			Result:            nil,
			Gas:               0,
			ServiceBlobs:      []types.ServiceBlob{},
			StorageKeyVal:     storageKeyVal,
		}
	}

	var serialized []byte
	encoder := types.NewEncoder()

	// Encode t
	// encoded, err := encoder.Encode(&time)
	encoded, err := encoder.EncodeUint(uint64(timeslot))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)

	// Encode s
	// encoded, err = encoder.Encode(&serviceId)
	encoded, err = encoder.EncodeUint(uint64(serviceId))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	// Encode |o|
	encoded, err = encoder.EncodeUint(uint64(len(operandOrDeferTransfers)))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	F := Omegas{}
	F[FetchOp] = HostCallFunctions[FetchOp] // added 0.6.6
	F[ReadOp] = wrapWithG(HostCallFunctions[ReadOp])
	F[WriteOp] = wrapWithG(HostCallFunctions[WriteOp])
	F[LookupOp] = wrapWithG(HostCallFunctions[LookupOp])
	F[GasOp] = HostCallFunctions[GasOp]
	F[InfoOp] = wrapWithG(HostCallFunctions[InfoOp])
	F[BlessOp] = HostCallFunctions[BlessOp]
	F[AssignOp] = HostCallFunctions[AssignOp]
	F[DesignateOp] = HostCallFunctions[DesignateOp]
	F[CheckpointOp] = HostCallFunctions[CheckpointOp]
	F[NewOp] = HostCallFunctions[NewOp]
	F[UpgradeOp] = HostCallFunctions[UpgradeOp]
	F[TransferOp] = HostCallFunctions[TransferOp]
	F[EjectOp] = HostCallFunctions[EjectOp]
	F[QueryOp] = HostCallFunctions[QueryOp]
	F[SolicitOp] = HostCallFunctions[SolicitOp]
	F[ForgetOp] = HostCallFunctions[ForgetOp]
	F[YieldOp] = HostCallFunctions[YieldOp]
	F[ProvideOp] = HostCallFunctions[ProvideOp]
	F[100] = logHostCall

	newPartialState := partialState.DeepCopy()
	newStorageKeyVal := storageKeyVal.DeepCopy()
	serviceAccount := newPartialState.ServiceAccounts[serviceId]
	addition := HostCallArgs{
		GeneralArgs: GeneralArgs{
			ServiceAccount:      &serviceAccount,
			ServiceId:           &serviceId,
			ServiceAccountState: &newPartialState.ServiceAccounts,
			CoreId:              nil,
			StorageKeyVal:       &newStorageKeyVal,
		},
		// storageKeyVal can be seen as service storage state, what partialState do, the storageKeyVal will do the same
		AccumulateArgs: AccumulateArgs{
			ResultContextX:             I(newPartialState, serviceId, timeslot, eta, &newStorageKeyVal),
			ResultContextY:             I(partialState, serviceId, timeslot, eta, &storageKeyVal),
			Eta:                        eta,
			OperandOrDeferredTransfers: operandOrDeferTransfers,
			Timeslot:                   timeslot,
		},
	}

	resultM := Psi_M(StandardCodeFormat(code), 5, types.Gas(gas), Argument(serialized), F, addition)
	partialState, deferredTransfer, result, gas, serviceBlobs, storageKeyVal := C(types.Gas(resultM.Gas), resultM.ReasonOrBytes, AccumulateArgs{
		ResultContextX: resultM.Addition.AccumulateArgs.ResultContextX,
		ResultContextY: resultM.Addition.AccumulateArgs.ResultContextY,
	})

	return Psi_A_ReturnType{
		PartialStateSet:   partialState,
		DeferredTransfers: deferredTransfer,
		Result:            result,
		Gas:               gas,
		ServiceBlobs:      serviceBlobs,
		StorageKeyVal:     storageKeyVal,
	}
}

// (B.12) G
func G(o OmegaOutput, serviceAccount types.ServiceAccount) OmegaOutput {
	o.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[o.Addition.AccumulateArgs.ResultContextX.ServiceId] = serviceAccount
	return o
}

func wrapWithG(original Omega) Omega {
	return func(input OmegaInput) OmegaOutput {
		output := original(input)
		serviceAccount := input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[input.Addition.AccumulateArgs.ResultContextX.ServiceId]
		return G(output, serviceAccount)
	}
}

// (B.13) C
func C(gas types.Gas, reasonOrBytes any, resultContext AccumulateArgs) (types.PartialStateSet, []types.DeferredTransfer, *types.OpaqueHash, types.Gas, types.ServiceBlobs, types.StateKeyVals) {
	serviceBlobs := make(types.ServiceBlobs, 0)
	switch reasonOrBytes := reasonOrBytes.(type) {
	case error: // system error
		for _, v := range resultContext.ResultContextY.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		return resultContext.ResultContextY.PartialState, resultContext.ResultContextY.DeferredTransfers, resultContext.ResultContextY.Exception, gas, serviceBlobs, *resultContext.ResultContextY.StorageKeyVal
	case []byte:
		var h types.OpaqueHash
		if len(reasonOrBytes) != len(h) {
			return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, resultContext.ResultContextX.Exception, gas, serviceBlobs, *resultContext.ResultContextX.StorageKeyVal
		}
		copy(h[:], reasonOrBytes[:len(h)])
		opaqueHash := &h
		for _, v := range resultContext.ResultContextX.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, opaqueHash, gas, serviceBlobs, *resultContext.ResultContextX.StorageKeyVal
	default:
		if reasonOrBytes == OUT_OF_GAS || reasonOrBytes == PANIC {
			for _, v := range resultContext.ResultContextY.ServiceBlobs {
				serviceBlobs = append(serviceBlobs, v)
			}
			return resultContext.ResultContextY.PartialState, resultContext.ResultContextY.DeferredTransfers, resultContext.ResultContextY.Exception, gas, serviceBlobs, *resultContext.ResultContextY.StorageKeyVal
		}
		for _, v := range resultContext.ResultContextX.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, resultContext.ResultContextX.Exception, gas, serviceBlobs, *resultContext.ResultContextX.StorageKeyVal
	}
}

// (B.10)
func I(partialState types.PartialStateSet, serviceId types.ServiceId, ht types.TimeSlot, eta types.Entropy, storageKeyVal *types.StateKeyVals) ResultContext {
	serialized := []byte{}
	encoder := types.NewEncoder()

	encoded, err := encoder.EncodeUint(uint64(serviceId))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	encoded, err = encoder.Encode(&eta)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	encoded, err = encoder.EncodeUint(uint64(ht))
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)

	hash := hash.Blake2bHash(serialized)

	var result types.ServiceId
	decoder := types.NewDecoder()
	err = decoder.Decode(hash[:], &result)
	if err != nil {
		panic(err)
	}

	var modValue types.ServiceId = (1 << 32) - types.MinimumServiceIndex - (1 << 8) // 2^32 - S - 2^8
	var addValue types.ServiceId = types.MinimumServiceIndex                        // 2^8
	result = check((result%modValue)+addValue, partialState.ServiceAccounts)

	return ResultContext{
		ServiceId:         serviceId,
		PartialState:      partialState,
		ImportServiceId:   result,
		DeferredTransfers: []types.DeferredTransfer{},
		Exception:         nil,
		ServiceBlobs:      make(map[types.OpaqueHash]types.ServiceBlob),
		StorageKeyVal:     storageKeyVal,
	}
}

type Psi_A_ReturnType struct {
	PartialStateSet   types.PartialStateSet
	DeferredTransfers []types.DeferredTransfer
	Result            *types.OpaqueHash
	Gas               types.Gas
	ServiceBlobs      []types.ServiceBlob
	StorageKeyVal     types.StateKeyVals
}

// (B.7)
type ResultContext struct {
	ServiceId         types.ServiceId                        // s
	PartialState      types.PartialStateSet                  // u
	ImportServiceId   types.ServiceId                        // i
	DeferredTransfers []types.DeferredTransfer               // t
	Exception         *types.OpaqueHash                      // y
	ServiceBlobs      map[types.OpaqueHash]types.ServiceBlob // p   v0.6.5
	StorageKeyVal     *types.StateKeyVals                    // add this for fuzzer
}

func (origin *ResultContext) DeepCopy() ResultContext {
	// ServiceId
	copiedServiceId := origin.ServiceId

	// PartialState
	copiedPartialState := origin.PartialState.DeepCopy()

	// ImportServiceId
	copiedImportServiceId := origin.ImportServiceId

	// DeferredTransfers
	copiedDeferredTransfers := make([]types.DeferredTransfer, len(origin.DeferredTransfers))
	for k, v := range origin.DeferredTransfers {
		copiedTrasfer := types.DeferredTransfer{
			SenderID:   v.SenderID,
			ReceiverID: v.ReceiverID,
			Balance:    v.Balance,
			Memo:       [128]byte{},
			GasLimit:   v.GasLimit,
		}
		copy(copiedTrasfer.Memo[:], v.Memo[:])
		copiedDeferredTransfers[k] = copiedTrasfer
	}

	// Exception
	var copiedException *types.OpaqueHash
	if origin.Exception != nil {
		exceptionVal := *origin.Exception
		copiedException = &exceptionVal
	}

	// ServiceBlobs
	copiedServiceBlobs := make(map[types.OpaqueHash]types.ServiceBlob)
	for k, v := range origin.ServiceBlobs {
		var copiedServiceBlob types.ServiceBlob
		copiedServiceBlob.ServiceID = v.ServiceID
		copiedServiceBlob.Blob = make([]byte, len(v.Blob))
		copy(copiedServiceBlob.Blob, v.Blob)
		copiedServiceBlobs[k] = copiedServiceBlob
	}

	// StorageKeyVal
	copiedStorageKeyVal := origin.StorageKeyVal.DeepCopy()

	return ResultContext{
		ServiceId:         copiedServiceId,
		PartialState:      copiedPartialState,
		ImportServiceId:   copiedImportServiceId,
		DeferredTransfers: copiedDeferredTransfers,
		Exception:         copiedException,
		ServiceBlobs:      copiedServiceBlobs,
		StorageKeyVal:     &copiedStorageKeyVal,
	}
}

package PVM

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// OperationType Enum
type OperationType int

const (
	// ----------------- General Functions -----------------
	GasOp    OperationType = iota // gas = 0
	FetchOp                       // fetch = 1
	LookupOp                      // lookup = 2
	ReadOp                        // read = 3
	WriteOp                       // write = 4
	InfoOp                        // info = 5

	// ----------------- Refine Functions -----------------
	HistoricalLookupOp // historical_lookup = 6
	ExportOp           // export = 7
	MachineOp          // machine = 8
	PeekOp             // peek = 9
	PokeOp             // poke = 10
	PagesOp            // pages = 11
	InvokeOp           // invoke = 12
	ExpungeOp          // expunge = 13

	// ----------------- Accumulate Functions -----------------
	BlessOp      // bless = 14
	AssignOp     // assign = 15
	DesignateOp  // designate = 16
	CheckpointOp // checkpoint = 17
	NewOp        // new = 18
	UpgradeOp    // upgrade = 19
	TransferOp   // transfer = 20
	EjectOp      // eject = 21
	QueryOp      // query = 22
	SolicitOp    // solicit = 23
	ForgetOp     // forget = 24
	YieldOp      // yield = 25
	ProvideOp    // provide = 26
	LogOp        = OperationType(100)

	MaxOperationType OperationType = 100
)

type GeneralArgs struct {
	ServiceAccount      *types.ServiceAccount
	ServiceId           *types.ServiceId
	ServiceAccountState *types.ServiceAccountState
	CoreId              *types.CoreIndex
	StorageKeyVal       *types.StateKeyVals
}

type AccumulateArgs struct {
	ResultContextX             ResultContext
	ResultContextY             ResultContext
	Timeslot                   types.TimeSlot
	Eta                        types.Entropy                     // italic n / eta_0, used in fetch
	OperandOrDeferredTransfers []types.OperandOrDeferredTransfer // o, used in fetch
}

type RefineArgs struct {
	WorkItemIndex       *uint                   // i
	WorkPackage         *types.WorkPackage      // p
	AuthOutput          *types.ByteSequence     // r
	ImportSegments      [][]types.ExportSegment // overline{bold{i}}
	ExportSegmentOffset uint                    // zeta
	ExtrinsicDataMap    ExtrinsicDataMap        // extrinsic data map
	IntegratedPVMMap    IntegratedPVMMap        // D ( N -> M ) : N -> (p(program_code), u, i)
	ExportSegment       []types.ExportSegment   // e
	// ServiceID           types.ServiceId         // s
	TimeSlot   types.TimeSlot          // t
	Extrinsics [][]types.ExtrinsicSpec // overline{x}, used in fetch
}

type HostCallArgs struct {
	GeneralArgs
	AccumulateArgs
	RefineArgs
	*Program
}

func getPtr[T any](v T) *T { return &v }

var hostCallName = []string{
	0:   "gas",
	1:   "fetch",
	2:   "lookup",
	3:   "read",
	4:   "write",
	5:   "info",
	6:   "historicalLookup",
	7:   "export",
	8:   "machine",
	9:   "peek",
	10:  "poke",
	11:  "pages",
	12:  "invoke",
	13:  "expunge",
	14:  "bless",
	15:  "assign",
	16:  "designate",
	17:  "checkpoint",
	18:  "new",
	19:  "upgrade",
	20:  "transfer",
	21:  "eject",
	22:  "query",
	23:  "solicit",
	24:  "forget",
	25:  "yield",
	26:  "provide",
	100: "log",
}

var HostCallFunctions Omegas = func() Omegas {
	f := make([]Omega, MaxOperationType+1)
	f[GasOp] = gas
	f[FetchOp] = fetch
	f[LookupOp] = lookup
	f[ReadOp] = read
	f[WriteOp] = write
	f[InfoOp] = info
	f[HistoricalLookupOp] = historicalLookup
	f[ExportOp] = export
	f[MachineOp] = machine
	f[PeekOp] = peek
	f[PokeOp] = poke
	f[PagesOp] = pages
	f[InvokeOp] = invoke
	f[ExpungeOp] = expunge
	f[BlessOp] = bless
	f[AssignOp] = assign
	f[DesignateOp] = designate
	f[CheckpointOp] = checkpoint
	f[NewOp] = new
	f[UpgradeOp] = upgrade
	f[TransferOp] = transfer
	f[EjectOp] = eject
	f[QueryOp] = query
	f[SolicitOp] = solicit
	f[ForgetOp] = forget
	f[YieldOp] = yield
	f[ProvideOp] = provide
	f[LogOp] = logHostCall
	return f
}()

var (
	readWrapWithG   Omega = wrapWithG(read)
	writeWrapWithG  Omega = wrapWithG(write)
	lookupWrapWithG Omega = wrapWithG(lookup)
	infoWrapWithG   Omega = wrapWithG(info)
)

func hostCallException(input OmegaInput) (output OmegaOutput) {
	// non-defined host call
	input.Interpreter.Registers[7] = WHAT
	input.Interpreter.Gas -= 10
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// 0.7.2
func hostCallOutOfGas(input OmegaInput) (output OmegaOutput) {
	return OmegaOutput{
		ExitReason: ExitOOG,
		Addition:   input.Addition,
	}
}

func chargeGasAndCheck(input *OmegaInput) *OmegaOutput {
	input.Interpreter.Gas -= 10
	if input.Interpreter.Gas < 0 {
		return &OmegaOutput{
			ExitReason: ExitOOG,
			Addition:   input.Addition,
		}
	}
	return nil
}

// Gas Function（ΩG）, gas = 0
func gas(input OmegaInput) OmegaOutput {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	input.Interpreter.Registers[7] = uint64(input.Interpreter.Gas)
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

type fetchHandler func(OmegaInput, *types.Encoder) ([]byte, error)

var fetchHandlers = [16]fetchHandler{
	0:  fetchConstants,
	1:  fetchEta,
	2:  fetchAuthOutput,
	3:  fetchExtrinsicAt,
	4:  fetchExtrinsicForWorkItem,
	5:  fetchImportSegmentAt,
	6:  fetchImportSegmentForWorkItem,
	7:  fetchWorkPackage,
	8:  fetchAuthorizerConfig,
	9:  fetchAuthorization,
	10: fetchWorkPackageContext,
	11: fetchWorkPackageItems,
	12: fetchWorkItemAt,
	13: fetchWorkItemPayload,
	14: fetchOperandOrDeferredTransfers,
	15: fetchOperandOrDeferredTransferAt,
}

func fetchConstants(input OmegaInput, _ *types.Encoder) ([]byte, error) {
	return getFetchConstantsData(), nil
}

func fetchEta(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if reflect.ValueOf(input.Addition.Eta).IsZero() {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.Eta)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 1 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchAuthOutput(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.AuthOutput == nil {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.AuthOutput)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 2 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchExtrinsicAt(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.Extrinsics) == 0 {
		return nil, nil
	}
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.Extrinsics)) {
		return nil, nil
	}
	w12 := input.Interpreter.Registers[12]
	if w12 >= uint64(len(input.Addition.Extrinsics[w11])) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.Extrinsics[w11][w12])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 3 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchExtrinsicForWorkItem(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.Extrinsics) == 0 || input.Addition.WorkItemIndex == nil {
		return nil, nil
	}
	i := *input.Addition.WorkItemIndex
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.Extrinsics[i])) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.Extrinsics[i][w11])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 4 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchImportSegmentAt(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.ImportSegments) == 0 {
		return nil, nil
	}
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.ImportSegments)) {
		return nil, nil
	}
	w12 := input.Interpreter.Registers[12]
	if w12 >= uint64(len(input.Addition.ImportSegments[w11])) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.ImportSegments[w11][w12])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 5 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchImportSegmentForWorkItem(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.ImportSegments) == 0 || input.Addition.WorkItemIndex == nil {
		return nil, nil
	}
	i := *input.Addition.WorkItemIndex
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.ImportSegments[i])) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.ImportSegments[i][w11])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 6 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchWorkPackage(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.WorkPackage)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 7 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchAuthorizerConfig(input OmegaInput, _ *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	return []byte(input.Addition.WorkPackage.AuthorizerConfig), nil
}

func fetchAuthorization(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.WorkPackage.Authorization)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 9 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchWorkPackageContext(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.WorkPackage.Context)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 10 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchWorkPackageItems(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	buffer, err := enc.EncodeUint(uint64(len(input.Addition.WorkPackage.Items)))
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 11 encode error: %v", err)
		return nil, err
	}
	for _, w := range input.Addition.WorkPackage.Items {
		sw, err := S(enc, w)
		if err != nil {
			pvmLogger.Errorf("fetch host-call case 11 S func error: %v", err)
			return nil, err
		}
		buffer = append(buffer, sw...)
	}
	return buffer, nil
}

func fetchWorkItemAt(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.WorkPackage.Items)) {
		return nil, nil
	}
	val, err := S(enc, input.Addition.WorkPackage.Items[w11])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 12 S func error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchWorkItemPayload(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if input.Addition.WorkPackage == nil {
		return nil, nil
	}
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.WorkPackage.Items)) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.WorkPackage.Items[w11].Payload)
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 13 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

func fetchOperandOrDeferredTransfers(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.OperandOrDeferredTransfers) == 0 {
		return nil, nil
	}
	buffer, err := enc.EncodeUint(uint64(len(input.Addition.OperandOrDeferredTransfers)))
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 14 encode uint error: %v", err)
		return nil, err
	}
	for _, o := range input.Addition.OperandOrDeferredTransfers {
		b, err := enc.Encode(&o)
		if err != nil {
			pvmLogger.Errorf("fetch host-call case 14 encode error: %v", err)
			return nil, err
		}
		buffer = append(buffer, b...)
	}
	return buffer, nil
}

func fetchOperandOrDeferredTransferAt(input OmegaInput, enc *types.Encoder) ([]byte, error) {
	if len(input.Addition.OperandOrDeferredTransfers) == 0 {
		return nil, nil
	}
	w11 := input.Interpreter.Registers[11]
	if w11 >= uint64(len(input.Addition.OperandOrDeferredTransfers)) {
		return nil, nil
	}
	val, err := enc.Encode(&input.Addition.OperandOrDeferredTransfers[w11])
	if err != nil {
		pvmLogger.Errorf("fetch host-call case 15 encode error: %v", err)
		return nil, err
	}
	return val, nil
}

// fetch = 1
func fetch(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	encoder := types.NewEncoder()
	idx := input.Interpreter.Registers[10]
	var v *[]byte
	var val []byte
	var err error
	if idx < uint64(len(fetchHandlers)) {
		val, err = fetchHandlers[idx](input, encoder)
		if err == nil && val != nil {
			v = &val
		}
	}

	if err != nil {
		v = nil
	}

	var dataLength uint64
	if v != nil {
		dataLength = uint64(len(*v))
	}
	o := input.Interpreter.Registers[7]
	f := min(input.Interpreter.Registers[8], dataLength)
	l := min(input.Interpreter.Registers[9], dataLength-f)
	// nothing to write, don't need to check memory access
	if l == 0 && v != nil {
		input.Interpreter.Registers[7] = dataLength
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	// need to first check writable
	if !isWriteable(o, l, *input.Interpreter.Memory) && v != nil {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if v = nil
	if v == nil {
		input.Interpreter.Registers[7] = NONE

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	input.Interpreter.Memory.Write(o, (*v)[f:f+l])
	input.Interpreter.Registers[7] = dataLength

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// ΩL(ϱ, ω, μ, s, s, d) , lookup = 2
func lookup(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	serviceID := *input.Addition.ServiceId
	serviceAccount := *input.Addition.ServiceAccount
	delta := *input.Addition.ServiceAccountState

	var a *types.ServiceAccount
	if input.Interpreter.Registers[7] == 0xffffffffffffffff || input.Interpreter.Registers[7] == uint64(serviceID) {
		a = &serviceAccount
	} else if value, exists := delta[types.ServiceId(input.Interpreter.Registers[7])]; exists {
		a = &value
	}

	h, o := input.Interpreter.Registers[8], input.Interpreter.Registers[9]
	if !isReadable(h, 32, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	preimageRawData := input.Interpreter.Memory.Read(h, 32)

	var v *types.ByteSequence
	var f uint64
	var l uint64
	if a != nil {
		if preimage, preimageExists := a.PreimageLookup[types.OpaqueHash(preimageRawData)]; preimageExists {
			v = &preimage
		}

		if v != nil {
			f = min(input.Interpreter.Registers[10], uint64(len(*v)))
			l = min(input.Interpreter.Registers[11], uint64(len(*v))-f)
		}
	}

	if !isWriteable(o, l, *input.Interpreter.Memory) && l != 0 {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	if v == nil {
		input.Interpreter.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.Interpreter.Registers[7] = uint64(len(*v))
	if l != 0 {
		input.Interpreter.Memory.Write(o, (*v)[f:f+l])
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// ΩR(ϱ, ω, μ, s, s, d) , read = 3
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(italic): ServiceId
d: ServiceAccountState (map[ServiceId]ServiceAccount)
*/
func read(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	serviceID := *input.Addition.GeneralArgs.ServiceId
	delta := *input.Addition.GeneralArgs.ServiceAccountState
	var sStar uint64
	// assign s*
	if input.Interpreter.Registers[7] == 0xffffffffffffffff {
		sStar = uint64(serviceID)
	} else {
		sStar = input.Interpreter.Registers[7]
	}
	// assign ko, kz, o first and check v = panic ?
	// since v = panic is the first condition to check
	ko, kz, o := input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10]
	if !isReadable(ko, kz, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	var a types.ServiceAccount
	// assign a
	if sStar == uint64(serviceID) {
		a = delta[serviceID]
	} else if value, exists := delta[types.ServiceId(sStar)]; exists {
		a = value
		serviceID = types.ServiceId(sStar)
	} else {
		// a = nil , v not panic, => v = nil
		input.Interpreter.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// v = a_s[k]?  ,  a = nil is checked, only check k in Key(a_s)
	// first compute k , mu_ko...+kz
	storageRawKey := input.Interpreter.Memory.Read(ko, kz)
	v, exists := a.StorageDict[string(storageRawKey)]
	storageValueFromKeyVal := getStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
	// v = nil
	if !exists {
		if storageValueFromKeyVal == nil { // check storage state key-val
			input.Interpreter.Registers[7] = NONE
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		} else {
			v = *storageValueFromKeyVal

			// store the unknown storage item to state
			a.StorageDict[string(storageRawKey)] = v

			input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
			// remove from storageKeyVal
			removeStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
		}
	}

	f := min(input.Interpreter.Registers[11], uint64(len(v)))
	l := min(input.Interpreter.Registers[12], uint64(len(v))-f)
	// nothing to write, don't need to check memory access
	if l == 0 {
		input.Interpreter.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// first check not writable, then check v = nil (not exists)
	if !isWriteable(o, l, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.Interpreter.Registers[7] = uint64(len(v))
	input.Interpreter.Memory.Write(o, v[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// ΩW (ϱ, ω, μ, s, s) , write = 4
func write(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	ko, kz, vo, vz := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10]
	if !isReadable(ko, kz, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}
	// compute \mathbb{k}
	storageRawKey := input.Interpreter.Memory.Read(ko, kz)

	serviceID := *input.Addition.GeneralArgs.ServiceId
	a := *input.Addition.GeneralArgs.ServiceAccount

	value, storageRawKeyExists := a.StorageDict[string(storageRawKey)]
	storageRawData := getStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
	var l uint64
	var footprintItems types.U32
	var footprintOctets types.U64
	if storageRawKeyExists {
		footprintItems, footprintOctets = service_account.CalcStorageItemfootprint(string(storageRawKey), value)
		l = uint64(len(value))
	} else if !storageRawKeyExists && storageRawData != nil {
		footprintItems, footprintOctets = service_account.CalcStorageItemfootprint(string(storageRawKey), *storageRawData)
		l = uint64(len(*storageRawData))
	} else {
		l = NONE
	}

	encodedKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageRawKey, nil)

	if vz == 0 { // remove storage
		delete(a.StorageDict, string(storageRawKey))
		removeStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)

		// direct update items, octets
		a.ServiceInfo.Items -= footprintItems
		a.ServiceInfo.Bytes -= footprintOctets
	} else if isReadable(vo, vz, *input.Interpreter.Memory) { // storage append/update
		storageRawData := input.Interpreter.Memory.Read(vo, vz)
		a.StorageDict[string(storageRawKey)] = storageRawData
		removeStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
		// compute items, octets , check a_t > a_b first
		newItems := a.ServiceInfo.Items - footprintItems
		newOctets := a.ServiceInfo.Bytes - footprintOctets

		storageItems, storageOctets := service_account.CalcStorageItemfootprint(string(storageRawKey), storageRawData)
		newItems += storageItems
		newOctets += storageOctets
		newMinBalance := service_account.CalcThresholdBalance(newItems, newOctets, a.ServiceInfo.DepositOffset) // a_t
		if newMinBalance > a.ServiceInfo.Balance {
			input.Interpreter.Registers[7] = FULL
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}
		pvmLogger.Debugf("write storage key: 0x%x, val: 0x%x", encodedKey.Key, storageRawData)
		// update items, octets
		a.ServiceInfo.Items = newItems
		a.ServiceInfo.Bytes = newOctets
	} else {
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// update service state
	(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = a
	(*input.Addition.GeneralArgs.ServiceAccount) = a
	if input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts != nil {
		input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
	}

	input.Interpreter.Registers[7] = l

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// ΩR(ϱ, ω, μ, s, d) , info = 5
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(italic): ServiceId
d: ServiceAccountState (map[ServiceId]ServiceAccount)
*/
func info(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	serviceID := *input.Addition.ServiceId
	delta := *input.Addition.ServiceAccountState

	var a types.ServiceAccount
	if input.Interpreter.Registers[7] == 0xffffffffffffffff {
		a = delta[serviceID]
	} else {
		value, exist := delta[types.ServiceId(input.Interpreter.Registers[7])]
		if exist {
			a = value
		} else {
			// v = nil , l = 0 -> don't need to check writeable
			input.Interpreter.Registers[7] = NONE
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}
	}

	minBalance := service_account.CalcThresholdBalance(a.ServiceInfo.Items, a.ServiceInfo.Bytes, a.ServiceInfo.DepositOffset)
	var v types.ByteSequence
	encoder := types.NewEncoder()
	// a_c
	encoded, _ := encoder.Encode(&a.ServiceInfo.CodeHash)
	v = append(v, encoded...)
	// a_b
	encoded, _ = encoder.Encode(&a.ServiceInfo.Balance)
	v = append(v, encoded...)
	// a_t
	encoded, _ = encoder.Encode(&minBalance)
	v = append(v, encoded...)
	// a_g
	encoded, _ = encoder.Encode(&a.ServiceInfo.MinItemGas)
	v = append(v, encoded...)
	// a_m
	encoded, _ = encoder.Encode(&a.ServiceInfo.MinMemoGas)
	v = append(v, encoded...)
	// a_o
	encoded, _ = encoder.Encode(&a.ServiceInfo.Bytes)
	v = append(v, encoded...)
	// a_i
	encoded, _ = encoder.Encode(&a.ServiceInfo.Items)
	v = append(v, encoded...)
	// a_f
	encoded, _ = encoder.Encode(&a.ServiceInfo.DepositOffset)
	v = append(v, encoded...)
	// a_r
	encoded, _ = encoder.Encode(&a.ServiceInfo.CreationSlot)
	v = append(v, encoded...)
	// a_a
	encoded, _ = encoder.Encode(&a.ServiceInfo.LastAccumulationSlot)
	v = append(v, encoded...)
	// a_p
	encoded, _ = encoder.Encode(&a.ServiceInfo.ParentService)
	v = append(v, encoded...)
	f := min(input.Interpreter.Registers[9], uint64(len(v)))
	l := min(input.Interpreter.Registers[10], uint64(len(v))-f)
	o := input.Interpreter.Registers[8]
	// nothing to write
	if l == 0 {
		input.Interpreter.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	// if mathbf{N}_{o..._l} \not in mathbf{V}^*_mu
	if !isWriteable(o, l, *input.Interpreter.Memory) { // v = ∇ not defined
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.Interpreter.Registers[7] = uint64(len(v))
	input.Interpreter.Memory.Write(o, v[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// log = 100 , [JIP-1](https://hackmd.io/@polkadot/jip1)
func logHostCall(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	level := input.Interpreter.Registers[7]
	message := input.Interpreter.Memory.Read(input.Interpreter.Registers[10], input.Interpreter.Registers[11])
	levelStr := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}

	if level > 4 {
		pvmLogger.Errorf("logHostCall level not supported")
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	timeFormat := time.RFC3339
	timeStamp := time.Now().Format(timeFormat)
	var logMsg string
	if input.Interpreter.Registers[8] == 0 && input.Interpreter.Registers[9] == 0 {
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v][%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreId), derefernceOrNil(input.Addition.ServiceId), string(message))
	} else {
		target := input.Interpreter.Memory.Read(input.Interpreter.Registers[8], input.Interpreter.Registers[9])
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v][%s][%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreId), derefernceOrNil(input.Addition.ServiceId), target, string(message))
	}

	input.Interpreter.Registers[7] = WHAT
	pvmLogger.Debugf("%v", logMsg)
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// Encoding of a work item, used in the fetch function.
// This is added because the encoding for WorkItem used in fetch
// is a little different from the default encoding
func S(encoder *types.Encoder, item types.WorkItem) ([]byte, error) {
	return encoder.EncodeMany(
		getPtr(types.U32(item.Service)),             // w_s
		&item.CodeHash,                              // w_h
		getPtr(types.U64(item.RefineGasLimit)),      // w_g
		getPtr(types.U64(item.AccumulateGasLimit)),  // w_a
		getPtr(types.U16(item.ExportCount)),         // w_e
		getPtr(types.U16(len(item.ImportSegments))), // |w_i|
		getPtr(types.U16(len(item.Extrinsic))),      // |w_x|
		getPtr(types.U32(len(item.Payload))),        // |w_y|
	)
}

// B.14
func check(serviceID types.ServiceId, serviceAccountState types.ServiceAccountState) types.ServiceId {
	for {
		if _, accountExists := serviceAccountState[serviceID]; !accountExists {
			return serviceID
		}

		serviceID = (serviceID-types.MinimumServiceIndex+1)%(1<<32-(1<<8)-types.MinimumServiceIndex) + types.MinimumServiceIndex
	}
}

// 0.7.0 later, fuzzer (forks) needs to recover state, storage, part of lookupData cannot be recover,
// Thus, needs to check storage, part of lookupData from KeyVal
// return storage val and add storage state into ResultContextX
func getStorageFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceId, storageKey types.ByteSequence) *types.ByteSequence {
	requestedStorageStateKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageKey, nil)
	for _, v := range *keyVal {
		if v.Key == requestedStorageStateKey.Key {
			return &v.Value
		}
	}

	return nil
}

func removeStorageFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceId, storageKey types.ByteSequence) {
	requestedStorageStateKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageKey, nil)
	for k, v := range *keyVal {
		if v.Key == requestedStorageStateKey.Key {
			pvmLogger.Debugf("remove storage key: 0x%x\n", requestedStorageStateKey.Key)
			if k < len(*keyVal)-1 { // not the last index
				*keyVal = append((*keyVal)[:k], (*keyVal)[k+1:]...)
			} else {
				*keyVal = (*keyVal)[:k]
			}
			return
		}
	}
}

func getLookupItemFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceId, lookupKey types.LookupMetaMapkey) []byte {
	lookupStateKey := merklization.EncodeDelta4Key(serviceID, lookupKey)
	for k, v := range *keyVal {
		if v.Key == lookupStateKey {
			// remove from key-val
			if k < len(*keyVal)-1 { // not the last index
				*keyVal = append((*keyVal)[:k], (*keyVal)[k+1:]...)
			} else {
				*keyVal = (*keyVal)[:k]
			}
			return v.Value
		}
	}

	return nil
}

func derefernceOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

var (
	fetchConstantsOnce sync.Once
	fetchConstantsData []byte
)

// when chainspec is import is needed, this function can be moved to chainspec package and input the bytes into PVM entry point
func getFetchConstantsData() []byte {
	fetchConstantsOnce.Do(func() {
		encoder := types.NewEncoder()
		val, err := encoder.EncodeMany(
			getPtr(types.U64(types.AdditionalMinBalancePerItem)),      // B_I
			getPtr(types.U64(types.AdditionalMinBalancePerOctet)),     // B_L
			getPtr(types.U64(types.BasicMinBalance)),                  // B_S
			getPtr(types.U16(types.CoresCount)),                       // C
			getPtr(types.U32(types.UnreferencedPreimageTimeslots)),    // D
			getPtr(types.U32(types.EpochLength)),                      // E
			getPtr(types.U64(types.MaxAccumulateGas)),                 // G_A
			getPtr(types.U64(types.IsAuthorizedGas)),                  // G_I
			getPtr(types.U64(types.MaxRefineGas)),                     // G_R
			getPtr(types.U64(types.TotalGas)),                         // G_T
			getPtr(types.U16(types.MaxBlocksHistory)),                 // H
			getPtr(types.U16(types.MaximumWorkItems)),                 // I
			getPtr(types.U16(types.MaximumDependencyItems)),           // J
			getPtr(types.U16(types.MaxTicketsPerBlock)),               // K
			getPtr(types.U32(types.MaxLookupAge)),                     // L
			getPtr(types.U16(types.TicketsPerValidator)),              // N
			getPtr(types.U16(types.AuthPoolMaxSize)),                  // O
			getPtr(types.U16(types.SlotPeriod)),                       // P
			getPtr(types.U16(types.AuthQueueSize)),                    // Q
			getPtr(types.U16(types.RotationPeriod)),                   // R
			getPtr(types.U16(types.MaxExtrinsics)),                    // T
			getPtr(types.U16(types.WorkReportTimeout)),                // U
			getPtr(types.U16(types.ValidatorsCount)),                  // V
			getPtr(types.U32(types.MaxIsAuthorizedCodeSize)),          // W_A
			getPtr(types.U32(types.MaxTotalSize)),                     // W_B
			getPtr(types.U32(types.MaxServiceCodeSize)),               // W_C
			getPtr(types.U32(types.ECBasicSize)),                      // W_E
			getPtr(types.U32(types.MaxImportCount)),                   // W_M
			getPtr(types.U32(types.ECPiecesPerSegment)),               // W_P
			getPtr(types.U32(types.WorkReportOutputBlobsMaximumSize)), // W_R
			getPtr(types.U32(types.TransferMemoSize)),                 // W_T
			getPtr(types.U32(types.MaxExportCount)),                   // W_X
			getPtr(types.U32(types.SlotSubmissionEnd)),                // Y
		)
		if err != nil {
			panic(err)
		}
		fetchConstantsData = val
	})
	return fetchConstantsData
}

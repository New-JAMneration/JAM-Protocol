package PVM

import (
	"fmt"
	"math"
	"math/bits"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// OperationType Enum
type OperationType int

const (
	// ----------------- General Functions (B.5) -----------------
	GasOp      OperationType = 0 // Ω_G
	GrowHeapOp OperationType = 1 // Ω_♊
	FetchOp    OperationType = 2 // Ω_Y
	LookupOp   OperationType = 3 // Ω_L
	ReadOp     OperationType = 4 // Ω_R
	WriteOp    OperationType = 5 // Ω_W
	InfoOp     OperationType = 6 // Ω_I

	// ----------------- Refine Functions (B.6) -----------------
	HistoricalLookupOp OperationType = 7  // Ω_H
	ExportOp           OperationType = 8  // Ω_E
	MachineOp          OperationType = 9  // Ω_M
	PeekOp             OperationType = 10 // Ω_P
	PokeOp             OperationType = 11 // Ω_O
	PagesOp            OperationType = 12 // Ω_Z
	InvokeOp           OperationType = 13 // Ω_K
	ExpungeOp          OperationType = 14 // Ω_X

	// ----------------- Accumulate Functions (B.7) -----------------
	BlessOp      OperationType = 15 // Ω_B
	AssignOp     OperationType = 16 // Ω_A
	DesignateOp  OperationType = 17 // Ω_D
	CheckpointOp OperationType = 18 // Ω_C
	NewOp        OperationType = 19 // Ω_N
	UpgradeOp    OperationType = 20 // Ω_U
	TransferOp   OperationType = 21 // Ω_T
	EjectOp      OperationType = 22 // Ω_J
	QueryOp      OperationType = 23 // Ω_Q
	SolicitOp    OperationType = 24 // Ω_S
	ForgetOp     OperationType = 25 // Ω_F
	YieldOp      OperationType = 26 // Ω_Taurus
	ProvideOp    OperationType = 27 // Ω_Aries
	LogOp        OperationType = 100

	MaxOperationType OperationType = 100
)

type GeneralArgs struct {
	ServiceAccount      *types.ServiceAccount
	ServiceID           *types.ServiceID
	ServiceAccountState *types.ServiceAccountState
	CoreID              *types.CoreIndex
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
	// ServiceID           types.ServiceID         // s
	TimeSlot   types.TimeSlot          // t
	Extrinsics [][]types.ExtrinsicSpec // overline{x}, used in fetch
}

type HostCallArgs struct {
	GeneralArgs
	AccumulateArgs
	RefineArgs
	*Program
	AccumulateTrace *AccumulateTraceContext
	// CodeHash identifies the program code being run; the recompiler backend uses
	// it as the key for the cross-invocation compiled-program cache. Zero value
	// (e.g. is_authorized) bypasses the cache. Callers set it before Psi_M.
	CodeHash types.OpaqueHash
}

func getPtr[T any](v T) *T { return &v }

var hostCallName = []string{
	0:   "gas",
	1:   "grow_heap",
	2:   "fetch",
	3:   "lookup",
	4:   "read",
	5:   "write",
	6:   "info",
	7:   "historicalLookup",
	8:   "export",
	9:   "machine",
	10:  "peek",
	11:  "poke",
	12:  "pages",
	13:  "invoke",
	14:  "expunge",
	15:  "bless",
	16:  "assign",
	17:  "designate",
	18:  "checkpoint",
	19:  "new",
	20:  "upgrade",
	21:  "transfer",
	22:  "eject",
	23:  "query",
	24:  "solicit",
	25:  "forget",
	26:  "yield",
	27:  "provide",
	100: "log",
}

// HostCallName returns the human-readable host-call opcode name.
func HostCallName(op int) string {
	if op >= 0 && op < len(hostCallName) {
		return hostCallName[op]
	}
	return "unknown"
}

var HostCallFunctions Omegas = func() Omegas {
	f := make([]Omega, MaxOperationType+1)
	f[GasOp] = gas
	f[GrowHeapOp] = growHeap
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
	if result := chargeGasAndCheck(&input, HostGasUnknown); result != nil {
		return *result
	}
	input.VM.Registers[7] = WHAT
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// HostCallOutOfGas and HostCallException are exported fallbacks for JIT host-call dispatch.
var (
	HostCallOutOfGas  Omega = hostCallOutOfGas
	HostCallException Omega = hostCallException
)

// 0.7.2
func hostCallOutOfGas(input OmegaInput) (output OmegaOutput) {
	return OmegaOutput{
		ExitReason: ExitOOG,
		Addition:   input.Addition,
	}
}

// chargeGasAndCheck deducts cost from gas and returns OOG output if negative.
func chargeGasAndCheck(input *OmegaInput, cost Gas) *OmegaOutput {
	*input.VM.Gas -= cost
	if *input.VM.Gas < 0 {
		return &OmegaOutput{
			ExitReason: ExitOOG,
			Addition:   input.Addition,
		}
	}
	return nil
}

// Gas Function（ΩG）, gas = 0
func gas(input OmegaInput) OmegaOutput {
	if result := chargeGasAndCheck(&input, HostGasGas); result != nil { // M_G
		return *result
	}
	input.VM.Registers[7] = uint64(*input.VM.Gas)
	return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
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
	if input.Addition.Eta == (types.Entropy{}) {
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
	w11 := input.VM.Registers[11]
	if w11 >= uint64(len(input.Addition.Extrinsics)) {
		return nil, nil
	}
	w12 := input.VM.Registers[12]
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
	w11 := input.VM.Registers[11]
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
	w11 := input.VM.Registers[11]
	if w11 >= uint64(len(input.Addition.ImportSegments)) {
		return nil, nil
	}
	w12 := input.VM.Registers[12]
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
	w11 := input.VM.Registers[11]
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
	w11 := input.VM.Registers[11]
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
	w11 := input.VM.Registers[11]
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
	w11 := input.VM.Registers[11]
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

// growHeap = 1 | Ω_♊(g, ω, μ, jam_blob)
// ω₇ = desired heap-top page index; ω'₇ = resulting heap-top page index.
// g = M_{♊,c} + (ω₇ − h) · M_{♊,p}
func growHeap(input OmegaInput) OmegaOutput {
	n := input.VM.Registers[7]       // ω₇
	h := input.VM.Mem.HeapPages()    // h = a + c
	b := input.VM.Mem.HeapMaxPages() // b

	// ω₇ ≤ h ∨ ω₇ > b: no growth needed or invalid → charge M_{♊,c} only
	if n <= h || n > b {
		if result := chargeGasAndCheck(&input, HostGasGrowHeapConst); result != nil {
			input.VM.Registers[7] = h
			return *result
		}
		input.VM.Registers[7] = h
		return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
	}

	g := HostGasGrowHeapConst + Gas(n-h)*HostGasGrowHeapPage
	if result := chargeGasAndCheck(&input, g); result != nil {
		input.VM.Registers[7] = h
		return *result
	}

	input.VM.Mem.GrowHeapTo(n)
	input.VM.Registers[7] = n
	return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
}

// fetch = 2
func fetch(input OmegaInput) (output OmegaOutput) {
	discriminator := input.VM.Registers[10]
	if result := chargeGasAndCheck(&input, fetchCost(discriminator, input.VM.Registers[9])); result != nil {
		return *result
	}

	encoder := types.NewEncoder()
	var v *[]byte
	var val []byte
	var err error
	if discriminator < uint64(len(fetchHandlers)) {
		val, err = fetchHandlers[discriminator](input, encoder)
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
	o := input.VM.Registers[7]
	f := min(input.VM.Registers[8], dataLength)
	l := min(input.VM.Registers[9], dataLength-f)
	// nothing to write, don't need to check memory access
	if l == 0 && v != nil {
		input.VM.Registers[7] = dataLength
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	// need to first check writable
	if !input.VM.Mem.IsWriteable(o, l) && v != nil {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if v = nil
	if v == nil {
		input.VM.Registers[7] = NONE

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	input.VM.Mem.Write(o, (*v)[f:f+l])
	input.VM.Registers[7] = dataLength

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

func fetchCost(discriminator, length uint64) Gas {
	constant, rate := FetchGasCost(discriminator)
	return addGas(constant, MemGas(rate, length))
}

// addGas saturates at MaxInt64 so huge linear terms never wrap.
func addGas(parts ...Gas) Gas {
	var sum Gas
	for _, part := range parts {
		if part <= 0 {
			continue
		}
		if sum > math.MaxInt64-part {
			return math.MaxInt64
		}
		sum += part
	}
	return sum
}

// unitGasCost is const + rate·n with saturation (pages / bless / designate).
func unitGasCost(constant, rate Gas, n uint64) Gas {
	if rate <= 0 || n == 0 {
		return constant
	}
	hi, lo := bits.Mul64(uint64(rate), n)
	if hi != 0 || lo > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return addGas(constant, Gas(lo))
}

// lookup = 3
func lookup(input OmegaInput) (output OmegaOutput) {
	// g = M_{L,c} + 𝒢(M_{L,ℓ}, z); z = ω₁₁
	cost := addGas(HostGasLookupConst, MemGas(HostGasLookupOctets, input.VM.Registers[11]))
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}

	serviceID := *input.Addition.ServiceID
	serviceAccount := *input.Addition.ServiceAccount
	delta := *input.Addition.ServiceAccountState

	var a *types.ServiceAccount
	if input.VM.Registers[7] == 0xffffffffffffffff || input.VM.Registers[7] == uint64(serviceID) {
		a = &serviceAccount
	} else if value, exists := delta[types.ServiceID(input.VM.Registers[7])]; exists {
		a = &value
	}

	h, o := input.VM.Registers[8], input.VM.Registers[9]
	if !input.VM.Mem.IsReadable(h, 32) {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	preimageRawData := input.VM.Mem.Read(h, 32)

	var v *types.ByteSequence
	var f uint64
	var l uint64
	if a != nil {
		if preimage, preimageExists := a.PreimageLookup[types.OpaqueHash(preimageRawData)]; preimageExists {
			v = &preimage
		}

		if v != nil {
			f = min(input.VM.Registers[10], uint64(len(*v)))
			l = min(input.VM.Registers[11], uint64(len(*v))-f)
		}
	}

	if !input.VM.Mem.IsWriteable(o, l) && l != 0 {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	if v == nil {
		input.VM.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.VM.Registers[7] = uint64(len(*v))
	if l != 0 {
		input.VM.Mem.Write(o, (*v)[f:f+l])
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// read = 4
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(italic): ServiceID
d: ServiceAccountState (map[ServiceID]ServiceAccount)
*/
func read(input OmegaInput) (output OmegaOutput) {
	// g = M_{R,c} + 𝒢(M_{R,k,ℓ}, k_Z) + 𝒢(M_{R,v,ℓ}, v_Z)
	ko, kz, o := input.VM.Registers[8], input.VM.Registers[9], input.VM.Registers[10]
	vZ := input.VM.Registers[12]
	cost := addGas(
		HostGasReadConst,
		MemGas(HostGasReadKeyOctets, kz),
		MemGas(HostGasReadValueOctets, vZ),
	)
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}

	serviceID := *input.Addition.GeneralArgs.ServiceID
	delta := *input.Addition.GeneralArgs.ServiceAccountState
	var sStar uint64
	// assign s*
	if input.VM.Registers[7] == 0xffffffffffffffff {
		sStar = uint64(serviceID)
	} else {
		sStar = input.VM.Registers[7]
	}
	// assign ko, kz, o first and check v = panic ?
	// since v = panic is the first condition to check
	if !input.VM.Mem.IsReadable(ko, kz) {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	var a types.ServiceAccount
	callerServiceID := serviceID
	// assign a
	if sStar == uint64(serviceID) {
		a = delta[serviceID]
	} else if value, exists := delta[types.ServiceID(sStar)]; exists {
		a = value
		serviceID = types.ServiceID(sStar)
	} else {
		// a = nil , v not panic, => v = nil
		input.VM.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// v = a_s[k]?  ,  a = nil is checked, only check k in Key(a_s)
	// first compute k , mu_ko...+kz
	storageRawKey := input.VM.Mem.Read(ko, kz)
	v, exists := a.StorageDict[string(storageRawKey)]
	storageValueFromKeyVal := getStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
	// v = nil
	if !exists {
		if storageValueFromKeyVal == nil { // check storage state key-val
			input.VM.Registers[7] = NONE
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		} else {
			v = *storageValueFromKeyVal

			// Only cache the unmatched pool entry into StorageDict (and remove from pool)
			// when the caller is reading its OWN storage. For cross-service reads,
			// ΩR must be side-effect free — otherwise the target service's storage
			// entry silently disappears from the global unmatched pool and is lost
			// if the target service isn't part of the accumulating set this block.
			if callerServiceID == serviceID {
				a.StorageDict[string(storageRawKey)] = v
				input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
				removeStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
			}
		}
	}

	f := min(input.VM.Registers[11], uint64(len(v)))
	l := min(input.VM.Registers[12], uint64(len(v))-f)
	// nothing to write, don't need to check memory access
	if l == 0 {
		input.VM.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// first check not writable, then check v = nil (not exists)
	if !input.VM.Mem.IsWriteable(o, l) {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.VM.Registers[7] = uint64(len(v))
	input.VM.Mem.Write(o, v[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// write = 5
func write(input OmegaInput) (output OmegaOutput) {
	// g = M_{W,c} + 𝒢(M_{W,k,ℓ}, k_Z) + 𝒢(M_{W,v,ℓ}, v_Z)
	ko, kz, vo, vz := input.VM.Registers[7], input.VM.Registers[8], input.VM.Registers[9], input.VM.Registers[10]
	cost := addGas(
		HostGasWriteConst,
		MemGas(HostGasWriteKeyOctets, kz),
		MemGas(HostGasWriteValOctets, vz),
	)
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}
	if !input.VM.Mem.IsReadable(ko, kz) {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}
	// compute \mathbb{k}
	storageRawKey := input.VM.Mem.Read(ko, kz)

	serviceID := *input.Addition.GeneralArgs.ServiceID
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
	} else if input.VM.Mem.IsReadable(vo, vz) { // storage append/update
		storageRawData := input.VM.Mem.Read(vo, vz)

		// compute items, octets , check a_t > a_b first (GP: a_minbalance > a_balance → FULL, s' = s)
		newItems := a.ServiceInfo.Items - footprintItems
		newOctets := a.ServiceInfo.Bytes - footprintOctets

		storageItems, storageOctets := service_account.CalcStorageItemfootprint(string(storageRawKey), storageRawData)
		newItems += storageItems
		newOctets += storageOctets
		newMinBalance := service_account.CalcThresholdBalance(newItems, newOctets, a.ServiceInfo.DepositOffset) // a_t
		if newMinBalance > a.ServiceInfo.Balance {
			input.VM.Registers[7] = FULL
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}

		// balance check passed, now apply the storage mutation
		a.StorageDict[string(storageRawKey)] = storageRawData
		removeStorageFromKeyVal(input.Addition.GeneralArgs.StorageKeyVal, serviceID, storageRawKey)
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

	input.VM.Registers[7] = l

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// info = 6
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(italic): ServiceID
d: ServiceAccountState (map[ServiceID]ServiceAccount)
*/
func info(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input, HostGasInfo); result != nil { // M_I
		return *result
	}

	serviceID := *input.Addition.ServiceID
	delta := *input.Addition.ServiceAccountState

	var a types.ServiceAccount
	if input.VM.Registers[7] == 0xffffffffffffffff {
		a = delta[serviceID]
	} else {
		value, exist := delta[types.ServiceID(input.VM.Registers[7])]
		if exist {
			a = value
		} else {
			// v = nil , l = 0 -> don't need to check writeable
			input.VM.Registers[7] = NONE
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}
	}

	minBalance := service_account.CalcThresholdBalance(a.ServiceInfo.Items, a.ServiceInfo.Bytes, a.ServiceInfo.DepositOffset)
	encoder := types.NewEncoder()
	v, _ := encoder.EncodeMany(
		&a.ServiceInfo.CodeHash,             // a_c
		&a.ServiceInfo.Balance,              // a_b
		&minBalance,                         // a_t
		&a.ServiceInfo.MinItemGas,           // a_g
		&a.ServiceInfo.MinMemoGas,           // a_m
		&a.ServiceInfo.Bytes,                // a_o
		&a.ServiceInfo.Items,                // a_i
		&a.ServiceInfo.DepositOffset,        // a_f
		&a.ServiceInfo.CreationSlot,         // a_r
		&a.ServiceInfo.LastAccumulationSlot, // a_a
		&a.ServiceInfo.ParentService,        // a_p
	)
	f := min(input.VM.Registers[9], uint64(len(v)))
	l := min(input.VM.Registers[10], uint64(len(v))-f)
	o := input.VM.Registers[8]
	// nothing to write
	if l == 0 {
		input.VM.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	// if mathbf{N}_{o..._l} \not in mathbf{V}^*_mu
	if !input.VM.Mem.IsWriteable(o, l) { // v = ∇ not defined
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.VM.Registers[7] = uint64(len(v))
	input.VM.Mem.Write(o, v[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// log = 100 , [JIP-1](https://hackmd.io/@polkadot/jip1)
// g = 10 (fixed; not in Gray Paper)
// Output registers: {} (none modified per spec)
// Side-effects: "No side-effects if memory access is invalid."
func logHostCall(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input, HostGasLog); result != nil {
		return *result
	}

	level := input.VM.Registers[7]
	msgPtr, msgLen := input.VM.Registers[10], input.VM.Registers[11]

	if !input.VM.Mem.IsReadable(msgPtr, msgLen) {
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	message := input.VM.Mem.Read(msgPtr, msgLen)
	levelStr := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}

	if level > 4 {
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	timeFormat := time.RFC3339
	timeStamp := time.Now().Format(timeFormat)
	var logMsg string
	if input.VM.Registers[8] == 0 && input.VM.Registers[9] == 0 {
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v][%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreID), derefernceOrNil(input.Addition.ServiceID), string(message))
	} else {
		tgtPtr, tgtLen := input.VM.Registers[8], input.VM.Registers[9]
		if !input.VM.Mem.IsReadable(tgtPtr, tgtLen) {
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}
		target := input.VM.Mem.Read(tgtPtr, tgtLen)
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v][%s][%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreID), derefernceOrNil(input.Addition.ServiceID), target, string(message))
	}

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
func check(serviceID types.ServiceID, serviceAccountState types.ServiceAccountState) types.ServiceID {
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
func getStorageFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceID, storageKey types.ByteSequence) *types.ByteSequence {
	requestedStorageStateKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageKey, nil)
	for _, v := range *keyVal {
		if v.Key == requestedStorageStateKey.Key {
			return &v.Value
		}
	}

	return nil
}

func removeStorageFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceID, storageKey types.ByteSequence) {
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

func getLookupItemFromKeyVal(keyVal *types.StateKeyVals, serviceID types.ServiceID, lookupKey types.LookupMetaMapkey) []byte {
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

// memGasUnit is the octet granularity of the linear gas term: rates in
// gas_const.go are quoted per this many octets.
const memGasUnit = 1024

// MemGas is 𝒢(L, ℓ) (formula B.17): the memory term of
// a host-call gas cost, ⌈L × ℓ / 1024⌉, for a rate L of gas per 1024 octets over
// a length ℓ.
//
// Callers pass ℓ straight from a guest register, so the product can exceed the
// range of Gas. Because such a cost is unpayable regardless of its exact value,
// the result saturates at the maximum instead of wrapping.
func MemGas(rate Gas, length uint64) Gas {
	if rate <= 0 || length == 0 {
		return 0
	}

	hi, lo := bits.Mul64(uint64(rate), length)
	if hi != 0 {
		return math.MaxInt64
	}

	rounded := lo + memGasUnit - 1
	if rounded < lo {
		return math.MaxInt64
	}

	return Gas(rounded / memGasUnit)
}
package PVM

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
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
)

type HistoryState struct {
	PreviousGas       Gas
	PreviousRegisters Registers
	PreviousMemory    PageMap
}

type OmegaInput struct {
	Operation OperationType // operation type
	Gas       Gas           // gas counter
	Registers Registers     // PVM registers
	Memory    Memory        // memory
	Addition  HostCallArgs  // Extra parameter for each host-call function
}
type OmegaOutput struct {
	ExitReason   error        // Exit reason
	NewGas       Gas          // New Gas
	NewRegisters Registers    // New Register
	NewMemory    Memory       // New Memory
	Addition     HostCallArgs // addition host-call context
}

// Ω⟨X⟩
type (
	Omega  func(OmegaInput) OmegaOutput
	Omegas map[OperationType]Omega
)

type GeneralArgs struct {
	ServiceAccount      types.ServiceAccount
	ServiceId           *types.ServiceId
	ServiceAccountState types.ServiceAccountState
	CoreId              *types.CoreIndex
}

type AccumulateArgs struct {
	ResultContextX ResultContext
	ResultContextY ResultContext
	Timeslot       types.TimeSlot
	Eta            types.Entropy   // italic n / eta_0, used in fetch
	Operands       []types.Operand // o, used in fetch
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
	ServiceID           types.ServiceId         // s
	TimeSlot            types.TimeSlot          // t
	Extrinsics          [][]types.ExtrinsicSpec // overline{x}, used in fetch
}

type OnTransferArgs struct {
	DeferredTransfer []types.DeferredTransfer // bold{t}
}

type HostCallArgs struct {
	GeneralArgs
	AccumulateArgs
	RefineArgs
	OnTransferArgs
	Program
}

type Psi_H_ReturnType struct {
	ExitReason error        // exit reason
	Counter    uint32       // new instruction counter
	Gas        Gas          // gas remain
	Reg        Registers    // new registers
	Ram        Memory       // new memory
	Addition   HostCallArgs // addition host-call context
}

// JIT version of (A.34) Ψ_H
func HostCall(program Program, pc ProgramCounter, gas types.Gas, reg Registers, ram Memory, omegas Omegas, addition HostCallArgs,
) (psi_result Psi_H_ReturnType) {
	var exitReason error
	var pcPrime ProgramCounter
	var gasPrime Gas
	var regPrime Registers
	var memPrime Memory

	if GasChargingMode == "blockBased" {
		exitReason, pcPrime, gasPrime, regPrime, memPrime = BlockBasedInvoke(program, pc, Gas(gas), reg, ram)
	} else {
		exitReason, pcPrime, gasPrime, regPrime, memPrime = SingleStepInvoke(program, pc, Gas(gas), reg, ram)
	}

	reason := exitReason.(*PVMExitReason)
	logger.SetShowLine(true)
	if reason.Reason == HALT || reason.Reason == PANIC || reason.Reason == OUT_OF_GAS || reason.Reason == PAGE_FAULT {
		psi_result.ExitReason = PVMExitTuple(reason.Reason, nil)
		psi_result.Counter = uint32(pcPrime)
		psi_result.Gas = gasPrime
		psi_result.Reg = regPrime
		psi_result.Ram = memPrime
		psi_result.Addition = addition
	} else if reason.Reason == HOST_CALL {
		var input OmegaInput
		input.Operation = OperationType(*reason.HostCall)
		input.Gas = gasPrime
		input.Registers = regPrime
		input.Memory = ram
		input.Addition = addition

		omega := omegas[input.Operation]
		if omega == nil {
			omega = hostCallException
		}
		omega_result := omega(input)
		var pvmExit *PVMExitReason
		if !errors.As(omega_result.ExitReason, &pvmExit) {
			logger.Errorf("%s host-call error : %v",
				hostCallName[input.Operation], omega_result.ExitReason)
			return
		}
		omega_reason := omega_result.ExitReason.(*PVMExitReason)
		logger.Debugf("%s host-call return: %s, gas : %d -> %d\nRegisters: %v\n",
			hostCallName[input.Operation], omega_reason, gasPrime, omega_result.NewGas, omega_result.NewRegisters)
		if omega_reason.Reason == PAGE_FAULT {
			psi_result.Counter = uint32(pcPrime)
			psi_result.Gas = gasPrime
			psi_result.Reg = regPrime
			psi_result.Ram = memPrime
			psi_result.ExitReason = PVMExitTuple(PAGE_FAULT, *omega_reason.FaultAddr)
			psi_result.Addition = addition
		} else if omega_reason.Reason == CONTINUE {
			return HostCall(program, pcPrime, types.Gas(omega_result.NewGas), omega_result.NewRegisters, omega_result.NewMemory, omegas, omega_result.Addition)
		} else if omega_reason.Reason == PANIC || omega_reason.Reason == OUT_OF_GAS || omega_reason.Reason == HALT {
			psi_result.ExitReason = omega_result.ExitReason
			psi_result.Counter = uint32(pcPrime)
			psi_result.Gas = omega_result.NewGas
			psi_result.Reg = omega_result.NewRegisters
			psi_result.Ram = omega_result.NewMemory
			psi_result.Addition = omega_result.Addition
		}
	}
	return
}

func getPtr[T any](v T) *T { return &v }

var hostCallName = map[OperationType]string{
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

var HostCallFunctions = map[OperationType]Omega{
	0:   gas,
	1:   fetch,
	2:   lookup,
	3:   read,
	4:   write,
	5:   info,
	6:   historicalLookup,
	7:   export,
	8:   machine,
	9:   peek,
	10:  poke,
	11:  pages,
	12:  invoke,
	13:  expunge,
	14:  bless,
	15:  assign,
	16:  designate,
	17:  checkpoint,
	18:  new,
	19:  upgrade,
	20:  transfer,
	21:  eject,
	22:  query,
	23:  solicit,
	24:  forget,
	25:  yield,
	26:  provide,
	100: logHostCall,
}

func hostCallException(input OmegaInput) (output OmegaOutput) {
	// non-defined host call
	input.Registers[7] = WHAT
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       input.Gas - 10,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// Gas Function（ΩG）, gas = 0
func gas(input OmegaInput) OmegaOutput {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(newGas)
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// fetch = 1
func fetch(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	var (
		v   []byte
		err error
	)
	encoder := types.NewEncoder()
	switch input.Registers[10] {
	case 0:
		v, err = encoder.EncodeMany(
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
			logger.Errorf("fetch host-call case 0 encode error: %v", err)
			return OmegaOutput{
				ExitReason:   err,
				NewGas:       newGas,
				NewRegisters: input.Registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		}
	case 1:
		if reflect.ValueOf(input.Addition.Eta).IsZero() {
			break
		}
		v, err = encoder.Encode(&input.Addition.Eta)
	case 2:
		if input.Addition.AuthOutput == nil {
			break
		}

		v, err = encoder.Encode(&input.Addition.AuthOutput)
	case 3:
		if len(input.Addition.Extrinsics) == 0 {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.Extrinsics)) {
			break
		}

		w12 := input.Registers[12]
		if w12 >= uint64(len(input.Addition.Extrinsics[w11])) {
			break
		}

		v, err = encoder.Encode(&input.Addition.Extrinsics[w11][w12])
	case 4:
		// check \bar{x}
		if len(input.Addition.Extrinsics) == 0 {
			break
		}
		// check i

		if input.Addition.WorkItemIndex == nil {
			break
		}

		i := *input.Addition.WorkItemIndex

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.Extrinsics[i])) {
			break
		}

		v, err = encoder.Encode(input.Addition.Extrinsics[i][w11])
	case 5:
		if len(input.Addition.ImportSegments) == 0 {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.ImportSegments)) {
			break
		}

		w12 := input.Registers[12]
		if w12 >= uint64(len(input.Addition.ImportSegments[w11])) {
			break
		}

		v, err = encoder.Encode(input.Addition.ImportSegments[w11][w12])
	case 6:
		// check \bar{i}
		if len(input.Addition.ImportSegments) == 0 {
			break
		}

		// check i
		if input.Addition.WorkItemIndex == nil {
			break
		}

		i := *input.Addition.WorkItemIndex
		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.ImportSegments[i])) {
			break
		}

		v, err = encoder.Encode(input.Addition.ImportSegments[i][w11])
	case 7:
		if input.Addition.WorkPackage == nil {
			break
		}

		v, err = encoder.Encode(*input.Addition.WorkPackage)
	case 8:
		if input.Addition.WorkPackage == nil {
			break
		}

		v, err = encoder.EncodeMany(
			&input.Addition.WorkPackage.AuthCodeHash,
			&input.Addition.WorkPackage.AuthorizerConfig,
		)
	case 9:
		if input.Addition.WorkPackage == nil {
			break
		}

		v, err = encoder.Encode(input.Addition.WorkPackage.Authorization)
	case 10:
		if input.Addition.WorkPackage == nil {
			break
		}

		v, err = encoder.Encode(input.Addition.WorkPackage.Context)
	case 11:
		if input.Addition.WorkPackage == nil {
			break
		}

		buffer, err := encoder.Encode(types.U64(len(input.Addition.WorkPackage.Items)))
		if err != nil {
			break
		}

		for _, w := range input.Addition.WorkPackage.Items {
			sw, err := S(encoder, w)
			if err != nil {
				break
			}

			buffer = append(buffer, sw...)
		}

		v = buffer
	case 12:
		if input.Addition.WorkPackage == nil {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.WorkPackage.Items)) {
			break
		}

		v, err = S(encoder, input.Addition.WorkPackage.Items[w11])
	case 13:
		if input.Addition.WorkPackage == nil {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.WorkPackage.Items)) {
			break
		}

		v, err = encoder.Encode(input.Addition.WorkPackage.Items[w11].Payload)
	case 14:
		if len(input.Addition.Operands) == 0 {
			break
		}

		buffer, err := encoder.EncodeUint(uint64((len(input.Addition.Operands))))
		if err != nil {
			break
		}

		for _, o := range input.Addition.Operands {
			bytes, err := encoder.Encode(&o)
			if err != nil {
				break
			}

			buffer = append(buffer, bytes...)
		}

		v = buffer
	case 15:
		if len(input.Addition.Operands) == 0 {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.Operands)) {
			break
		}

		v, err = encoder.Encode(input.Addition.Operands[w11])
	case 16:
		if len(input.Addition.DeferredTransfer) == 0 {
			break
		}

		buffer, err := encoder.Encode(types.U64(len(input.Addition.DeferredTransfer)))
		if err != nil {
			break
		}

		for _, t := range input.Addition.DeferredTransfer {
			bytes, err := encoder.Encode(t)
			if err != nil {
				break
			}

			buffer = append(buffer, bytes...)
		}

		v = buffer
	case 17:
		if len(input.Addition.DeferredTransfer) == 0 {
			break
		}

		w11 := input.Registers[11]
		if w11 >= uint64(len(input.Addition.DeferredTransfer)) {
			break
		}

		v, err = encoder.Encode(input.Addition.DeferredTransfer[w11])
	}

	if err != nil {
		v = nil
	}

	dataLength := uint64(len(v))
	o := input.Registers[7]
	f := min(input.Registers[8], dataLength)
	l := min(input.Registers[9], dataLength-f)
	// nothing to write, don't need to check memory access
	if l == 0 {
		input.Registers[7] = dataLength
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// need to first check writable
	if !isWriteable(o, l, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if v = nil
	if len(v) == 0 {
		input.Registers[7] = NONE

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	input.Memory.Write(o, l, v[f:])
	input.Registers[7] = dataLength

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// ΩL(ϱ, ω, μ, s, s, d) , lookup = 2
func lookup(input OmegaInput) (output OmegaOutput) {
	serviceID := input.Addition.ResultContextX.ServiceId
	serviceAccount := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	delta := input.Addition.ResultContextX.PartialState.ServiceAccounts

	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	var a types.ServiceAccount
	if input.Registers[7] == 0xffffffffffffffff || input.Registers[7] == uint64(serviceID) {
		a = serviceAccount
	} else if value, exists := delta[types.ServiceId(input.Registers[7])]; exists {
		a = value
	} else {
		new_registers := input.Registers
		new_registers[7] = NONE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h, o := input.Registers[8], input.Registers[9]
	if !isReadable(h, 32, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	var concated_bytes []byte
	for address := uint32(h); address < uint32(h)+32; address++ {
		page := address / ZP
		index := address % ZP
		concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
	}
	v, exist := a.PreimageLookup[types.OpaqueHash(concated_bytes)]
	if !exist {
		new_registers := input.Registers
		new_registers[7] = NONE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	f := min(input.Registers[10], uint64(len(v)))
	l := min(input.Registers[11], uint64(len(v))-f)

	// nothing to write, don't need to check memory access
	if l == 0 {
		input.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if !isWriteable(o, l, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	} else {
		new_registers := input.Registers
		new_registers[7] = uint64(len(v))
		new_memory := input.Memory

		for offset := uint32(0); offset < uint32(l); offset++ {
			address := uint32(offset + uint32(o))
			page := address / ZP
			index := address % ZP
			new_memory.Pages[page].Value[index] = v[uint32(f)+offset]
		}
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    new_memory,
			Addition:     input.Addition,
		}
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
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	delta := input.Addition.ResultContextX.PartialState.ServiceAccounts

	var sStar uint64
	// assign s*
	if input.Registers[7] == 0xffffffffffffffff {
		sStar = uint64(serviceID)
	} else {
		sStar = input.Registers[7]
	}
	// assign ko, kz, o first and check v = panic ?
	// since v = panic is the first condition to check
	ko, kz, o := input.Registers[8], input.Registers[9], input.Registers[10]
	if !isReadable(ko, kz, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
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
		new_registers := input.Registers
		new_registers[7] = NONE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// v = a_s[k]?  ,  a = nil is checked, only check k in Key(a_s)
	// first compute k , mu_ko...+kz
	storageRawKey := input.Memory.Read(ko, kz)
	v, exists := a.StorageDict[string(storageRawKey)]
	storageValueFromKeyVal := input.Addition.ResultContextX.getStorageFromKeyVal(serviceID, storageRawKey)

	// v = nil
	if !exists {
		if storageValueFromKeyVal == nil { // check storage state key-val
			new_registers := input.Registers
			new_registers[7] = NONE
			return OmegaOutput{
				ExitReason:   PVMExitTuple(CONTINUE, nil),
				NewGas:       newGas,
				NewRegisters: new_registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		} else {
			v = *storageValueFromKeyVal

			// store the unknown storage item to state
			a.StorageDict[string(storageRawKey)] = v
			input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
			// remove from storageKeyVal
			input.Addition.ResultContextX.removeStorageFromKeyVal(serviceID, storageRawKey)
		}
	}

	f := min(input.Registers[11], uint64(len(v)))
	l := min(input.Registers[12], uint64(len(v))-f)
	// nothing to write, don't need to check memory access
	if l == 0 {
		input.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// first check not writable, then check v = nil (not exists)
	if !isWriteable(o, l, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	new_registers := input.Registers
	new_registers[7] = uint64(len(v))
	new_memory := input.Memory
	for i := uint32(0); i < uint32(l); i++ {
		address := i + uint32(o)
		page := address / ZP
		index := address % ZP
		new_memory.Pages[page].Value[index] = v[uint32(f)+i]
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: new_registers,
		NewMemory:    new_memory,
		Addition:     input.Addition,
	}
}

// ΩW (ϱ, ω, μ, s, s) , write = 4
func write(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10

	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	ko, kz, vo, vz := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]
	if !isReadable(ko, kz, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// compute \mathbb{k}
	storageRawKey := input.Memory.Read(ko, kz)

	serviceID := input.Addition.ResultContextX.ServiceId
	a := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]

	value, storageRawKeyExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID].StorageDict[string(storageRawKey)]

	storageRawData := input.Addition.ResultContextX.getStorageFromKeyVal(serviceID, storageRawKey)

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

	if vz == 0 { // remove storage
		delete(a.StorageDict, string(storageRawKey))
		input.Addition.ResultContextX.removeStorageFromKeyVal(serviceID, storageRawKey)

		// direct update items, octets
		a.ServiceInfo.Items -= footprintItems
		a.ServiceInfo.Bytes -= footprintOctets
	} else if isReadable(vo, vz, input.Memory) { // storage append/update
		storageRawData := input.Memory.Read(vo, vz)
		a.StorageDict[string(storageRawKey)] = storageRawData
		input.Addition.ResultContextX.removeStorageFromKeyVal(serviceID, storageRawKey)

		// compute items, octets , check a_t > a_b first
		newItems := a.ServiceInfo.Items - footprintItems
		newOctets := a.ServiceInfo.Bytes - footprintOctets

		storageItems, storageOctets := service_account.CalcStorageItemfootprint(string(storageRawKey), storageRawData)
		newItems += storageItems
		newOctets += storageOctets

		newMinBalance := service_account.CalcThresholdBalance(newItems, newOctets, a.ServiceInfo.DepositOffset) // a_t
		if newMinBalance > a.ServiceInfo.Balance {
			input.Registers[7] = FULL
			return OmegaOutput{
				ExitReason:   PVMExitTuple(CONTINUE, nil),
				NewGas:       newGas,
				NewRegisters: input.Registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		}
		// update items, octets
		a.ServiceInfo.Items = newItems
		a.ServiceInfo.Bytes = newOctets

	} else {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// update service state
	input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
	input.Registers[7] = l

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
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
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	delta := input.Addition.ResultContextX.PartialState.ServiceAccounts

	var a types.ServiceAccount
	var empty bool
	empty = true
	if input.Registers[7] == 0xffffffffffffffff {
		value, exist := delta[types.ServiceId(serviceID)]
		if exist {
			a = value
			empty = false
		}
	} else {
		value, exist := delta[types.ServiceId(input.Registers[7])]
		if exist {
			a = value
			empty = false
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

	f := min(input.Registers[9], uint64(len(v)))
	l := min(input.Registers[10], uint64(len(v))-f)
	o := input.Registers[8]
	// nothing to write
	if l == 0 {
		input.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// if mathbf{N}_{o..._l} \not in mathbf{V}^*_mu
	if !isWriteable(o, l, input.Memory) { // v = ∇ not defined
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// if a is nil => v = nil
	if empty {
		input.Registers[7] = NONE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(len(v))
	input.Memory.Write(o, l, v)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// historical_lookup = 6
func historicalLookup(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	// first check v panic, then assign a
	h, o := input.Registers[8], input.Registers[9]

	offset := uint64(32)
	if !isReadable(h, offset, input.Memory) { // not readable, return panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	codeHash := types.OpaqueHash(input.Memory.Read(h, offset))

	// assign a
	var a types.ServiceAccount
	var v types.ByteSequence

	a, accountExists := input.Addition.ServiceAccountState[input.Addition.ServiceID]
	if accountExists && input.Registers[7] == 0xffffffffffffffff {
		v = service_account.HistoricalLookup(a, input.Addition.TimeSlot, codeHash)
	} else if a, accountExists := input.Addition.ServiceAccountState[types.ServiceId(input.Registers[7])]; accountExists {
		v = service_account.HistoricalLookup(a, input.Addition.TimeSlot, codeHash)
	} else {
		// otherwise if a = nil => v = nil, here will not check writeable first, since no need to write in memory
		input.Registers[7] = NONE

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	f := min(input.Registers[10], uint64(len(v)))
	l := min(input.Registers[11], uint64(len(v))-f)

	if l == 0 {
		input.Registers[7] = uint64(len(v))
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if !isWriteable(o, l, input.Memory) { // not writeable, return panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(len(v))

	offset = l
	input.Memory.Write(o, offset, v)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// export = 7
func export(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	p := input.Registers[7]
	z := min(input.Registers[8], types.SegmentSize)

	if !isReadable(p, z, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	segmentLength := input.Addition.ExportSegmentOffset + uint(len(input.Addition.ExportSegment))
	// otherwise if ζ + |e| >= W_M
	if segmentLength > types.MaxExportCount {
		input.Registers[7] = FULL
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// data = mu_p...+z
	data := input.Memory.Read(p, z)
	x := zeroPadding(data, types.SegmentSize)
	exportSegment := types.ExportSegment{}
	copy(exportSegment[:], x)

	input.Registers[7] = uint64(input.Addition.ExportSegmentOffset) + uint64(segmentLength)
	input.Addition.ExportSegment = append(input.Addition.ExportSegment, exportSegment)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// machine = 8
func machine(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	po, pz, i := input.Registers[7], input.Registers[8], input.Registers[9]
	// pz = offset
	if !isReadable(po, pz, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	p := input.Memory.Read(po, pz)

	// find first i not in K(m)
	n := uint64(0)
	for ; n <= ^uint64(0); n++ {
		if _, pvmTypeExists := input.Addition.IntegratedPVMMap[n]; !pvmTypeExists {
			break
		}
	}

	var u Memory
	_, exitReason := DeBlobProgramCode(p)
	// otherwise if deblob(p) = PANIC
	if exitReason.(*PVMExitReason).Reason == PANIC {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	input.Registers[7] = n
	input.Addition.IntegratedPVMMap[n] = IntegratedPVMType{
		ProgramCode: ProgramCode(p),
		Memory:      u,
		PC:          ProgramCounter(i),
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// peek = 9
func peek(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n, o, s, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if z == 0 {
		input.Registers[7] = OK
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// z = offset
	if !isWriteable(o, z, input.Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if N_s...+z not subset of \mathbf{V}_m[n]_u
	// can be simplify to check readable, if not readable => Inaccessible
	if !isReadable(s, z, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	// read data from m[n]_u first
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	data := integratedPVMType.Memory.Read(s, z)
	// write data into memory
	input.Memory.Write(o, z, data)

	input.Registers[7] = OK
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// poke = 10
func poke(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n, s, o, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if !isReadable(s, z, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if N_o...+z not subset of \mathbf{V}_m[n]_u
	if !isWriteable(o, z, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	// read data from memory first
	data := input.Memory.Read(s, z)
	// write data into m[n]_u
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	integratedPVMType.Memory.Write(o, z, data)
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// pages = 11 , GP 0.6.7 void is renamed pages
func pages(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n, p, c := input.Registers[7], input.Registers[8], input.Registers[9]
	// u = panic
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		// u = panic
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if p < 16 or p + c >= 2^32 / ZP or i in N_p...+c : (u_A)_i = nil
	if p < 16 || p+c >= (1<<32)/ZP || !isReadable(p, c, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise : ok
	for i := uint32(p); i < uint32(c); i++ {
		input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
			Value:  make([]byte, ZP),
			Access: MemoryInaccessible,
		}
	}

	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// invoke = 12
func invoke(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n, o := input.Registers[7], input.Registers[8]

	offset := uint64(112)
	// g = panic
	if !isWriteable(o, offset, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in M
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// assign g, w  |  g => gas , w => registers[13]   , 8(gas) + 8(uint64) * 13 = 112
	var gas uint64
	var w Registers

	// first read data from memory
	data := input.Memory.Read(o, offset)

	decoder := types.NewDecoder()
	// decode gas
	err := decoder.Decode(data[:8], gas)
	if err != nil {
		log.Printf("host-call function \"invoke\" decode gas error : %v", err)
	}
	// decode registers
	for i := uint64(1); i < offset/8; i++ {
		err = decoder.Decode(data[8*i:8*(i+1)], w[i-1])
		if err != nil {
			log.Printf("host-call function \"invoke\" decode register:%d error : %v", i-1, err)
		}
	}
	// psi
	input.Addition.Program.InstructionData = input.Addition.IntegratedPVMMap[n].ProgramCode

	var c error
	var pcPrime ProgramCounter
	var gasPrime Gas
	var wPrime Registers
	var uPrime Memory

	if GasChargingMode == "blockBased" {
		c, pcPrime, gasPrime, wPrime, uPrime = BlockBasedInvoke(input.Addition.Program, input.Addition.IntegratedPVMMap[n].PC, Gas(gas), w, input.Addition.IntegratedPVMMap[n].Memory)
	} else {
		c, pcPrime, gasPrime, wPrime, uPrime = SingleStepInvoke(input.Addition.Program, input.Addition.IntegratedPVMMap[n].PC, Gas(gas), w, input.Addition.IntegratedPVMMap[n].Memory)
	}

	// mu* = mu
	encoder := types.NewEncoder()
	data = types.ByteSequence(make([]byte, offset))
	encoded, _ := encoder.Encode(gasPrime)
	copy(data, encoded)
	for i := uint64(1); i < offset/8; i++ {
		encoded, _ := encoder.Encode(wPrime[i-1])
		copy(data[8*i:8*(i+1)], encoded)
	}
	// write data into memory (mu)
	input.Memory.Write(o, offset, data)

	// m* = m
	tmp := input.Addition.IntegratedPVMMap[n]
	tmp.Memory = uPrime
	if c.(*PVMExitReason).Reason == HOST_CALL {
		tmp.PC = pcPrime + 1
	} else {
		tmp.PC = pcPrime
	}
	input.Addition.IntegratedPVMMap[n] = tmp

	switch c.(*PVMExitReason).Reason {
	case HOST_CALL:
		input.Registers[7] = INNERHOST
		input.Registers[8] = *c.(*PVMExitReason).HostCall

	case PAGE_FAULT:
		input.Registers[7] = INNERFAULT
		input.Registers[8] = *c.(*PVMExitReason).FaultAddr

	case OUT_OF_GAS:
		input.Registers[7] = INNEROOG

	case PANIC:
		input.Registers[7] = INNERPANIC

	case HALT:
		input.Registers[7] = INNERHALT

	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// expunge = 13
func expunge(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n := input.Registers[7]
	// n not in K(m)
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Registers[7] = WHO

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(input.Addition.IntegratedPVMMap[n].PC)
	// m ˋ n
	delete(input.Addition.IntegratedPVMMap, n)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// bless = 14
func bless(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	m, a, v, o, n := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10], input.Registers[11]

	// if N_{a...+4C} not readable
	offset := uint64(4 * types.CoresCount)
	if !isReadable(a, offset, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// \mathbb{a}
	rawData := input.Memory.Read(a, uint64(4*types.CoresCount))
	var assignData types.ServiceIdList
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, assignData)
	if err != nil {
		log.Printf("host-call function \"bless\" decode assignData error : %v", err)
	}

	offset = uint64(12 * n)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if x_s ≠ (x_u)_m
	if input.Addition.ResultContextX.ServiceId != input.Addition.ResultContextX.PartialState.Bless {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// (m, a, v) \not in N_s
	limit := uint64(1 << 32)
	if m >= limit || a >= limit || v >= limit {
		input.Registers[7] = WHO

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// otherwise
	// read data from memory, might cross many pages
	rawData = input.Memory.Read(o, offset)

	// s -> g this will update into (x_u)_x => partialState.Chi_g, decode rawData
	alwaysAccum := types.AlwaysAccumulateMap{}
	err = decoder.Decode(rawData, alwaysAccum)
	if err != nil {
		log.Printf("host-call function \"bless\" decode alwaysAccum error : %v", err)
	}

	input.Registers[7] = OK

	input.Addition.ResultContextX.PartialState.Assign = assignData
	input.Addition.ResultContextX.PartialState.AlwaysAccum = alwaysAccum

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// assign = 15
func assign(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	c, o, a := input.Registers[7], input.Registers[8], input.Registers[9]

	offset := uint64(32 * types.AuthQueueSize)
	if !isReadable(o, offset, input.Memory) { // not readable, panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if c >= C
	if c >= uint64(types.CoresCount) {
		input.Registers[7] = CORE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if x_s ≠ (x_u)_a[c]
	if input.Addition.ResultContextX.ServiceId != input.Addition.ResultContextX.PartialState.Assign[c] {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	rawData := input.Memory.Read(o, offset)

	// decode rawData , authQueue = mathbb{q}
	authQueue := types.AuthQueue{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, &authQueue)
	if err != nil {
		log.Printf("host-call function \"assign\" decode error : %v", err)
	}

	input.Addition.ResultContextX.PartialState.Authorizers[c] = authQueue
	input.Addition.ResultContextX.PartialState.Assign[c] = types.ServiceId(a)
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// designate = 16
func designate(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o := input.Registers[7]

	offset := uint64(336 * types.ValidatorsCount)
	if !isReadable(o, offset, input.Memory) { // not readable, panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if x_s ≠ (x_u)_v
	if input.Addition.ResultContextX.ServiceId != input.Addition.ResultContextX.PartialState.Designate {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// 336 * types.ValidatorsCount might cross many pages
	rawData := input.Memory.Read(o, offset) // bold{v}

	validatorsData := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, validatorsData)
	if err != nil {
		log.Printf("host-call function \"designate\" decode validatorsData error : %v", err)
	}

	input.Addition.ResultContextX.PartialState.ValidatorKeys = validatorsData
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// checkpoint = 17
func checkpoint(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	input.Addition.ResultContextY = input.Addition.ResultContextX
	input.Registers[7] = uint64(newGas)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// new = 18
func new(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, l, g, m, f := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10], input.Registers[11]

	offset := uint64(32)
	// if c = ∇
	if !(isReadable(o, offset, input.Memory) && l < (1<<32)) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if f ≠ 0 and x_s ≠ (x_u)_m
	if f != 0 && input.Addition.ResultContextX.ServiceId != input.Addition.ResultContextY.PartialState.Bless {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	c := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	s, sExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !sExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Printf("host-call function \"new\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	var cDecoded types.U32
	decoder := types.NewDecoder()
	err := decoder.Decode(c, &cDecoded)
	if err != nil {
		log.Printf("host-call function \"new\" decode error %v: ", err)
	}

	// new an account
	a := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			CodeHash:             types.OpaqueHash(c),                     // c
			Balance:              0,                                       // b, will be updated later
			MinItemGas:           types.Gas(g),                            // g
			MinMemoGas:           types.Gas(m),                            // m
			CreationSlot:         input.Addition.AccumulateArgs.Timeslot,  // r
			DepositOffset:        types.U64(0),                            // f
			LastAccumulationSlot: types.TimeSlot(0),                       // a
			ParentService:        input.Addition.ResultContextX.ServiceId, // p
		},
		PreimageLookup: types.PreimagesMapEntry{}, // p
		LookupDict: types.LookupMetaMapEntry{ // l
			types.LookupMetaMapkey{
				Hash:   types.OpaqueHash(c),
				Length: types.U32(l),
			}: types.TimeSlotSet{},
		},
		StorageDict: types.Storage{}, // s
	}

	derive := service_account.GetServiceAccountDerivatives(a)
	at := derive.Minbalance
	a.ServiceInfo.Items = derive.Items
	a.ServiceInfo.Bytes = derive.Bytes
	a.ServiceInfo.Balance = at
	// s_b = (x_s)_b - at
	newBalance := s.ServiceInfo.Balance - at
	// otherwise if s_b < (x_s)_t, transfer a_t tokens to new service, so need to check balance(b) > minBalance()
	minBalance := service_account.CalcThresholdBalance(s.ServiceInfo.Items, s.ServiceInfo.Bytes, s.ServiceInfo.DepositOffset)
	if s.ServiceInfo.Balance < minBalance {
		input.Registers[7] = CASH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	importServiceID := input.Addition.ResultContextX.ImportServiceId

	s.ServiceInfo.Balance = newBalance
	// reg[7] = x_i
	input.Registers[7] = uint64(importServiceID)
	// i* = check(i)
	iStar := check((1<<8)+(importServiceID-(1<<8)+42)%(1<<32-1<<9), input.Addition.ResultContextX.PartialState.ServiceAccounts)
	input.Addition.ResultContextX.ImportServiceId = iStar
	// mathbb{d} : x_i -> a
	input.Addition.ResultContextX.PartialState.ServiceAccounts[importServiceID] = a
	// mathbb{d} : x_s -> s
	input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = s

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// upgrade = 19
func upgrade(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, g, m := input.Registers[7], input.Registers[8], input.Registers[9]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	c := input.Memory.Read(o, offset)

	input.Registers[7] = OK

	serviceID := input.Addition.ResultContextX.ServiceId
	// x_bold{s} = (x_u)_d[x_s]
	if serviceAccount, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		serviceAccount.ServiceInfo.CodeHash = types.OpaqueHash(c)
		serviceAccount.ServiceInfo.MinItemGas = types.Gas(g)
		serviceAccount.ServiceInfo.MinMemoGas = types.Gas(m)
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = serviceAccount
	} else {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Printf("host-call function \"upgrade\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// transfer = 20
func transfer(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	transferGas := Gas(input.Registers[9])
	if newGas < transferGas {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas -= transferGas

	d, a, l, o := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if !isReadable(o, uint64(types.TransferMemoSize), input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// m
	rawData := input.Memory.Read(o, types.TransferMemoSize)

	if accountD, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceId(d)]; !accountExists {
		// not exist
		input.Registers[7] = WHO

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	} else if l < uint64(accountD.ServiceInfo.MinMemoGas) {
		input.Registers[7] = LOW

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	serviceID := input.Addition.ResultContextX.ServiceId
	if accountS, accountSExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountSExists {
		b := accountS.ServiceInfo.Balance - types.U64(a) // b = (x_s)_b - a
		minBalance := service_account.CalcThresholdBalance(accountS.ServiceInfo.Items, accountS.ServiceInfo.Bytes, accountS.ServiceInfo.DepositOffset)
		if a < uint64(minBalance) {
			input.Registers[7] = CASH

			return OmegaOutput{
				ExitReason:   PVMExitTuple(CONTINUE, nil),
				NewGas:       newGas,
				NewRegisters: input.Registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		}

		t := types.DeferredTransfer{
			SenderID:   serviceID,
			ReceiverID: types.ServiceId(d),
			Balance:    types.U64(a),
			Memo:       [128]byte(rawData),
			GasLimit:   types.Gas(l),
		}

		accountS.ServiceInfo.Balance = b
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = accountS
		input.Addition.ResultContextX.DeferredTransfers = append(input.Addition.ResultContextX.DeferredTransfers, t)
	} else {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Printf("host-call function \"transfer\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// eject = 21
func eject(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	d, o := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId

	accountD, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceId(d)]
	if !(types.ServiceId(d) != serviceID && accountExists) {
		// bold{d} = panic => CONTINUE, WHO
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// else : d = account
	serviceIDSerialized := utils.SerializeFixedLength(types.U32(serviceID), types.U32(32))
	// not sure need to add d_b first or not
	if !bytes.Equal(accountD.ServiceInfo.CodeHash[:], serviceIDSerialized) {
		// d_c not equal E_32(x_s)
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	l := max(81, accountD.ServiceInfo.Bytes) - 81 // a_o

	lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(l)} // x_bold{s}_l
	lookupData, lookupDataExists := accountD.LookupDict[lookupKey]

	if accountD.ServiceInfo.Items != 2 || !lookupDataExists {
		input.Registers[7] = HUH

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	timeslot := input.Addition.Timeslot
	lookupDataLength := len(lookupData)

	if lookupDataLength == 2 {
		if lookupData[1] < timeslot-types.TimeSlot(types.UnreferencedPreimageTimeslots) {
			if accountS, accountSExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountSExists {

				accountS.ServiceInfo.Balance += accountD.ServiceInfo.Balance // s'_b
				input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = accountS

				delete(input.Addition.ResultContextX.PartialState.ServiceAccounts, types.ServiceId(d))
				input.Registers[7] = OK

				return OmegaOutput{
					ExitReason:   PVMExitTuple(CONTINUE, nil),
					NewGas:       newGas,
					NewRegisters: input.Registers,
					NewMemory:    input.Memory,
					Addition:     input.Addition,
				}
			}
			// according GP, no need to check the service exists => it should in ServiceAccountState
			log.Printf("host-call function \"eject\" serviceID : %d not in ServiceAccount state", serviceID)
		}
	}

	input.Registers[7] = HUH

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// query = 22
func query(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	// x_bold{s} = (x_u)_d[x_s]
	account, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !accountExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Printf("host-call function \"query\" serviceID : %d not in ServiceAccount state", serviceID)
	}
	lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
	lookupData, lookupDataExists := account.LookupDict[lookupKey]
	if lookupDataExists {
		// a = lookupData[h,z]
		switch len(lookupData) {
		case 0:
			input.Registers[7], input.Registers[8] = 0, 0
		case 1:
			input.Registers[7] = 1 + uint64(1<<32)*uint64(lookupData[0])
			input.Registers[8] = 0
		case 2:
			input.Registers[7] = 2 + uint64(1<<32)*uint64(lookupData[0])
			input.Registers[8] = uint64(lookupData[1])
		case 3:
			input.Registers[7] = 3 + uint64(1<<32)*uint64(lookupData[0])
			input.Registers[8] = uint64(lookupData[1]) + uint64(1<<32)*uint64(lookupData[2])
		}
	} else {
		// a = panic
		input.Registers[7] = NONE
		input.Registers[8] = 0

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// solicit = 23
func solicit(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		lookupData, lookupDataExists := a.LookupDict[lookupKey]
		itemFootprintItems, itemFootprintOctets := service_account.CalcLookupItemfootprint(lookupKey)

		newFootprintItems := a.ServiceInfo.Items
		newFootprintOctets := a.ServiceInfo.Bytes

		if !lookupDataExists {
			// a_l[(h,z)] = []
			a.LookupDict[lookupKey] = make(types.TimeSlotSet, 0)
			itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
			newFootprintItems += itemFootprintItems
			newFootprintOctets += itemFootprintOctets

		} else if lookupDataExists && len(lookupData) == 2 {
			// a_l[(h,z)] = (x_s)_l[(h,z)] 艹 t   艹 = concat
			// first take off the lookup item footprints
			newFootprintItems -= itemFootprintItems
			newFootprintOctets -= itemFootprintOctets
			lookupData = append(lookupData, timeslot)
			// re-compute the item footprints
			itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
			a.LookupDict[lookupKey] = lookupData
			newFootprintItems += itemFootprintItems
			newFootprintOctets += itemFootprintOctets
		} else {
			// a = panic
			input.Registers[7] = HUH

			return OmegaOutput{
				ExitReason:   PVMExitTuple(CONTINUE, nil),
				NewGas:       newGas,
				NewRegisters: input.Registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		}

		newMinBalance := service_account.CalcThresholdBalance(newFootprintItems, newFootprintOctets, a.ServiceInfo.DepositOffset)

		// a_b < a_t
		if a.ServiceInfo.Balance < newMinBalance {
			input.Registers[7] = FULL
			return OmegaOutput{
				ExitReason:   PVMExitTuple(CONTINUE, nil),
				NewGas:       newGas,
				NewRegisters: input.Registers,
				NewMemory:    input.Memory,
				Addition:     input.Addition,
			}
		}
		//
		input.Registers[7] = OK
		// LookupDict is updated, service items and service Bytes should be updated
		a.ServiceInfo.Items = newFootprintItems
		a.ServiceInfo.Bytes = newFootprintOctets
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
	} else {
		log.Printf("host-call function \"solicit\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// forget = 24
func forget(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := input.Memory.Read(o, offset)
	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	// x_bold{s} = (x_u)_d[x_s] check service exists
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		if lookupData, lookupDataExists := a.LookupDict[lookupKey]; lookupDataExists {
			lookupDataLength := len(lookupData)
			itemFootprintItems, itemFootprintOctets := service_account.CalcLookupItemfootprint(lookupKey)

			newFootprintItems := a.ServiceInfo.Items
			newFootprintOctets := a.ServiceInfo.Bytes
			if lookupDataLength == 0 || (lookupDataLength == 2 && lookupDataLength > 1 && int(lookupData[1]) < int(timeslot)-int(types.UnreferencedPreimageTimeslots)) {
				// delete (h,z) from a_l
				expectedRemoveLookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)}
				delete(a.LookupDict, expectedRemoveLookupKey) // if key not exist, delete do nothing
				// delete (h) from a_p
				delete(a.PreimageLookup, types.OpaqueHash(h))
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
			} else if lookupDataLength == 1 {
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
				// a_l[h,z] = [x,t]
				lookupData = append(lookupData, timeslot)
				a.LookupDict[lookupKey] = lookupData
				itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
				newFootprintItems += itemFootprintItems
				newFootprintOctets += itemFootprintOctets

			} else if lookupDataLength == 3 && lookupDataLength > 1 && int(lookupData[1]) < int(timeslot)-int(types.UnreferencedPreimageTimeslots) {
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
				// a_l[h,z] = [w,t]
				lookupData[0] = lookupData[2]
				lookupData[1] = timeslot
				lookupData = lookupData[:2]
				a.LookupDict[lookupKey] = lookupData
				itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
				newFootprintItems += itemFootprintItems
				newFootprintOctets += itemFootprintOctets
			} else { // otherwise, panic
				input.Registers[7] = HUH
				return OmegaOutput{
					ExitReason:   PVMExitTuple(CONTINUE, nil),
					NewGas:       newGas,
					NewRegisters: input.Registers,
					NewMemory:    input.Memory,
					Addition:     input.Addition,
				}
			}
			// x'_s = a
			a.ServiceInfo.Items = newFootprintItems
			a.ServiceInfo.Bytes = newFootprintOctets
			input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a

			input.Registers[7] = OK
		} else { // otherwise : lookupData (x_s)_l[h,z] not exist
			input.Registers[7] = HUH
		}
	} else {
		log.Printf("host-call function \"forget\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// yield = 25
func yield(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o := input.Registers[7]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := input.Memory.Read(o, offset)
	opaqueHash := types.OpaqueHash(h)
	input.Addition.ResultContextX.Exception = &opaqueHash
	// copy(input.Addition.ResultContextX.Exception[:], h)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// provide = 26
func provide(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       0,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	o, z := input.Registers[8], input.Registers[9]
	// i = panic
	offset := uint64(z)
	if !isReadable(o, offset, input.Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// i = mu_o...+z
	i := input.Memory.Read(o, z)

	// s* = s or s = omega_7
	var sStar types.ServiceId
	if input.Registers[7] == 0xffffffffffffffff {
		sStar = *input.Addition.ServiceId
	} else {
		sStar = types.ServiceId(input.Registers[7])
	}

	// a = d[s*] or nil,  d = (x_u)_d
	account, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceId(sStar)]
	if !accountExists {
		// otherwise if a = nil
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	lookupKey := types.LookupMetaMapkey{
		Hash:   hash.Blake2bHash(i),
		Length: types.U32(z),
	}

	// otherwise if a_l[H(i), z] not in []
	if lookupData, lookupDataExists := account.LookupDict[lookupKey]; lookupDataExists && len(lookupData) != 0 {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	serviceBlob := types.ServiceBlob{
		ServiceID: sStar,
		Blob:      i,
	}

	encoder := types.NewEncoder()
	serialized, _ := encoder.Encode(sStar)
	encoded, _ := encoder.Encode(i)
	serialized = append(serialized, encoded...)
	hashKey := hash.Blake2bHash(serialized)
	// golang can not have slice in map key, so use hash instead
	if _, hashExists := input.Addition.ResultContextX.ServiceBlobs[hashKey]; hashExists {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise OK
	input.Addition.ResultContextX.ServiceBlobs[hashKey] = serviceBlob
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// log = 100 , [JIP-1](https://hackmd.io/@polkadot/jip1)
func logHostCall(input OmegaInput) (output OmegaOutput) {
	level := input.Registers[7]
	message := input.Memory.Read(input.Registers[10], input.Registers[11])
	levelStr := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}

	if level > 4 {
		logger.Errorf("logHostCall level not supported")
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       input.Gas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	timeStamp := time.RFC3339
	var logMsg string
	if input.Registers[8] == 0 && input.Registers[9] == 0 {
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v] [message:%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreId), derefernceOrNil(input.Addition.ServiceId), string(message))
	} else {
		target := input.Memory.Read(input.Registers[8], input.Registers[9])
		logMsg = fmt.Sprintf("%s [%s][core:%v][service:%v] [target:0x%x] [message:%s]\n", timeStamp, levelStr[level],
			derefernceOrNil(input.Addition.CoreId), derefernceOrNil(input.Addition.ServiceId), target, string(message))
	}
	logger.Debugf("%v", logMsg)
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       input.Gas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
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

// zero is removed in GP 0.6.7
/*
func zero(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10)
	if input.Gas < gasFee {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(OUT_OF_GAS, nil),
			NewGas:       input.Gas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	newGas := input.Gas - gasFee

	n, p, c := input.Registers[7], input.Registers[8], input.Registers[9]

	if p < 16 || (p+c) >= (1<<32)/ZP {
		input.Registers[7] = HUH
		return OmegaOutput{
			// exitReason is ncessary to keep PVM running, according previous setting, HUH, WHO is also CONTINUE
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		// u = panic
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// u = m[n]u
	for i := uint32(p); i < uint32(c); i++ {
		input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
			Value:  make([]byte, ZP),
			Access: MemoryReadWrite,
		}
	}

	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}
*/

// B.14
func check(serviceID types.ServiceId, serviceAccountState types.ServiceAccountState) types.ServiceId {
	for {
		if _, accountExists := serviceAccountState[serviceID]; !accountExists {
			return serviceID
		}

		serviceID = (serviceID-(1<<8)+1)%(1<<32-1<<9) + (1 << 8)
	}
}

// 0.7.0 later, fuzzer needs to recover state, storage cannot be recover,
// Thus, needs to check storage from KeyVal
// return storage val and add storage state into ResultContextX
func (r *ResultContext) getStorageFromKeyVal(serviceID types.ServiceId, storageKey types.ByteSequence) *types.ByteSequence {
	requestedStorageStateKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageKey, nil)
	for _, v := range r.StorageKeyVal {
		if v.Key == requestedStorageStateKey.Key {
			return &v.Value
		}
	}

	return nil
}

func (r *ResultContext) removeStorageFromKeyVal(serviceID types.ServiceId, storageKey types.ByteSequence) {
	requestedStorageStateKey := merklization.WrapEncodeDelta2KeyVal(serviceID, storageKey, nil)

	for k, v := range r.StorageKeyVal {
		if v.Key == requestedStorageStateKey.Key {
			if k < len(r.StorageKeyVal)-1 { // not the last index
				copy(r.StorageKeyVal[k:], r.StorageKeyVal[k+1:])
				return
			}
		}
	}
}

func derefernceOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

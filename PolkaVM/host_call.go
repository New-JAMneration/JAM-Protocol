package PolkaVM

import (
	"bytes"
	"log"

	service "github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// OperationType Enum
type OperationType int

const (
	GasOp              OperationType = iota // gas = 0
	LookupOp                                // lookup = 1
	ReadOp                                  // read = 2
	WriteOp                                 // write = 3
	InfoOp                                  // info = 4
	BlessOp                                 // bless = 5
	AssignOp                                // assign = 6
	DesignateOp                             // designate = 7
	CheckpointOp                            // checkpoint = 8
	NewOp                                   // new = 9
	UpgradeOp                               // upgrade = 10
	TransferOp                              // transfer = 11
	EjectOp                                 // eject = 12
	QueryOp                                 // query = 13
	SolicitOp                               // solicit = 14
	ForgetOp                                // forget = 15
	YieldOp                                 // yield = 16
	HistoricalLookupOp                      // historical_lookup = 17
	FetchOp                                 // fetch = 18
	ExportOp                                // export = 19
	MachineOp                               // machine = 20
	PeekOp                                  // peek = 21
	PokeOp                                  // poke = 22
	ZeroOp                                  // zero = 23
	VoidOp                                  // void = 24
	InvokeOp                                // invoke = 25
	ExpungeOp                               // expunge = 26
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
	ServiceId           types.ServiceId
	ServiceAccountState types.ServiceAccountState
}

type AccumulateArgs struct {
	ResultContextX ResultContext
	ResultContextY ResultContext
	Timeslot       types.TimeSlot
}

type RefineArgs struct {
	RefineMap          any
	ImportSegment      any
	ExportSegment      any
	ExportSegmentIndex any
}

type HostCallArgs struct {
	GeneralArgs
	AccumulateArgs
	RefineArgs
}

type Psi_H_ReturnType struct {
	ExitReason error        // exit reason
	Counter    uint32       // new instruction counter
	Gas        Gas          // gas remain
	Reg        Registers    // new registers
	Ram        Memory       // new memory
	Addition   HostCallArgs // addition host-call context
}

// (A.34) Ψ_H
func Psi_H(
	program StandardProgram,
	counter ProgramCounter, // program counter
	gas Gas, // gas counter
	reg Registers, // registers
	ram Memory, // memory
	omegas Omegas, // jump table
	addition HostCallArgs, // host-call context
) (
	psi_result Psi_H_ReturnType,
) {
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(program.ProgramBlob.InstructionData, counter, gas, reg, ram)

	reason := exitreason_prime.(*PVMExitReason)
	if reason.Reason == HALT || reason.Reason == PANIC || reason.Reason == OUT_OF_GAS || reason.Reason == PAGE_FAULT {
		psi_result.ExitReason = PVMExitTuple(reason.Reason, nil)
		psi_result.Counter = uint32(counter_prime)
		psi_result.Gas = gas_prime
		psi_result.Reg = reg_prime
		psi_result.Ram = memory_prime
		psi_result.Addition = addition
	} else if reason.Reason == HOST_CALL {
		var input OmegaInput
		input.Operation = OperationType(*reason.HostCall)
		input.Gas = gas_prime
		input.Registers = reg_prime
		input.Memory = ram
		input.Addition = addition
		omega := omegas[input.Operation]
		if omega == nil {
			omega = omegas[27]
		}
		omega_result := omega(input)
		omega_reason := omega_result.ExitReason.(*PVMExitReason)
		if omega_reason.Reason == PAGE_FAULT {
			psi_result.Counter = uint32(counter_prime)
			psi_result.Gas = gas_prime
			psi_result.Reg = reg_prime
			psi_result.Ram = memory_prime
			psi_result.ExitReason = PVMExitTuple(PAGE_FAULT, *omega_reason.FaultAddr)
			psi_result.Addition = addition
		} else if omega_reason.Reason == CONTINUE {
			return Psi_H(program, counter_prime, omega_result.NewGas, omega_result.NewRegisters, omega_result.NewMemory, omegas, omega_result.Addition)
		} else if omega_reason.Reason == PANIC || omega_reason.Reason == OUT_OF_GAS || omega_reason.Reason == HALT {
			psi_result.ExitReason = omega_result.ExitReason
			psi_result.Counter = uint32(counter_prime)
			psi_result.Gas = omega_result.NewGas
			psi_result.Reg = omega_result.NewRegisters
			psi_result.Ram = omega_result.NewMemory
			psi_result.Addition = omega_result.Addition
		}
	}
	return
}

var HostCallFunctions = [27]Omega{
	0:  gas,
	1:  lookup,
	2:  read,
	3:  write,
	4:  info,
	5:  bless,
	6:  assign,
	7:  designate,
	8:  checkpoint,
	9:  new,
	10: upgrade,
	11: transfer,
	12: eject,
	13: query,
	14: solicit,
	15: forget,
	16: yield,
}

// Gas Function（ΩG）
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
	register := input.Registers
	register[7] = uint64(newGas)
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: register,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// ΩL(ϱ, ω, μ, s, s, d)
func lookup(input OmegaInput) (output OmegaOutput) {
	serviceID := input.Addition.ServiceId
	serviceAccount := input.Addition.ServiceAccount
	delta := input.Addition.ServiceAccountState

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
	f := min(input.Registers[10], uint64(len(concated_bytes)))
	l := min(input.Registers[11], uint64(len(concated_bytes))-f)

	if !isWriteable(o, l, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	} else {
		new_registers := input.Registers
		new_registers[7] = uint64(len(concated_bytes))
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

// ΩR(ϱ, ω, μ, s, s, d)
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(斜): ServiceId
d: ServiceAccountState (map[ServiceId]ServiceAccount)
*/
func read(input OmegaInput) (output OmegaOutput) {
	serviceID := input.Addition.ServiceId
	serviceAccount := input.Addition.ServiceAccount
	delta := input.Addition.ServiceAccountState

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

	var s_star uint64
	var a types.ServiceAccount
	if input.Registers[7] == 0xffffffffffffffff {
		s_star = uint64(serviceID)
		a = serviceAccount
	} else if value, exists := delta[types.ServiceId(s_star)]; exists {
		s_star = input.Registers[7]
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
	ko, kz, o := input.Registers[8], input.Registers[9], input.Registers[10]
	if !isReadable(ko, kz, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	var concated_bytes []byte
	concated_bytes = append(concated_bytes, utilities.SerializeFixedLength(types.U64(s_star), 4)...)
	for address := uint32(ko); address < uint32(ko+kz); address++ {
		page := address / ZP
		index := address % ZP
		concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
	}
	k := hash.Blake2bHash(concated_bytes)
	v, exists := a.StorageDict[k]
	f := min(input.Registers[11], uint64(len(concated_bytes)))
	l := min(input.Registers[12], uint64(len(concated_bytes))-f)
	if !exists {
		new_registers := input.Registers
		new_registers[7] = NONE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	} else if !isWriteable(o, l, input.Memory) {
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
}

// ΩW (ϱ, ω, μ, s, s)
func write(input OmegaInput) (output OmegaOutput) {
	serviceID := input.Addition.ServiceId
	serviceAccount := input.Addition.ServiceAccount

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

	ko, kz, vo, vz := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]
	if !isReadable(ko, kz, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	var concated_bytes []byte
	concated_bytes = append(concated_bytes, utilities.SerializeFixedLength(types.U64(serviceID), 4)...)
	for address := uint32(ko); address < uint32(ko+kz); address++ {
		page := address / ZP
		index := address % ZP
		concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
	}
	k := hash.Blake2bHash(concated_bytes)
	var a types.ServiceAccount
	if vz == 0 {
		a = serviceAccount
		delete(a.StorageDict, k)
	} else if isReadable(vo, vz, input.Memory) {
		var concated_bytes []byte
		concated_bytes = append(concated_bytes, utilities.SerializeFixedLength(types.U64(serviceID), 4)...)
		for address := uint32(vo); address < uint32(vo+vz); address++ {
			page := address / ZP
			index := address % ZP
			concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
		}
		a = serviceAccount
		a.StorageDict[k] = concated_bytes
	} else {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if a.ServiceInfo.Balance < service.GetSerivecAccountDerivatives(types.ServiceId(serviceID)).Minbalance {
		new_registers := input.Registers
		new_registers[7] = FULL
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	value, exists := serviceAccount.StorageDict[k]
	var l uint64
	if exists {
		l = uint64(len(value))
	} else {
		l = NONE
	}
	new_registers := input.Registers
	new_registers[7] = l
	// TODO update s to a
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: new_registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// ΩR(ϱ, ω, μ, s, d)
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(斜): ServiceId
d: ServiceAccountState (map[ServiceId]ServiceAccount)
*/
func info(input OmegaInput) (output OmegaOutput) {
	serviceID := input.Addition.ServiceId
	delta := input.Addition.ServiceAccountState

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

	var t types.ServiceAccount
	var empty bool
	empty = true
	if input.Registers[7] == 0xffffffffffffffff {
		value, exist := delta[types.ServiceId(serviceID)]
		if exist {
			t = value
			empty = false
		}
	} else {
		value, exist := delta[types.ServiceId(input.Registers[7])]
		if exist {
			t = value
			empty = false
		}
	}
	if empty {
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
	derivatives := service.GetSerivecAccountDerivatives(types.ServiceId(serviceID))

	var serialized_bytes types.ByteSequence
	encoder := types.NewEncoder()
	// t_c
	encoded, _ := encoder.Encode(&t.ServiceInfo.CodeHash)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_b
	encoded, _ = encoder.EncodeUint(uint64(t.ServiceInfo.Balance))
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_t
	encoded, _ = encoder.EncodeUint(uint64(derivatives.Minbalance))
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_g
	encoded, _ = encoder.EncodeUint(uint64(t.ServiceInfo.MinItemGas))
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_m
	encoded, _ = encoder.EncodeUint(uint64(t.ServiceInfo.MinMemoGas))
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_o
	encoded, _ = encoder.EncodeUint(uint64(derivatives.Bytes))
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_i
	encoded, _ = encoder.EncodeUint(uint64(derivatives.Items))
	serialized_bytes = append(serialized_bytes, encoded...)

	o := input.Registers[8]

	if !isWriteable(o, uint64(len(serialized_bytes)), input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	} else {
		new_memory := input.Memory
		for i := 0; i < len(serialized_bytes); i++ {
			address := uint32(int(o) + i)
			page := address / ZP
			index := address % ZP
			new_memory.Pages[page].Value[index] = serialized_bytes[i]
		}

		new_registers := input.Registers
		new_registers[7] = OK
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: new_registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
}

// bless = 5
func bless(input OmegaInput) (output OmegaOutput) {
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

	m, a, v, o, n := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10], input.Registers[11]

	offset := uint64(12 * n)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
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
	rawData := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	// read data from memory, might cross many pages
	for dataLength := uint64(0); dataLength < offset; {
		rawLength := ZP - pageIndex // data length read from current page
		copy(rawData[dataLength:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		pageNumber++
		pageIndex = 0
		dataLength += rawLength
	}

	// s -> g this will update into (x_u)_x => partialState.Chi_g, decode rawData
	alwaysAccum := types.AlwaysAccumulateMap{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, alwaysAccum)
	if err != nil {
		log.Fatalf("host-call function \"bless\" decode alwaysAccum error : %v", err)
	}

	input.Registers[7] = OK

	input.Addition.ResultContextX.PartialState.Privileges = types.Privileges{
		Bless:       types.ServiceId(m),
		Assign:      types.ServiceId(a),
		Designate:   types.ServiceId(v),
		AlwaysAccum: alwaysAccum,
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// assign = 6
func assign(input OmegaInput) (output OmegaOutput) {
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

	o := input.Registers[8]

	offset := uint64(32 * types.AuthQueueSize)
	if !isReadable(o, offset, input.Memory) { // not readable, panic
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// w7 >= C
	if input.Registers[7] >= uint64(types.CoresCount) {
		input.Registers[7] = CORE
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	rawData := types.ByteSequence(make([]byte, offset)) // bold{c}
	pageNumber := o / ZP
	pageIndex := o % ZP

	// read data from memory, might cross many pages
	for dataLength := uint64(0); dataLength < offset; {
		rawLength := ZP - pageIndex // data length read from current page
		copy(rawData[dataLength:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		pageNumber++
		pageIndex = 0
		dataLength += rawLength
	}
	// decode rawData
	authQueue := types.AuthQueue{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, authQueue)
	if err != nil {
		log.Fatalf("host-call function \"assign\" decode error : %v", err)
	}

	input.Addition.ResultContextX.PartialState.Authorizers[input.Registers[7]] = authQueue
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// designate = 7
func designate(input OmegaInput) (output OmegaOutput) {
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

	o := input.Registers[7]

	offset := uint64(336 * types.ValidatorsCount)
	if !isReadable(o, offset, input.Memory) { // not readable, panic
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	rawData := types.ByteSequence(make([]byte, offset)) // bold{v}
	// 336 * types.ValidatorsCount might cross many pages
	pageNumber := o / ZP
	pageIndex := o % ZP

	for dataLength := uint64(0); dataLength < offset; {
		rawLength := ZP - pageIndex // data length read from current page
		copy(rawData[dataLength:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		pageNumber++
		pageIndex = 0
		dataLength += rawLength
	}

	validatorsData := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, validatorsData)
	if err != nil {
		log.Fatalf("host-call function \"designate\" decode validatorsData error : %v", err)
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

// checkpoint = 8
func checkpoint(input OmegaInput) (output OmegaOutput) {
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

// new = 9
func new(input OmegaInput) (output OmegaOutput) {
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

	o, l, g, m := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	offset := uint64(32)
	if !(isReadable(o, offset, input.Memory) && l < (1<<32)) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	c := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP
	// reda data from memory, might only cross one page
	if ZP-pageIndex < uint64(offset) {
		copy(c, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(c[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(c, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	s, sExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !sExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Fatalf("host-call function \"new\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	if s.ServiceInfo.Balance < service.GetSerivecAccountDerivatives(serviceID).Minbalance {
		input.Registers[7] = CASH

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	var cDecoded types.U32
	decoder := types.NewDecoder()
	err := decoder.Decode(c, cDecoded)
	if err != nil {
		log.Fatalf("host-call function \"new\" decode error %v: ", err)
	}

	accountDer := service.GetSerivecAccountDerivatives(types.ServiceId(cDecoded))
	at := accountDer.Minbalance
	// s_b = (x_s)_b - at
	s.ServiceInfo.Balance -= at
	// new an account
	serviceinfo := types.ServiceInfo{
		CodeHash:   types.OpaqueHash(c), // c
		Balance:    at,                  // b
		MinItemGas: types.Gas(g),        // g
		MinMemoGas: types.Gas(m),        // m
	}
	lookupKey := types.LookupMetaMapkey{
		Hash:   types.OpaqueHash(c),
		Length: types.U32(l),
	}
	lookupMetaMapEntry := types.LookupMetaMapEntry{
		lookupKey: types.TimeSlotSet{},
	}

	a := types.ServiceAccount{
		ServiceInfo:    serviceinfo,
		PreimageLookup: types.PreimagesMapEntry{}, // p
		LookupDict:     lookupMetaMapEntry,        // l
		StorageDict:    types.Storage{},           // s
	}

	importServiceID := input.Addition.ResultContextX.ImportServiceId
	// reg[7] = x_i
	input.Registers[7] = uint64(importServiceID)
	// x_i = check(i)
	i := (1 << 8) + (importServiceID-(1<<8)+42)%(1<<32-1<<9)
	input.Addition.ResultContextX.ImportServiceId = check(i, input.Addition.ResultContextX.PartialState.ServiceAccounts)
	// x_i -> a
	input.Addition.ResultContextX.PartialState.ServiceAccounts[importServiceID] = a
	// x_s -> s
	input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = s

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// upgrade = 10
func upgrade(input OmegaInput) (output OmegaOutput) {
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

	o, g, m := input.Registers[7], input.Registers[8], input.Registers[9]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	c := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < uint64(offset) { // cross page
		copy(c, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(c[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(c, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

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
		log.Fatalf("host-call function \"upgrade\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// transfer = 11
func transfer(input OmegaInput) (output OmegaOutput) {
	gasFee := Gas(10) + Gas(input.Registers[9])
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

	d, a, l, o := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if !isReadable(o, uint64(types.TransferMemoSize), input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}
	// m
	rawData := types.ByteSequence(make([]byte, types.TransferMemoSize))
	pageNumber := o / ZP
	pageIndex := o % ZP

	for dataLength := uint64(0); dataLength < types.TransferMemoSize; {
		rawLength := ZP - pageIndex // data length read from current page
		copy(rawData[dataLength:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		pageNumber++
		pageIndex = 0
		dataLength += rawLength
	}

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
		if b < service.GetSerivecAccountDerivatives(serviceID).Minbalance {
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
		log.Fatalf("host-call function \"transfer\" serviceID : %d not in ServiceAccount state", serviceID)
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

// eject = 12
func eject(input OmegaInput) (output OmegaOutput) {
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

	d, o := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < offset { // cross one page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

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
	seviceIDSerialized := utils.SerializeFixedLength(types.U32(serviceID), types.U32(32))
	// not sure need to add d_b first or not
	if !bytes.Equal(accountD.ServiceInfo.CodeHash[:], seviceIDSerialized) {
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
			} else {
				// according GP, no need to check the service exists => it should in ServiceAccountState
				log.Fatalf("host-call function \"eject\" serviceID : %d not in ServiceAccount state", serviceID)
			}
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

// query = 13
func query(input OmegaInput) (output OmegaOutput) {
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

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := types.ByteSequence(make([]byte, 32))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < offset { // cross one page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	// x_bold{s} = (x_u)_d[x_s]
	account, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !accountExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Fatalf("host-call function \"query\" serviceID : %d not in ServiceAccount state", serviceID)
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

// solicit = 14
func solicit(input OmegaInput) (output OmegaOutput) {
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

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < offset { // cross one page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		lookupData, lookupDataExists := a.LookupDict[lookupKey]
		// a_l[(h,z)] = [] => no changes, do not need to implement
		if lookupDataExists && len(lookupData) == 2 {
			// a_l[(h,z)] = (x_s)_l[(h,z)] 艹 t   艹 = concat
			lookupData = append(lookupData, timeslot)
			a.LookupDict[lookupKey] = lookupData
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
		// a_b < a_t
		if a.ServiceInfo.Balance < service.GetSerivecAccountDerivatives(serviceID).Minbalance {
			input.Registers[7] = FULL
		}

		// else
		input.Registers[7] = OK
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
	} else {
		log.Fatalf("host-call function \"solicit\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// forget = 15
func forget(input OmegaInput) (output OmegaOutput) {
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

	o, z := input.Registers[7], input.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) { // not readable, return
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < offset { // cross one page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	// x_bold{s} = (x_u)_d[x_s] check service exists
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		if lookupData, lookupDataExists := a.LookupDict[lookupKey]; lookupDataExists {
			lookupDataLength := len(lookupData)

			if lookupDataLength == 0 || lookupDataLength == 2 {
				if lookupData[1] < timeslot-types.TimeSlot(types.UnreferencedPreimageTimeslots) {
					// delete (h,z) from a_l
					expectedRemoveLookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)}
					delete(a.LookupDict, expectedRemoveLookupKey) // if key not exist, delete do nothing
					// delete (h) from a_p
					delete(a.PreimageLookup, types.OpaqueHash(h))
				}
			} else if lookupDataExists && lookupDataLength == 1 {
				// a_l[h,z] = [x,t]
				lookupData = append(lookupData, timeslot)
				a.LookupDict[lookupKey] = lookupData
			} else if lookupDataExists && lookupDataLength == 3 {
				if lookupData[1] < timeslot-types.TimeSlot(types.UnreferencedPreimageTimeslots) {
					// a_l[h,z] = [w,t]
					lookupData[0] = lookupData[2]
					lookupData[1] = timeslot
					lookupData = lookupData[:2]
					a.LookupDict[lookupKey] = lookupData
				}
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
			input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a

			input.Registers[7] = OK
		} else { // otherwise : lookupData (x_s)_l[h,z] not exist
			input.Registers[7] = HUH
		}
	} else {
		log.Fatalf("host-call function \"forget\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// yield = 16
func yield(input OmegaInput) (output OmegaOutput) {
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

	o := input.Registers[7]

	offset := uint64(32)
	if !isReadable(o, offset, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h := types.ByteSequence(make([]byte, offset))
	pageNumber := o / ZP
	pageIndex := o % ZP

	if ZP-pageIndex < offset { // cross one page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber+1)].Value[:ZP-pageIndex])
	} else {
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+offset])
	}

	input.Registers[7] = OK

	copy(input.Addition.ResultContextX.Exception[:], h)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// B.14
func check(serviceID types.ServiceId, serviceAccountState types.ServiceAccountState) types.ServiceId {
	if _, accountExists := serviceAccountState[serviceID]; !accountExists {
		return check((serviceID-(1<<8)+1)%(1<<32-1<<9)+(1<<8), serviceAccountState)
	}

	return serviceID
}

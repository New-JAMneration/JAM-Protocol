package PolkaVM

import (
	"bytes"
	"log"

	// service "github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	FetchOp                                 // fetch = 18;

	ExportOp  // export = 19
	MachineOp // machine = 20
	PeekOp    // peek = 21
	PokeOp    // poke = 22
	ZeroOp    // zero = 23
	VoidOp    // void = 24
	InvokeOp  // invoke = 25
	ExpungeOp // expunge = 26
	ProvideOp // provide = 27
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
	RefineInput                            // i, p(package), o, bold{i}, zeta
	IntegratedPVMMap                       // D ( N -> M ) : N -> (p(program_code), u, i)
	ExportSegment    []types.ExportSegment // e
	ServiceID        types.ServiceId       // s
	TimeSlot         types.TimeSlot        // t
	ExtrinsicDataMap                       // extrinsic data map
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
	gas types.Gas, // gas counter
	reg Registers, // registers
	ram Memory, // memory
	omegas Omegas, // jump table
	addition HostCallArgs, // host-call context
) (
	psi_result Psi_H_ReturnType,
) {
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(program.ProgramBlob.InstructionData, counter, Gas(gas), reg, ram)

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
			return Psi_H(program, counter_prime, types.Gas(omega_result.NewGas), omega_result.NewRegisters, omega_result.NewMemory, omegas, omega_result.Addition)
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

var HostCallFunctions = [29]Omega{
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
<<<<<<< HEAD
	17: historicalLookup,
	18: fetch,
	19: export,
	20: machine,
	21: peek,
	22: poke,
	23: zero,
	24: void,
	25: invoke,
	26: expunge,
<<<<<<< HEAD
=======

	27: provide,
>>>>>>> e8b6214 (PVM update v0.6.5)
=======
	27: provide,
>>>>>>> e3e8407da15c5f41c0b06b8d0dd7c7e578c98728
}

func onTransferHostCallException(input OmegaInput) (output OmegaOutput) {
	input.Registers[7] = WHAT
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       input.Gas - 10,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
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

	f := min(input.Registers[10], uint64(len(v)))
	l := min(input.Registers[11], uint64(len(v))-f)

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

	var sStar uint64
	var a types.ServiceAccount
	// s* = ?
	if input.Registers[7] == 0xffffffffffffffff {
		sStar = uint64(serviceID)
	} else {
		sStar = input.Registers[7]
	}
	// assign ko, kz, o first and check v = panic ?
	// since v = panic is the first condition to check
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
	// a = ?
	if sStar == uint64(serviceID) {
		a = serviceAccount
	} else if value, exists := delta[types.ServiceId(sStar)]; exists {
		a = value
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
	// first compute k
	encoder := types.NewEncoder()
	serviceStar := types.ServiceId(sStar)
	concated_bytes, _ := encoder.Encode(&serviceStar)
	// mu_ko...+kz
	for address := uint32(ko); address < uint32(ko+kz); address++ {
		page := address / ZP
		index := address % ZP
		concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
	}

	k := hash.Blake2bHash(concated_bytes)
	v, exists := a.StorageDict[k]
	f := min(input.Registers[11], uint64(len(concated_bytes)))
	l := min(input.Registers[12], uint64(len(concated_bytes))-f)
	// first check not writable, then check v = nil (not exists)
	if !isWriteable(o, l, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// v = nil
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

	encoder := types.NewEncoder()
	concated_bytes, _ := encoder.Encode(&serviceID)
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
		concated_bytes = []byte{}
		for address := uint32(vo); address < uint32(vo+vz); address++ {
			page := address / ZP
			index := address % ZP
			concated_bytes = append(concated_bytes, input.Memory.Pages[page].Value[index])
		}
		a = serviceAccount
		// need extra storage space :
		// check a_t > a_b : storage need gas, balance is not enough for storage
		if a.ServiceInfo.Balance < service_account.GetServiceAccountDerivatives(a).Minbalance {
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

	value, exists := serviceAccount.StorageDict[k]
	var l uint64
	if exists {
		l = uint64(len(value))
	} else {
		l = NONE
	}
	new_registers := input.Registers
	new_registers[7] = l

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

	derivatives := service_account.GetServiceAccountDerivatives(t)

	var serialized_bytes types.ByteSequence
	encoder := types.NewEncoder()
	// t_c
	encoded, _ := encoder.Encode(&t.ServiceInfo.CodeHash)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_b
	encoded, _ = encoder.Encode(&t.ServiceInfo.Balance)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_t
	encoded, _ = encoder.Encode(&derivatives.Minbalance)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_g
	encoded, _ = encoder.Encode(&t.ServiceInfo.MinItemGas)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_m
	encoded, _ = encoder.Encode(&t.ServiceInfo.MinMemoGas)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_o
	encoded, _ = encoder.Encode(&derivatives.Bytes)
	serialized_bytes = append(serialized_bytes, encoded...)
	// t_i
	encoded, _ = encoder.Encode(&derivatives.Items)
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
	}
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
		NewMemory:    new_memory,
		Addition:     input.Addition,
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
	// read data from memory, might cross many pages
	rawData := input.Memory.Read(o, offset)

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

	rawData := input.Memory.Read(o, offset)

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

	// 336 * types.ValidatorsCount might cross many pages
	rawData := input.Memory.Read(o, offset) // bold{v}

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

	c := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	s, sExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !sExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		log.Fatalf("host-call function \"new\" serviceID : %d not in ServiceAccount state", serviceID)
	}
	// s_b < (x_s)_t
	if s.ServiceInfo.Balance < service_account.GetServiceAccountDerivatives(s).Minbalance {
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
	err := decoder.Decode(c, &cDecoded)
	if err != nil {
		log.Fatalf("host-call function \"new\" decode error %v: ", err)
	}

	// new an account
	a := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			CodeHash:   types.OpaqueHash(c), // c
			Balance:    0,                   // b, will be updated later
			MinItemGas: types.Gas(g),        // g
			MinMemoGas: types.Gas(m),        // m
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
	at := service_account.GetServiceAccountDerivatives(a).Minbalance
	a.ServiceInfo.Balance = at

	// s_b = (x_s)_b - at
	s.ServiceInfo.Balance -= at

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
		if b < service_account.GetServiceAccountDerivatives(accountS).Minbalance {
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
			log.Fatalf("host-call function \"eject\" serviceID : %d not in ServiceAccount state", serviceID)
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

	h := input.Memory.Read(o, offset)

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

	h := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		lookupData, lookupDataExists := a.LookupDict[lookupKey]

		if !lookupDataExists {
			// a_l[(h,z)] = []
			a.LookupDict[lookupKey] = make(types.TimeSlotSet, 0)
		} else if lookupDataExists && len(lookupData) == 2 {
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
		if a.ServiceInfo.Balance < service_account.GetServiceAccountDerivatives(a).Minbalance {
			// rollback the changes made to the lookup dict
			if lookupDataExists {
				a.LookupDict[lookupKey] = lookupData[:2]
			} else {
				delete(a.LookupDict, lookupKey)
			}

			input.Registers[7] = FULL
		} else {
			input.Registers[7] = OK
			input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
		}
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

	h := input.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceId
	timeslot := input.Addition.Timeslot
	// x_bold{s} = (x_u)_d[x_s] check service exists
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		if lookupData, lookupDataExists := a.LookupDict[lookupKey]; lookupDataExists {
			lookupDataLength := len(lookupData)

			if lookupDataLength == 0 || (lookupDataLength == 2 && lookupDataLength > 1 && lookupData[1] < timeslot-types.TimeSlot(types.UnreferencedPreimageTimeslots)) {
				// delete (h,z) from a_l
				expectedRemoveLookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)}
				delete(a.LookupDict, expectedRemoveLookupKey) // if key not exist, delete do nothing
				// delete (h) from a_p
				delete(a.PreimageLookup, types.OpaqueHash(h))
			} else if lookupDataLength == 1 {
				// a_l[h,z] = [x,t]
				lookupData = append(lookupData, timeslot)
				a.LookupDict[lookupKey] = lookupData
			} else if lookupDataLength == 3 && lookupDataLength > 1 && lookupData[1] < timeslot-types.TimeSlot(types.UnreferencedPreimageTimeslots) {
				// a_l[h,z] = [w,t]
				lookupData[0] = lookupData[2]
				lookupData[1] = timeslot
				lookupData = lookupData[:2]
				a.LookupDict[lookupKey] = lookupData
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

	h := input.Memory.Read(o, offset)

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

// historical_lookup = 17
func historicalLookup(input OmegaInput) (output OmegaOutput) {
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

	// first check v panic, then assign a
	h, o := input.Registers[8], input.Registers[9]

	offset := uint64(32)
	if !isReadable(h, offset, input.Memory) { // not readable, return panic
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

	if !isWriteable(o, l, input.Memory) { // not writeable, return panic
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
	input.Memory.Write(f, offset, v)

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// fetch = 18
func fetch(input OmegaInput) (output OmegaOutput) {
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

	// pre-processing
	encoder := types.NewEncoder()
	// condition reg[10] == 3
	// w_11 < |p_w|
	condition31 := input.Registers[11] < uint64(len(input.Addition.WorkPackage.Items))
	// workItem3 = p_w[w_11]
	workItem3 := input.Addition.WorkPackage.Items[input.Registers[11]]
	// w_12 < |p_w[w_11]x|
	condition32 := input.Registers[12] < uint64(len(workItem3.Extrinsic))
	// extrinsic data exists
	extrinsicData3, condition33 := input.Addition.ExtrinsicDataMap[workItem3.Extrinsic[input.Registers[12]].Hash]

	// condition reg[10] == 4
	// workItem4 = p_w[i]
	workItem4 := input.Addition.WorkPackage.Items[input.Addition.WorkItemIndex]

	// condition4 := extrinsicSpec4.Hash == workItem4.Extrinsic[input.Registers[11]].Hash && extrinsicSpec4.Len == workItem4.Extrinsic[input.Registers[11]].Len
	extrinsicData4, condition4 := input.Addition.ExtrinsicDataMap[workItem4.Extrinsic[input.Registers[11]].Hash]

	var v []byte
	var dataLength uint64

	switch {
	case input.Registers[10] == 0:
		wp, _ := encoder.Encode(input.Addition.WorkPackage)
		// v = |E(p)
		v = wp
		dataLength = uint64(len(v))

	case input.Registers[10] == 1:
		// v = o
		v = input.Addition.AuthOutput
		dataLength = uint64(len(v))

	case input.Registers[10] == 2 && input.Registers[11] < uint64(len(input.Addition.WorkPackage.Items)):
		// v = p_w[w_11]y
		v = input.Addition.WorkPackage.Items[input.Registers[11]].Payload
		dataLength = uint64(len(v))

	case input.Registers[10] == 3 && condition31 && condition32 && condition33:
		// v = x = p_w[w_11]_x
		v = extrinsicData3
		dataLength = uint64(len(v))

	case input.Registers[10] == 4 && input.Registers[11] < uint64(len(input.Addition.WorkPackage.Items[input.Addition.WorkItemIndex].Extrinsic)) && condition4:
		// v = x = p_w[i]x
		v = extrinsicData4
		dataLength = uint64(len(v))

	case input.Registers[10] == 5 && input.Registers[11] < uint64(len(input.Addition.ImportSegments)) && input.Registers[12] < uint64(len(input.Addition.ImportSegments[input.Registers[11]])):
		v = input.Addition.ImportSegments[input.Registers[11]][input.Registers[12]][:]
		dataLength = uint64(len(v))

	case input.Registers[10] == 6 && input.Registers[11] < uint64(len(input.Addition.ImportSegments[input.Addition.WorkItemIndex])):
		v = input.Addition.ImportSegments[input.Addition.WorkItemIndex][input.Registers[11]][:]
		dataLength = uint64(len(v))

	case input.Registers[10] == 7:
		v = input.Addition.WorkPackage.Authorizer.Params
		dataLength = uint64(len(input.Addition.WorkPackage.Authorizer.Params))
	default: // default = nil
		dataLength = 0
	}

	o := input.Registers[7]
	f := min(input.Registers[8], dataLength)
	l := min(input.Registers[9], dataLength-f)

	// need to first check writable
	if !isWriteable(o, l, input.Memory) {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if v = nil
	if v == nil {
		input.Registers[7] = NONE

		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Memory.Write(f, l, v[f:])
	input.Registers[7] = dataLength

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// export = 19
func export(input OmegaInput) (output OmegaOutput) {
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

	p := input.Registers[7]
	z := min(input.Registers[8], types.SegmentSize)

	if !isReadable(p, z, input.Memory) { // not readable, return
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

// machine = 20
func machine(input OmegaInput) (output OmegaOutput) {
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

	po, pz, i := input.Registers[7], input.Registers[8], input.Registers[9]
	// pz = offset
	if !isReadable(po, pz, input.Memory) { // not readable, return
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

// peek = 21
func peek(input OmegaInput) (output OmegaOutput) {
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

	n, o, s, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	// z = offset
	if !isWriteable(o, z, input.Memory) { // not writeable, return
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

// poke = 22
func poke(input OmegaInput) (output OmegaOutput) {
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

	n, s, o, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if !isReadable(s, z, input.Memory) { // not readable, return
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

// zero = 23
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

// void = 24
func void(input OmegaInput) (output OmegaOutput) {
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

// invoke = 25
func invoke(input OmegaInput) (output OmegaOutput) {
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

	n, o := input.Registers[7], input.Registers[8]

	offset := uint64(112)
	// g = panic
	if !isWriteable(o, offset, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
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
		log.Fatalf("host-call function \"invoke\" decode gas error : %v", err)
	}
	// decode registers
	for i := uint64(1); i < offset/8; i++ {
		err = decoder.Decode(data[8*i:8*(i+1)], w[i-1])
		if err != nil {
			log.Fatalf("host-call function \"invoke\" decode register:%d error : %v", i-1, err)
		}
	}
	// psi
	c, pcPrime, gasPrime, wPrime, uPrime := SingleStepInvoke(input.Addition.IntegratedPVMMap[n].ProgramCode, input.Addition.IntegratedPVMMap[n].PC, Gas(gas), w, input.Addition.IntegratedPVMMap[n].Memory)

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

// expunge = 26
func expunge(input OmegaInput) (output OmegaOutput) {
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

// B.14
func check(serviceID types.ServiceId, serviceAccountState types.ServiceAccountState) types.ServiceId {
	for {
		if _, accountExists := serviceAccountState[serviceID]; !accountExists {
			return serviceID
		}

		serviceID = (serviceID-(1<<8)+1)%(1<<32-1<<9) + (1 << 8)
	}
}

func provide(input OmegaInput) (output OmegaOutput) {
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

	o, z := input.Registers[8], input.Registers[9]
	// i = panic
	offset := uint64(z)
	if !isReadable(o, offset, input.Memory) {
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
		sStar = input.Addition.ServiceId
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

	serviceBlob := ServiceBlob{
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

package PolkaVM

import (
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
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
	Addition  []any         // Extra parameter for each host-call function
}
type OmegaOutput struct {
	ExitReason   error     // Exit reason
	NewGas       Gas       // New Gas
	NewRegisters Registers // New Register
	NewMemory    Memory    // New Memory
	Addition     []any     // addition host-call context
}

// Ω⟨X⟩
type Omega func(OmegaInput) OmegaOutput

type Psi_H_ReturnType struct {
	ExitReason error     // exit reason
	Counter    uint32    // new instruction counter
	Gas        Gas       // gas remain
	Reg        Registers // new registers
	Ram        Memory    // new memory
	Addition   []any     // addition host-call context
}

// (A.31) Ψ_H
func Psi_H(
	counter ProgramCounter, // program counter
	gas Gas, // gas counter
	reg Registers, // registers
	ram Memory, // memory
	omega Omega, // jump table
	addition []any, // host-call context
	program StandardProgram,
) (
	psi_result Psi_H_ReturnType,
) {
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(program.ProgramBlob.InstructionData, counter, gas, reg, ram)
	fmt.Println(exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime)
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
			return Psi_H(ProgramCounter(skip(int(counter_prime), program.ProgramBlob.Bitmasks)), omega_result.NewGas, omega_result.NewRegisters, omega_result.NewMemory, omega, omega_result.Addition, program)
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

var hostCallFunctions = [26]Omega{
	0: gas,
	1: lookup,
	2: read,
	3: write,
	4: info,
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

func getServiceID(addition []any) (uint64, error) {
	if len(addition) == 0 {
		return 0, errors.New("serviceID not found in Addition")
	}

	serviceID, ok := addition[0].(uint64)
	if !ok {
		return 0, errors.New("serviceID is not of type uint64")
	}

	return serviceID, nil
}

// ΩL(ϱ, ω, μ, s, s, d)
func lookup(input OmegaInput) (output OmegaOutput) {
	serviceID, err := getServiceID(input.Addition)
	if err != nil {
		fmt.Println("Addition context error")
		return output
	}
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
	delta := store.GetInstance().GetPriorStates().GetDelta()
	serviceAccount := delta[types.ServiceId(serviceID)]
	var a types.ServiceAccount
	if input.Registers[7] == 0xffffffffffffffff || input.Registers[7] == serviceID {
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
	serviceID, err := getServiceID(input.Addition)
	if err != nil {
		fmt.Println("Addition context error")
		return output
	}
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

	delta := store.GetInstance().GetPriorStates().GetDelta()
	serviceAccount := delta[types.ServiceId(serviceID)]
	var s_star uint64
	var a types.ServiceAccount
	if input.Registers[7] == 0xffffffffffffffff {
		s_star = serviceID
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
	f := min(input.Registers[10], uint64(len(concated_bytes)))
	l := min(input.Registers[11], uint64(len(concated_bytes))-f)
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
	serviceID, err := getServiceID(input.Addition)
	if err != nil {
		fmt.Println("Addition context error")
		return output
	}
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

	delta := store.GetInstance().GetPriorStates().GetDelta()
	serviceAccount := delta[types.ServiceId(serviceID)]

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
		delta[types.ServiceId(serviceID)] = a
		store.GetInstance().GetPriorStates().SetDelta(delta)
	} else {
		return OmegaOutput{
			ExitReason:   PVMExitTuple(CONTINUE, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if a.Balance < service_account.GetSerivecAccountDerivatives(types.ServiceId(serviceID)).Minbalance {
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

// ΩR(ϱ, ω, μ, s, s)
/*
ϱ: gas
ω: registers
μ:  memory
s: ServiceAccount
s(斜): ServiceId
d: ServiceAccountState (map[ServiceId]ServiceAccount)
*/
func info(input OmegaInput) (output OmegaOutput) {
	serviceID, err := getServiceID(input.Addition)
	if err != nil {
		fmt.Println("Addition context error")
		return output
	}
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

	delta := store.GetInstance().GetPriorStates().GetDelta()

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
	derivatives := service_account.GetSerivecAccountDerivatives(types.ServiceId(serviceID))

	var serialized_bytes types.ByteSequence
	serialized_bytes = append(serialized_bytes, utilities.SerializeByteSequence(t.CodeHash[:])...)
	serialized_bytes = append(serialized_bytes, utilities.SerializeU64(t.Balance)...)
	serialized_bytes = append(serialized_bytes, utilities.SerializeU64(derivatives.Minbalance)...)
	serialized_bytes = append(serialized_bytes, utilities.SerializeU64(types.U64(t.MinItemGas))...)
	serialized_bytes = append(serialized_bytes, utilities.SerializeU64(types.U64(t.MinMemoGas))...)
	result := utilities.WrapDictionaryKeyMap(t.LookupDict)
	encoded := result.Serialize()
	serialized_bytes = append(serialized_bytes, encoded...)
	serialized_bytes = append(serialized_bytes, utilities.SerializeU64(types.U64(derivatives.Items))...)
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

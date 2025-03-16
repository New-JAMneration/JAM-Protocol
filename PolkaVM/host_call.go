package PolkaVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	// 0: gas,
	// 1: lookup,
	// 2: read,
	// 3: write,
	// 4: info,
	15: forget,
	16: bless,
}

// 15: forget
func forget(input OmegaInput) OmegaOutput {
	newGas := input.Gas - 10
	o, z := input.Registers[7], input.Registers[8]
	var h types.ByteSequence
	pageNumber := o / ZP
	pageIndex := o % ZP

	if !isReadable(o, 32, input.Memory) { // memory not readable, return panic
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h = make([]byte, 32)
	if ZP-pageIndex < 32 { // cross page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber)].Value[:32-(ZP-pageIndex)])
	} else {
		copy(h[:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+32])
	}

	var a types.ServiceAccount
	t := input.Addition[2]

	// x_bold{s} = x_s = (x_u)_d[x_s] = serviceState_d[x_s]  B.7

	/*
		type ServiceAccount struct {
			StorageDict    map[OpaqueHash]ByteSequence   // a_s
			PreimageLookup map[OpaqueHash]ByteSequence   // a_p
			LookupDict     map[DictionaryKey]TimeSlotSet // a_l
			CodeHash       OpaqueHash                    // a_c
			Balance        U64                           // a_b
			MinItemGas     Gas                           // a_g
			MinMemoGas     Gas                           // a_m
		}
	*/

	return OmegaOutput{}
}

// 16: bless
func bless(input OmegaInput) OmegaOutput {
	newGas := input.Gas - 10
	o := input.Registers[7]
	var h types.ByteSequence
	pageNumber := o / ZP
	pageIndex := o % ZP
	if !isReadable(o, 32, input.Memory) { // memory not readable
		return OmegaOutput{
			ExitReason:   PVMExitTuple(PANIC, nil),
			NewGas:       newGas,
			NewRegisters: input.Registers, // don't have to return if the value not changed ?
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	h = make([]byte, 32)
	if ZP-pageIndex < 32 { // cross page
		copy(h, input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:])
		copy(h[ZP-pageIndex:], input.Memory.Pages[uint32(pageNumber)].Value[:32-(ZP-pageIndex)])
	} else {
		copy(h[:], input.Memory.Pages[uint32(pageNumber)].Value[pageIndex:pageIndex+32])
	}

	input.Registers[7] = OK
	if resultContext, ok := input.Addition[0].(ResultContext); ok {
		resultContext.additionBytes = h
		input.Addition[0] = resultContext
	}

	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

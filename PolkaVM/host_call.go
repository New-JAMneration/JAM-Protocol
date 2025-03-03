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
	PreviousAccount   types.ServiceAccountState
}

type OmegaInput struct {
	Instruction  uint64                    // opcode
	Operation    OperationType             // operation type
	Gas          Gas                       // gas counter
	Registers    Registers                 // PVM registers
	Memory       PageMap                   // memory
	AccountState types.ServiceAccountState // State
	History      *HistoryState             // For storing history state
	Args         []any                     // Extra parameter for each host-call function
}
type OmegaOutput struct {
	ExitReason      error                     // Exit reason
	NewGas          Gas                       // New Gas
	NewRegisters    Registers                 // New Register
	NewMemory       PageMap                   // New Memory
	NewAccountState types.ServiceAccountState // New State
	Counter         uint32                    // For calculate run count
	History         *HistoryState             // For roll back
	Addition        []any                     // addition host-call context
}

// Ω⟨X⟩
type Omega func(OmegaInput) OmegaOutput

type Psi_H_ReturnType struct {
	ExitReason error     // exit reason
	Counter    uint32    // new instruction counter
	Gas        Gas       // gas remain
	Reg        Registers // new registers
	Ram        Memory    // new memory
	Addition   any       // addition host-call context
}

// (A.31) Ψ_H
func Psi_H(
	code ProgramCode, // program code
	counter ProgramCounter, // program counter
	gas Gas, // gas counter
	reg Registers, // registers
	ram Memory, // memory
	omega Omega, // jump table
	addition any, // host-call context
	program StandardProgram,
) (
	psi_result Psi_H_ReturnType,
) {
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(code, counter, gas, reg, ram)
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
		omega_result := omega(*reason.FaultAddr, gas_prime, reg_prime, ram, addition)
		omega_reason := omega_result.ExitReason.(*PVMExitReason)
		if omega_reason.Reason == PAGE_FAULT {
			psi_result.Counter = uint32(counter_prime)
			psi_result.Gas = gas_prime
			psi_result.Reg = reg_prime
			psi_result.Ram = memory_prime
			psi_result.ExitReason = PVMExitTuple(PAGE_FAULT, *omega_reason.FaultAddr)
			psi_result.Addition = addition
		} else if omega_reason.Reason == CONTINUE {
			return Psi_H(code, ProgramCounter(skip(int(counter_prime), program.ProgramBlob.Bitmasks)), omega_result.GasRemain, omega_result.Register, omega_result.Ram, omega, omega_result.Addition, program)
		} else if omega_reason.Reason == PANIC || omega_reason.Reason == OUT_OF_GAS || omega_reason.Reason == HALT {
			psi_result.ExitReason = omega_result.ExitReason
			psi_result.Counter = uint32(counter_prime)
			psi_result.Gas = omega_result.GasRemain
			psi_result.Reg = omega_result.Register
			psi_result.Ram = omega_result.Ram
			psi_result.Addition = omega_result.Addition
		}
	}
	return
}

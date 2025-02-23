package PolkaVM

import "fmt"

type Psi_H_ReturnType struct {
	ExitReason error     // exit reason
	Counter    uint64    // new instruction counter
	Gas        Gas       // gas remain
	Reg        Registers // new registers
	Ram        Memory    // new memory
	Addition   any       // addition host-call context
}

type OmegaReturnType struct {
	ExitReason error     // exit reason
	GasRemain  Gas       // gas remain
	Register   Registers // new registers
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
) (
	psi_result Psi_H_ReturnType,
) {
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(code, counter, gas, reg, ram)
	fmt.Println(exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime)
	reason := exitreason_prime.(*PVMExitReason)
	if reason.Reason == HALT || reason.Reason == PANIC || reason.Reason == OUT_OF_GAS || reason.Reason == PAGE_FAULT {
		psi_result.ExitReason = PVMExitTuple(reason.Reason, nil)
		psi_result.Counter = uint64(counter_prime)
		psi_result.Gas = gas_prime
		psi_result.Reg = reg_prime
		psi_result.Ram = memory_prime
		psi_result.Addition = addition
	} else if reason.Reason == HOST_CALL {
		omega_result := omega(*reason.FaultAddr, gas_prime, reg_prime, ram, addition)
		omega_reason := omega_result.ExitReason.(*PVMExitReason)
		if omega_reason.Reason == PAGE_FAULT {
			psi_result.Counter = uint64(counter_prime)
			psi_result.Gas = gas_prime
			psi_result.Reg = reg_prime
			psi_result.Ram = memory_prime
			psi_result.ExitReason = PVMExitTuple(PAGE_FAULT, *omega_reason.FaultAddr)
			psi_result.Addition = addition
		} else if omega_reason.Reason == CONTINUE {
			return Psi_H(code, ProgramCounter(counter_prime), omega_result.GasRemain, omega_result.Register, omega_result.Ram, omega, omega_result.Addition)
			// NEED TO USE SKIP??
		} else if omega_reason.Reason == PANIC || omega_reason.Reason == OUT_OF_GAS || omega_reason.Reason == HALT {
			psi_result.ExitReason = omega_result.ExitReason
			psi_result.Counter = uint64(counter_prime)
			psi_result.Gas = omega_result.GasRemain
			psi_result.Reg = omega_result.Register
			psi_result.Ram = omega_result.Ram
			psi_result.Addition = omega_result.Addition
		}
	}
	return
}

// (A.32) Ω⟨X⟩
type Omega func(
	uint64, // instruction
	Gas, // gas counter
	Registers, // registers
	Memory, // memory
	any, // host-call context
) OmegaReturnType

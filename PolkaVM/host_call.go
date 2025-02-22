package PolkaVM

import "fmt"

type Psi_H_ReturnType struct {
	Pagefault        bool            // exit reason is page fault (for handling multiple return type)
	ExitReason       ExitReasonTypes // exit reason
	Counter          uint64          // new instruction counter
	Gas              Gas             // gas remain
	Reg              Registers       // new registers
	Ram              Memory          // new memory
	Addition         any             // addition host-call context
	PagefaultAddress uint64          // page fault address, only use for page fault
}

type OmegaReturnType struct {
	Pagefault        bool            // exit reason is page fault (for handling multiple return type)
	ExitReason       ExitReasonTypes // exit reason
	GasRemain        Gas             // gas remain
	Register         Registers       // new registers
	Ram              Memory          // new memory
	Addition         any             // addition host-call context
	PagefaultAddress uint64          // page fault address, only use for page fault
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
	exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime := SingleStepInvoke(code, uint32(counter), gas, reg, ram)
	fmt.Println(exitreason_prime, counter_prime, gas_prime, reg_prime, memory_prime)
	if exitreason_prime == HALT || exitreason_prime == PANIC || exitreason_prime == OUT_OF_GAS || exitreason_prime == PAGE_FAULT {
		psi_result.Pagefault = false
		psi_result.ExitReason = exitreason_prime
		psi_result.Counter = uint64(counter_prime)
		psi_result.Gas = gas_prime
		psi_result.Reg = reg_prime
		psi_result.Ram = memory_prime
		psi_result.Addition = addition
	} else if exitreason_prime == HOST_CALL {
		var inst uint64 // TODO How to get the address(h)
		omega_result := omega(inst, gas_prime, reg_prime, ram, addition)
		if omega_result.Pagefault {
			psi_result.Pagefault = true
			psi_result.PagefaultAddress = omega_result.PagefaultAddress
			psi_result.Counter = uint64(counter_prime)
			psi_result.Gas = gas_prime
			psi_result.Reg = reg_prime
			psi_result.Ram = memory_prime
			psi_result.ExitReason = PAGE_FAULT
			psi_result.Addition = addition
		} else if omega_result.ExitReason == CONTINUE {
			return Psi_H(code, ProgramCounter(counter_prime+1), omega_result.GasRemain, omega_result.Register, omega_result.Ram, omega, omega_result.Addition)
			// TODO HOW TO USE SKIP??
		} else if omega_result.ExitReason == PANIC || omega_result.ExitReason == OUT_OF_GAS || omega_result.ExitReason == HALT {
			psi_result.Pagefault = false
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

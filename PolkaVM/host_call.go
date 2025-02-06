package PolkaVM

type Psi_H_ReturnType struct {
	Pagefault        bool            // exit reason is page fault (for handling multiple return type)
	ExitReason       ExitReasonTypes // exit reason
	Counter          uint64          // new instruction counter
	Gas              uint64          // gas remain
	Reg              Registers       // new registers
	Ram              PageMap         // new memory
	Addition         any             // addition host-call context
	PagefaultAddress uint64          // page fault address, only use for page fault
}

type OmegaReturnType struct {
	Pagefault        bool            // exit reason is page fault (for handling multiple return type)
	ExitReason       ExitReasonTypes // exit reason
	GasRemain        Gas             // gas remain
	Register         Registers       // new registers
	Ram              PageMap         // new memory
	Addition         any             // addition host-call context
	PagefaultAddress uint64          // page fault address, only use for page fault
}

// (A.31) Ψ_H
func Psi_H(
	code ProgramCode, // program code
	counter ProgramCounter, // program counter
	gas Gas, // gas counter
	reg Registers, // registers
	ram PageMap, // memory
	omega Omega, // jump table
	addition any, // host-call context
) (
	psi_result Psi_H_ReturnType,
) {
	// TODO: Implement Ψ_H function.
	return
}

// (A.32) Ω⟨X⟩
type Omega func(
	uint64, // instruction
	Gas, // gas counter
	Registers, // registers
	PageMap, // memory
	any, // host-call context
) OmegaReturnType

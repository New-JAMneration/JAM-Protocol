//go:build linux && amd64

package recompiler

import "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"

// PVMRegCount is the number of PVM general-purpose registers (RA, SP, T0–T2, S0–S1, A0–A5).
const PVMRegCount = 13

// PVMToX86 maps each PVM register index to its dedicated x86-64 native register.
// All 13 PVM registers are statically assigned — zero spill.
//
// Allocation rationale:
//   - High-frequency registers (RA, SP, T0–T2, A5) get non-REX registers (saves 1 byte/instruction)
//   - Lower-frequency registers (S0, S1, A0–A4) get R8–R14
//   - RA/SP control-region slots use disp8 offsets (see pvmRegSlot) for division spill paths
var PVMToX86 = [PVMRegCount]asm.Register{
	/* RA = 0  */ asm.RAX,
	/* SP = 1  */ asm.RDX,
	/* T0 = 2  */ asm.RBX,
	/* T1 = 3  */ asm.RSI,
	/* T2 = 4  */ asm.RDI,
	/* S0 = 5  */ asm.R8,
	/* S1 = 6  */ asm.R9,
	/* A0 = 7  */ asm.R10,
	/* A1 = 8  */ asm.R11,
	/* A2 = 9  */ asm.R12,
	/* A3 = 10 */ asm.R13,
	/* A4 = 11 */ asm.R14,
	/* A5 = 12 */ asm.RBP,
}

// Reserved registers (not allocated to any PVM register):
const (
	RegGuestBase = asm.R15 // R15 = guest memory base; control region at [R15 - offset]
	RegScratch   = asm.RCX // scratch for DIV (needs CL), shifts, address calculation
	// RSP is the native stack pointer — implicitly reserved
)

// PVMReg returns the x86-64 register assigned to a PVM register index.
func PVMReg(index uint8) asm.Register {
	return PVMToX86[index]
}

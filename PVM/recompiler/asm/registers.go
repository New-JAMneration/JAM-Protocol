// Package asm provides a minimal x86-64 assembler for the PVM JIT recompiler.
package asm

// Register represents an x86-64 general-purpose register (0–15).
type Register uint8

const (
	RAX Register = 0
	RCX Register = 1
	RDX Register = 2
	RBX Register = 3
	RSP Register = 4
	RBP Register = 5
	RSI Register = 6
	RDI Register = 7
	R8  Register = 8
	R9  Register = 9
	R10 Register = 10
	R11 Register = 11
	R12 Register = 12
	R13 Register = 13
	R14 Register = 14
	R15 Register = 15
)

// IsExtended reports whether the register requires a REX prefix bit (R8–R15).
func (r Register) IsExtended() bool { return r >= 8 }

// Lo3 returns the low 3 bits of the register number (for ModR/M / SIB encoding).
func (r Register) Lo3() uint8 { return uint8(r) & 0x07 }

// IsCalleeSaved reports whether the register is callee-saved in the System V AMD64 ABI.
func (r Register) IsCalleeSaved() bool {
	switch r {
	case RBX, RBP, R12, R13, R14, R15:
		return true
	}
	return false
}

// String returns the conventional name.
func (r Register) String() string {
	return regNames[r]
}

var regNames = [16]string{
	"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi",
	"r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15",
}

// ConditionCode is an x86-64 condition code for Jcc / SETcc / CMOVcc.
type ConditionCode uint8

const (
	// overflow flag (OF)
	CondO  ConditionCode = 0x00 // overflow
	CondNO ConditionCode = 0x01 // not overflow
	// sign flag (SF)
	CondB  ConditionCode = 0x02 // below (unsigned <), CF=1
	CondAE ConditionCode = 0x03 // above or equal (unsigned >=), CF=0
	// zero flag (ZF)
	CondEQ ConditionCode = 0x04 // equal, ZF=1
	CondNE ConditionCode = 0x05 // not equal, ZF=0
	// carry flag (CF)
	CondBE ConditionCode = 0x06 // below or equal (unsigned <=)
	CondA  ConditionCode = 0x07 // above (unsigned >)
	// sign flag (SF)
	CondS  ConditionCode = 0x08 // sign (negative)
	CondNS ConditionCode = 0x09 // not sign (positive)
	// less than flag (LT)
	CondLT ConditionCode = 0x0C // less (signed <), SF≠OF
	// greater than or equal flag (GE)
	CondGE ConditionCode = 0x0D // greater or equal (signed >=), SF=OF
	// less than or equal flag (LE)
	CondLE ConditionCode = 0x0E // less or equal (signed <=)
	// greater than flag (GT)
	CondGT ConditionCode = 0x0F // greater (signed >)
)

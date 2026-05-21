//go:build linux && amd64

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

func threeRegFromMeta(instr *PVM.InstrMeta) (dReg, aReg, bReg asm.Register) {
	return PVMReg(instr.Dst), PVMReg(instr.Src[0]), PVMReg(instr.Src[1])
}

func twoRegFromMeta(instr *PVM.InstrMeta) (dReg, aReg asm.Register) {
	return PVMReg(instr.Dst), PVMReg(instr.Src[0])
}

func twoRegImmFromMeta(instr *PVM.InstrMeta) (aReg, bReg asm.Register, vX uint64) {
	return PVMReg(instr.Dst), PVMReg(instr.Src[0]), instr.Imm[0]
}

func oneRegImmFromMeta(instr *PVM.InstrMeta) (aReg asm.Register, vX uint64) {
	return PVMReg(instr.Dst), instr.Imm[0]
}

func twoImmFromMeta(instr *PVM.InstrMeta) (vX, vY uint64) {
	return instr.Imm[0], instr.Imm[1]
}

func oneRegTwoImmFromMeta(instr *PVM.InstrMeta) (aReg asm.Register, vX, vY uint64) {
	return PVMReg(instr.Src[0]), instr.Imm[0], instr.Imm[1]
}

func twoRegTwoImmFromMeta(instr *PVM.InstrMeta) (aReg, bReg asm.Register, vX, vY uint64) {
	return PVMReg(instr.Dst), PVMReg(instr.Src[0]), instr.Imm[0], instr.Imm[1]
}

func branchTwoRegFromMeta(instr *PVM.InstrMeta) (aReg, bReg asm.Register, targetPC PVM.ProgramCounter) {
	return PVMReg(instr.Src[0]), PVMReg(instr.Src[1]), PVM.ProgramCounter(instr.Imm[0])
}

func branchOneRegImmFromMeta(instr *PVM.InstrMeta) (aReg asm.Register, vX uint64, targetPC PVM.ProgramCounter) {
	return PVMReg(instr.Dst), instr.Imm[0], PVM.ProgramCounter(instr.Imm[1])
}

func fallthroughPC(instr *PVM.InstrMeta) PVM.ProgramCounter {
	return instr.PC + PVM.ProgramCounter(instr.SkipLen) + 1
}

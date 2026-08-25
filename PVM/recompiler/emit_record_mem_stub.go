//go:build linux && amd64 && cgo && !pvmtrace

package recompiler

import "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"

// Production builds omit debug mem-trace stores into the control region.
func emitRecordMemAccessImmVal(_ *asm.Assembler, _ uint32, _ uint64) {}

func emitRecordGuestAddrFromScratch(_ *asm.Assembler) {}

func emitRecordMemValFromReg(_ *asm.Assembler, _ asm.Register) {}

func emitRecordMemValImm(_ *asm.Assembler, _ uint64) {}

func emitRecordMemAccessImm(_ *asm.Assembler, _ uint32, _ asm.Register) {}

func emitRecordMemAccessReg(_ *asm.Assembler, _, _ asm.Register) {}

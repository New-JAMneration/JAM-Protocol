//go:build linux && amd64

package recompiler

import "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"

// emitRecordMemAccessImmVal writes a compile-time guest address and value for debug trace.
func emitRecordMemAccessImmVal(a *asm.Assembler, guestAddr uint32, val uint64) {
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetMemAccessAddr), int32(guestAddr))
	a.MovImm64ToReg(RegScratch, val)
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessVal), RegScratch)
}

// emitRecordGuestAddrFromScratch writes the 32-bit guest address in RegScratch.
func emitRecordGuestAddrFromScratch(a *asm.Assembler) {
	a.StoreDword(RegGuestBase, -int32(OffsetMemAccessAddr), RegScratch)
}

// emitRecordMemValFromReg writes the post-access register value for debug trace.
func emitRecordMemValFromReg(a *asm.Assembler, valReg asm.Register) {
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessVal), valReg)
}

// emitRecordMemValImm writes a compile-time value; guest addr must already be in the control region.
func emitRecordMemValImm(a *asm.Assembler, val uint64) {
	a.MovImm64ToReg(RegScratch, val)
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessVal), RegScratch)
}

// emitRecordMemAccessImm writes guest address and value to the control region for debug trace.
func emitRecordMemAccessImm(a *asm.Assembler, guestAddr uint32, valReg asm.Register) {
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetMemAccessAddr), int32(guestAddr))
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessVal), valReg)
}

// emitRecordMemAccessReg writes guest address (register) and value to the control region.
func emitRecordMemAccessReg(a *asm.Assembler, guestAddrReg, valReg asm.Register) {
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessAddr), guestAddrReg)
	a.MovRegToMem(RegGuestBase, -int32(OffsetMemAccessVal), valReg)
}

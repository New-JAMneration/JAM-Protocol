//go:build linux && amd64 && trace

package recompiler

import (
	"encoding/binary"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

// traceReadGuestMem reads width bytes at guest addr when the JIT segment check
// confirms readability. Stack/heap pages are readable from Go; PROT_NONE pages
// are skipped to avoid SIGSEGV.
func traceReadGuestMem(ctx *JITContext, addr uint32, width int) (uint64, bool) {
	if width <= 0 {
		return 0, false
	}
	gm := ctx.GuestMemory()
	a := uint64(addr)
	l := uint64(width)
	if !gm.IsReadable(a, l) {
		return 0, false
	}
	b := gm.Read(a, l)
	switch width {
	case 1:
		return uint64(b[0]), true
	case 2:
		return uint64(binary.LittleEndian.Uint16(b)), true
	case 4:
		return uint64(binary.LittleEndian.Uint32(b)), true
	case 8:
		return binary.LittleEndian.Uint64(b), true
	default:
		return 0, false
	}
}

// traceRecordedMemAccess returns load/store addr and value for trace streams.
// Only successful guest reads/writes are recorded, matching interpreter trace.
func traceRecordedMemAccess(ctx *JITContext, opcode byte, recordedAddr uint32) (loadAddr uint32, loadVal uint64, storeAddr uint32, storeVal uint64) {
	if recordedAddr == 0 {
		return 0, 0, 0, 0
	}
	width, ok := PVM.MemAccessWidth(opcode)
	if !ok {
		return 0, 0, 0, 0
	}

	if PVM.IsStoreOpcode(opcode) {
		if v, ok := traceReadGuestMem(ctx, recordedAddr, width); ok {
			return 0, 0, recordedAddr, v
		}
		// Mirror interpreter: only record stores after a successful write.
		return 0, 0, 0, 0
	}

	if PVM.IsLoadOpcode(opcode) {
		if v, ok := traceReadGuestMem(ctx, recordedAddr, width); ok {
			return recordedAddr, v, 0, 0
		}
		// Mirror interpreter: only record loads after a successful read.
		return 0, 0, 0, 0
	}
	return 0, 0, 0, 0
}

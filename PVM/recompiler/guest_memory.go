//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"golang.org/x/sys/unix"
)

// InitFromProgram decodes a PVM program blob and maps its segments into the
// mmap'd guest memory region with correct mprotect permissions.
// Guest memory is a contiguous 4GB []byte — all subsequent access (from both
// JIT native code and Go host calls) is simply guestMem[addr : addr+size].
func (ctx *JITContext) InitFromProgram(p PVM.StandardCodeFormat, a PVM.Argument) (PVM.Instructions, PVM.Registers, error) {
	c, o, w, z, s, err := PVM.DecodeSerializedValues(p)
	if err != nil {
		return nil, PVM.Registers{}, fmt.Errorf("decode program: %w", err)
	}

	if 5*PVM.ZZ+uint64(PVM.Z(len(o)))+uint64(PVM.Z(len(w)+int(z)*int(PVM.ZP)+int(s)+PVM.ZI)) > 1<<32 {
		return nil, PVM.Registers{}, fmt.Errorf("memory layout exceeds 4GB")
	}

	readOnlyStart := uint32(PVM.ZZ)
	// readOnlyEnd := readOnlyStart + uint32(len(o))
	readOnlyPadding := readOnlyStart + PVM.P(len(o))
	readWriteStart := 2*PVM.ZZ + PVM.Z(len(o))
	// readWriteEnd := readWriteStart + uint32(len(w))
	readWritePadding := readWriteStart + PVM.P(len(w)) + uint32(z)*PVM.ZP
	heapStart := readWritePadding

	stackEnd := uint32(1<<32 - 2*PVM.ZZ - PVM.ZI)
	stackStart := stackEnd - PVM.P(int(s))
	argumentStart := uint32(1<<32 - PVM.ZZ - PVM.ZI)
	argumentEnd := argumentStart + uint32(len(a))
	argumentPadding := argumentEnd + PVM.P(len(a))

	if err := ctx.mapSegment(readOnlyStart, readOnlyPadding, o, unix.PROT_READ); err != nil {
		return nil, PVM.Registers{}, err
	}
	if err := ctx.mapSegment(readWriteStart, readWritePadding, w, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		return nil, PVM.Registers{}, err
	}
	if err := ctx.mapSegment(stackStart, stackEnd, nil, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		return nil, PVM.Registers{}, err
	}
	if err := ctx.mapSegment(argumentStart, argumentPadding, a, unix.PROT_READ); err != nil {
		return nil, PVM.Registers{}, err
	}

	ctx.WriteHeapPointer(uint64(heapStart))
	ctx.heapLimit = uint64(stackStart)
	// Record segment boundaries for segment-aware GuestMemory checks (Layer 1).
	ctx.seg = guestSegments{
		roStart:       uint64(readOnlyStart),
		roEnd:         uint64(readOnlyPadding),
		rwStart:       uint64(readWriteStart),
		rwPaddingEnd:  uint64(readWritePadding),
		stackStart:    uint64(stackStart),
		stackEnd:      uint64(stackEnd),
		argStart:      uint64(argumentStart),
		argEnd:        uint64(argumentPadding),
	}

	var regs PVM.Registers
	regs[0] = uint64(1<<32 - 1<<16)
	regs[1] = uint64(1<<32 - 2*PVM.ZZ - PVM.ZI)
	regs[7] = uint64(1<<32 - PVM.ZZ - PVM.ZI)
	regs[8] = uint64(len(a))

	return c, regs, nil
}

// mapSegment sets mprotect on a byte range in guest memory and copies content into it.
// If content is non-empty and the target prot is read-only, temporarily grants RW for the copy.
func (ctx *JITContext) mapSegment(start, end uint32, content []byte, prot int) error {
	if end <= start {
		return nil
	}

	region := ctx.guestMem[start:end]

	if len(content) > 0 && prot&unix.PROT_WRITE == 0 {
		if err := unix.Mprotect(region, unix.PROT_READ|unix.PROT_WRITE); err != nil {
			return fmt.Errorf("mprotect 0x%08x..0x%08x for write: %w", start, end, err)
		}
		copy(region, content)
		if err := unix.Mprotect(region, prot); err != nil {
			return fmt.Errorf("mprotect 0x%08x..0x%08x: %w", start, end, err)
		}
		ctx.markPageRange(start, end, pageAccessFromProt(prot))
		return nil
	}

	if err := unix.Mprotect(region, prot); err != nil {
		return fmt.Errorf("mprotect 0x%08x..0x%08x: %w", start, end, err)
	}
	if len(content) > 0 {
		copy(region, content)
	}
	ctx.markPageRange(start, end, pageAccessFromProt(prot))
	return nil
}

// SetPageAccess changes the hardware protection of a guest memory page.
func (ctx *JITContext) SetPageAccess(pageNum uint32, prot int) error {
	byteOffset := uint64(pageNum) * PVM.ZP
	if byteOffset+PVM.ZP > GuestMemorySize {
		return fmt.Errorf("page %d (offset 0x%x) exceeds 4GB guest memory", pageNum, byteOffset)
	}
	if err := unix.Mprotect(ctx.guestMem[byteOffset:byteOffset+PVM.ZP], prot); err != nil {
		return err
	}
	ctx.setPageAccess(pageNum, pageAccessFromProt(prot))
	return nil
}

func (ctx *JITContext) ensurePages() {
	if ctx.pages == nil {
		ctx.pages = make(map[uint32]pageAccess)
	}
}

func (ctx *JITContext) markPageRange(start, end uint32, access pageAccess) {
	if end <= start {
		return
	}
	ctx.ensurePages()
	for addr := start; addr < end; addr += PVM.ZP {
		ctx.pages[addr/PVM.ZP] = access
	}
}

func (ctx *JITContext) setPageAccess(pageNum uint32, access pageAccess) {
	ctx.ensurePages()
	ctx.pages[pageNum] = access
}

func (ctx *JITContext) getPageAccess(pageNum uint32) pageAccess {
	if ctx.pages == nil {
		return pageInaccessible
	}
	access, ok := ctx.pages[pageNum]
	if !ok {
		return pageInaccessible
	}
	return access
}

func pageAccessFromProt(prot int) pageAccess {
	if prot&unix.PROT_READ == 0 {
		return pageInaccessible
	}
	if prot&unix.PROT_WRITE != 0 {
		return pageReadWrite
	}
	return pageReadOnly
}

// pagesReadable matches PVM isReadable: every page in range must be allocated.
func (ctx *JITContext) pagesReadable(start, length uint64) bool {
	if length == 0 {
		return true
	}
	if length > (1<<32) || start > (1<<32)-length {
		return false
	}
	startPage := uint32(start / PVM.ZP)
	endPage := uint32((start + length - 1) / PVM.ZP)
	for p := startPage; p <= endPage; p++ {
		if ctx.getPageAccess(p) == pageInaccessible {
			return false
		}
	}
	return true
}

// pagesWriteable matches PVM isWriteable: every page must be MemoryReadWrite.
func (ctx *JITContext) pagesWriteable(start, length uint64) bool {
	if length == 0 {
		return true
	}
	if length > (1<<32) || start > (1<<32)-length {
		return false
	}
	startPage := uint32(start / PVM.ZP)
	endPage := uint32((start + length - 1) / PVM.ZP)
	for p := startPage; p <= endPage; p++ {
		if ctx.getPageAccess(p) != pageReadWrite {
			return false
		}
	}
	return true
}

// GuestMemory returns the PVM.GuestMemory view of this context for omega / R().
// The returned value is a single pointer (pointer-shaped), so assigning it to a
// GuestMemory interface is allocation-free.
func (ctx *JITContext) GuestMemory() PVM.GuestMemory { return jitGuestMemory{ctx: ctx} }

// jitGuestMemory implements PVM.GuestMemory over the flat mmap'd guest memory,
// using the recorded segment boundaries for permission checks (Layer 1). See
// guestSegments for why this Go-side check is mandatory (recover() cannot catch
// the hardware SIGSEGV from touching a PROT_NONE page).
type jitGuestMemory struct {
	ctx *JITContext
}

func (g jitGuestMemory) IsReadable(addr, length uint64) bool {
	return g.ctx.pagesReadable(addr, length)
}

func (g jitGuestMemory) IsWriteable(addr, length uint64) bool {
	return g.ctx.pagesWriteable(addr, length)
}

// Read returns a freshly-allocated copy of [addr, addr+length); the caller must
// have confirmed IsReadable. A copy (not a slice into guestMem) is required
// because R() and some omega calls retain the bytes after the JIT context — and
// thus guestMem — is unmapped.
func (g jitGuestMemory) Read(addr, length uint64) []byte {
	out := make([]byte, length)
	copy(out, g.ctx.guestMem[addr:addr+length])
	return out
}

// Write copies data into [addr, addr+len(data)); caller must have confirmed IsWriteable.
func (g jitGuestMemory) Write(addr uint64, data []byte) {
	copy(g.ctx.guestMem[addr:addr+uint64(len(data))], data)
}


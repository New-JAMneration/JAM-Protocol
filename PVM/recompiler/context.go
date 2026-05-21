//go:build linux && amd64

// Package recompiler implements the PVM JIT recompiler's memory management layer,
// including executable memory (W^X), guest memory (unified mmap with slice aliasing),
// code cache, and control region accessors for the Generic Sandbox memory layout.
package recompiler

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"golang.org/x/sys/unix"
)

// Generic Sandbox mmap layout sizes
const (
	ControlRegionSize = 4096                   // 4KB control region before guest memory
	GuestMemorySize   = 4 * 1024 * 1024 * 1024 // 4GB PVM address space
	GuardPageSize     = 4096                   // 4KB guard page after guest memory
	TotalMmapSize     = ControlRegionSize + GuestMemorySize + GuardPageSize
)

// Control region field offsets (negative from R15 = guestBase).
// Fields are laid out from the end of the control region backward,
// with the most frequently accessed fields closest to R15 (smallest offset).
const (
	OffsetReturnStack = 8   // R15 - 8:   uintptr (signal handler RSP restore)
	OffsetReturnAddr  = 16  // R15 - 16:  uintptr (signal handler return address)
	OffsetHeapPointer = 24  // R15 - 24:  uint64
	OffsetExitPC      = 32  // R15 - 32:  uint32 (+ 4B padding)
	OffsetExitReason  = 40  // R15 - 40:  PVM.ExitReason (uint64)
	OffsetGas         = 48  // R15 - 48:  int64 — within disp8 range (-128..+127)
	OffsetRegisters   = 152 // R15 - 152: [13]uint64 = 104 bytes (R15-152 .. R15-49)
)

// guestSegments records the guest-memory segment boundaries captured during
// InitFromProgram. They drive the segment-aware GuestMemory permission checks
// (Layer 1): reading an inaccessible (PROT_NONE) page in Go would raise a
// hardware SIGSEGV that recover() cannot catch, so omega / R() must reject
// out-of-segment ranges in Go before touching guestMem. Layer 1 is kept a
// subset of the mprotect state (Layer 2), so anything Layer 1 admits is safe to
// touch.
//
// GuestMemory permission checks (Layer 1) mirror interpreter isReadable/isWriteable:
// walk the page table (ctx.pages) populated by mapSegment and sbrk SetPageAccess.
// stackStart == JITContext.heapLimit.
type guestSegments struct {
	roStart, roEnd        uint64 // read-only program code/data (padded)
	rwStart, rwPaddingEnd uint64 // RW data + z-page reservation (heapStart)
	stackStart, stackEnd  uint64
	argStart, argEnd      uint64 // argument (read-only, padded)
}

// JITContext holds the unified mmap region (control region + guest memory + guard page)
// and provides Go-side accessors for the control region fields.
type JITContext struct {
	rawMem         []byte         // full mmap'd region returned by unix.Mmap
	guestBasePtr   unsafe.Pointer // points to rawMem[ControlRegionSize] — R15's value in JIT code
	controlMem     []byte         // rawMem[0 : ControlRegionSize]
	guestMem       []byte         // rawMem[ControlRegionSize : ControlRegionSize+GuestMemorySize]
	executableMem  *ExecutableMemory
	trampolineAddr uintptr               // cached entry trampoline for this executable memory
	heapLimit      uint64                // stackStart; sbrk must not grow past this (A.36 / instSbrkMeta)
	seg            guestSegments         // segment boundaries (init layout metadata)
	pages          map[uint32]pageAccess // Layer-1 page permissions for host-call checks
}

// pageAccess mirrors PVM.MemoryAccess for the recompiler page table.
type pageAccess uint8

const (
	pageInaccessible pageAccess = iota
	pageReadOnly
	pageReadWrite
)

// NewJITContext allocates the unified mmap region for the Generic Sandbox layout:
//
//	[control region 4KB] [guest memory 4GB] [guard page 4KB]
func NewJITContext() (*JITContext, error) {
	rawMem, err := unix.Mmap(
		-1, 0, TotalMmapSize,
		unix.PROT_NONE,
		unix.MAP_ANON|unix.MAP_PRIVATE|unix.MAP_NORESERVE,
	)
	if err != nil {
		return nil, fmt.Errorf("mmap unified region (%d bytes): %w", TotalMmapSize, err)
	}

	if err := unix.Mprotect(rawMem[:ControlRegionSize], unix.PROT_READ|unix.PROT_WRITE); err != nil {
		unix.Munmap(rawMem)
		return nil, fmt.Errorf("mprotect control region: %w", err)
	}

	ctx := &JITContext{
		rawMem:       rawMem,
		guestBasePtr: unsafe.Pointer(&rawMem[ControlRegionSize]),
		controlMem:   rawMem[:ControlRegionSize],
		guestMem:     rawMem[ControlRegionSize : ControlRegionSize+GuestMemorySize],
	}
	return ctx, nil
}

// Close releases the unified mmap region.
func (ctx *JITContext) Close() error {
	if ctx.rawMem == nil {
		return nil
	}
	err := unix.Munmap(ctx.rawMem)
	ctx.rawMem = nil
	ctx.controlMem = nil
	ctx.guestMem = nil
	ctx.guestBasePtr = nil
	return err
}

// GuestBase returns the guest memory base address as a uintptr (R15's value in JIT code).
func (ctx *JITContext) GuestBase() uintptr { return uintptr(ctx.guestBasePtr) }

// GuestMem returns the guest memory byte slice (4GB).
func (ctx *JITContext) GuestMem() []byte { return ctx.guestMem }

// controlPtr returns a pointer into the control region at a negative offset from guestBase.
// Uses unsafe.Add which is the sanctioned way to do pointer arithmetic in Go.
func (ctx *JITContext) controlPtr(negOffset int) unsafe.Pointer {
	return unsafe.Add(ctx.guestBasePtr, -negOffset)
}

// --- Control region read/write helpers ---

func (ctx *JITContext) ReadGas() PVM.Gas {
	return *(*PVM.Gas)(ctx.controlPtr(OffsetGas))
}

// ReadGasInto is designed for host-call hot path to keep the snapshot heap-free.
func (ctx *JITContext) ReadGasInto(dst *PVM.Gas) {
	*dst = *(*PVM.Gas)(ctx.controlPtr(OffsetGas))
}

func (ctx *JITContext) WriteGas(gas PVM.Gas) {
	*(*PVM.Gas)(ctx.controlPtr(OffsetGas)) = gas
}

func (ctx *JITContext) ReadExitReason() PVM.ExitReason {
	return *(*PVM.ExitReason)(ctx.controlPtr(OffsetExitReason))
}

func (ctx *JITContext) WriteExitReason(reason PVM.ExitReason) {
	*(*PVM.ExitReason)(ctx.controlPtr(OffsetExitReason)) = reason
}

func (ctx *JITContext) ReadExitPC() PVM.ProgramCounter {
	return *(*PVM.ProgramCounter)(ctx.controlPtr(OffsetExitPC))
}

func (ctx *JITContext) WriteExitPC(pc PVM.ProgramCounter) {
	*(*PVM.ProgramCounter)(ctx.controlPtr(OffsetExitPC)) = pc
}

func (ctx *JITContext) ReadHeapPointer() uint64 {
	return *(*uint64)(ctx.controlPtr(OffsetHeapPointer))
}

func (ctx *JITContext) WriteHeapPointer(hp uint64) {
	*(*uint64)(ctx.controlPtr(OffsetHeapPointer)) = hp
}

// SetExecutableMemory attaches an ExecutableMemory to this context.
func (ctx *JITContext) SetExecutableMemory(em *ExecutableMemory) {
	ctx.executableMem = em
	ctx.trampolineAddr = 0
}

// ReadRegisters reads all 13 PVM registers from the control region.
// Registers are stored as [13]uint64 starting at guestBase - OffsetRegisters.
func (ctx *JITContext) ReadRegisters() PVM.Registers {
	var regs PVM.Registers
	ctx.ReadRegistersInto(&regs)
	return regs
}

// ReadRegistersInto reads all 13 PVM registers from the control region
// into the caller's buffer, keeping the snapshot heap-free.
func (ctx *JITContext) ReadRegistersInto(dst *PVM.Registers) {
	regBase := ControlRegionSize - OffsetRegisters
	for i := range 13 {
		off := regBase + pvmRegSlot[i]*8
		dst[i] = binary.LittleEndian.Uint64(ctx.rawMem[off : off+8])
	}
}

// WriteRegisters writes all 13 PVM registers to the control region.
func (ctx *JITContext) WriteRegisters(regs PVM.Registers) {
	regBase := ControlRegionSize - OffsetRegisters
	for i := range 13 {
		off := regBase + pvmRegSlot[i]*8
		binary.LittleEndian.PutUint64(ctx.rawMem[off:off+8], regs[i])
	}
}

// ReadRegister reads a single PVM register by index from the control region.
func (ctx *JITContext) ReadRegister(idx uint8) uint64 {
	if idx > 12 {
		return 0
	}
	off := ControlRegionSize - OffsetRegisters + pvmRegSlot[idx]*8
	return binary.LittleEndian.Uint64(ctx.rawMem[off : off+8])
}

// Memory access recording for debug single-step mode.
// Native code writes addr (u32) + val (u64) into the control region at fixed offsets.
// These are only valid for the instruction that just executed.

const (
	OffsetMemAccessAddr = 160 // R15 - 160: uint32 (+ 4B padding)
	OffsetMemAccessVal  = 168 // R15 - 168: uint64

	// JIT djump metadata pointers (Go-heap / rodata addresses; read-only during invoke).
	OffsetDjumpTable    = 176 // R15 - 176: uintptr — jump table rodata in ExecutableMemory
	OffsetDjumpBitmask  = 184 // R15 - 184: uintptr — bitmask rodata in ExecutableMemory
	OffsetDjumpDispatch = 192 // R15 - 192: uintptr — PC→native dispatch table ([]uintptr)
)

// HasMemAccess returns true if the last instruction recorded a memory access.
func (ctx *JITContext) HasMemAccess() bool {
	off := ControlRegionSize - OffsetMemAccessAddr
	addr := binary.LittleEndian.Uint32(ctx.rawMem[off : off+4])
	return addr != 0
}

// ReadMemAccess reads the memory access addr and value from the control region.
func (ctx *JITContext) ReadMemAccess() (addr uint32, val uint64) {
	offA := ControlRegionSize - OffsetMemAccessAddr
	offV := ControlRegionSize - OffsetMemAccessVal
	addr = binary.LittleEndian.Uint32(ctx.rawMem[offA : offA+4])
	val = binary.LittleEndian.Uint64(ctx.rawMem[offV : offV+8])
	return
}

// ClearMemAccess resets the memory access fields to zero.
func (ctx *JITContext) ClearMemAccess() {
	offA := ControlRegionSize - OffsetMemAccessAddr
	offV := ControlRegionSize - OffsetMemAccessVal
	binary.LittleEndian.PutUint32(ctx.rawMem[offA:offA+4], 0)
	binary.LittleEndian.PutUint64(ctx.rawMem[offV:offV+8], 0)
}

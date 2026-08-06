package PVM

// GuestMemory is the single view of guest memory for shared JAM semantics
// (omega host-calls in host_call_*.go and the R() finaliser in
// argument_invocation.go). It decouples those semantics from any concrete
// memory representation so they are written exactly once.
//
// Two backends implement it:
//   - interpreter: pagedGuestMemory, wrapping the paged Memory type (this file).
//   - recompiler:  wrapping JITContext (flat mmap + segments), see
//     PVM/recompiler/guest_memory.go.
//
// Contract:
//   - Read must be preceded by a successful IsReadable check, and Write by a
//     successful IsWriteable check.
//   - IsReadable/IsWriteable are segment-aware (not a flat 4GB bound), so backends
//     with hardware-protected pages (recompiler) can reject inaccessible ranges
//     in Go before any native access faults.
//   - Read returns a freshly-allocated copy that the caller may retain. Both
//     backends copy: the interpreter via Memory.Read, the recompiler because its
//     guest memory is unmapped when the JIT context closes, so a slice into it
//     would dangle once R()/omega return.
//
// Escape note: every implementation MUST be pointer-shaped (exactly one pointer
// field, or methods on a pointer type) so converting it to a GuestMemory value
// is allocation-free (Go's direct-interface optimisation). Build the wrapper
// once per Psi_M invocation and store it, rather than rebuilding on each memory
// op. This interface is for the host-call / R() path only — the hot instruction
// loop must keep using concrete types directly.
type GuestMemory interface {
	IsReadable(addr, length uint64) bool
	IsWriteable(addr, length uint64) bool
	Read(addr, length uint64) []byte // caller must have checked IsReadable
	Write(addr uint64, data []byte)  // caller must have checked IsWriteable

	// B.5 Ω_♊: heap state in page indices.
	HeapPages() uint64              // h: current heap-top page index
	HeapMaxPages() uint64           // b: max possible heap-top page index
	GrowHeapTo(targetPage uint64)   // expand heap to targetPage (caller ensures h < target ≤ b)
}

// pagedGuestMemory adapts the interpreter's paged Memory to GuestMemory.
// It is a single pointer (pointer-shaped) so conversion to GuestMemory does not
// allocate. Its behaviour matches the pre-existing isReadable / isWriteable /
// Memory.Read / Memory.Write helpers exactly, so wiring it into omega and R()
// is a pure refactor with no semantic change.
type pagedGuestMemory struct {
	mem *Memory
}

// NewPagedGuestMemory wraps a paged Memory as a GuestMemory. The interpreter
// backend uses this; mem must remain valid for the lifetime of the wrapper.
func NewPagedGuestMemory(mem *Memory) GuestMemory {
	return pagedGuestMemory{mem: mem}
}

func (p pagedGuestMemory) IsReadable(addr, length uint64) bool {
	if p.mem == nil {
		return false
	}
	return isReadable(addr, length, *p.mem)
}

func (p pagedGuestMemory) IsWriteable(addr, length uint64) bool {
	if p.mem == nil {
		return false
	}
	return isWriteable(addr, length, *p.mem)
}

// Read returns a freshly-allocated copy of [addr, addr+length); caller must
// have confirmed IsReadable. Note Memory.Read returns types.ByteSequence, which
// is assignable to the interface's []byte (slices are unnamed types).
func (p pagedGuestMemory) Read(addr, length uint64) []byte {
	return p.mem.Read(addr, length)
}

// Write copies data into [addr, addr+len(data)); caller must have confirmed
// Writeable. Mirrors Memory.Write.
func (p pagedGuestMemory) Write(addr uint64, data []byte) {
	p.mem.Write(addr, data)
}

func (p pagedGuestMemory) HeapPages() uint64 {
	return p.mem.heapPointer / uint64(ZP)
}

func (p pagedGuestMemory) HeapMaxPages() uint64 {
	return p.mem.heapLimit / uint64(ZP)
}

// GrowHeapTo expands the heap so pages [h..targetPage) become writable.
// Caller has already verified h < targetPage ≤ b.
func (p pagedGuestMemory) GrowHeapTo(targetPage uint64) {
	mem := p.mem
	newHP := targetPage * uint64(ZP)
	oldBound := P(int(mem.heapPointer))
	newBound := P(int(newHP))
	if newHP > uint64(oldBound) {
		allocateMemorySegment(mem, uint32(mem.heapPointer), uint32(newBound), nil, MemoryReadWrite)
	}
	mem.heapPointer = newHP
}

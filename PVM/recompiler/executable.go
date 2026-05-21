//go:build linux && amd64

package recompiler

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

const DefaultExecutableSize = 16 * 1024 * 1024 // 16MB

// ExecutableMemory manages a JIT code region as a **dual mapping** of one
// memfd-backed allocation:
//
//   - rwMem  (PROT_READ|PROT_WRITE) — code is *written* here
//   - rxMem  (PROT_READ|PROT_EXEC)  — code is *executed* here (base = rxBase)
//
// Both views alias the same physical pages (MAP_SHARED on the same fd), so code
// written through rwMem is immediately runnable through rxBase with **no
// per-write mprotect**. This removes the per-block W^X toggle that the profile
// showed to be ~49% of total time (PERFORMANCE.md §1.2 / §4.0): QEMU re-translated
// the whole 16MB arena on every PROT_EXEC flip.
//
// Trade-off: the pages are simultaneously writable (rwMem) and executable (rxMem)
// at different virtual addresses — a relaxation of strict W^X, acceptable for this
// JIT sandbox because guest code cannot obtain the rwMem address.
//
// INVARIANT: every callable/stored address — CompiledBlock.NativeAddr, the djump
// dispatch entries, the entry trampoline, and the signal-handler fault window —
// MUST come from GetPtr, i.e. the rxBase (executable) view, because that is where
// execution actually happens. Writes go only through Write → rwMem.
//
// amd64 note: x86-64 keeps the instruction cache coherent with data writes, so no
// explicit icache flush / barrier is needed between Write (rwMem) and execution
// (rxBase) on the same core; cross-core coherency is guaranteed by hardware.
type ExecutableMemory struct {
	fd     int     // backing memfd
	rwMem  []byte  // PROT_READ|PROT_WRITE view — code written here
	rxMem  []byte  // PROT_READ|PROT_EXEC view — kept for Munmap
	rxBase uintptr // &rxMem[0]; base of the executable view (callable addresses)
	size   int
	used   int
}

// NewExecutableMemory allocates a memfd-backed dual mapping of the given size.
// Both an RW and an RX mapping of the same backing pages are created up front,
// so no mprotect is ever needed during compilation or execution.
func NewExecutableMemory(size int) (*ExecutableMemory, error) {
	if size <= 0 {
		size = DefaultExecutableSize
	}

	fd, err := unix.MemfdCreate("jit-code", unix.MFD_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("memfd_create jit arena: %w", err)
	}
	if err := unix.Ftruncate(fd, int64(size)); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("ftruncate jit arena (%d bytes): %w", size, err)
	}

	rwMem, err := unix.Mmap(fd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("mmap RW view (%d bytes): %w", size, err)
	}
	rxMem, err := unix.Mmap(fd, 0, size, unix.PROT_READ|unix.PROT_EXEC, unix.MAP_SHARED)
	if err != nil {
		unix.Munmap(rwMem)
		unix.Close(fd)
		return nil, fmt.Errorf("mmap RX view (%d bytes): %w", size, err)
	}

	return &ExecutableMemory{
		fd:     fd,
		rwMem:  rwMem,
		rxMem:  rxMem,
		rxBase: uintptr(unsafe.Pointer(&rxMem[0])),
		size:   size,
		used:   0,
	}, nil
}

// Write appends native code through the writable view and returns the byte
// offset where it was placed. No mprotect: the executable view (GetPtr) sees the
// bytes immediately because both views map the same physical pages.
func (em *ExecutableMemory) Write(code []byte) (offset int, err error) {
	if em.used+len(code) > em.size {
		return 0, fmt.Errorf("executable memory full: need %d bytes, have %d free",
			len(code), em.size-em.used)
	}
	offset = em.used
	copy(em.rwMem[offset:], code)
	em.used += len(code)
	return offset, nil
}

// GetPtr returns the callable (executable-view) address at the given byte offset.
func (em *ExecutableMemory) GetPtr(offset int) uintptr {
	return em.rxBase + uintptr(offset)
}

// Used returns the number of bytes currently written.
func (em *ExecutableMemory) Used() int { return em.used }

// Size returns the total capacity in bytes.
func (em *ExecutableMemory) Size() int { return em.size }

// Reset discards all written code (INT3-fill through the writable view) and
// resets the cursor. The executable view sees the INT3 immediately (same pages).
func (em *ExecutableMemory) Reset() error {
	for i := range em.rwMem[:em.used] {
		em.rwMem[i] = 0xCC // INT3 — trap on accidental execution
	}
	em.used = 0
	return nil
}

// Close unmaps both views and closes the backing fd.
func (em *ExecutableMemory) Close() error {
	var firstErr error
	if em.rwMem != nil {
		if err := unix.Munmap(em.rwMem); err != nil && firstErr == nil {
			firstErr = err
		}
		em.rwMem = nil
	}
	if em.rxMem != nil {
		if err := unix.Munmap(em.rxMem); err != nil && firstErr == nil {
			firstErr = err
		}
		em.rxMem = nil
		em.rxBase = 0
	}
	if em.fd != 0 {
		if err := unix.Close(em.fd); err != nil && firstErr == nil {
			firstErr = err
		}
		em.fd = 0
	}
	return firstErr
}

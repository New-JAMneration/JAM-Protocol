//go:build linux && amd64

package recompiler

import (
	"sync"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CompiledProgram is the per-CodeHash compiled artifact shared across
// invocations: the executable code arena, the PC→block cache, djump support, and
// the pre-emitted entry trampoline. Per-invocation guest state lives in a fresh
// JITContext bound via bindContext.
type CompiledProgram struct {
	program *PVM.Program
	em      *ExecutableMemory
	cache   *CodeCache
	djump   *djumpSupport // nil if the program has no jump table
	tramp   uintptr       // entry trampoline in em (0 ⇒ emitted lazily, uncached path)

	// mu serializes lazy block compilation (em append + cache.Put + dispatch
	// store) when the artifact is shared by concurrent invocations of the same
	// code. Already-compiled blocks execute lock-free.
	mu sync.Mutex
}

// buildCompiledProgram creates a fresh artifact with its own executable arena, a
// pre-emitted entry trampoline, and (if the program uses a jump table) djump
// support. Used both for the cached store and the uncached (zero-hash) path.
func buildCompiledProgram(program *PVM.Program) (*CompiledProgram, error) {
	em, err := NewExecutableMemory(0)
	if err != nil {
		return nil, err
	}
	tramp, err := emitEntryTrampolineInto(em)
	if err != nil {
		_ = em.Close()
		return nil, err
	}
	cache := NewCodeCache()
	cache.BindExecutableMemory(em)

	cp := &CompiledProgram{program: program, em: em, cache: cache, tramp: tramp}

	if jt := program.JumpTable; jt.Length > 0 && jt.Size > 0 && len(jt.Data) > 0 {
		d, err := buildDjumpSupport(em, program)
		if err != nil {
			_ = em.Close()
			return nil, err
		}
		cp.djump = d
	}
	return cp, nil
}

// newUncachedCompiledProgram wraps an existing executable arena (e.g. a test's
// JITContext em) without pre-emitting the trampoline or djump support; those are
// built lazily on first use (getTrampolineAddr / ensureDjumpSupport). Backs
// NewRecompiler for non-Psi_M callers.
func newUncachedCompiledProgram(program *PVM.Program, em *ExecutableMemory) *CompiledProgram {
	return &CompiledProgram{program: program, em: em, cache: NewCodeCache()}
}

// bindContext points a fresh JITContext at this artifact's executable arena,
// entry trampoline, and djump rodata/dispatch (stamped into the control region).
// Must run for every invocation, including cache hits that do no compilation.
func (cp *CompiledProgram) bindContext(ctx *JITContext) {
	ctx.SetExecutableMemory(cp.em) // also resets ctx.trampolineAddr to 0
	ctx.trampolineAddr = cp.tramp
	if cp.djump != nil {
		ctx.setDjumpPointers(cp.djump.tableAddr, cp.djump.bitmaskAddr, cp.djump.dispatchBase)
	}
}

// close releases the executable arena. Only for uncached artifacts (zero-hash);
// cached artifacts are never removed from the store — they live for the process
// lifetime (eviction is not implemented yet).
func (cp *CompiledProgram) close() {
	if cp != nil && cp.em != nil {
		_ = cp.em.Close()
		cp.em = nil
	}
}

func emitEntryTrampolineInto(em *ExecutableMemory) (uintptr, error) {
	a := asm.NewAssembler()
	EmitEntryTrampoline(a)
	code, err := a.Finalize()
	if err != nil {
		return 0, err
	}
	offset, err := em.Write(code)
	if err != nil {
		return 0, err
	}
	return em.GetPtr(offset), nil
}

// --- cross-invocation artifact store (keyed by CodeHash, single-flight) ---

type artifactEntry struct {
	ready chan struct{} // closed once cp/err are set
	cp    *CompiledProgram
	err   error
}

var artifactStore = struct {
	mu sync.Mutex
	m  map[types.OpaqueHash]*artifactEntry
}{m: make(map[types.OpaqueHash]*artifactEntry)}

// acquireCompiledProgram returns the cached compiled artifact for hash, building
// it at most once per distinct hash even under concurrent callers (single-flight).
// A zero hash is not cached: it returns a fresh artifact with cached=false and the
// caller must close() it. Cached artifacts are never evicted; they live for the
// process lifetime.
func acquireCompiledProgram(hash types.OpaqueHash, program *PVM.Program) (cp *CompiledProgram, cached bool, err error) {
	var zero types.OpaqueHash
	if hash == zero {
		cp, err = buildCompiledProgram(program)
		return cp, false, err
	}

	artifactStore.mu.Lock()
	if e, ok := artifactStore.m[hash]; ok {
		artifactStore.mu.Unlock()
		<-e.ready
		return e.cp, true, e.err
	}
	e := &artifactEntry{ready: make(chan struct{})}
	artifactStore.m[hash] = e
	artifactStore.mu.Unlock()

	e.cp, e.err = buildCompiledProgram(program)
	close(e.ready)
	if e.err != nil {
		// Don't keep a failed entry; allow a retry on the next call.
		artifactStore.mu.Lock()
		delete(artifactStore.m, hash)
		artifactStore.mu.Unlock()
	}
	return e.cp, true, e.err
}

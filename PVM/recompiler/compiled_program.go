//go:build linux && amd64 && cgo

package recompiler

import (
	"os"
	"strconv"
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
	hash    types.OpaqueHash
	program *PVM.Program
	em      *ExecutableMemory
	cache   *CodeCache
	djump   *djumpSupport // dispatch table (block chaining) + djump rodata; nil only for empty code
	tramp   uintptr       // entry trampoline in em (0 ⇒ emitted lazily, uncached path)

	// mu serializes lazy block compilation (em append + cache.Put + dispatch
	// store) when the artifact is shared by concurrent invocations of the same
	// code. Already-compiled blocks execute lock-free.
	mu sync.Mutex

	// refCount and lastUsed are guarded by programStore.mu. refCount is the number
	// of in-flight invocations using this artifact; eviction only reclaims
	// artifacts with refCount == 0, so an arena a goroutine is executing is never
	// Munmapped. lastUsed is the store sequence at the most recent acquire; the
	// reclaimable artifact with the smallest lastUsed is the LRU victim.
	refCount int32
	lastUsed uint64
}

// buildCompiledProgram creates a fresh artifact with its own executable arena, a
// pre-emitted entry trampoline, and djump support (PC→native dispatch for block
// chaining + jump-table rodata). Used both for the cached store and the uncached
// (zero-hash) path.
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

	if len(program.Bitmasks) > 0 {
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

// close releases the executable arena (Munmap). Called for uncached artifacts
// (zero-hash) after use, and by the store when it evicts a cached artifact.
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

// --- cross-invocation artifact store (keyed by CodeHash) ---
//
// Single-flight build (one compile per distinct hash, even concurrently) + LRU
// eviction with a soft, refcount-protected cap. Eviction triggers only on insert
// when over cap; victim selection is an O(n) scan (n <= cap, rare) for the
// reclaimable (refCount==0) artifact with the smallest lastUsed — no linked list,
// since the LRU is touched once per invocation, not per block.

const defaultMaxCachedPrograms = 256

type buildState struct {
	ready chan struct{} // closed once cp/err are set
	cp    *CompiledProgram
	err   error
}

type programStore struct {
	mu       sync.Mutex
	built    map[types.OpaqueHash]*CompiledProgram
	inflight map[types.OpaqueHash]*buildState
	seq      uint64 // monotonic LRU clock
	max      int    // soft cap on cached artifacts
}

var theProgramStore = newProgramStore()

func newProgramStore() *programStore {
	return &programStore{
		built:    make(map[types.OpaqueHash]*CompiledProgram),
		inflight: make(map[types.OpaqueHash]*buildState),
		max:      cacheMaxFromEnv(),
	}
}

// cacheMaxFromEnv reads JIT_CACHE_MAX_PROGRAMS (positive int) or falls back to
// the default. The cap is soft (in-use artifacts are never evicted).
func cacheMaxFromEnv() int {
	if v := os.Getenv("JIT_CACHE_MAX_PROGRAMS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultMaxCachedPrograms
}

// acquireCompiledProgram returns the cached compiled artifact for hash, building
// it at most once per distinct hash even under concurrent callers (single-flight),
// and increments its refcount (pair with releaseCompiledProgram). A zero hash is
// not cached: it returns a fresh artifact with cached=false and the caller must
// close() it.
func acquireCompiledProgram(hash types.OpaqueHash, program *PVM.Program) (cp *CompiledProgram, cached bool, err error) {
	var zero types.OpaqueHash
	if hash == zero {
		cp, err = buildCompiledProgram(program)
		return cp, false, err
	}
	cp, err = theProgramStore.acquire(hash, program)
	return cp, true, err
}

func (s *programStore) acquire(hash types.OpaqueHash, program *PVM.Program) (*CompiledProgram, error) {
	s.mu.Lock()
	if cp := s.built[hash]; cp != nil {
		cp.refCount++
		s.touchLocked(cp)
		s.mu.Unlock()
		return cp, nil
	}
	if bs := s.inflight[hash]; bs != nil {
		// Another goroutine is building this hash; wait, then retry the lookup
		// (it may have been evicted between the build finishing and our retry).
		s.mu.Unlock()
		<-bs.ready
		if bs.err != nil {
			return nil, bs.err
		}
		return s.acquire(hash, program)
	}
	bs := &buildState{ready: make(chan struct{})}
	s.inflight[hash] = bs
	s.mu.Unlock()

	cp, err := buildCompiledProgram(program) // expensive; outside the lock
	if cp != nil {
		cp.hash = hash
	}

	s.mu.Lock()
	delete(s.inflight, hash)
	bs.cp, bs.err = cp, err
	var evicted []*CompiledProgram
	if err == nil {
		cp.refCount = 1 // this acquire's reference
		s.touchLocked(cp)
		s.built[hash] = cp
		evicted = s.evictLocked()
	}
	s.mu.Unlock()
	close(bs.ready)

	for _, v := range evicted {
		v.close() // Munmap outside the lock; victims are already removed + refCount==0
	}
	return cp, err
}

// touchLocked marks cp as most-recently-used. Caller holds s.mu.
func (s *programStore) touchLocked(cp *CompiledProgram) {
	s.seq++
	cp.lastUsed = s.seq
}

// evictLocked removes least-recently-used, reclaimable (refCount==0) artifacts
// until the cache is within its cap, returning them to be close()d outside the
// lock. In-use artifacts are skipped (soft cap). Caller holds s.mu.
func (s *programStore) evictLocked() []*CompiledProgram {
	var evicted []*CompiledProgram
	for len(s.built) > s.max {
		var victim *CompiledProgram
		for _, cp := range s.built {
			if cp.refCount != 0 {
				continue
			}
			if victim == nil || cp.lastUsed < victim.lastUsed {
				victim = cp
			}
		}
		if victim == nil {
			break // every artifact is in use — exceed the cap rather than evict one
		}
		delete(s.built, victim.hash)
		evicted = append(evicted, victim)
	}
	return evicted
}

// releaseCompiledProgram drops one reference taken by acquireCompiledProgram.
// The artifact becomes evictable once its refcount reaches zero.
func releaseCompiledProgram(cp *CompiledProgram) {
	theProgramStore.release(cp)
}

func (s *programStore) release(cp *CompiledProgram) {
	if cp == nil {
		return
	}
	s.mu.Lock()
	if cp.refCount > 0 {
		cp.refCount--
	}
	s.mu.Unlock()
}

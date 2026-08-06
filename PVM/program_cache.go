package PVM

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Cross-invocation program cache, shared by both PVM backends.
//
// Deblobbing is identical baseline work for the interpreter and recompiler;
// service code might repeat across blocks, so caching the deblob'd *Program by CodeHash
// makes decode once-per-distinct-code instead of once-per-invocation. Both backends
// share this cache so interpreter-vs-recompiler comparisons stay fair: only the
// recompiler's extra native-code generation (cached separately) should differ.
//
// After decode, *Program is read-only (LookupBlock/BlockContaining are pure reads),
// so entries are safe to share across goroutines. Nothing is evicted.

type programCacheEntry struct {
	ready   chan struct{} // closed once program/reason are set (single-flight)
	program *Program
	reason  ExitReason
}

var programCache = struct {
	mu sync.Mutex
	m  map[types.OpaqueHash]*programCacheEntry
}{m: make(map[types.OpaqueHash]*programCacheEntry)}

// GetOrDeblobProgram returns the deblob'd program for programCode, caching by
// CodeHash across invocations. A zero hash bypasses the cache (e.g. is_authorized).
// Each distinct hash decodes at most once, even under concurrency (single-flight);
// failures are not cached and may be retried. The returned *Program is read-only.
//
// Only the entry-point-independent half of deblob is cached (𝔳_blob). 𝔳_inst for
// pc is checked on every call—the same code may be entered at different counters.
func GetOrDeblobProgram(hash types.OpaqueHash, programCode []byte, pc uint64) (*Program, ExitReason) {
	var zero types.OpaqueHash
	if hash == zero {
		p, reason := deblobValidatedProgram(programCode)
		if reason != ExitContinue {
			return nil, reason
		}
		return validEntry(&p, pc)
	}

	programCache.mu.Lock()
	if e, ok := programCache.m[hash]; ok {
		programCache.mu.Unlock()
		<-e.ready
		if e.reason != ExitContinue {
			return e.program, e.reason
		}
		return validEntry(e.program, pc)
	}
	e := &programCacheEntry{ready: make(chan struct{})}
	programCache.m[hash] = e
	programCache.mu.Unlock()

	p, reason := deblobValidatedProgram(programCode)
	if reason == ExitContinue {
		e.program = &p
	}
	e.reason = reason
	close(e.ready)

	if reason != ExitContinue {
		// Drop failed entries so the next call can retry.
		programCache.mu.Lock()
		delete(programCache.m, hash)
		programCache.mu.Unlock()
		return e.program, reason
	}
	return validEntry(e.program, pc)
}

// validEntry applies 𝔳_inst(c, k, ι) for pc on an already-decoded program.
// Invalid entry points return ExitPanic, matching DeBlobProgramCode.
func validEntry(p *Program, pc uint64) (*Program, ExitReason) {
	if !p.ValidInstructionAt(pc) {
		pvmLogger.Errorf("instruction counter %d is not a valid entry point", pc)
		return nil, ExitPanic
	}
	return p, ExitContinue
}

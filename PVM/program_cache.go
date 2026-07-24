package PVM

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Cross-invocation program cache — shared by BOTH PVM backends.
//
// Deblobbing a program (DeBlobProgramCode) is identical baseline work for the
// interpreter and the recompiler and currently runs once per invocation. The
// same service code recurs across blocks, so caching the deblob'd *Program by
// CodeHash makes it once-per-distinct-code. Caching it for both backends keeps
// the interpreter-vs-recompiler comparison fair: only the recompiler's extra
// per-invocation native-code generation (cached separately) should differ.
//
// The Program is fully precomputed by DeBlobProgramCode and only read afterwards
// (LookupBlock/BlockContaining are pure reads), so a single *Program is safe to
// share read-only across goroutines. Cached entries are not evicted.

type programCacheEntry struct {
	ready   chan struct{} // closed once program/reason are set (single-flight)
	program *Program
	reason  ExitReason
}

var programCache = struct {
	mu sync.Mutex
	m  map[types.OpaqueHash]*programCacheEntry
}{m: make(map[types.OpaqueHash]*programCacheEntry)}

// GetOrDeblobProgram returns the deblob'd program for programCode, caching it by
// hash (CodeHash) across invocations. A zero hash bypasses the cache (e.g.
// is_authorized). DeBlobProgramCode runs at most once per distinct hash even
// under concurrent callers (single-flight); a deblob failure is not cached so it
// can be retried. The returned *Program is read-only and safe to share.
func GetOrDeblobProgram(hash types.OpaqueHash, programCode []byte) (*Program, ExitReason) {
	var zero types.OpaqueHash
	if hash == zero {
		p, reason := DeBlobProgramCode(programCode)
		if reason != ExitContinue {
			return nil, reason
		}
		return &p, ExitContinue
	}

	programCache.mu.Lock()
	if e, ok := programCache.m[hash]; ok {
		programCache.mu.Unlock()
		<-e.ready
		return e.program, e.reason
	}
	e := &programCacheEntry{ready: make(chan struct{})}
	programCache.m[hash] = e
	programCache.mu.Unlock()

	p, reason := DeBlobProgramCode(programCode)
	if reason == ExitContinue {
		e.program = &p
	}
	e.reason = reason
	close(e.ready)

	if reason != ExitContinue {
		// Don't keep a failed entry; allow a retry on the next call.
		programCache.mu.Lock()
		delete(programCache.m, hash)
		programCache.mu.Unlock()
	}
	return e.program, reason
}

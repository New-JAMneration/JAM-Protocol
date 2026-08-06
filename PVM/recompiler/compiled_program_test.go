//go:build linux && amd64 && cgo

package recompiler

import (
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func storeTestHash(b byte) types.OpaqueHash {
	var h types.OpaqueHash
	h[0] = b
	return h
}

func newTestStore(max int) *programStore {
	return &programStore{
		built:    make(map[types.OpaqueHash]*CompiledProgram),
		inflight: make(map[types.OpaqueHash]*buildState),
		max:      max,
	}
}

// storeTestProgram returns a minimal valid *PVM.Program for building artifacts.
// The store keys by hash and never inspects the program's identity, so the same
// program can back several distinct cache entries.
func storeTestProgram(t *testing.T) *PVM.Program {
	t.Helper()
	instBytes, boundaries := buildMinimalInstr(1, PVM.GetOpcodeInfo(1)) // fallthrough
	blob := buildBlobExact(instBytes, boundaries)
	prog, exitReason := PVM.DeBlobProgramCode(blob, 0)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}
	return &prog
}

func TestProgramStoreEviction(t *testing.T) {
	prog := storeTestProgram(t)
	s := newTestStore(2)
	t.Cleanup(func() {
		for _, cp := range s.built {
			cp.close()
		}
	})

	h1, h2, h3 := storeTestHash(1), storeTestHash(2), storeTestHash(3)

	// Fill to cap, releasing each so it is reclaimable.
	cp1, err := s.acquire(h1, prog)
	if err != nil {
		t.Fatalf("acquire h1: %v", err)
	}
	s.release(cp1)
	cp2, err := s.acquire(h2, prog)
	if err != nil {
		t.Fatalf("acquire h2: %v", err)
	}
	s.release(cp2)
	if len(s.built) != 2 {
		t.Fatalf("want 2 cached, got %d", len(s.built))
	}

	// A 3rd insert over cap evicts the LRU (h1) and closes its arena.
	cp3, err := s.acquire(h3, prog)
	if err != nil {
		t.Fatalf("acquire h3: %v", err)
	}
	s.release(cp3)
	if len(s.built) != 2 {
		t.Fatalf("after evict want 2 cached, got %d", len(s.built))
	}
	if _, ok := s.built[h1]; ok {
		t.Fatal("h1 (LRU) should have been evicted")
	}
	if cp1.em != nil {
		t.Fatal("evicted artifact's arena should be closed (em == nil)")
	}
	if _, ok := s.built[h2]; !ok {
		t.Fatal("h2 should remain")
	}
	if _, ok := s.built[h3]; !ok {
		t.Fatal("h3 should remain")
	}

	// Re-acquiring an evicted hash rebuilds a fresh artifact.
	cp1b, err := s.acquire(h1, prog)
	if err != nil {
		t.Fatalf("re-acquire h1: %v", err)
	}
	if cp1b == cp1 {
		t.Fatal("re-acquire after eviction should build a new artifact")
	}
	s.release(cp1b)
}

func TestProgramStoreEvictionSkipsInUse(t *testing.T) {
	prog := storeTestProgram(t)
	s := newTestStore(1)
	t.Cleanup(func() {
		for _, cp := range s.built {
			cp.close()
		}
	})

	held := storeTestHash(1)

	// Acquire and HOLD a reference (do not release): refCount stays > 0.
	cpHeld, err := s.acquire(held, prog)
	if err != nil {
		t.Fatalf("acquire held: %v", err)
	}

	// Push several more distinct hashes over the cap of 1. Each is released, so
	// they are reclaimable, but the held one must never be evicted (soft cap).
	for b := byte(2); b <= 5; b++ {
		cp, err := s.acquire(storeTestHash(b), prog)
		if err != nil {
			t.Fatalf("acquire %d: %v", b, err)
		}
		s.release(cp)
	}

	if _, ok := s.built[held]; !ok {
		t.Fatal("in-use artifact (refCount > 0) must not be evicted (soft cap)")
	}
	if cpHeld.em == nil {
		t.Fatal("in-use artifact's arena must not be closed")
	}
	s.release(cpHeld)
}

package blockchain

import (
	"sync"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// makeValidatorSet builds a validator set of the configured size whose
// bandersnatch keys are derived from seeds salted by `salt`. Different salts
// therefore produce different (but individually valid) rings, letting us drive
// the cache's hit / rebuild / eviction paths deterministically.
func makeValidatorSet(t *testing.T, salt uint16) types.ValidatorsData {
	t.Helper()
	vs := make(types.ValidatorsData, types.ValidatorsCount)
	for i := range vs {
		seed := make([]byte, 32)
		seed[0] = byte(salt)
		seed[1] = byte(salt >> 8)
		seed[2] = byte(i)
		seed[3] = byte(i >> 8)
		pub, err := vrf.GetPublicKeyFromSecret(seed)
		if err != nil {
			t.Fatalf("derive bandersnatch key (salt=%d i=%d): %v", salt, i, err)
		}
		copy(vs[i].Bandersnatch[:], pub)
	}
	return vs
}

func cacheLen() int {
	cache.RLock()
	defer cache.RUnlock()
	return len(cache.verifiers)
}

// TestGetVerifierWrongSizeErrors covers the guard against a mismatched ring
// size; it does not touch the cache or the FFI.
func TestGetVerifierWrongSizeErrors(t *testing.T) {
	ClearVerifierCache()
	if _, err := GetVerifier(types.ValidatorsData{}); err == nil {
		t.Fatal("expected an error for a validator set of the wrong size")
	}
	if cacheLen() != 0 {
		t.Fatalf("expected empty cache, got %d entries", cacheLen())
	}
}

// TestGetVerifierCacheHit verifies that the same validator set returns the very
// same cached verifier and does not add a second entry.
func TestGetVerifierCacheHit(t *testing.T) {
	ClearVerifierCache()
	vs := makeValidatorSet(t, 1)

	first, err := GetVerifier(vs)
	if err != nil {
		t.Fatalf("first GetVerifier: %v", err)
	}
	second, err := GetVerifier(vs)
	if err != nil {
		t.Fatalf("second GetVerifier: %v", err)
	}
	if first != second {
		t.Fatal("expected a cache hit to return the same verifier pointer")
	}
	if cacheLen() != 1 {
		t.Fatalf("expected 1 cached entry, got %d", cacheLen())
	}
}

// TestGetVerifierDistinctSetsRebuild verifies that a different validator set
// misses the cache and produces a distinct verifier.
func TestGetVerifierDistinctSetsRebuild(t *testing.T) {
	ClearVerifierCache()
	a := makeValidatorSet(t, 1)
	b := makeValidatorSet(t, 2)

	va, err := GetVerifier(a)
	if err != nil {
		t.Fatalf("GetVerifier(a): %v", err)
	}
	vb, err := GetVerifier(b)
	if err != nil {
		t.Fatalf("GetVerifier(b): %v", err)
	}
	if va == vb {
		t.Fatal("expected distinct validator sets to yield distinct verifiers")
	}
	if cacheLen() != 2 {
		t.Fatalf("expected 2 cached entries, got %d", cacheLen())
	}
}

// TestGetVerifierEvictionBoundedByMax feeds more distinct validator sets than
// the cache bound and asserts the cache never exceeds it. Building many rings is
// only cheap in tiny mode, so the test is skipped for large validator counts.
func TestGetVerifierEvictionBoundedByMax(t *testing.T) {
	if types.ValidatorsCount > 16 {
		t.Skipf("skipping eviction test for validator count %d (too expensive)", types.ValidatorsCount)
	}
	ClearVerifierCache()

	for salt := uint16(1); salt <= ringVerifierCacheMax+5; salt++ {
		if _, err := GetVerifier(makeValidatorSet(t, salt)); err != nil {
			t.Fatalf("GetVerifier(salt=%d): %v", salt, err)
		}
		if got := cacheLen(); got > ringVerifierCacheMax {
			t.Fatalf("cache size %d exceeded bound %d", got, ringVerifierCacheMax)
		}
	}
}

// TestClearVerifierCache verifies the cache is emptied after a clear.
func TestClearVerifierCache(t *testing.T) {
	ClearVerifierCache()
	if _, err := GetVerifier(makeValidatorSet(t, 1)); err != nil {
		t.Fatalf("GetVerifier: %v", err)
	}
	ClearVerifierCache()
	if cacheLen() != 0 {
		t.Fatalf("expected empty cache after clear, got %d", cacheLen())
	}
}

// TestGetVerifierConcurrent exercises the RWMutex path under the race detector
// with a mix of shared and distinct validator sets.
func TestGetVerifierConcurrent(t *testing.T) {
	ClearVerifierCache()
	sets := []types.ValidatorsData{
		makeValidatorSet(t, 1),
		makeValidatorSet(t, 2),
		makeValidatorSet(t, 3),
	}

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 32; i++ {
				if _, err := GetVerifier(sets[(g+i)%len(sets)]); err != nil {
					t.Errorf("concurrent GetVerifier: %v", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()

	if got := cacheLen(); got != len(sets) {
		t.Fatalf("expected %d cached entries, got %d", len(sets), got)
	}
}

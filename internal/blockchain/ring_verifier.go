package blockchain

import (
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"

	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// ringVerifierCacheMax bounds the number of distinct ring verifiers kept in
// memory. Each entry holds a Rust-side verifier that retains the whole ring of
// bandersnatch keys (tens of KB in full mode), so the bound is deliberately
// small; the number of validator sets in flight at once is only a handful.
const ringVerifierCacheMax = 32

// ringVerifierCache maps a hash of the validator set's bandersnatch keys to its
// ring verifier. Keying on the key-set content (rather than the epoch) lets
// forks that share a validator set reuse the same verifier, so a fork-restore
// no longer has to drop the cache and pay for a rebuild.
type ringVerifierCache struct {
	sync.RWMutex
	verifiers map[types.OpaqueHash]*vrf.Verifier
}

var cache = &ringVerifierCache{
	verifiers: make(map[types.OpaqueHash]*vrf.Verifier, ringVerifierCacheMax),
}

// ClearVerifierCache drops all cached verifiers (used to reset state between
// fuzz runs). It only releases the Go references; see GetVerifier for why the
// underlying Rust verifiers must not be Free()d here.
func ClearVerifierCache() {
	cache.Lock()
	defer cache.Unlock()
	cache.verifiers = make(map[types.OpaqueHash]*vrf.Verifier, ringVerifierCacheMax)
}

// GetVerifier returns the ring verifier for the given validator set, building
// and caching it on a miss. The cache key is a hash of the concatenated
// bandersnatch public keys, which alone determine the verifier: the ring
// commitment is a pure function of those keys (GP 0.8.0 eq 6.13, z = O([k_b])),
// so the epoch is irrelevant and forks sharing a validator set hit the cache.
//
// A cached verifier is handed out under a read lock, so a concurrent goroutine
// may still hold the pointer when its entry is later evicted. We therefore
// never call Verifier.Free() on eviction; dropping the Go reference lets the GC
// finalizer (registered in vrf.NewVerifier) reclaim the Rust heap once no
// goroutine holds the pointer anymore.
func GetVerifier(gammaK types.ValidatorsData) (*vrf.Verifier, error) {
	if len(gammaK) != types.ValidatorsCount {
		return nil, fmt.Errorf("gammaK size %d is not equal to validators count %d", len(gammaK), types.ValidatorsCount)
	}

	// Concatenate the ring's bandersnatch keys; this is both the cache key
	// material and the input to the verifier builder on a miss.
	keySize := len(types.BandersnatchPublic{})
	ring := make([]byte, 0, len(gammaK)*keySize)
	for _, v := range gammaK {
		ring = append(ring, v.Bandersnatch[:]...)
	}
	key := hash.Blake2bHash(ring)

	// Fast path: read lock.
	cache.RLock()
	if v, ok := cache.verifiers[key]; ok {
		cache.RUnlock()
		return v, nil
	}
	cache.RUnlock()

	// Slow path: write lock.
	cache.Lock()
	defer cache.Unlock()

	// Re-check after acquiring the write lock.
	if v, ok := cache.verifiers[key]; ok {
		return v, nil
	}

	ringVerifier, err := vrf.NewVerifier(ring, uint(len(gammaK)))
	if err != nil {
		return nil, fmt.Errorf("failed to create ring verifier: %w", err)
	}

	// Evict everything when at capacity. Overflow is not expected (the working
	// set of validator sets is tiny), so a simple clear-all keeps the common
	// path free of bookkeeping; dropped references are reclaimed by the finalizer.
	if len(cache.verifiers) >= ringVerifierCacheMax {
		cache.verifiers = make(map[types.OpaqueHash]*vrf.Verifier, ringVerifierCacheMax)
	}
	cache.verifiers[key] = ringVerifier
	return ringVerifier, nil
}

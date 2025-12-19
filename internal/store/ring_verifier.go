package store

import (
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"

	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

type ringVerifierCache struct {
	sync.RWMutex
	epoch types.TimeSlot
	*vrf.Verifier
}

var cache = &ringVerifierCache{}

// ClearVerifierCache clears all cached verifiers (for testing or cleanup)
func ClearVerifierCache() {
	cache.Lock()
	defer cache.Unlock()
	cache.Free()
}

func (c *ringVerifierCache) Free() {
	if c.Verifier != nil {
		c.Verifier.Free()
		c.Verifier = nil
	}
	c.epoch = 0
}

func GetVerifier(epoch types.TimeSlot, gammaK types.ValidatorsData) (*vrf.Verifier, error) {
	if len(gammaK) != types.ValidatorsCount {
		return nil, fmt.Errorf("gammaK size %d is not equal to validators count %d", len(gammaK), types.ValidatorsCount)
	}

	// First path: read lock
	// Try to get both cached verifiers
	cache.RLock()
	if cache.epoch == epoch && cache.Verifier != nil {
		cache.RUnlock()
		return cache.Verifier, nil
	}
	cache.RUnlock()

	// Second path: write lock
	cache.Lock()
	defer cache.Unlock()

	// Double check
	if cache.epoch == epoch && cache.Verifier != nil {
		return cache.Verifier, nil
	}

	// epoch transition or not initialized: free old verifiers
	cache.Free()

	// create ring verifier
	keySize := len(types.BandersnatchPublic{})
	ring := make([]byte, 0, len(gammaK)*keySize)
	for _, v := range gammaK {
		ring = append(ring, v.Bandersnatch[:]...)
	}

	ringVerifier, err := vrf.NewVerifier(ring, uint(len(gammaK)))
	if err != nil {
		return nil, fmt.Errorf("failed to create ring verifier: %w", err)
	}

	// update cache and return
	cache.epoch = epoch
	cache.Verifier = ringVerifier
	return ringVerifier, nil
}

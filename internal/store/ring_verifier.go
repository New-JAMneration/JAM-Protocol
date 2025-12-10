package store

import (
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"

	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// TODO: rename gammaKVerifier to ringVerifier and remove kappaVerifier
type RingVerifier struct {
	mu             sync.RWMutex
	CurrentEpoch   types.TimeSlot
	gammaKVerifier *vrf.Verifier
	kappaVerifier  *vrf.Verifier
}

// Global verifier cache (similar to globalStore pattern)
var ringVerifierCache = &RingVerifier{
	CurrentEpoch: types.TimeSlot(0),
}

// ClearVerifierCache clears all cached verifiers (for testing or cleanup)
func ClearVerifierCache() {
	ringVerifierCache.mu.Lock()
	defer ringVerifierCache.mu.Unlock()
	if ringVerifierCache.gammaKVerifier != nil {
		ringVerifierCache.gammaKVerifier.Free()
		ringVerifierCache.gammaKVerifier = nil
	}
	if ringVerifierCache.kappaVerifier != nil {
		ringVerifierCache.kappaVerifier.Free()
		ringVerifierCache.kappaVerifier = nil
	}
	ringVerifierCache.CurrentEpoch = types.TimeSlot(0)
}

func GetVerifiers(epoch types.TimeSlot, gammaK, kappa types.ValidatorsData) (*vrf.Verifier, *vrf.Verifier, error) {
	ringVerifierCache.mu.Lock()
	defer ringVerifierCache.mu.Unlock()

	// Try to get both cached verifiers
	if ringVerifierCache.CurrentEpoch == epoch &&
		ringVerifierCache.gammaKVerifier != nil &&
		ringVerifierCache.kappaVerifier != nil {
		return ringVerifierCache.gammaKVerifier, ringVerifierCache.kappaVerifier, nil
	}

	// epoch transition or not initialized: free old verifiers
	if ringVerifierCache.gammaKVerifier != nil {
		ringVerifierCache.gammaKVerifier.Free()
		ringVerifierCache.gammaKVerifier = nil
	}
	if ringVerifierCache.kappaVerifier != nil {
		ringVerifierCache.kappaVerifier.Free()
		ringVerifierCache.kappaVerifier = nil
	}

	// create gammaK verifier
	if len(gammaK) != types.ValidatorsCount {
		return nil, nil, fmt.Errorf("gammaK size %v is not equal to validators count %v", len(gammaK), types.ValidatorsCount)
	}
	gammaKRing := []byte{}
	for _, v := range gammaK {
		gammaKRing = append(gammaKRing, []byte(v.Bandersnatch[:])...)
	}
	gammaKVerifier, err := vrf.NewVerifier(gammaKRing, uint(len(gammaK)))
	if err != nil {
		return nil, nil, err
	}

	// create kappa verifier
	kappaRing := []byte{}
	for _, v := range kappa {
		kappaRing = append(kappaRing, []byte(v.Bandersnatch[:])...)
	}
	kappaVerifier, err := vrf.NewVerifier(kappaRing, uint(len(kappa)))
	if err != nil {
		gammaKVerifier.Free()
		return nil, nil, err
	}

	// update cache and return
	ringVerifierCache.CurrentEpoch = epoch
	ringVerifierCache.gammaKVerifier = gammaKVerifier
	ringVerifierCache.kappaVerifier = kappaVerifier
	return gammaKVerifier, kappaVerifier, nil
}

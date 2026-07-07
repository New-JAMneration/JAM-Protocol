package blockchain

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamhash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

func testGammaK(seed byte) types.ValidatorsData {
	gammaK := make(types.ValidatorsData, types.ValidatorsCount)
	for i := range gammaK {
		b := byte(i) + seed
		gammaK[i].Bandersnatch[0] = b
		gammaK[i].Ed25519[0] = b + 1
		gammaK[i].Bls[0] = b + 2
		gammaK[i].Metadata[0] = b + 3
	}
	return gammaK
}

func TestRingVerifierCacheKeyIncludesEpochAndGammaKContents(t *testing.T) {
	gammaK := testGammaK(1)

	key, err := newRingVerifierCacheKey(42, gammaK)
	if err != nil {
		t.Fatalf("newRingVerifierCacheKey failed: %v", err)
	}

	sameKey, err := newRingVerifierCacheKey(42, gammaK)
	if err != nil {
		t.Fatalf("newRingVerifierCacheKey failed: %v", err)
	}
	if key != sameKey {
		t.Fatalf("same epoch and gammaK produced different keys")
	}

	differentEpochKey, err := newRingVerifierCacheKey(43, gammaK)
	if err != nil {
		t.Fatalf("newRingVerifierCacheKey failed: %v", err)
	}
	if key == differentEpochKey {
		t.Fatalf("different epoch produced the same key")
	}

	differentGammaK := append(types.ValidatorsData(nil), gammaK...)
	differentGammaK[0].Metadata[0] ^= 0xff
	differentGammaKKey, err := newRingVerifierCacheKey(42, differentGammaK)
	if err != nil {
		t.Fatalf("newRingVerifierCacheKey failed: %v", err)
	}
	if key == differentGammaKKey {
		t.Fatalf("different gammaK contents produced the same key")
	}
}

func TestRestoreWithStateKeepsVerifierCache(t *testing.T) {
	ClearVerifierCache()
	t.Cleanup(ClearVerifierCache)

	gammaK := testGammaK(2)
	key, err := newRingVerifierCacheKey(7, gammaK)
	if err != nil {
		t.Fatalf("newRingVerifierCacheKey failed: %v", err)
	}
	verifier := &vrf.Verifier{}

	cache.Lock()
	cache.key = key
	cache.Verifier = verifier
	cache.Unlock()

	cs := newChainState()
	block := types.Block{Header: types.Header{Slot: 1}}
	headerHash, err := jamhash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		t.Fatalf("ComputeBlockHeaderHash failed: %v", err)
	}

	if err := cs.restoreWithState(headerHash, block, types.State{}, nil); err != nil {
		t.Fatalf("restoreWithState failed: %v", err)
	}

	cache.RLock()
	defer cache.RUnlock()
	if cache.key != key {
		t.Fatalf("restoreWithState changed verifier cache key")
	}
	if cache.Verifier != verifier {
		t.Fatalf("restoreWithState cleared verifier cache")
	}
}

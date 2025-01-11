package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// This is borne out with V = 1, 023 validators and C = 341 cores.
var (
	V = types.ValidatorsCount
	C = types.CoresCount
	E = types.EpochLength
	R = 10
)

// GuranatorAssignments is a struct that contains a slice of CoreIndex and Ed25519Public
// (11.18) G ∈ (⟦N_C⟧N_V , ⟦H_K ⟧N_V )
type GuranatorAssignments struct {
	CoreAssignments []types.CoreIndex
	PublicKeys      []types.Ed25519Public
}

// (11.19) R(c, n) = [(x + n) mod C | x ∈ c]
func rotateCores(in []types.U32, n types.U32) []types.U32 {
	out := make([]types.U32, len(in))
	for i, x := range in {
		out[i] = (x + n) % types.U32(C)
	}
	return out
}

// (11.20)
func permute(e types.Entropy, currentSlot types.TimeSlot) []types.CoreIndex {
	base := make([]types.U32, V)
	for i := 0; i < V; i++ {
		c := (C * uint32(i)) / V
		base[i] = types.U32(c)
	}

	shuffled := shuffle.Shuffle(base, types.OpaqueHash(e))

	subEpoch := (uint32(currentSlot) % E) / R

	// R(...) call
	rotatedU32 := rotateCores(shuffled, types.U32(subEpoch))

	// Convert back to []types.CoreIndex
	rotated := make([]types.CoreIndex, len(rotatedU32))
	for i, v := range rotatedU32 {
		rotated[i] = types.CoreIndex(v)
	}
	return rotated
}

// (11.21) G(e, t, k) = (P(e, t), H_K)
func NewGuranatorAssignments(
	epochEntropy types.Entropy,
	currentSlot types.TimeSlot,
	validators types.ValidatorsData,
) GuranatorAssignments {

	// 1. get the core assignments
	coreAssignments := permute(epochEntropy, currentSlot)

	// 2. get the public keys
	result := safrole.ReplaceOffenderKeys(validators)
	pubKeys := make([]types.Ed25519Public, len(result))

	for i, v := range result {
		pubKeys[i] = v.Ed25519
	}

	return GuranatorAssignments{
		CoreAssignments: coreAssignments,
		PublicKeys:      pubKeys,
	}
}

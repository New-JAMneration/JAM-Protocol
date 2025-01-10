package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// This is borne out with V = 1, 023 validators and C = 341 cores.
const (
	V = 1023
	C = 341
)

// GuranatorAssignments is a struct that contains a slice of CoreIndex and Ed25519Public
// (11.18) G ∈ (⟦N_C⟧N_V , ⟦H_K ⟧N_V )
type GuranatorAssignments struct {
	CoreAssignments []types.CoreIndex
	PublicKeys      []types.Ed25519Public
}

// (11.19) R(c, n) = [(x + n) mod C | x ∈ c]
// U16 -> coreIndex is U16
func rotateCores(in []types.U16, n types.U16) []types.U16 {
	out := make([]types.U16, len(in))
	for i, x := range in {
		out[i] = (x + n) % types.U16(C)
	}
	return out
}

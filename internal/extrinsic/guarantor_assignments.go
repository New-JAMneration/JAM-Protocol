package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	ReportsErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// This is borne out with V = 1, 023 validators and C = 341 cores.
/*
var (
	V = types.ValidatorsCount
	C = types.CoresCount
	E = types.EpochLength
	R = types.RotationPeriod
)
*/
// GuranatorAssignments is a struct that contains a slice of CoreIndex and Ed25519Public
// (11.18) G ∈ (⟦N_C⟧N_V , ⟦H_K ⟧N_V )
type GuranatorAssignments struct {
	CoreAssignments []types.CoreIndex
	PublicKeys      []types.Validator
}

// (11.19) R(c, n) = [(x + n) mod C | x ∈ c]
func rotateCores(in []types.U32, n types.U32) []types.U32 {
	out := make([]types.U32, len(in))
	for i, x := range in {
		out[i] = (x + n) % types.U32(types.CoresCount)
	}
	return out
}

// (11.20)
func permute(e types.Entropy, currentSlot types.TimeSlot) []types.CoreIndex {
	base := make([]types.U32, types.ValidatorsCount)
	for i := 0; i < types.ValidatorsCount; i++ {
		c := (types.CoresCount * i) / types.ValidatorsCount
		base[i] = types.U32(c)
	}

	shuffled := shuffle.Shuffle(base, types.OpaqueHash(e))

	subEpoch := (int(currentSlot) % types.EpochLength) / types.RotationPeriod

	// R(...) call
	rotatedU32 := rotateCores(shuffled, types.U32(subEpoch))

	// Convert back to []types.CoreIndex
	rotated := make([]types.CoreIndex, len(rotatedU32))
	for i, v := range rotatedU32 {
		rotated[i] = types.CoreIndex(v)
	}
	return rotated
}

func NewGuranatorAssignments(
	epochEntropy types.Entropy,
	currentSlot types.TimeSlot,
	validators types.ValidatorsData,
) GuranatorAssignments {
	// 1. get the core assignments
	coreAssignments := permute(epochEntropy, currentSlot)
	// 2. get the public keys
	result := safrole.ReplaceOffenderKeys(validators)
	pubKeys := make([]types.Validator, len(result))

	for i, v := range result {
		pubKeys[i].Ed25519 = v.Ed25519
		pubKeys[i].Bandersnatch = v.Bandersnatch
		pubKeys[i].Bls = v.Bls
		pubKeys[i].Metadata = v.Metadata
	}

	return GuranatorAssignments{
		CoreAssignments: coreAssignments,
		PublicKeys:      pubKeys,
	}
}

// (11.21) G(e, t, k) = (P(e, t), H_K)
// G ≡ (P (η′2, τ ′), Φ(κ′))
func GFunc(offendersMap map[types.Ed25519Public]bool) (GuranatorAssignments, error) {
	state := store.GetInstance().GetPosteriorStates()
	etaPrime := state.GetEta()

	// (η′2, κ′)
	e := etaPrime[2]
	validators := state.GetKappa()

	for _, validator := range validators {
		if _, offenderExists := offendersMap[validator.Ed25519]; offenderExists {
			err := ReportsErrorCode.BannedValidator
			return GuranatorAssignments{}, &err
		}
	}

	return NewGuranatorAssignments(e, state.GetTau(), validators), nil
}

// (11.22) G∗ ≡ (P (e, τ ′ − R), Φ(k))
func GStarFunc(offendersMap map[types.Ed25519Public]bool) (GuranatorAssignments, error) {
	state := store.GetInstance().GetPosteriorStates()
	var e types.Entropy
	var validators types.ValidatorsData

	etaPrime := state.GetEta()
	if (int(state.GetTau())-types.RotationPeriod)/types.EpochLength == int(state.GetTau())/types.EpochLength {
		// (η′2, κ′)
		e = etaPrime[2]
		validators = state.GetKappa()
		for _, validator := range validators {
			if _, offenderExists := offendersMap[validator.Ed25519]; offenderExists {
				err := ReportsErrorCode.BannedValidator
				return GuranatorAssignments{}, &err
			}
		}
	} else {
		// (η′3, λ′)
		e = etaPrime[3]
		validators = state.GetLambda()
		for _, validator := range validators {
			if _, offenderExists := offendersMap[validator.Ed25519]; offenderExists {
				err := ReportsErrorCode.BannedValidator
				return GuranatorAssignments{}, &err
			}
		}
	}

	return NewGuranatorAssignments(e, state.GetTau()-types.TimeSlot(types.RotationPeriod), validators), nil
}

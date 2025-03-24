package accumulation

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

/*
This section implements the preimage accumulation logic:
For all upload preimage_extrinsics, we should check available preimages
After that, adding those preimages into delta anchored with time for each service id

// Pathway: Get Ep, sort and remove duplicated
// -> Eps
// -> compared with elements in delta
// -> form P
// -> for service id, adding preimages and tauPrime in delta preimage lookup
// -> set delta^daggerdagger to delta^prime
*/

// v0.6.4 (12.36) R function determines whether a preimage should be integrated
func ShouldIntegratePreimage(d types.ServiceAccountState, s types.ServiceId, h types.OpaqueHash, l types.U32) bool {
	// Check if the preimage hash is not in the service account's preimage map
	account, isInAccount := d[s]
	if !isInAccount {
		// If account does not exist in service, return false
		return false
	}
	_, isInPreimageMap := account.PreimageLookup[h]

	// Check if the preimage hash and length's time slot set is empty
	lookupKey := types.LookupMetaMapkey{
		Hash:   h,
		Length: l,
	}
	timeSlotSet, isInLookupDict := account.LookupDict[lookupKey]

	// Condition: hash does not exist in preimage map, and lookup time slot set is empty
	return !isInPreimageMap && (!isInLookupDict || len(timeSlotSet) == 0)
}

// IntegratePreimage filters the preimage extrinsics and returns only those that should be integrated
// v0.6.4 (12.37): ∀(s, p) ∈ EP ∶ R(δ, s, H(p), |p|)
func IntegratePreimage(eps types.PreimagesExtrinsic, d types.ServiceAccountState) types.PreimagesExtrinsic {
	j := 0
	for i, ep := range eps {
		// Calculate preimage hash and length
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		// Check if the preimage should be integrated
		if ShouldIntegratePreimage(d, ep.Requester, preimageHash, preimageLength) {
			eps[j] = eps[i]
			j++
		}
	}
	eps = eps[:j]
	return eps
}

// v0.6.4 (12.38) P = {(s, p) | (s, p)∈ EP , R(δ‡, s, H(p), |p|)}
// FilterPreimageExtrinsics filtered extrinsics that should be integrated
func FilterPreimageExtrinsics(eps types.PreimagesExtrinsic, deltaDoubleDagger types.ServiceAccountState) types.PreimagesExtrinsic {
	filteredEps := IntegratePreimage(eps, deltaDoubleDagger)
	// TODO: add filter (sorted and then remove duplicated)
	return filteredEps
}

// UpdateDeltaWithExtrinsicPreimage updates the deltaDoubleDagger state with filtered preimages
// It integrates preimages into deltaDoubleDagger using the provided tauPrime time slot
// v0.6.4 (12.39)
func UpdateDeltaWithExtrinsicPreimage(eps types.PreimagesExtrinsic, deltaDoubleDagger types.ServiceAccountState, tauPrime types.TimeSlot) types.ServiceAccountState {
	for _, ep := range eps {
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))
		lookupKey := types.LookupMetaMapkey{
			Hash:   preimageHash,
			Length: preimageLength,
		}
		deltaDoubleDagger[ep.Requester].LookupDict[lookupKey] = types.TimeSlotSet{tauPrime}
		deltaDoubleDagger[ep.Requester].PreimageLookup[preimageHash] = ep.Blob
	}

	return deltaDoubleDagger
}

// ProcessPreimageExtrinsics is the main unified function for handling preimage extrinsics
// It combines filtering and delta state updates in a single call for external use
// v0.6.4 (12.38-12.39)
func ProcessPreimageExtrinsics() {
	// Get store instance and required states
	s := store.GetInstance()
	eps := s.GetProcessingBlockPointer().GetPreimagesExtrinsic()
	deltaDoubleDagger := s.GetIntermediateStates().GetDeltaDoubleDagger()
	tauPrime := s.GetPosteriorStates().GetTau()

	// Filter preimage extrinsics
	filteredEps := FilterPreimageExtrinsics(eps, deltaDoubleDagger)

	// Update deltaDoubleDagger with filtered preimages
	newDeltaDoubleDagger := UpdateDeltaWithExtrinsicPreimage(filteredEps, deltaDoubleDagger, tauPrime)

	// Update new double-dagger to posterior state
	s.GetPosteriorStates().SetDelta(newDeltaDoubleDagger)
}

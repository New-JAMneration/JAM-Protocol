package accumulation

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"sort"

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
	// Check for existence of the service account
	account, isInAccount := d[s]
	if !isInAccount || account.PreimageLookup == nil || account.LookupDict == nil {
		// If account does not exist or maps are uninitialized, return false
		return false
	}

	// Check if the preimage hash is not in the service account's preimage map
	_, isInPreimageMap := account.PreimageLookup[h]

	// Construct lookup key
	lookupKey := types.LookupMetaMapkey{
		Hash:   h,
		Length: l,
	}

	// Check if the preimage hash and length's time slot set is empty
	timeSlotSet, exists := account.LookupDict[lookupKey]
	if !exists {
		// If lookup key doesn't exist in the dictionary, consider the time slot set as empty
		return false
	}

	// Condition: hash does not exist in preimage map, and lookup time slot set is empty
	return !isInPreimageMap && (len(timeSlotSet) == 0)
}

// IntegratePreimage filters the preimage extrinsics and returns only those that should be integrated
// v0.6.4 (12.37): ∀(s, p) ∈ EP ∶ R(δ, s, H(p), |p|)
func IntegratePreimage(eps types.PreimagesExtrinsic, d types.ServiceAccountState) (types.PreimagesExtrinsic, error) {
	j := 0
	for i, ep := range eps {
		// Calculate preimage hash and length
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		// Check if the preimage should be integrated
		if ShouldIntegratePreimage(d, ep.Requester, preimageHash, preimageLength) {
			eps[j] = eps[i]
			j++
		} else if !ShouldIntegratePreimage(d, ep.Requester, preimageHash, preimageLength) {
			return nil, errors.New("preimage is not solicited")
		}
	}
	eps = eps[:j]
	return eps, nil
}

// v0.6.4 (12.38) P = {(s, p) | (s, p)∈ EP , R(δ‡, s, H(p), |p|)}
// FilterPreimageExtrinsics filtered extrinsics that should be integrated
func FilterPreimageExtrinsics(eps types.PreimagesExtrinsic, deltaDoubleDagger types.ServiceAccountState) (types.PreimagesExtrinsic, error) {
	// If eps is empty, return empty slice
	if len(eps) == 0 {
		log.Printf("Nothing is provided in Eps")
		return eps, nil
	}
	originEps := make(types.PreimagesExtrinsic, len(eps))
	copy(originEps, eps)
	sort.Sort(&eps)

	if !reflect.DeepEqual(eps, originEps) {
		return nil, errors.New("eps is not sorted")
	}
	// // Then, remove the duplicates
	// j := 0
	// for i := 1; i < len(eps); i++ {
	// 	if eps[i].Requester != eps[j].Requester || !bytes.Equal(eps[i].Blob, eps[j].Blob) {
	// 		j++
	// 		eps[j] = eps[i]
	// 	}
	// }
	// eps = eps[:j+1]
	// If eps have duplicates, return error
	for i := 1; i < len(eps); i++ {
		if eps[i].Requester == eps[i-1].Requester && bytes.Equal(eps[i].Blob, eps[i-1].Blob) {
			return nil, errors.New("eps have duplicates")
		}
	}

	filteredEps, err := IntegratePreimage(eps, deltaDoubleDagger)
	if err != nil {
		return nil, err
	}
	return filteredEps, nil
}

// UpdateDeltaWithExtrinsicPreimage updates the deltaDoubleDagger state with filtered preimages
// It integrates preimages into deltaDoubleDagger using the provided tauPrime time slot
// v0.6.4 (12.39)
func UpdateDeltaWithExtrinsicPreimage(eps types.PreimagesExtrinsic, deltaDoubleDagger types.ServiceAccountState, tauPrime types.TimeSlot) (types.ServiceAccountState, error) {
	for _, ep := range eps {
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))
		lookupKey := types.LookupMetaMapkey{
			Hash:   preimageHash,
			Length: preimageLength,
		}

		// Check if ServiceId exists in deltaDoubleDagger
		serviceAccount, exists := deltaDoubleDagger[ep.Requester]
		if !exists {
			return nil, errors.New("service account not found")
		} else {
			// Ensure map fields are initialized
			if serviceAccount.LookupDict == nil {
				return nil, errors.New("lookupDict not initialized")
			}
			if serviceAccount.PreimageLookup == nil {
				return nil, errors.New("preimageLookup not initialized")
			}
		}

		// Update map
		serviceAccount.LookupDict[lookupKey] = types.TimeSlotSet{tauPrime}
		serviceAccount.PreimageLookup[preimageHash] = ep.Blob

		// Write updated serviceAccount back to deltaDoubleDagger
		deltaDoubleDagger[ep.Requester] = serviceAccount
	}

	return deltaDoubleDagger, nil
}

// ProcessPreimageExtrinsics is the main unified function for handling preimage extrinsics
// It combines filtering and delta state updates in a single call for external use
// v0.6.4 (12.38-12.39)
func ProcessPreimageExtrinsics() error {
	// Get store instance and required states
	s := store.GetInstance()
	eps := s.GetProcessingBlockPointer().GetPreimagesExtrinsic()
	deltaDoubleDagger := s.GetIntermediateStates().GetDeltaDoubleDagger()
	tauPrime := s.GetPosteriorStates().GetTau()

	// Filter preimage extrinsics
	filteredEps, err := FilterPreimageExtrinsics(eps, deltaDoubleDagger)
	if err != nil {
		return err
	}

	// Update deltaDoubleDagger with filtered preimages
	newDeltaDoubleDagger, err := UpdateDeltaWithExtrinsicPreimage(filteredEps, deltaDoubleDagger, tauPrime)
	if err != nil {
		return err
	}

	// Update new double-dagger to posterior state
	s.GetPosteriorStates().SetDelta(newDeltaDoubleDagger)
	return nil
}

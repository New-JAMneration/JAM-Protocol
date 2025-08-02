package accumulation

import (
	"bytes"
	"errors"
	"log"
	"maps"

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

	// If eps is not sorted, return error
	for i := 1; i < len(eps); i++ {
		if eps[i-1].Requester > eps[i].Requester {
			return nil, errors.New("eps is not sorted by Requester")
		}

		if eps[i-1].Requester == eps[i].Requester && bytes.Compare(eps[i-1].Blob, eps[i].Blob) > 0 {
			return nil, errors.New("eps.Requester is not sorted by Blob")
		}
	}

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
	eps := s.GetLatestBlock().Extrinsic.Preimages
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

// Provide is the preimage integration function (different from IntegratePreimage despite re-using the word "integrate")
// It transforms a dictionary of service states and a set of service/hash pairs into a new dictionary of service states.
// (map[N_s]A, (N_s, Y)) -> map[N_s]A
// v0.6.5 (12.18)
func Provide(d types.ServiceAccountState, eps types.ServiceBlobs) (types.ServiceAccountState, error) {
	dPrime := maps.Clone(d)

	for _, serviceblob := range eps {
		serviceId := serviceblob.ServiceID
		serviceAccount, found := d[serviceId]
		if !found {
			continue
		}

		lookupKey := types.LookupMetaMapkey{
			Hash:   hash.Blake2bHash(serviceblob.Blob),
			Length: types.U32(len(serviceblob.Blob)),
		}
		if timeSlotSet, found := serviceAccount.LookupDict[lookupKey]; found && len(timeSlotSet) > 0 {
			continue
		}

		tauPrime := store.GetInstance().GetPosteriorStates().GetTau()
		serviceAccount.LookupDict[lookupKey] = types.TimeSlotSet{tauPrime}
		serviceAccount.PreimageLookup[lookupKey.Hash] = serviceblob.Blob
		dPrime[serviceId] = serviceAccount
	}

	return dPrime, nil
}

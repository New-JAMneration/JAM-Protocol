package accumulation

import (
	"bytes"
	"errors"
	"log"
	"maps"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	PreimageErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/preimages"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
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
func ShouldIntegratePreimage(d types.ServiceAccountState, s types.ServiceId, h types.OpaqueHash, l types.U32, keyVals *types.StateKeyVals, parseToState bool) bool {
	// Check for existence of the service account
	account, isInAccount := d[s]
	if !isInAccount || account.PreimageLookup == nil || account.LookupDict == nil {
		// If account does not exist or maps are uninitialized, return false
		// log.Println("ServiceAccount not found or maps are uninitialized")
		return false
	}

	// Check if the preimage hash is not in the service account's preimage map
	_, isInPreimageMap := account.PreimageLookup[h]

	// Construct lookup key
	lookupKey := types.LookupMetaMapkey{
		Hash:   h,
		Length: l,
	}

	// Check if the lookupKey have been set before(time slot set is not empty)
	timeSlotSet, lookupKeyExists := account.LookupDict[lookupKey]
	if !lookupKeyExists {
		if parseToState {
			return lookupAndRemoveKeyVal(keyVals, lookupKey, s)
			// only parseToState == true (filter deltaDoubleDagger) needs to remove keyVal and parse to service lookupDict
		} else {
			return lookupInKeyVal(*keyVals, lookupKey, s)
		}
	}

	// Condition: hash does not exist in preimage map, and lookup time slot set is empty
	return !isInPreimageMap && (len(timeSlotSet) == 0)
}

func lookupInKeyVal(keyVals types.StateKeyVals, lookupKey types.LookupMetaMapkey, serviceId types.ServiceId) bool {
	if len(keyVals) == 0 {
		return false
	}

	lookupStateKey := m.EncodeDelta4Key(serviceId, lookupKey)
	for _, v := range keyVals {
		if v.Key == lookupStateKey {
			if len(v.Value) == 1 && v.Value[0] == 0 {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func lookupAndRemoveKeyVal(keyVals *types.StateKeyVals, lookupKey types.LookupMetaMapkey, serviceId types.ServiceId) bool {
	if len(*keyVals) == 0 {
		return false
	}

	lookupStateKey := m.EncodeDelta4Key(serviceId, lookupKey)
	for k, v := range *keyVals {
		if v.Key == lookupStateKey {
			if len(v.Value) == 1 && v.Value[0] == 0 {
				// remove lookupData from keyval
				if k < len(*keyVals)-1 { // not the last index
					*keyVals = append((*keyVals)[:k], (*keyVals)[k+1:]...)
				} else {
					*keyVals = (*keyVals)[:k]
				}
				return true
			} else {
				return false
			}
		}
	}

	return false
}

// v0.7.0 (12.39, 12.40)  for all: E_P: Y(δ, s, H(d), |d|)
// Validate Preimage Extrinsics with prior state service preimage and lookupDict
func validatePreimageExtrinsics(eps types.PreimagesExtrinsic, delta types.ServiceAccountState, keyVals *types.StateKeyVals) *types.ErrorCode {
	// If eps is empty, return empty slice
	if len(eps) == 0 {
		return nil
	}
	// 12.39
	err := validateSortUnique(eps)
	if err != nil {
		return err
	}
	// 12.40
	for _, ep := range eps {
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		if !ShouldIntegratePreimage(delta, ep.Requester, preimageHash, preimageLength, keyVals, false) {
			errCode := PreimageErrorCode.PreimageUnneeded
			return &errCode
		}
	}
	return nil
}

// v0.7.0 (12.39)
func validateSortUnique(eps types.PreimagesExtrinsic) *types.ErrorCode {
	// If eps is not sorted, return error
	for i := 1; i < len(eps); i++ {
		if eps[i-1].Requester > eps[i].Requester {
			// log.Println("eps is not sorted by Requester")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}

		if eps[i-1].Requester == eps[i].Requester && bytes.Compare(eps[i-1].Blob, eps[i].Blob) > 0 {
			// log.Println("eps.Requester is not sorted by Blob")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
	}

	// If eps have duplicates, return error
	for i := 1; i < len(eps); i++ {
		if eps[i].Requester == eps[i-1].Requester && bytes.Equal(eps[i].Blob, eps[i-1].Blob) {
			// log.Println("eps have duplicates")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
	}

	return nil
}

func filterPreimageExtrinsics(eps types.PreimagesExtrinsic, d types.ServiceAccountState, keyVals *types.StateKeyVals) (types.PreimagesExtrinsic, types.ServiceAccountState) {
	j := 0
	for i, ep := range eps {
		// Calculate preimage hash and length
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		// Check if the preimage should be integrated
		if ShouldIntegratePreimage(d, ep.Requester, preimageHash, preimageLength, keyVals, true) {
			eps[j] = eps[i]
			j++

			lookupData := types.LookupMetaMapkey{
				Hash:   preimageHash,
				Length: preimageLength,
			}
			requestService := d[ep.Requester]
			requestService.LookupDict[lookupData] = make(types.TimeSlotSet, 0)
		}

	}
	eps = eps[:j]
	return eps, d
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
// v0.7.0 (12.38-12.43)
func ProcessPreimageExtrinsics() *types.ErrorCode {
	// Get store instance and required states
	s := store.GetInstance()
	eps := s.GetLatestBlock().Extrinsic.Preimages
	delta := s.GetPriorStates().GetDelta()
	deltaDoubleDagger := s.GetIntermediateStates().GetDeltaDoubleDagger()
	keyVals := s.GetStorageKeyVals()
	tauPrime := s.GetPosteriorStates().GetTau()
	// validate E_P and prior state service preimage, lookupDict
	err := validatePreimageExtrinsics(eps, delta, &keyVals)
	if err != nil {
		return err
	}

	// Filter preimage extrinsics, integrate lookup keyvals into dict
	filteredEps, updatedLookupServiceAccount := filterPreimageExtrinsics(eps, deltaDoubleDagger, &keyVals)

	// Update deltaDoubleDagger with filtered preimages
	newDeltaDoubleDagger, UpdateErr := UpdateDeltaWithExtrinsicPreimage(filteredEps, updatedLookupServiceAccount, tauPrime)
	if UpdateErr != nil {
		log.Println("UpdateDeltaWithExtrinsicPreimageErr:", UpdateErr)
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

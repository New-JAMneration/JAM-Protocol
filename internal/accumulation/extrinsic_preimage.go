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
)

// (12.22) Y — predicate function
// Y(δ, s, i) ≡ (ℋ(i), |i|) ∉ K(δ[s]_L) ∧ ℋ(i) ∉ K(δ[s]_P)
func Y(d types.ServiceAccountState, s types.ServiceId, i []byte) bool {
	ds, ok := d[s]
	if !ok || ds.PreimageLookup == nil || ds.LookupDict == nil {
		return false
	}

	// Compute ℋ(i) and |i|
	h := hash.Blake2bHash(i)
	l := types.U32(len(i))

	// Construct lookup key
	lookupKey := types.LookupMetaMapkey{Hash: h, Length: l}

	// Check presence
	timeSlotSet, inLookup := ds.LookupDict[lookupKey]
	if !inLookup {
		return false
	}
	_, isInPreimageMap := ds.PreimageLookup[h]

	// True if both absent
	return !isInPreimageMap && len(timeSlotSet) == 0
}

// (12.38) δ′ = I(δ‡, E_P)
// Main entry point for preimage integration
func ProcessPreimageExtrinsics() *types.ErrorCode {
	s := store.GetInstance()
	eps := s.GetLatestBlock().Extrinsic.Preimages
	deltaDoubleDagger := s.GetIntermediateStates().GetDeltaDoubleDagger()

	// Check E_P sorted & unique
	for i := 1; i < len(eps); i++ {
		if eps[i-1].Requester > eps[i].Requester {
			log.Println("eps is not sorted by Requester")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
		if eps[i-1].Requester == eps[i].Requester && bytes.Compare(eps[i-1].Blob, eps[i].Blob) > 0 {
			log.Println("eps.Requester is not sorted by Blob")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
		if eps[i].Requester == eps[i-1].Requester && bytes.Equal(eps[i].Blob, eps[i-1].Blob) {
			log.Println("eps have duplicates")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
	}

	EP := make(types.ServiceBlobs, 0, len(eps))
	for _, ep := range eps {
		EP = append(EP, types.ServiceBlob{
			ServiceID: ep.Requester,
			Blob:      ep.Blob,
		})
	}

	// (12.38) integrate preimage
	dPrime, err := I(deltaDoubleDagger, EP)
	if err != nil {
		log.Println("IntegratePreimageErr:", err)
		errCode := PreimageErrorCode.PreimageUnneeded
		return &errCode
	}

	// Update δ′ to posterior state
	s.GetPosteriorStates().SetDelta(dPrime)
	return nil
}

// (12.21) I — preimage integration function
func I(d types.ServiceAccountState, p types.ServiceBlobs) (types.ServiceAccountState, error) {
	dPrime := maps.Clone(d)
	tauPrime := store.GetInstance().GetPosteriorStates().GetTau()
	unneededFound := false
	for _, pair := range p {
		s := pair.ServiceID
		i := pair.Blob
		// check predicate Y
		if !Y(d, s, i) {
			unneededFound = true
			continue
		}

		// Compute hash and length
		hash := hash.Blake2bHash(i)
		len := types.U32(len(i))
		key := types.LookupMetaMapkey{Hash: hash, Length: len}

		ds := dPrime[s]

		// ensure maps initialized
		if ds.LookupDict == nil {
			ds.LookupDict = make(map[types.LookupMetaMapkey]types.TimeSlotSet)
		}
		if ds.PreimageLookup == nil {
			ds.PreimageLookup = make(types.PreimagesMapEntry)
		}

		// apply integration
		ds.LookupDict[key] = types.TimeSlotSet{tauPrime}
		ds.PreimageLookup[hash] = i
		dPrime[s] = ds
	}
	if unneededFound {
		return dPrime, errors.New("preimage is not solicited")
	}
	return dPrime, nil
}

package accumulation

import (
	"bytes"
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
	_, inLookup := ds.LookupDict[lookupKey]
	_, inPreimage := ds.PreimageLookup[h]

	// True if both absent
	return !inLookup && !inPreimage
}

// (12.38) δ′ = I(δ‡, E_P)
// Main entry point for preimage integration
func ProcessPreimageExtrinsics() *types.ErrorCode {
	s := store.GetInstance()
	eps := s.GetLatestBlock().Extrinsic.Preimages
	deltaDoubleDagger := s.GetIntermediateStates().GetDeltaDoubleDagger()

	// Sanity: ensure E_P sorted & unique
	for i := 1; i < len(eps); i++ {
		if eps[i-1].Requester > eps[i].Requester ||
			(eps[i-1].Requester == eps[i].Requester &&
				bytes.Compare(eps[i-1].Blob, eps[i].Blob) >= 0) {
			log.Println("E_P not sorted or contains duplicates")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
	}
	// (12.37) disregard irrelevant preimages
	EP := FilterRelevantPreimages(deltaDoubleDagger, eps)

	// (12.38) integrate via I()
	dPrime, err := I(deltaDoubleDagger, EP)
	if err != nil {
		log.Println("Preimage integration error:", err)
		errCode := PreimageErrorCode.PreimageUnneeded
		return &errCode
	}

	// Update δ′ to posterior state
	s.GetPosteriorStates().SetDelta(dPrime)
	return nil
}

// (12.37) FilterRelevantPreimages
// Disregard preimages no longer useful due to accumulation effects.
// ∀(s, i) ∈ E_P : Y(δ, s, i)
func FilterRelevantPreimages(delta types.ServiceAccountState, eps types.PreimagesExtrinsic) types.ServiceBlobs {
	filtered := make(types.ServiceBlobs, 0, len(eps))
	for _, ep := range eps {
		if Y(delta, ep.Requester, ep.Blob) {
			filtered = append(filtered, types.ServiceBlob{
				ServiceID: ep.Requester,
				Blob:      ep.Blob,
			})
		}
	}
	return filtered
}

// (12.21) I — preimage integration function
func I(d types.ServiceAccountState, p types.ServiceBlobs) (types.ServiceAccountState, error) {
	dPrime := maps.Clone(d)
	tauPrime := store.GetInstance().GetPosteriorStates().GetTau()

	for _, pair := range p {
		s := pair.ServiceID
		i := pair.Blob

		// check predicate Y
		if !Y(d, s, i) {
			continue
		}

		// Compute hash and length
		h := hash.Blake2bHash(i)
		l := types.U32(len(i))
		key := types.LookupMetaMapkey{Hash: h, Length: l}

		// ensure account exists
		ds, ok := dPrime[s]
		if !ok {
			continue
		}

		// ensure maps initialized
		if ds.LookupDict == nil {
			ds.LookupDict = make(map[types.LookupMetaMapkey]types.TimeSlotSet)
		}
		if ds.PreimageLookup == nil {
			ds.PreimageLookup = make(types.PreimagesMapEntry)
		}

		// apply integration
		ds.LookupDict[key] = types.TimeSlotSet{tauPrime}
		ds.PreimageLookup[h] = i
		dPrime[s] = ds
	}

	return dPrime, nil
}

package accumulation

import (
	"bytes"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	PreimageErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/preimages"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
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

// v0.6.4 (12.36) R function determines whether a preimage should be integrated.
//
// After the globalKV refactor (Method A: full load at deserialization),
// the legacy unmatchedKeyVals fallback pool no longer exists; lookup-meta
// presence is resolved directly through GetPreimageMeta against globalKV.
func ShouldIntegratePreimage(d types.ServiceAccountState, s types.ServiceID, h types.OpaqueHash, l types.U32) bool {
	account, isInAccount := d[s]
	if !isInAccount {
		return false
	}

	_, isInPreimageMap := account.PreimageLookup[h]

	lookupStateKey, err := m.NewPreimageMetaStateKey(s, h, l)
	if err != nil {
		return false
	}
	timeSlotSet, lookupKeyExists := account.GetPreimageMeta(lookupStateKey)
	if !lookupKeyExists {
		return false
	}

	// Integrate iff the preimage blob is not yet present AND the lookup
	// entry is in the "requested but unsupplied" state (empty timeslot set).
	return !isInPreimageMap && len(timeSlotSet) == 0
}

// v0.7.0 (12.39, 12.40)  for all: E_P: Y(δ, s, H(d), |d|)
// Validate Preimage Extrinsics with prior state service preimage and lookupDict.
func ValidatePreimageExtrinsics(eps types.PreimagesExtrinsic, delta types.ServiceAccountState) error {
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

		if !ShouldIntegratePreimage(delta, ep.Requester, preimageHash, preimageLength) {
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
			// logger.Errorf("eps is not sorted by Requester")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}

		if eps[i-1].Requester == eps[i].Requester && bytes.Compare(eps[i-1].Blob, eps[i].Blob) >= 0 {
			// logger.Errorf("eps.Requester is not sorted by Blob or have duplicates")
			errCode := PreimageErrorCode.PrimagesNotSortedUnique
			return &errCode
		}
	}

	return nil
}

func filterPreimageExtrinsics(eps types.PreimagesExtrinsic, d types.ServiceAccountState) (types.PreimagesExtrinsic, types.ServiceAccountState) {
	// Build a new slice for δ integration. Do not compact in-place: eps may share
	// backing storage with block.Extrinsic.Preimages; in-place compaction leaves a
	// stale len and corrupts E_P used later for π statistics (GP §13.5).
	filtered := make(types.PreimagesExtrinsic, 0, len(eps))
	for _, ep := range eps {
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		if ShouldIntegratePreimage(d, ep.Requester, preimageHash, preimageLength) {
			filtered = append(filtered, ep)
			// No need to "lift" the lookup entry from a fallback pool any
			// more — Method A guarantees the entry is already in globalKV.
		}
	}
	return filtered, d
}

// UpdateDeltaWithExtrinsicPreimage updates the deltaDoubleDagger state with filtered preimages.
// It integrates preimages into deltaDoubleDagger using the provided tauPrime time slot.
// v0.6.4 (12.39)
//
// Mirrors the preimage-integration step from the reference graypaper rules
// (eq. 12.21): for every preimage that passed ShouldIntegratePreimage we
// stamp a_l[(H(i), |i|)] = [τ ′] in globalKV and add the blob to a_p.
// InsertPreimageMeta is used (rather than UpdatePreimageMeta) so this step is
// idempotent: if the lookup entry already exists with empty timeslots
// (the normal solicited case) the value is overwritten without touching
// the a_i / a_o counters; if it somehow does not exist yet, the entry is
// created and the counters are charged. This matches the behaviour of the
// pre-refactor map[…]TimeSlotSet assignment and avoids a nil-state hazard
// that bubbled up into ProcessPreimageExtrinsics' SetDelta call.
func UpdateDeltaWithExtrinsicPreimage(eps types.PreimagesExtrinsic, deltaDoubleDagger types.ServiceAccountState, tauPrime types.TimeSlot) (types.ServiceAccountState, error) {
	for _, ep := range eps {
		preimageHash := hash.Blake2bHash(ep.Blob)
		preimageLength := types.U32(len(ep.Blob))

		serviceAccount, exists := deltaDoubleDagger[ep.Requester]
		if !exists {
			return nil, errors.New("service account not found")
		}
		if serviceAccount.PreimageLookup == nil {
			serviceAccount.PreimageLookup = make(types.PreimagesMapEntry)
		}

		stateKey, err := m.NewPreimageMetaStateKey(ep.Requester, preimageHash, preimageLength)
		if err != nil {
			return nil, err
		}
		if err := serviceAccount.InsertPreimageMeta(stateKey, uint64(preimageLength), types.TimeSlotSet{tauPrime}); err != nil {
			return nil, err
		}
		serviceAccount.PreimageLookup[preimageHash] = ep.Blob

		deltaDoubleDagger[ep.Requester] = serviceAccount
	}

	return deltaDoubleDagger, nil
}

// ProcessPreimageExtrinsics is the main unified function for handling preimage extrinsics
// It combines filtering and delta state updates in a single call for external use
// v0.7.0 (12.38-12.43)
func ProcessPreimageExtrinsics() error {
	// Get cs instance and required states
	cs := blockchain.GetInstance()
	eps := cs.GetLatestBlock().Extrinsic.Preimages
	deltaDoubleDagger := cs.GetIntermediateStates().GetDeltaDoubleDagger()
	tauPrime := cs.GetPosteriorStates().GetTau()

	// Filter preimage extrinsics, integrate lookup keyvals into dict
	filteredEps, updatedLookupServiceAccount := filterPreimageExtrinsics(eps, deltaDoubleDagger)

	// Update deltaDoubleDagger with filtered preimages
	newDeltaDoubleDagger, UpdateErr := UpdateDeltaWithExtrinsicPreimage(filteredEps, updatedLookupServiceAccount, tauPrime)
	if UpdateErr != nil {
		// Log the error but keep the prior delta intact: an error from
		// UpdateDeltaWithExtrinsicPreimage returns (nil, err), so blindly
		// publishing newDeltaDoubleDagger would wipe every service account
		// from the posterior state.
		logger.Errorf("UpdateDeltaWithExtrinsicPreimageErr: %v", UpdateErr)
		return nil
	}

	// Update new double-dagger to posterior state
	cs.GetPosteriorStates().SetDelta(newDeltaDoubleDagger)
	return nil
}

// Provide is the preimage integration function (different from IntegratePreimage despite re-using the word "integrate")
// It transforms a dictionary of service states and a set of service/hash pairs into a new dictionary of service states.
// (map[N_s]A, (N_s, Y)) -> map[N_s]A
// v0.6.5 (12.18)
//
// Uses InsertPreimageMeta (idempotent) instead of UpdatePreimageMeta so that
// a lookup entry which had timeslots == [] is overwritten with [τ ′] in
// place — see the UpdateDeltaWithExtrinsicPreimage comment for the same
// rationale.
func Provide(d types.ServiceAccountState, eps types.ServiceBlobs) (types.ServiceAccountState, error) {
	tauPrime := blockchain.GetInstance().GetPosteriorStates().GetTau()
	for _, serviceblob := range eps {
		serviceID := serviceblob.ServiceID
		serviceAccount, found := d[serviceID]
		if !found {
			continue
		}

		preimageHash := hash.Blake2bHash(serviceblob.Blob)
		preimageLength := types.U32(len(serviceblob.Blob))
		stateKey, err := m.NewPreimageMetaStateKey(serviceID, preimageHash, preimageLength)
		if err != nil {
			return nil, err
		}
		timeSlotSet, found := serviceAccount.GetPreimageMeta(stateKey)
		if !found || len(timeSlotSet) > 0 {
			// No matching solicit, or the preimage was already provided.
			continue
		}

		if err := serviceAccount.InsertPreimageMeta(stateKey, uint64(preimageLength), types.TimeSlotSet{tauPrime}); err != nil {
			return nil, err
		}
		if serviceAccount.PreimageLookup == nil {
			serviceAccount.PreimageLookup = make(types.PreimagesMapEntry)
		}
		serviceAccount.PreimageLookup[preimageHash] = serviceblob.Blob
		d[serviceID] = serviceAccount
	}

	return d, nil
}

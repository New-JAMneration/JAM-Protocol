package service_account

import (
	"fmt"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// (9.4) This function was integrated into HistoricalLookup function.
// serviceID is required after the globalKV refactor to construct the
// preimage-meta StateKey when validating that a_p entries are still
// referenced by a_l.
func FetchCodeByHash(serviceID types.ServiceID, account types.ServiceAccount, codeHash types.OpaqueHash) (metadata types.ByteSequence, code types.ByteSequence, err error) {
	/*
		∀a ∈ A ∶  E(↕a_meta,a_code) ≡⎧ a_p[a_codeHash] if a_codeHash ∈ a_p
		                             ⎨ ∅               otherwise
	*/

	// check if CodeHash exists in PreimageLookup
	if bytes, exists := account.PreimageLookup[codeHash]; exists {
		return DecodeMetaCode(bytes)
	}
	// if we can't fetch metadata and code, then validate function should return error
	err = ValidatePreimageLookupDict(serviceID, account)
	if err != nil {
		logger.Error("ValidatePreimageLookupDict failed and you'll get nil return")
		return nil, nil, err
	}
	// default is empty
	return nil, nil, nil
}

func DecodeMetaCode(bytes types.ByteSequence) (metadata types.ByteSequence, code types.ByteSequence, err error) {
	metaCode := types.MetaCode{}
	decoder := types.NewDecoder()
	err = decoder.Decode(bytes, &metaCode)
	if err != nil {
		logger.Errorf("Failed to deserialize code and metadata: %v, return nil\n", err)
		return nil, nil, err
	}
	/*
		if metaCode.Metadata != nil {
			// print metadata
			logger.Debugf("Metadata of fetched code is %v\n", metaCode.Metadata)
		}
	*/
	return metaCode.Metadata, metaCode.Code, nil
}

// (9.6) Invariant: This is definition, not real used formula
// but I implement this for debugging/validation
// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)
//
// serviceID is required after the globalKV refactor to construct the
// preimage-meta StateKey when checking K(a_l) membership.
func ValidatePreimageLookupDict(serviceID types.ServiceID, account types.ServiceAccount) error {
	for codeHash, preimage := range account.PreimageLookup {
		// // h = H(p)
		preimageHash := hash.Blake2bHash(utils.ByteSequenceWrapper{Value: preimage}.Serialize())
		if codeHash != preimageHash {
			return fmt.Errorf("\nCodeHash: 0x%x \nshould equal to PreimageHash: 0x%x", codeHash, preimageHash)
		}
		if !existsInLookupDict(serviceID, account, codeHash, preimage) {
			return fmt.Errorf("\nCodeHash: 0x%x, Preimage: 0x%x not found in LookupDict keysize", codeHash, preimage)
		}
	}
	return nil
}

// (h, |p|) ∈ K(a_l)
//
// Looks up the preimage-meta entry in globalKV via NewPreimageMetaStateKey
// (post-refactor SOT). serviceID is part of the StateKey, hence required.
func existsInLookupDict(serviceID types.ServiceID, account types.ServiceAccount, codeHash types.OpaqueHash, preimage types.ByteSequence) bool {
	stateKey, err := merklization.NewPreimageMetaStateKey(serviceID, codeHash, types.U32(len(preimage)))
	if err != nil {
		return false
	}
	_, exists := account.GetPreimageMeta(stateKey)
	return exists
}

// (9.7) historicalLookup Lambda Λ, which is the exact definition of (9.5).
//
// serviceID is required after the globalKV refactor to derive the
// preimage-meta StateKey for a_l[h, |a_p[h]|].
func HistoricalLookup(serviceID types.ServiceID, account types.ServiceAccount, timestamp types.TimeSlot, h types.OpaqueHash) (bytes types.ByteSequence) {
	/*
		Λ(a, t, h) ≡
			a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
			∅      otherwise
	*/
	// h, |a_p[h]|
	preimage, preimageExists := account.PreimageLookup[h]
	if !preimageExists {
		return nil
	}
	lookupStateKey, err := merklization.NewPreimageMetaStateKey(serviceID, h, types.U32(len(preimage)))
	if err != nil {
		return nil
	}

	// a_l[ h, |a_p[h]| ]
	l, _ := account.GetPreimageMeta(lookupStateKey)

	// a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
	if isValidTime(l, timestamp) {
		return preimage
	}

	// ∅      otherwise
	return nil
}

// MaxHistoricalTimeslotsForPreimageMeta is the cap on the length of an
// a_l[h,z] entry per the graypaper: l ∈ ⟦N_T⟧_{:3}.
const MaxHistoricalTimeslotsForPreimageMeta = 3

// AddPreimage installs preimage blob p on the given service account while
// preserving the GP §9.6 invariant:
//
//	∀a ∈ A, (h ↦ d) ∈ a_p ⇒ h = H(d) ∧ (h, |d|) ∈ K(a_l)
//
// Branches:
//   - PreimageLookup already contains h AND a_l[h,|p|] exists: append
//     currentTimeslot to a_l[h,|p|] if its length is below the cap; the
//     counters are untouched because the footprint was paid for at solicit
//     time.
//   - PreimageLookup already contains h but a_l[h,|p|] is missing (edge case
//     where a forget already removed the meta but the blob still lingers):
//     silently no-op, refusing to re-break the invariant via writes here.
//   - PreimageLookup does NOT contain h: store the blob and create a fresh
//     a_l entry of [currentTimeslot] via InsertPreimageMeta (which charges
//     the +2 items / +(81+|p|) octets footprint).
func AddPreimage(account *types.ServiceAccount, serviceID types.ServiceID, p []byte, currentTimeslot types.TimeSlot) error {
	h := hash.Blake2bHash(types.ByteSequence(p))
	stateKey, err := merklization.NewPreimageMetaStateKey(serviceID, h, types.U32(len(p)))
	if err != nil {
		return err
	}

	if _, exists := account.PreimageLookup[h]; exists {
		meta, ok := account.GetPreimageMeta(stateKey)
		if !ok {
			// Edge case: blob present but meta missing. Silently no-op.
			return nil
		}
		if len(meta) < MaxHistoricalTimeslotsForPreimageMeta {
			meta = append(meta, currentTimeslot)
			if err := account.UpdatePreimageMeta(stateKey, meta); err != nil {
				return err
			}
		}
		return nil
	}

	// Fresh insertion: store blob, create meta with single timeslot.
	if account.PreimageLookup == nil {
		account.PreimageLookup = make(types.PreimagesMapEntry)
	}
	account.PreimageLookup[h] = types.ByteSequence(p)
	return account.InsertPreimageMeta(stateKey, uint64(len(p)), types.TimeSlotSet{currentTimeslot})
}

// I
func isValidTime(l types.TimeSlotSet, t types.TimeSlot) bool {
	/*
		I(l, t) =
			false             if [] = l
			x ≤ t             if [x] = l
			x ≤ t < y         if [x, y] = l
			x ≤ t < y ∨ z ≤ t if [x, y, z] = l
	*/
	switch len(l) {
	case 0:
		return false
	case 1:
		return l[0] <= t
	case 2:
		return l[0] <= t && t < l[1]
	case 3:
		return (l[0] <= t && t < l[1]) || l[2] <= t
	default:
		// ⟦N_T⟧_{∶3}
		return false
	}
}

// CalcKeys returns a_i — the number of items in the account.
// After the global-KV refactor this is an O(1) read from the incremental
// counter maintained by InsertStorage / InsertPreimageMeta / DeleteStorage /
// DeletePreimageMeta. GP §9.8: a_i ≡ 2*|a_l| + |a_s|.
func CalcKeys(account types.ServiceAccount) types.U32 {
	return types.U32(account.GetTotalNumberOfItems())
}

// CalcOctets returns a_o — the total number of octets occupied by the
// account's storage and preimage meta. After the global-KV refactor this is
// an O(1) read from the incremental counter.
// GP §9.8: a_o ≡ Σ(81+z) + Σ(34+|x|+|key|).
func CalcOctets(account types.ServiceAccount) types.U64 {
	return types.U64(account.GetTotalNumberOfOctets())
}

// CalcThresholdBalance is preserved as a free function for the PVM host
// calls that already pass (aI, aO, aF) directly into it.
//
// Internally we delegate to the new ServiceAccount.ThresholdBalance() method
// by temporarily setting the counters on a stack-allocated ServiceAccount,
// keeping a single source of truth for the GP §9.8 formula
// (a_t ≡ max(0, B_S + B_I*a_i + B_L*a_o − a_f)).
func CalcThresholdBalance(aI types.U32, aO types.U64, aF types.U64) types.U64 {
	var stub types.ServiceAccount
	stub.SetTotalNumberOfItems(uint32(aI))
	stub.SetTotalNumberOfOctets(uint64(aO))
	stub.ServiceInfo.DepositOffset = aF
	result, err := stub.ThresholdBalance()
	if err != nil {
		// Overflow in threshold balance is a programmer error; falling back
		// to 0 here mirrors how the host calls treat an unsatisfiable
		// balance (FULL outcome) without panicking.
		return 0
	}
	return result
}

/*
	a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
	a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈Value(a_s)}	34 + |x| ]
*/
// compute how many items a_i(keys) and a_o(ocetes) the lookupItem has
func CalcLookupItemfootprint(lookupItem types.LookupMetaMapkey) (types.U32, types.U64) {
	return 2, 81 + types.U64(lookupItem.Length)
}

// compute how many items a_i(keys) and a_o(ocetes) the storageItem has
func CalcStorageItemfootprint(storageRawKey string, storageData types.ByteSequence) (types.U32, types.U64) {
	return 1, 34 + types.U64(len(storageRawKey)) + types.U64(len(storageData))
}

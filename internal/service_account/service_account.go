package service_account

import (
	"fmt"
	"log"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (9.4) This function was integrated into HistoricalLookup function
func FetchCodeByHash(account types.ServiceAccount, codeHash types.OpaqueHash) (metadata types.ByteSequence, code types.ByteSequence, err error) {
	/*
		∀a ∈ A ∶  E(↕a_meta,a_code) ≡⎧ a_p[a_codeHash] if a_codeHash ∈ a_p
		                             ⎨ ∅               otherwise
	*/

	// check if CodeHash exists in PreimageLookup
	if bytes, exists := account.PreimageLookup[codeHash]; exists {
		return DecodeMetaCode(bytes)
	}
	// if we can't fetch metadata and code, then validate function should return error
	err = ValidatePreimageLookupDict(account)
	if err != nil {
		log.Printf("ValidatePreimageLookupDict failed and you'll get nil return")
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
		log.Printf("Failed to deserialize code and metadata: %v, return nil\n", err)
		return nil, nil, err
	}
	/*
		if metaCode.Metadata != nil {
			// print metadata
			log.Printf("Metadata of fetched code is %v\n", metaCode.Metadata)
		}
	*/
	return metaCode.Metadata, metaCode.Code, nil
}

// (9.6) Invariant: This is definition, not real used formula
// but I implement this for debugging/validation
// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)
func ValidatePreimageLookupDict(account types.ServiceAccount) error {
	for codeHash, preimage := range account.PreimageLookup {
		// // h = H(p)
		preimageHash := hash.Blake2bHash(utils.ByteSequenceWrapper{Value: preimage}.Serialize())
		if codeHash != preimageHash {
			return fmt.Errorf("\nCodeHash: %v \nshould equal to PreimageHash: %v", codeHash, preimageHash)
		}
		if !existsInLookupDict(account, codeHash, preimage) {
			return fmt.Errorf("\nCodeHash: %v, Preimage: %v not found in LookupDict keysize", codeHash, preimage)
		}
	}
	return nil
}

// (h, |p|) ∈ K(a_l)
func existsInLookupDict(account types.ServiceAccount, codeHash types.OpaqueHash, preimage types.ByteSequence) bool {
	key := types.LookupMetaMapkey{
		Hash:   codeHash,
		Length: types.U32(len(preimage)),
	}
	_, exists := account.LookupDict[key]
	return exists
}

// (9.7) historicalLookup Lambda Λ, which is the exact definition of (9.5)
func HistoricalLookup(account types.ServiceAccount, timestamp types.TimeSlot, hash types.OpaqueHash) (bytes types.ByteSequence) {
	/*
		Λ(a, t, h) ≡
			a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
			∅      otherwise
	*/
	// h, |a_p[h]|
	lookupkey := types.LookupMetaMapkey{
		Hash:   hash,
		Length: types.U32(len(account.PreimageLookup[hash])),
	}

	// a_l[ h, |a_p[h]| ]
	l := account.LookupDict[lookupkey]

	// a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
	if bytes, exists := account.PreimageLookup[hash]; exists && isValidTime(l, timestamp) {
		/*
			∀a ∈ A ∶  E(↕a_meta,a_code) ≡⎧ a_p[a_codeHash] if a_codeHash ∈ a_p
			                             ⎨ ∅               otherwise
		*/
		return bytes
	}

	// ∅      otherwise
	return nil
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

// (9.8) You can use this function to get account derivatives
// TODO: Items(a_i) and Bytes(a_o) are stored in ServiceInfo,
// Maybe we can update a_t only
func GetServiceAccountDerivatives(account types.ServiceAccount) (accountDer types.ServiceAccountDerivatives) {
	/*
		∀a ∈ V(δ) ∶
		⎧ a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
		⎪ a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈V(a_s)}	34 + |x| ]
		⎨ a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
		⎩
	*/

	// calculate derivative invariants
	var (
		Items      = CalcKeys(account)
		Bytes      = CalcOctets(account)
		Minbalance = CalcThresholdBalance(Items, Bytes, account.ServiceInfo.DepositOffset)
	)
	accountDer = types.ServiceAccountDerivatives{
		Items:      Items,
		Bytes:      Bytes,
		Minbalance: Minbalance,
	}
	return accountDer
}

// a_i: calculate number of items(keys) in storage
func CalcKeys(account types.ServiceAccount) types.U32 {
	/*
		a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
	*/
	return types.U32(2*len(account.LookupDict) + len(account.StorageDict))
}

// a_o: calculate total number of octets(datas) used in storage
func CalcOctets(account types.ServiceAccount) types.U64 {
	/*
		a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈Value(a_s)}	34 + |x| ]
	*/
	// calculate all (81 + preiamge lookup length in keysize)
	keyContribution := 0
	for key := range account.LookupDict {
		keyContribution += 81 + int(key.Length)
	}

	//  calculate all [ 32(size of key) + size of data ]
	stateContribution := 0
	for x, y := range account.StorageDict {
		stateContribution += 34 + len(y) + len(x)
	}

	return types.U64(keyContribution + stateContribution)
}

// a_t: calculate threshold(minimum) balance needed for any account in terms of storage footprint
func CalcThresholdBalance(aI types.U32, aO types.U64, aF types.U64) types.U64 {
	/*
		a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
	*/
	storage := types.U64(types.BasicMinBalance) + types.U64(types.U32(types.AdditionalMinBalancePerItem)*aI) + types.U64(types.AdditionalMinBalancePerOctet)*aO
	if storage < aF {
		// result < 0
		return 0
	}
	return aF
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

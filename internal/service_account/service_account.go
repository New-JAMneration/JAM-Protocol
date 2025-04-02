package service_account

import (
	"fmt"
	"log"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (9.4) This function was integrated into HistoricalLookupFunction
func FetchCodeByHash(account types.ServiceAccount, codeHash types.OpaqueHash) (metadata types.ByteSequence, code types.ByteSequence) {
	/*
		∀a ∈ A ∶  E(↕a_meta,a_code) ≡⎧ a_p[a_codeHash] if a_codeHash ∈ a_p
		                             ⎨ ∅               otherwise
	*/

	// check if CodeHash exists in PreimageLookup
	if bytes, exists := account.PreimageLookup[codeHash]; exists {
		metaCode := types.MetaCode{}
		decoder := types.NewDecoder()
		err := decoder.Decode(bytes, &metaCode)
		if err != nil {
			log.Fatalf("Failed to deserialize code and metadata: %v", err)
			return nil, nil
		}
		return metaCode.Metadata, metaCode.Code
	}
	// default is empty
	return nil, nil
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

// (9.7) historicalLookupFunction Lambda Λ, which is the exact definition of (9.5)
func HistoricalLookupFunction(account types.ServiceAccount, timestamp types.TimeSlot, hash types.OpaqueHash) (metadata types.ByteSequence, code types.ByteSequence) {
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
		metaCode := types.MetaCode{}
		decoder := types.NewDecoder()
		err := decoder.Decode(bytes, &metaCode)
		if err != nil {
			log.Printf("Failed to deserialize code and metadata: %v, return nil\n", err)
			return nil, nil
		}
		if metaCode.Metadata != nil {
			// print metadata
			log.Printf("Metadata is %v\n", metaCode.Metadata)
		}
		return metaCode.Metadata, metaCode.Code
	}

	// ∅      otherwise
	return nil, nil
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
func GetSerivecAccountDerivatives(account types.ServiceAccount) (accountDer types.ServiceAccountDerivatives) {
	/*
		∀a ∈ V(δ) ∶
		⎧ a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
		⎪ a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈V(a_s)}	32 + |x| ]
		⎨ a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
		⎩
	*/

	// calculate derivative invariants
	var (
		Items      = calcKeys(account)
		Bytes      = calcUsedOctets(account)
		Minbalance = calcThresholdBalance(Items, Bytes)
	)
	accountDer = types.ServiceAccountDerivatives{
		Items:      Items,
		Bytes:      Bytes,
		Minbalance: Minbalance,
	}
	return accountDer
}

// calculate number of items(keys) in storage
func calcKeys(account types.ServiceAccount) types.U32 {
	/*
		a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
	*/
	lookupDictKeySize := len(account.LookupDict) * 2
	storageDictKeySize := len(account.StorageDict)
	return types.U32(lookupDictKeySize + storageDictKeySize)
}

// calculate total number of octets(datas) used in storage
func calcUsedOctets(account types.ServiceAccount) types.U64 {
	/*
		a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈Value(a_s)}	32 + |x| ]
	*/
	// calculate all (81 + preiamge lookup length in keysize)
	keyContribution := 0
	for key := range account.LookupDict {
		keyContribution += 81 + int(key.Length)
	}

	//  calculate all [ 32(size of key) + size of data ]
	stateContribution := 0
	for _, x := range account.StorageDict {
		stateContribution += 32 + len(x)
	}

	return types.U64(keyContribution + stateContribution)
}

// calculate threshold(minimum) balance needed for any account in terms of storage footprint
func calcThresholdBalance(aI types.U32, aO types.U64) types.U64 {
	/*
		a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
	*/
	return types.U64(types.BasicMinBalance) + types.U64(types.U32(types.AdditionalMinBalancePerItem)*aI) + types.U64(types.AdditionalMinBalancePerOctet)*aO
}

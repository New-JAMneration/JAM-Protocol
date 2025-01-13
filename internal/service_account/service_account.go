package service_account

import (
	"fmt"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (9.4) // Remain Check this function
func CheckAccountExistence() {
	delta := store.GetInstance().GetPriorStates().GetState().Delta
	for id := range delta {
		// take out value(account) from delta
		/*
			type ServiceAccountState(delta) map[ServiceId]ServiceAccount
		*/
		account := delta[id]

		// check if CodeHash exists in PreimageLookup
		if value, exists := account.PreimageLookup[account.CodeHash]; exists {
			fmt.Println("CodeHash exists in PreimageLookup", id, value)
			// account.CodeHash = value
		} else {
			// default is empty
			account.CodeHash = types.OpaqueHash{}
		}

		// set value back to delta
		delta[id] = account
	}

	// set new delta to IntermediateStates
	store.GetInstance().GetIntermediateStates().SetDelta(delta)
}

// // (9.5)
// func historicalLookupFunction(account types.ServiceAccount, timestamp int64, hash types.OpaqueHash) types.ByteSequence {
// 	deltas := store.GetInstance().GetIntermediateStates().GetDelta()
// 	for _, block := range blocks {
// 		// 僅檢查在指定時間戳範圍內的區塊
// 		if block.Timestamp > timestamp {
// 			continue
// 		}

// 		// 遍歷區塊中的前影數據
// 		for _, preimage := range block.Preimages {
// 			if preimage.Hash == hash {
// 				return preimage.Data, true
// 			}
// 		}
// 	}

// 	// 若未找到匹配的前影，返回空值
// 	return string{}
// }

// (9.6) Invariant

// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)
func ValidateAccount(account types.ServiceAccount) error {
	for codeHash, preimage := range account.PreimageLookup {
		// h = H(p)
		preimageHash := hash.Blake2bHash(preimage)
		if codeHash != preimageHash || !existsInLookupDict(account, codeHash, preimage) {
			return fmt.Errorf("validation failed for CodeHash: %v, PreimageHash: %v", codeHash, preimageHash)
		}
	}
	return nil
}

// (h, |p|) ∈ K(a_l)
func existsInLookupDict(account types.ServiceAccount, codeHash types.OpaqueHash, preimage types.ByteSequence) bool {
	key := types.DictionaryKey{
		Hash:   codeHash,
		Length: types.U32(len(preimage)),
	}
	_, exists := account.LookupDict[key]
	return exists
}

// (9.7) historicalLookupFunction Lambda Λ
func HistoricalLookupFunction(account types.ServiceAccount, timestamp types.TimeSlot, hash types.OpaqueHash) types.ByteSequence {
	/*
		Λ(a, t, h) ≡
			a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
			∅      otherwise
	*/
	// h, |a_p[h]|
	lookupkey := types.DictionaryKey{
		Hash:   hash,
		Length: types.U32(len(account.PreimageLookup[hash])),
	}

	// a_l[ h, |a_p[h]| ]
	l := account.LookupDict[lookupkey]

	// a_p[h] if h ∈ Key(a_p) ∧ I( a_l[ h, |a_p[h]| ], t )
	if ifHashInKey(account, hash) && isValidTime(l, timestamp) {
		return account.PreimageLookup[hash]
	}

	// ∅      otherwise
	return types.ByteSequence{}
}

// h ∈ Key(a_p)
func ifHashInKey(account types.ServiceAccount, hash types.OpaqueHash) bool {
	_, exists := account.PreimageLookup[hash]
	return exists
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
		return false
	}
}

// (9.8) shoud be updated once values in service account changed
func UpdateSerivecAccount() {
	/*
		∀a ∈ V(δ) ∶
		⎧ a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
		⎪ a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈V(a_s)}	32 + |x| ]
		⎨ a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
		⎩
	*/
	delta := store.GetInstance().GetPriorStates().GetState().Delta
	for id := range delta {
		// take out value(account) from delta
		/*
			type ServiceAccountState(delta) map[ServiceId]ServiceAccount
		*/
		account := delta[id]
		account = types.ServiceAccount{
			Items:   calcKeys(account),
			Bytes:   calcUsedOctets(account),
			Balance: calcThresholdBalance(account),
		}
		// set value back to delta
		delta[id] = account
	}
	store.GetInstance().GetIntermediateStates().SetDelta(delta)
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
func calcThresholdBalance(account types.ServiceAccount) types.U64 {
	/*
		a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
	*/
	aI := calcKeys(account)
	aO := calcUsedOctets(account)
	return types.U64(types.BasicMinBalance) + types.U64(types.U32(types.AdditionalMinBalancePerItem)*aI) + types.U64(types.AdditionalMinBalancePerOctet)*aO
}

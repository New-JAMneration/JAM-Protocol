package PolkaVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (9.8) You can use this function to get account derivatives
func GetSerivecAccountDerivatives(account types.ServiceAccount) (accountDer service_account.ServiceAccountDerivatives) {
	/*
		∀a ∈ V(δ) ∶
		⎧ a_i ∈ N_2^32 ≡ 2*|a_l| + |a_s|
		⎪ a_o ∈ N_2^64 ≡ [ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈V(a_s)}	32 + |x| ]
		⎨ a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
		⎩
	*/

	// account := delta[id]
	// calculate derivative invariants
	var (
		Items      = calcKeys(account)
		Bytes      = calcUsedOctets(account)
		Minbalance = CalcThresholdBalance(Items, Bytes)
	)
	accountDer = service_account.ServiceAccountDerivatives{
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
func CalcThresholdBalance(aI types.U32, aO types.U64) types.U64 {
	/*
		a_t ∈ N_B ≡ B_S + B_I*a_i + B_L*a_o
	*/
	return types.U64(types.BasicMinBalance) + types.U64(types.U32(types.AdditionalMinBalancePerItem)*aI) + types.U64(types.AdditionalMinBalancePerOctet)*aO
}

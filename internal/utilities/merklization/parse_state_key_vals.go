package merklization

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (D.1) type 1
func C(stateIndex types.U8) types.StateKey {
	stateWrapper := StateWrapper{StateIndex: stateIndex}
	stateKey := stateWrapper.StateKeyConstruct()
	return stateKey
}

func isServiceInfo(stateKey types.StateKey) bool {
	return stateKey[0] == 0xFF
}

// (D.1) type 2
func parseServiceIdFromType2(stateKey types.StateKey) (types.ServiceId, error) {
	// Decode the service Id from the state key
	// service id = [k1, k3, k5, k7]
	encodedServiceId := types.ByteSequence{stateKey[1], stateKey[3], stateKey[5], stateKey[7]}
	decoder := types.NewDecoder()
	var serviceId types.ServiceId
	err := decoder.Decode(encodedServiceId, &serviceId)
	if err != nil {
		return types.ServiceId(0), err
	}

	return serviceId, nil
}

// (D.1) type 3
func parseServiceIdFromType3(stateKey types.StateKey) (types.ServiceId, error) {
	// Decode the service Id from the state key
	// service id = [k0, k2, k4, k6]
	encodedServiceId := types.ByteSequence{stateKey[0], stateKey[2], stateKey[4], stateKey[6]}
	decoder := types.NewDecoder()
	var serviceId types.ServiceId
	err := decoder.Decode(encodedServiceId, &serviceId)
	if err != nil {
		return types.ServiceId(0), err
	}

	return serviceId, nil
}

// TODO: Refactor this function
// return error type
// We need to serviceId, preimageKey, and preimageValue when we update the state
func isPreimage(stateKey types.StateKey, stateVal types.ByteSequence) bool {
	// 1. decode the value with preimage type
	// 2. get the servcieId
	// 3. create a state key by myself
	// 4. compare the statekey with input state key
	// 5. if the key is matched, them the value is preimage value.

	// The preimage value is a ByteSequence
	preimageValue := stateVal

	// Get ServiceId from state key
	serviceId, err := parseServiceIdFromType3(stateKey)
	if err != nil {
		fmt.Println("Error parsing service ID from state key:", err)
		// FIXME: return error
		return false
	}

	// Create preiage key
	preimageKey := hash.Blake2bHash(preimageValue)

	// Create a new state key using the preimage value
	preimageStateKeyVal := encodeDelta3KeyVal(serviceId, preimageKey, preimageValue)

	return preimageStateKeyVal.Key == stateKey
}

func StateKeyValsToState(stateKeyVals types.StateKeyVals) (types.State, error) {
	fmt.Println("Size of stateKeyVals:", len(stateKeyVals))

	var err error
	state := types.State{
		Delta: make(types.ServiceAccountState),
	}

	unmatchedStateKeyVals := make(map[types.StateKey]types.ByteSequence)

	for _, keyVal := range stateKeyVals {
		stateKey := keyVal.Key
		stateVal := keyVal.Value

		unmatchedStateKeyVals[stateKey] = stateVal

		switch stateKey {
		case C(1):
			fmt.Println("StateKey 1")
			fmt.Println(stateKey)
			// Decode the alpha
			alpha, err := decodeAlpha(stateVal)
			if err != nil {
				return state, err
			}
			state.Alpha = alpha
			delete(unmatchedStateKeyVals, stateKey)
		case C(2):
			fmt.Println("StateKey 2")
			fmt.Println(stateKey)
			// Decode the varphi
			varphi, err := decodeVarphi(stateVal)
			if err != nil {
				return state, err
			}
			state.Varphi = varphi
			delete(unmatchedStateKeyVals, stateKey)
		case C(3):
			fmt.Println("StateKey 3")
			fmt.Println(stateKey)
			// Decode the beta
			beta, err := decodeBeta(stateVal)
			if err != nil {
				return state, err
			}
			state.Beta = beta
			delete(unmatchedStateKeyVals, stateKey)
		case C(4):
			fmt.Println("StateKey 4")
			fmt.Println(stateKey)
			// Decode the gamma
			gamma, err := decodeGamma(stateVal)
			if err != nil {
				return state, err
			}
			state.Gamma = gamma
			delete(unmatchedStateKeyVals, stateKey)
		case C(5):
			fmt.Println("StateKey 5")
			fmt.Println(stateKey)
			// Decode the psi
			psi, err := decodePsi(stateVal)
			if err != nil {
				return state, err
			}
			state.Psi = psi
			delete(unmatchedStateKeyVals, stateKey)
		case C(6):
			fmt.Println("StateKey 6")
			fmt.Println(stateKey)
			// Decode the eta
			eta, err := decodeEta(stateVal)
			if err != nil {
				return state, err
			}
			state.Eta = eta
			delete(unmatchedStateKeyVals, stateKey)
		case C(7):
			fmt.Println("StateKey 7")
			fmt.Println(stateKey)
			// Decode the iota
			iota, err := decodeIota(stateVal)
			if err != nil {
				return state, err
			}
			state.Iota = iota
			delete(unmatchedStateKeyVals, stateKey)
		case C(8):
			fmt.Println("StateKey 8")
			fmt.Println(stateKey)
			// Decode the kappa
			kappa, err := decodeKappa(stateVal)
			if err != nil {
				return state, err
			}
			state.Kappa = kappa
			delete(unmatchedStateKeyVals, stateKey)
		case C(9):
			fmt.Println("StateKey 9")
			fmt.Println(stateKey)
			// Decode the lambda
			lambda, err := decodeLambda(stateVal)
			if err != nil {
				return state, err
			}
			state.Lambda = lambda
			delete(unmatchedStateKeyVals, stateKey)
		case C(10):
			fmt.Println("StateKey 10")
			fmt.Println(stateKey)
			// Decode the rho
			rho, err := decodeRho(stateVal)
			if err != nil {
				return state, err
			}
			state.Rho = rho
			delete(unmatchedStateKeyVals, stateKey)
		case C(11):
			fmt.Println("StateKey 11")
			fmt.Println(stateKey)
			// Decode the tau
			tau, err := decodeTau(stateVal)
			if err != nil {
				return state, err
			}
			state.Tau = tau
			delete(unmatchedStateKeyVals, stateKey)
		case C(12):
			fmt.Println("StateKey 12")
			fmt.Println(stateKey)
			// Decode the chi
			chi, err := decodeChi(stateVal)
			if err != nil {
				return state, err
			}
			state.Chi = chi
			delete(unmatchedStateKeyVals, stateKey)
		case C(13):
			fmt.Println("StateKey 13")
			fmt.Println(stateKey)
			// Decode the pi
			pi, err := decodePi(stateVal)
			if err != nil {
				return state, err
			}
			state.Pi = pi
			delete(unmatchedStateKeyVals, stateKey)
		case C(14):
			fmt.Println("StateKey 14")
			fmt.Println(stateKey)
			// Decode the theta
			theta, err := decodeTheta(stateVal)
			if err != nil {
				return state, err
			}
			state.Theta = theta
			delete(unmatchedStateKeyVals, stateKey)
		case C(15):
			fmt.Println("StateKey 15")
			fmt.Println(stateKey)
			// Decode the xi
			xi, err := decodeXi(stateVal)
			if err != nil {
				return state, err
			}
			state.Xi = xi
			delete(unmatchedStateKeyVals, stateKey)
		case C(16):
			fmt.Println("StateKey 16")
			fmt.Println(stateKey)
			// Decode the theta AccumulatedServiceOutput
			lastAccOut, err := decodeThetaAccOut(stateVal)
			if err != nil {
				return state, err
			}
			state.LastAccOut = lastAccOut
			delete(unmatchedStateKeyVals, stateKey)
		default:
			fmt.Println("Unknown StateKey")
			fmt.Println(stateKey)

			// C(255, s)
			if isServiceInfo(stateKey) {
				fmt.Println("StateKey Delta Service Info")
				fmt.Println(stateKey)

				// ServiceId
				serviceId, err := parseServiceIdFromType2(stateKey)
				if err != nil {
					return state, err
				}
				fmt.Println("ServiceId:", serviceId)

				// Decode the value
				serviceInfo, err := decodeServiceInfo(stateVal)
				if err != nil {
					return state, err
				}

				serviceAccount, exists := state.Delta[serviceId]

				if !exists {
					fmt.Println("Service account does not exist, creating a new one")
					serviceAccount = types.ServiceAccount{
						ServiceInfo:    serviceInfo,
						PreimageLookup: make(types.PreimagesMapEntry),
						LookupDict:     make(types.LookupMetaMapEntry),
						StorageDict:    make(types.Storage),
					}
				} else {
					fmt.Println("Service account exists, updating the service info")
					serviceAccount.ServiceInfo = serviceInfo
				}

				// Assign the service account back to the state
				state.Delta[serviceId] = serviceAccount
				delete(unmatchedStateKeyVals, stateKey)
			}

			if isPreimage(stateKey, stateVal) {
				fmt.Println("StateKey Delta Preimage")
				fmt.Println(stateKey)

				// ServiceId, PreimageKey, PreimageValue

				serviceId, err := parseServiceIdFromType3(stateKey)
				if err != nil {
					return state, err
				}

				fmt.Println("ServiceId:", serviceId)

				// PreimageValue
				preimageValue := stateVal

				// PreimageKey
				preimageKey := hash.Blake2bHash(preimageValue)

				// Assign the preimage value to the service account
				serviceAccount, exists := state.Delta[serviceId]

				// If the service account does not exist, create a new one
				// and initialize the preimage
				// Otherwise, update the existing service account

				if !exists {
					fmt.Println("Service account does not exist, creating a new one")
					serviceAccount = types.ServiceAccount{
						PreimageLookup: make(types.PreimagesMapEntry),
						LookupDict:     make(types.LookupMetaMapEntry),
						StorageDict:    make(types.Storage),
					}
				} else {
					fmt.Println("Service account exists, updating the preimage")
				}

				// Add a new preimage entry to the service account
				serviceAccount.PreimageLookup[preimageKey] = preimageValue

				// Assign the updated service account back to the state
				state.Delta[serviceId] = serviceAccount

				delete(unmatchedStateKeyVals, stateKey)
			}

			// INFO: Lookup 需要依賴 preimage 的資訊, Lookup 的 key = (preimage hash, preimage length)
			// 但我們無法確保 state key values 的排序
			// 有可能先遇到 lookup, 但此時 preimage 資訊還沒被解析
			// 因此, 需要先將所有的 state value 處理完, 獲得所有的 preimage 資訊,
			// 才能有效的反推出哪一個 preimage 對應到哪一個 lookup
		}
	}

	fmt.Println("Unmatched State Key Vals Size:", len(unmatchedStateKeyVals))

	// 獲得所有的 preimage 後, 即可製作出 lookup 的 key

	// 求出 lookup 的 key 之後, 再計算 lookup 的 state key
	// 比對 lookup 的 state key 是否與 input state key 相同
	// 若相同, decode input state value with LookupMetaMapEntry type
	// 將數值設定給 state.Delta[serviceId].LookupDict

	// Get all preimages
	// FIXME: preimage, lookup 應該是對應的, 數量一致？
	for serviceId, serviceAccount := range state.Delta {
		fmt.Println("ServiceId:", serviceId)

		for preimageKey, preimageValue := range serviceAccount.PreimageLookup {
			preimageStateKeyVal := encodeDelta3KeyVal(serviceId, preimageKey, preimageValue)

			// print preimage key with hex string
			fmt.Printf("Preimage Key: %x\n", preimageStateKeyVal.Key)

			// Lookup key = (preimage hash, preimage length)
			lookupKey := types.LookupMetaMapkey{
				Hash:   preimageKey,
				Length: types.U32(len(preimageValue)),
			}

			lookupStateKey := encodeDelta4KeyVal(serviceId, lookupKey, types.TimeSlotSet{})

			// find the lookup state key in unmatched state key vals
			if lookupStateVal, exists := unmatchedStateKeyVals[lookupStateKey.Key]; exists {
				fmt.Println("Lookup State Key Found:", lookupStateKey.Key)
				fmt.Println("Lookup State Value:", lookupStateVal)

				// Decode the lookup value => TimeSlotSet
				timeSlotSet := types.TimeSlotSet{}
				decoder := types.NewDecoder()
				err := decoder.Decode(lookupStateVal, &timeSlotSet)
				if err != nil {
					fmt.Println("Error decoding lookup state value:", err)
				}

				// Assign the lookup value to the service account
				serviceAccount.LookupDict[lookupKey] = timeSlotSet

				// Assign the updated service account back to the state
				state.Delta[serviceId] = serviceAccount

				delete(unmatchedStateKeyVals, lookupStateKey.Key)
			}
		}
	}

	fmt.Println("Unmatched State Key Vals Size after lookup:", len(unmatchedStateKeyVals))

	// 剩餘的 state key 就是 storage
	// storage value 可以 decode
	// 但我們無法有效推回 storage key
	// 只能知道 ServcieId 的 storage value

	// The remaining state keys are storage
	for stateKey, stateVal := range unmatchedStateKeyVals {
		serviceId, err := parseServiceIdFromType3(stateKey)
		if err != nil {
			fmt.Println("Error parsing service ID from state key:", err)
		}

		// We don't know the storage key, but we can assign the storage value to the service account
		serviceAccount, exists := state.Delta[serviceId]
		if !exists {
			fmt.Println("Service account does not exist, creating a new one for storage")
			serviceAccount = types.ServiceAccount{
				PreimageLookup: make(types.PreimagesMapEntry),
				LookupDict:     make(types.LookupMetaMapEntry),
				StorageDict:    make(types.Storage),
			}
		} else {
			fmt.Println("Service account exists, updating the storage")
		}

		// Assign the storage value to the service accoun
		storageKey := hash.Blake2bHash(stateVal)
		storageKeyString := fmt.Sprintf("%x", storageKey)

		serviceAccount.StorageDict[storageKeyString] = stateVal
		state.Delta[serviceId] = serviceAccount
		delete(unmatchedStateKeyVals, stateKey)

		// print storage key with hex string
		fmt.Printf("State Key: %x\n", stateKey)

		// print storage value
		fmt.Printf("Storage Value: %x\n", stateVal)
	}

	fmt.Println("Unmatched State Key Vals Size after storage:", len(unmatchedStateKeyVals))

	return state, err
}

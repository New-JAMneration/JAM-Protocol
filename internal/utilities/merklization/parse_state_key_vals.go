package merklization

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

func printStateKey(color string, stateKey types.StateKey) {
	cLog(color, fmt.Sprintf("State Key: 0x%x", stateKey))
}

func printStateValue(stateVal types.ByteSequence) {
	if len(stateVal) > 32 {
		cLog(Cyan, fmt.Sprintf("State Val: 0x%x...", stateVal[:32]))
	} else {
		cLog(Cyan, fmt.Sprintf("State Val: 0x%x", stateVal))
	}
}

func isPreimage(stateKey types.StateKey, stateVal types.ByteSequence) (bool, error) {
	var err error

	// The preimage value is a ByteSequence
	preimageValue := stateVal

	// Get ServiceId from state key
	serviceId, err := decodeServiceIdFromType3(stateKey)
	if err != nil {
		return false, fmt.Errorf("failed to parse service ID from state key: %w", err)
	}

	// Create preimage key (hash of the preimage value)
	preimageKey := hash.Blake2bHash(preimageValue)

	// Create a new state key using the serviceId, preimageKey, and preimageValue
	preimageStateKeyVal := encodeDelta3KeyVal(serviceId, preimageKey, preimageValue)

	isPreimage := preimageStateKeyVal.Key == stateKey

	return isPreimage, nil
}

func updateServiceInfo(state *types.State, serviceId types.ServiceId, serviceInfo types.ServiceInfo) {
	// Check if the service account exists
	serviceAccount, exists := state.Delta[serviceId]
	if !exists {
		serviceAccount = types.ServiceAccount{
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}
	}

	// Add or update the service info
	serviceAccount.ServiceInfo = serviceInfo

	// Assign the updated service account back to the state
	state.Delta[serviceId] = serviceAccount
}

func updatePreimage(state *types.State, serviceId types.ServiceId, preimageKey types.OpaqueHash, preimageValue types.ByteSequence) {
	// Check if the service account exists
	serviceAccount, exists := state.Delta[serviceId]
	if !exists {
		serviceAccount = types.ServiceAccount{
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}
	}

	// Add or update the preimage entry
	serviceAccount.PreimageLookup[preimageKey] = preimageValue

	// Assign the updated service account back to the state
	state.Delta[serviceId] = serviceAccount
}

func updateLookup(state *types.State, serviceId types.ServiceId, lookupKey types.LookupMetaMapkey, lookupValue types.TimeSlotSet) {
	// Check if the service account exists
	serviceAccount, exists := state.Delta[serviceId]
	if !exists {
		serviceAccount = types.ServiceAccount{
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}
	}

	// Add or update the lookup entry
	serviceAccount.LookupDict[lookupKey] = lookupValue

	// Assign the updated service account back to the state
	state.Delta[serviceId] = serviceAccount
}

func updateStorage(state *types.State, serviceId types.ServiceId, storageKey string, storageValue types.ByteSequence) {
	// Check if the service account exists
	serviceAccount, exists := state.Delta[serviceId]
	if !exists {
		serviceAccount = types.ServiceAccount{
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}
	}

	// Add or update the storage entry
	serviceAccount.StorageDict[storageKey] = storageValue

	// Assign the updated service account back to the state
	state.Delta[serviceId] = serviceAccount
}

func GetStateKeyValsDiff(a, b types.StateKeyVals) ([]types.StateKeyValDiff, error) {
	var diffs []types.StateKeyValDiff

	aMap := make(map[types.StateKey]types.ByteSequence)
	bMap := make(map[types.StateKey]types.ByteSequence)

	for _, kv := range a {
		aMap[kv.Key] = kv.Value
	}

	for _, kv := range b {
		bMap[kv.Key] = kv.Value
	}

	for key, valueA := range aMap {
		if valueB, exists := bMap[key]; exists {
			// the a key exists in b
			if !bytes.Equal(valueA, valueB) {
				diffs = append(diffs, types.StateKeyValDiff{
					Key:           key,
					ExpectedValue: valueA,
					ActualValue:   valueB,
				})
			}
			delete(bMap, key) // Remove from bMap to avoid false positives later
		} else {
			// the a key does not exist in b
			diffs = append(diffs, types.StateKeyValDiff{
				Key:           key,
				ExpectedValue: valueA,
				ActualValue:   nil,
			})
		}
	}

	// Remainig keys in bMap are those that were not in aMap
	for key, valueB := range bMap {
		diffs = append(diffs, types.StateKeyValDiff{
			Key:           key,
			ExpectedValue: nil,
			ActualValue:   valueB,
		})
	}

	return diffs, nil
}

func GetStateKeyValsDiffOld(expectedKeyVals, actualKeyVals types.StateKeyVals) ([]types.StateKeyValDiff, error) {
	// if the value is different, then return the key and the two values
	var diffs []types.StateKeyValDiff

	// convert statekeyvals to map for easy lookup
	expectedMap := make(map[types.StateKey]types.ByteSequence)
	actualMap := make(map[types.StateKey]types.ByteSequence)

	for _, kv := range expectedKeyVals {
		expectedMap[kv.Key] = kv.Value
	}

	for _, kv := range actualKeyVals {
		actualMap[kv.Key] = kv.Value
	}

	for expectedKey, expectedValue := range expectedMap {
		if actualValue, exists := actualMap[expectedKey]; exists {
			// Key exists in both expected and actual
			if !reflect.DeepEqual(expectedValue, actualValue) {
				diffs = append(diffs, types.StateKeyValDiff{
					Key:           expectedKey,
					ExpectedValue: expectedValue,
					ActualValue:   actualValue,
				})
			}
		} else {
			// Key exists in expected but not in actual
			diffs = append(diffs, types.StateKeyValDiff{
				Key:           expectedKey,
				ExpectedValue: expectedValue,
				ActualValue:   nil,
			})
		}
	}

	return diffs, nil
}

func StateKeyValsToState(stateKeyVals types.StateKeyVals) (types.State, error) {
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
			// Decode the alpha
			cLog(Yellow, "[C(1)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			alpha, err := decodeAlpha(stateVal)
			if err != nil {
				return state, err
			}
			state.Alpha = alpha
			delete(unmatchedStateKeyVals, stateKey)
		case C(2):
			// Decode the varphi
			cLog(Yellow, "[C(2)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			varphi, err := decodeVarphi(stateVal)
			if err != nil {
				return state, err
			}
			state.Varphi = varphi
			delete(unmatchedStateKeyVals, stateKey)
		case C(3):
			// Decode the beta
			cLog(Yellow, "[C(3)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			beta, err := decodeBeta(stateVal)
			if err != nil {
				return state, err
			}
			state.Beta = beta
			delete(unmatchedStateKeyVals, stateKey)
		case C(4):
			// Decode the gamma
			cLog(Yellow, "[C(4)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			gamma, err := decodeGamma(stateVal)
			if err != nil {
				return state, err
			}
			state.Gamma = gamma
			delete(unmatchedStateKeyVals, stateKey)
		case C(5):
			// Decode the psi
			cLog(Yellow, "[C(5)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			psi, err := decodePsi(stateVal)
			if err != nil {
				return state, err
			}
			state.Psi = psi
			delete(unmatchedStateKeyVals, stateKey)
		case C(6):
			// Decode the eta
			cLog(Yellow, "[C(6)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			eta, err := decodeEta(stateVal)
			if err != nil {
				return state, err
			}
			state.Eta = eta
			delete(unmatchedStateKeyVals, stateKey)
		case C(7):
			// Decode the iota
			cLog(Yellow, "[C(7)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			iota, err := decodeIota(stateVal)
			if err != nil {
				return state, err
			}
			state.Iota = iota
			delete(unmatchedStateKeyVals, stateKey)
		case C(8):
			// Decode the kappa
			cLog(Yellow, "[C(8)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			kappa, err := decodeKappa(stateVal)
			if err != nil {
				return state, err
			}
			state.Kappa = kappa
			delete(unmatchedStateKeyVals, stateKey)
		case C(9):
			// Decode the lambda
			cLog(Yellow, "[C(9)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			lambda, err := decodeLambda(stateVal)
			if err != nil {
				return state, err
			}
			state.Lambda = lambda
			delete(unmatchedStateKeyVals, stateKey)
		case C(10):
			// Decode the rho
			cLog(Yellow, "[C(10)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			rho, err := decodeRho(stateVal)
			if err != nil {
				return state, err
			}
			state.Rho = rho
			delete(unmatchedStateKeyVals, stateKey)
		case C(11):
			// Decode the tau
			cLog(Yellow, "[C(11)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			tau, err := decodeTau(stateVal)
			if err != nil {
				return state, err
			}
			state.Tau = tau
			delete(unmatchedStateKeyVals, stateKey)
		case C(12):
			// Decode the chi
			cLog(Yellow, "[C(12)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			chi, err := decodeChi(stateVal)
			if err != nil {
				return state, err
			}
			state.Chi = chi
			delete(unmatchedStateKeyVals, stateKey)
		case C(13):
			// Decode the pi
			cLog(Yellow, "[C(13)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			pi, err := decodePi(stateVal)
			if err != nil {
				return state, err
			}
			state.Pi = pi
			delete(unmatchedStateKeyVals, stateKey)
		case C(14):
			// Decode the theta
			cLog(Yellow, "[C(14)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			theta, err := decodeTheta(stateVal)
			if err != nil {
				return state, err
			}
			state.Theta = theta
			delete(unmatchedStateKeyVals, stateKey)
		case C(15):
			// Decode the xi
			cLog(Yellow, "[C(15)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			xi, err := decodeXi(stateVal)
			if err != nil {
				return state, err
			}
			state.Xi = xi
			delete(unmatchedStateKeyVals, stateKey)
		case C(16):
			// Decode the theta AccumulatedServiceOutput
			cLog(Yellow, "[C(16)]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			lastAccOut, err := decodeThetaAccOut(stateVal)
			if err != nil {
				return state, err
			}
			state.LastAccOut = lastAccOut
			delete(unmatchedStateKeyVals, stateKey)
		default:
			// C(255, s)
			isServiceInfo := stateKey[0] == 0xFF
			if isServiceInfo {
				cLog(Yellow, "[ServiceInfo]")
				printStateKey(Cyan, stateKey)
				printStateValue(stateVal)

				// ServiceId
				serviceId, err := decodeServiceIdFromType2(stateKey)
				if err != nil {
					return state, err
				}

				// Decode the value
				serviceInfo, err := decodeServiceInfo(stateVal)
				if err != nil {
					return state, err
				}

				// Update the service info in the state
				updateServiceInfo(&state, serviceId, serviceInfo)

				delete(unmatchedStateKeyVals, stateKey)
			}

			isPreimage, err := isPreimage(stateKey, stateVal)
			if err != nil {
				return state, err
			}

			if isPreimage {
				cLog(Yellow, "[Preimage]")
				printStateKey(Cyan, stateKey)
				printStateValue(stateVal)

				// ServiceId
				serviceId, err := decodeServiceIdFromType3(stateKey)
				if err != nil {
					return state, err
				}

				// PreimageValue
				preimageValue := stateVal

				// PreimageKey
				preimageKey := hash.Blake2bHash(preimageValue)

				// Update the preimage in the state
				updatePreimage(&state, serviceId, preimageKey, preimageValue)

				delete(unmatchedStateKeyVals, stateKey)
			}
		} // End of switch
	} // End of for loop

	// INFO: Lookup depends on preimage information. The key for Lookup is (preimage hash, preimage length).
	// However, we cannot guarantee the order of state key values.
	// It's possible to encounter a Lookup entry first, before its corresponding preimage information has been parsed.
	// Therefore, we need to process all state values first to obtain all preimage information,
	// so that after obtaining all preimages, we can construct the corresponding Lookup keys.

	// After updating the preimages, we can now process the Lookup entries.
	for serviceId, serviceAccount := range state.Delta {
		for preimageKey, preimageValue := range serviceAccount.PreimageLookup {
			cLog(Yellow, "[Lookup]")

			// Lookup key = (preimage hash, preimage length)
			lookupKey := types.LookupMetaMapkey{
				Hash:   preimageKey,
				Length: types.U32(len(preimageValue)),
			}

			// Create the lookup state key
			lookupStateKeyVal := encodeDelta4KeyVal(serviceId, lookupKey, types.TimeSlotSet{})
			lookupStateKey := lookupStateKeyVal.Key

			// Find the lookup state key in unmatchedStateKeyVals
			// If it exists, imply that we have a lookup entry for this preimage
			if lookupStateVal, exists := unmatchedStateKeyVals[lookupStateKey]; exists {
				printStateKey(Cyan, lookupStateKey)
				printStateValue(lookupStateVal)
				// Decode the lookup state value
				timeSlotSet := types.TimeSlotSet{}
				decoder := types.NewDecoder()
				err := decoder.Decode(lookupStateVal, &timeSlotSet)
				if err != nil {
					return state, fmt.Errorf("failed to decode lookup value: %w", err)
				}

				updateLookup(&state, serviceId, lookupKey, timeSlotSet)

				delete(unmatchedStateKeyVals, lookupStateKey)
			}
		}
	}

	// The remaining state keys are storage
	// FIXME: We cannot obtain the storage key from StateKeyVal.
	for stateKey, stateVal := range unmatchedStateKeyVals {
		cLog(Yellow, "[Storage]")
		serviceId, err := decodeServiceIdFromType3(stateKey)
		if err != nil {
			return state, fmt.Errorf("failed to parse service ID from state key: %w", err)
		}

		storageKey := fmt.Sprintf("%x", hash.Blake2bHash(stateVal))
		storageVal := stateVal

		storageStateKeyVal := encodeDelta2KeyVal(serviceId, types.ByteSequence(storageKey), storageVal)
		storageStateKey := storageStateKeyVal.Key

		if stateKey == storageStateKey {
			updateStorage(&state, serviceId, storageKey, storageVal)
			delete(unmatchedStateKeyVals, stateKey)
		} else {
			cLog(Red, fmt.Sprintf("Storage State Key mismatch: expected 0x%x, got 0x%x", stateKey, storageStateKey))
		}
	}

	cLog(Red, "--------------------")
	cLog(Red, "Unmatched State Keys:")
	for stateKey := range unmatchedStateKeyVals {
		printStateKey(Red, stateKey)
	}
	cLog(Red, "--------------------")

	return state, err
}

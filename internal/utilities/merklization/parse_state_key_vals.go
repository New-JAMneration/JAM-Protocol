package merklization

import (
	"bytes"
	"errors"
	"fmt"

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

func isServiceInfoKey(stateKey types.StateKey) bool {
	if stateKey[0] != 0xFF {
		return false
	}
	for i := 1; i < len(stateKey); i++ {
		if i == 1 || i == 3 || i == 5 || i == 7 {
			continue
		}
		if stateKey[i] != 0 {
			return false
		}
	}
	return true
}

func SingleKeyValToState(stateKey types.StateKey, stateVal types.ByteSequence) (interface{}, error) {
	switch stateKey {
	case C(1):
		// Decode the alpha
		cLog(Yellow, "[C(1)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		alpha, err := decodeAlpha(stateVal)
		if err != nil {
			return nil, err
		}
		return alpha, nil
	case C(2):
		// Decode the varphi
		cLog(Yellow, "[C(2)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		varphi, err := decodeVarphi(stateVal)
		if err != nil {
			return nil, err
		}
		return varphi, nil
	case C(3):
		// Decode the beta
		cLog(Yellow, "[C(3)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		beta, err := decodeBeta(stateVal)
		if err != nil {
			return nil, err
		}
		return beta, nil
	case C(4):
		// Decode the gamma
		cLog(Yellow, "[C(4)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		gamma, err := decodeGamma(stateVal)
		if err != nil {
			return nil, err
		}
		return gamma, nil
	case C(5):
		// Decode the psi
		cLog(Yellow, "[C(5)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		psi, err := decodePsi(stateVal)
		if err != nil {
			return nil, err
		}
		return psi, nil
	case C(6):
		// Decode the eta
		cLog(Yellow, "[C(6)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		eta, err := decodeEta(stateVal)
		if err != nil {
			return nil, err
		}
		return eta, nil
	case C(7):
		// Decode the iota
		cLog(Yellow, "[C(7)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		iota, err := decodeIota(stateVal)
		if err != nil {
			return nil, err
		}
		return iota, nil
	case C(8):
		// Decode the kappa
		cLog(Yellow, "[C(8)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		kappa, err := decodeKappa(stateVal)
		if err != nil {
			return nil, err
		}
		return kappa, nil
	case C(9):
		// Decode the lambda
		cLog(Yellow, "[C(9)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		lambda, err := decodeLambda(stateVal)
		if err != nil {
			return nil, err
		}
		return lambda, nil
	case C(10):
		// Decode the rho
		cLog(Yellow, "[C(10)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		rho, err := decodeRho(stateVal)
		if err != nil {
			return nil, err
		}
		return rho, nil
	case C(11):
		// Decode the tau
		cLog(Yellow, "[C(11)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		tau, err := decodeTau(stateVal)
		if err != nil {
			return nil, err
		}
		return tau, nil
	case C(12):
		// Decode the chi
		cLog(Yellow, "[C(12)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		chi, err := decodeChi(stateVal)
		if err != nil {
			return nil, err
		}
		return chi, nil
	case C(13):
		// Decode the pi
		cLog(Yellow, "[C(13)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		pi, err := decodePi(stateVal)
		if err != nil {
			return nil, err
		}
		return pi, nil
	case C(14):
		// Decode the theta
		cLog(Yellow, "[C(14)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		theta, err := decodeTheta(stateVal)
		if err != nil {
			return nil, err
		}
		return theta, nil
	case C(15):
		// Decode the xi
		cLog(Yellow, "[C(15)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		xi, err := decodeXi(stateVal)
		if err != nil {
			return nil, err
		}
		return xi, nil
	case C(16):
		// Decode the theta AccumulatedServiceOutput
		cLog(Yellow, "[C(16)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		lastAccOut, err := decodeThetaAccOut(stateVal)
		if err != nil {
			return nil, err
		}
		return lastAccOut, nil
	default:
		// C(255, s)
		isServiceInfo := isServiceInfoKey(stateKey)
		if isServiceInfo {
			cLog(Yellow, "[ServiceInfo]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			services := types.ServiceAccountState{}
			// ServiceId
			serviceId, err := DecodeServiceIdFromType2(stateKey)
			if err != nil {
				return services, err
			}

			// Decode the value
			serviceInfo, err := decodeServiceInfo(stateVal)
			if err != nil {
				return services, err
			}

			service := types.ServiceAccount{}
			service.ServiceInfo = serviceInfo
			services[serviceId] = service
			return service, err
		}

	}

	return nil, errors.New("unsupported stat-key")
}

func StateKeyValsToState(stateKeyVals types.StateKeyVals) (types.State, types.StateKeyVals, types.StateKeyVals, error) {
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
				return state, nil, nil, fmt.Errorf("failed to decode alpha: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode varphi: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode beta: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode gamma: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode psi: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode eta: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode iota: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode kappa: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode lambda: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode rho: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode tau: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode chi: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode pi: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode theta: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode xi: %w", err)
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
				return state, nil, nil, fmt.Errorf("failed to decode theta: %w", err)
			}
			state.LastAccOut = lastAccOut
			delete(unmatchedStateKeyVals, stateKey)
		default:
			// C(255, s)
			isServiceInfo := isServiceInfoKey(stateKey)
			if isServiceInfo {
				cLog(Yellow, "[ServiceInfo]")
				printStateKey(Cyan, stateKey)
				printStateValue(stateVal)

				// ServiceId
				serviceId, err := DecodeServiceIdFromType2(stateKey)
				if err != nil {
					return state, nil, nil, fmt.Errorf("failed to decode service ID: %w", err)
				}
				// Decode the value
				serviceInfo, err := decodeServiceInfo(stateVal)
				if err != nil {
					return state, nil, nil, fmt.Errorf("failed to decode service info of service ID %d: %w", serviceId, err)
				}
				// Update the service info in the state
				updateServiceInfo(&state, serviceId, serviceInfo)
				delete(unmatchedStateKeyVals, stateKey)
				continue
			}

			isPreimage, err := isPreimage(stateKey, stateVal)
			if err != nil {
				// This error should not happen
				// As the state key passed here should be service-related
				return state, nil, nil, fmt.Errorf("failed to check if preimage: %w", err)
			}

			if isPreimage {
				cLog(Yellow, "[Preimage]")
				printStateKey(Cyan, stateKey)
				printStateValue(stateVal)

				// ServiceId
				serviceId, err := decodeServiceIdFromType3(stateKey)
				if err != nil {
					return state, nil, nil, fmt.Errorf("failed to decode service ID: %w", err)
				}

				// PreimageValue
				preimageValue := stateVal

				// PreimageKey
				preimageKey := hash.Blake2bHash(preimageValue)

				// Update the preimage in the state
				updatePreimage(&state, serviceId, preimageKey, preimageValue)
				delete(unmatchedStateKeyVals, stateKey)
				continue
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
					return state, nil, nil, fmt.Errorf("failed to decode lookup value: %w", err)
				}

				updateLookup(&state, serviceId, lookupKey, timeSlotSet)
				delete(unmatchedStateKeyVals, lookupStateKey)
			}
		}
	}

	// Find the lookup entries that do not have corresponding preimages
	// INFO: Now, we try to fine the lookup entries using the value directly (TimeSlotSet)
	// If there is a better way to do this, please update me.
	unmatchedLookupKeyVals := types.StateKeyVals{}
	cLog(Yellow, "[Lookup without Preimage]")
	for stateKey, stateVal := range unmatchedStateKeyVals {
		// Try to decode the state value as TimeSlotSet
		// The first byte is `01` (A timeslotset)
		// And this should be able to decode using TimeSlotSet decoder

		// Check the length of stateVal
		if len(stateVal) == 0 {
			// Not a lookup entry
			continue
		}

		// Check the first byte for valid TimeSlotSet types
		if stateVal[0] != 0x01 && stateVal[0] != 0x02 && stateVal[0] != 0x03 {
			// Not a lookup entry
			continue
		}

		timeSlotSet := types.TimeSlotSet{}
		decoder := types.NewDecoder()
		err := decoder.Decode(stateVal, &timeSlotSet)
		if err != nil {
			// Not a lookup entry
			continue
		}

		// It's a lookup entry
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)

		// Append to unmatchedLookupDictEntries
		unmatchedLookupKeyVals = append(unmatchedLookupKeyVals, types.StateKeyVal{
			Key:   stateKey,
			Value: stateVal,
		})

		// Remove from unmatchedStateKeyVals
		delete(unmatchedStateKeyVals, stateKey)
	}

	storageStateKeyVals := make(types.StateKeyVals, 0, len(unmatchedStateKeyVals))
	for stateKey, stateVal := range unmatchedStateKeyVals {
		storageStateKeyVals = append(storageStateKeyVals, types.StateKeyVal{
			Key:   stateKey,
			Value: stateVal,
		})
	}

	return state, storageStateKeyVals, unmatchedLookupKeyVals, err
}

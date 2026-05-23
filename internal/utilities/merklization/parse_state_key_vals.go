package merklization

import (
	"bytes"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/google/go-cmp/cmp"
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
		logger.Debugf("%s%s%s", color, string, Reset)
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

// IsPreimage classifies a state-key/value pair as a preimage-lookup (delta3)
// entry by checking the invariant key == NewPreimageLookupStateKey(serviceID, H(value)).
// The computed preimage hash is returned so the caller can reuse it instead
// of hashing the value a second time.
func IsPreimage(stateKey types.StateKey, stateVal types.ByteSequence) (bool, types.OpaqueHash, error) {
	serviceID, err := DecodeServiceIDFromType3(stateKey)
	if err != nil {
		return false, types.OpaqueHash{}, fmt.Errorf("failed to parse service ID from state key: %w", err)
	}

	preimageHash := hash.Blake2bHash(stateVal)
	preimageStateKey, err := NewPreimageLookupStateKey(serviceID, preimageHash)
	if err != nil {
		return false, types.OpaqueHash{}, err
	}

	return preimageStateKey == stateKey, preimageHash, nil
}

// ensureServiceAccount returns a mutable reference-style accessor for the
// ServiceAccount in state.Delta[serviceID], lazily allocating state.Delta
// itself and the ServiceAccount when needed. The deserialization path uses
// struct literals (not NewServiceAccount) because the maps inside are
// initialized on demand below as we walk through each bucket.
func ensureServiceAccount(state *types.State, serviceID types.ServiceID) types.ServiceAccount {
	if state.Delta == nil {
		state.Delta = make(types.ServiceAccountState)
	}
	if existing, ok := state.Delta[serviceID]; ok {
		return existing
	}
	return types.ServiceAccount{}
}

func GetStateKeyValsDiff(a, b types.StateKeyVals) ([]types.StateKeyValDiff, error) {
	diffs := make([]types.StateKeyValDiff, 0, len(a)+len(b))

	aMap := make(map[types.StateKey]types.ByteSequence, len(a))
	bMap := make(map[types.StateKey]types.ByteSequence, len(b))

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

func IsServiceInfoKey(stateKey types.StateKey) bool {
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

// Decode a single state key-value pair to a state struct (except Type 3)
func SingleKeyValToState(stateKey types.StateKey, stateVal types.ByteSequence) (interface{}, error) {
	switch stateKey {
	case C(1):
		// Decode the alpha
		cLog(Yellow, "[C(1)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		alpha, err := decodeAlpha(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode alpha: %w", err)
		}
		return alpha, nil
	case C(2):
		// Decode the varphi
		cLog(Yellow, "[C(2)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		varphi, err := decodeVarphi(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode varphi: %w", err)
		}
		return varphi, nil
	case C(3):
		// Decode the beta
		cLog(Yellow, "[C(3)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		beta, err := decodeBeta(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode beta: %w", err)
		}
		return beta, nil
	case C(4):
		// Decode the gamma
		cLog(Yellow, "[C(4)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		gamma, err := decodeGamma(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gamma: %w", err)
		}
		return gamma, nil
	case C(5):
		// Decode the psi
		cLog(Yellow, "[C(5)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		psi, err := decodePsi(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode psi: %w", err)
		}
		return psi, nil
	case C(6):
		// Decode the eta
		cLog(Yellow, "[C(6)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		eta, err := decodeEta(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode eta: %w", err)
		}
		return eta, nil
	case C(7):
		// Decode the iota
		cLog(Yellow, "[C(7)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		iota, err := decodeIota(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode iota: %w", err)
		}
		return iota, nil
	case C(8):
		// Decode the kappa
		cLog(Yellow, "[C(8)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		kappa, err := decodeKappa(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode kappa: %w", err)
		}
		return kappa, nil
	case C(9):
		// Decode the lambda
		cLog(Yellow, "[C(9)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		lambda, err := decodeLambda(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode lambda: %w", err)
		}
		return lambda, nil
	case C(10):
		// Decode the rho
		cLog(Yellow, "[C(10)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		rho, err := decodeRho(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode rho: %w", err)
		}
		return rho, nil
	case C(11):
		// Decode the tau
		cLog(Yellow, "[C(11)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		tau, err := decodeTau(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tau: %w", err)
		}
		return tau, nil
	case C(12):
		// Decode the chi
		cLog(Yellow, "[C(12)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		chi, err := decodeChi(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode chi: %w", err)
		}
		return chi, nil
	case C(13):
		// Decode the pi
		cLog(Yellow, "[C(13)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		pi, err := decodePi(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pi: %w", err)
		}
		return pi, nil
	case C(14):
		// Decode the vartheta
		cLog(Yellow, "[C(14)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		vartheta, err := decodeVartheta(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vartheta: %w", err)
		}
		return vartheta, nil
	case C(15):
		// Decode the xi
		cLog(Yellow, "[C(15)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		xi, err := decodeXi(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode xi: %w", err)
		}
		return xi, nil
	case C(16):
		// Decode the theta AccumulatedServiceOutput
		cLog(Yellow, "[C(16)]")
		printStateKey(Cyan, stateKey)
		printStateValue(stateVal)
		theta, err := decodeTheta(stateVal)
		if err != nil {
			return nil, fmt.Errorf("failed to decode theta: %w", err)
		}
		return theta, nil
	default:
		// C(255, s)
		if IsServiceInfoKey(stateKey) {
			cLog(Yellow, "[ServiceInfo]")
			printStateKey(Cyan, stateKey)
			printStateValue(stateVal)
			delta := types.ServiceAccountState{}
			// ServiceID
			serviceID, err := DecodeServiceIDFromType2(stateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decode service ID: %w", err)
			}

			// Decode the value
			serviceInfo, err := DecodeServiceInfo(stateVal)
			if err != nil {
				return nil, fmt.Errorf("failed to decode service info: %w", err)
			}

			service := types.ServiceAccount{}
			service.ServiceInfo = serviceInfo
			delta[serviceID] = service
			return delta, nil
		}
	}

	return nil, fmt.Errorf("unsupported state-key: 0x%x", stateKey)
}

// StateKeyValsToState rebuilds an in-memory State from its serialized
// (StateKey, value) pairs in a single pass over the input.
//
// After the globalKV refactor (Method A) the storage (a_s) and preimage-meta
// (a_l) entries live together in account.globalKV — they no longer need to
// be stored in a separate "unmatched" fallback pool. The function still
// returns an empty types.StateKeyVals for source compatibility with existing
// callers; Step 7.5 will drop that second return value entirely.
//
// Dispatch (one pass, three buckets):
//
//  1. Chapter keys C(1)..C(16)               → typed state fields.
//  2. Service-info keys C(255, s)            → ServiceAccount.ServiceInfo;
//                                              the a_i / a_o counters are
//                                              initialised from the encoded
//                                              ServiceInfo.Items / Bytes so
//                                              we never have to recompute
//                                              them by walking globalKV.
//  3. IsPreimage(stateKey, value) == true    → ServiceAccount.PreimageLookup
//                                              (a_p blobs stay in their own
//                                              map; the hash returned by
//                                              IsPreimage is reused so we
//                                              don't hash the value twice).
//  4. Everything else (delta2 + delta4)      → ServiceAccount.globalKV.
//
// state.Delta, PreimageLookup, and globalKV are all lazy-initialised on
// first use so callers may feed in inputs that don't include every chapter.
func StateKeyValsToState(stateKeyVals types.StateKeyVals) (types.State, types.StateKeyVals, error) {
	state := types.State{
		Delta: make(types.ServiceAccountState),
	}

	for _, keyVal := range stateKeyVals {
		stateKey := keyVal.Key
		stateVal := keyVal.Value

		switch stateKey {
		case C(1):
			cLog(Yellow, "[C(1)]")
			alpha, err := decodeAlpha(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode alpha: %w", err)
			}
			state.Alpha = alpha
		case C(2):
			cLog(Yellow, "[C(2)]")
			varphi, err := decodeVarphi(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode varphi: %w", err)
			}
			state.Varphi = varphi
		case C(3):
			cLog(Yellow, "[C(3)]")
			beta, err := decodeBeta(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode beta: %w", err)
			}
			state.Beta = beta
		case C(4):
			cLog(Yellow, "[C(4)]")
			gamma, err := decodeGamma(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode gamma: %w", err)
			}
			state.Gamma = gamma
		case C(5):
			cLog(Yellow, "[C(5)]")
			psi, err := decodePsi(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode psi: %w", err)
			}
			state.Psi = psi
		case C(6):
			cLog(Yellow, "[C(6)]")
			eta, err := decodeEta(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode eta: %w", err)
			}
			state.Eta = eta
		case C(7):
			cLog(Yellow, "[C(7)]")
			iota, err := decodeIota(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode iota: %w", err)
			}
			state.Iota = iota
		case C(8):
			cLog(Yellow, "[C(8)]")
			kappa, err := decodeKappa(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode kappa: %w", err)
			}
			state.Kappa = kappa
		case C(9):
			cLog(Yellow, "[C(9)]")
			lambda, err := decodeLambda(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode lambda: %w", err)
			}
			state.Lambda = lambda
		case C(10):
			cLog(Yellow, "[C(10)]")
			rho, err := decodeRho(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode rho: %w", err)
			}
			state.Rho = rho
		case C(11):
			cLog(Yellow, "[C(11)]")
			tau, err := decodeTau(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode tau: %w", err)
			}
			state.Tau = tau
		case C(12):
			cLog(Yellow, "[C(12)]")
			chi, err := decodeChi(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode chi: %w", err)
			}
			state.Chi = chi
		case C(13):
			cLog(Yellow, "[C(13)]")
			pi, err := decodePi(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode pi: %w", err)
			}
			state.Pi = pi
		case C(14):
			cLog(Yellow, "[C(14)]")
			vartheta, err := decodeVartheta(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode vartheta: %w", err)
			}
			state.Vartheta = vartheta
		case C(15):
			cLog(Yellow, "[C(15)]")
			xi, err := decodeXi(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode xi: %w", err)
			}
			state.Xi = xi
		case C(16):
			cLog(Yellow, "[C(16)]")
			theta, err := decodeTheta(stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode theta: %w", err)
			}
			state.Theta = theta
		default:
			// C(255, s) — service info (delta1).
			if IsServiceInfoKey(stateKey) {
				cLog(Yellow, "[ServiceInfo]")
				serviceID, err := DecodeServiceIDFromType2(stateKey)
				if err != nil {
					return state, nil, fmt.Errorf("failed to decode service ID: %w", err)
				}
				serviceInfo, err := DecodeServiceInfo(stateVal)
				if err != nil {
					return state, nil, fmt.Errorf("failed to decode service info of service ID %d: %w", serviceID, err)
				}
				sa := ensureServiceAccount(&state, serviceID)
				sa.ServiceInfo = serviceInfo
				// Initialise counters from the wire-format mirror fields so
				// we don't have to recompute them by walking globalKV. Once
				// Step 8 introduces the codec-only struct the values come
				// straight from the decoder.
				sa.SetTotalNumberOfItems(uint32(serviceInfo.Items))
				sa.SetTotalNumberOfOctets(uint64(serviceInfo.Bytes))
				state.Delta[serviceID] = sa
				continue
			}

			// Type-3 key (storage, preimage-lookup, or preimage-meta).
			serviceID, err := DecodeServiceIDFromType3(stateKey)
			if err != nil {
				return state, nil, fmt.Errorf("failed to decode service ID: %w", err)
			}

			isPreimage, preimageHash, err := IsPreimage(stateKey, stateVal)
			if err != nil {
				return state, nil, fmt.Errorf("failed to classify state-key: %w", err)
			}
			if isPreimage {
				cLog(Yellow, "[Preimage]")
				sa := ensureServiceAccount(&state, serviceID)
				if sa.PreimageLookup == nil {
					sa.PreimageLookup = make(types.PreimagesMapEntry)
				}
				sa.PreimageLookup[preimageHash] = stateVal
				state.Delta[serviceID] = sa
				continue
			}

			// delta2 (storage) or delta4 (preimage meta) — both go into
			// globalKV. Distinguishing the two would require structural
			// knowledge that StateKey alone cannot provide, but the trie
			// representation does not need to distinguish them either.
			cLog(Yellow, "[globalKV]")
			sa := ensureServiceAccount(&state, serviceID)
			kv := sa.GetGlobalKVItems()
			if kv == nil {
				kv = make(map[types.StateKey][]byte)
			}
			kv[stateKey] = stateVal
			sa.SetGlobalKVItems(kv)
			state.Delta[serviceID] = sa
		} // End of switch
	} // End of for loop

	// Method A: storage / preimage-meta now live in globalKV, the fallback
	// pool is unused. Return an empty slice for source compatibility; the
	// second return value will be dropped in Step 7.5.
	return state, types.StateKeyVals{}, nil
}

func DebugStateKeyValsDiff(diffs []types.StateKeyValDiff) error {
	LookupStatekeyVal := make(map[types.StateKey]types.LookupMetaMapkey)
	unmatchedStateKeyVals := make([]types.StateKeyValDiff, 0, len(diffs))

	for _, v := range diffs {
		if v.ActualValue == nil && v.ExpectedValue == nil {
			logger.ColorDebug("ignore state key 0x%x because both actual and expected are nil", v.Key)
			continue
		}
		if state, keyExists := types.KeyValMap[v.Key]; keyExists { // Type 1: C(1)-C(16)
			// logger.ColorYellow("state: %s, key: %v", state, v.Key)
			// detailed state struct, except service-related
			act, err := SingleKeyValToState(v.Key, v.ActualValue)
			if err != nil {
				return fmt.Errorf("failed to SingleKeyValToState actual from state key 0x%x: %w", v.Key, err)
			}
			exp, err := SingleKeyValToState(v.Key, v.ExpectedValue)
			if err != nil {
				return fmt.Errorf("failed to SingleKeyValToState expected from state key 0x%x: %w", v.Key, err)
			}
			diff := cmp.Diff(exp, act)
			logger.ColorDebug("state %s exp/act diff: %+v", state, diff)
			continue
		} else if IsServiceInfoKey(v.Key) { // Type 2: C(255, s), ServiceInfo
			serviceID, err := DecodeServiceIDFromType2(v.Key)
			if err != nil {
				return fmt.Errorf("failed to decode service ID from state key 0x%x by type 2: %w", v.Key, err)
			}
			act, err := DecodeServiceInfo(v.ActualValue)
			if err != nil {
				return fmt.Errorf("failed to decode actual service info from state key 0x%x: %w", v.Key, err)
			}
			exp, err := DecodeServiceInfo(v.ExpectedValue)
			if err != nil {
				return fmt.Errorf("failed to decode expected service info from state key 0x%x: %w", v.Key, err)
			}
			diff := cmp.Diff(exp, act)
			logger.ColorDebug("ServiceID %d Info exp/act diff: %+v", serviceID, diff)
			continue
		} else { // Rest of the keys are service-related (type3)
			serviceID, err := DecodeServiceIDFromType3(v.Key)
			if err != nil {
				return fmt.Errorf("failed to decode service ID from state key 0x%x by type 3: %w", v.Key, err)
			}
			// a_p: Preimage
			if isPreimage, preimageKey, _ := IsPreimage(v.Key, v.ExpectedValue); isPreimage {
				logger.ColorDebug("serviceID %v state key 0x%x is preimage key 0x%x, value exp/act diff: %v", serviceID, v.Key, preimageKey, cmp.Diff(v.ExpectedValue, v.ActualValue))

				// Store the preimage and related lookup key for later lookup

				// Lookup key = (preimage hash, preimage length)
				lookupKey := types.LookupMetaMapkey{
					Hash:   preimageKey,
					Length: types.U32(len(v.ExpectedValue)),
				}

				// Create the lookup state key
				lookupStateKeyVal := EncodeDelta4KeyVal(serviceID, lookupKey, types.TimeSlotSet{})

				LookupStatekeyVal[lookupStateKeyVal.Key] = lookupKey
				continue
			}
			// Store the unmatched state key value (potentially Lookup or storage)
			unmatchedStateKeyVals = append(unmatchedStateKeyVals, v)

		} // End of preimage
	} // End of for _, v := range diffs

	for _, v := range unmatchedStateKeyVals { // a_l, a_s
		serviceID, err := DecodeServiceIDFromType3(v.Key)
		if err != nil {
			return fmt.Errorf("failed to decode service ID from state key 0x%x by type 3: %w", v.Key, err)
		}
		// a_l: Lookup
		// Find the lookup state key in unmatchedStateKeyVals
		// If it exists, imply that a lookup entry for this preimage
		if lookupKey, exists := LookupStatekeyVal[v.Key]; exists {
			logger.ColorDebug("serviceID %d state key 0x%x is Lookup entry hash 0x%x len %d, exp/act diff: %v",
				serviceID, v.Key, lookupKey.Hash, lookupKey.Length, cmp.Diff(v.ExpectedValue, v.ActualValue))
			// // Decode the lookup state value
			// timeSlotSet := types.TimeSlotSet{}
			// decoder := types.NewDecoder()
			// err := decoder.Decode(v.ExpectedValue, &timeSlotSet)
			// if err != nil {
			// 	return fmt.Errorf(": failed to decode lookup value: %w", err)
			// }
			continue
		}
		// a_s: Storage or unmatched lookup
		exp := v.ExpectedValue
		act := v.ActualValue
		if len(exp) > 8 {
			exp = exp[:8]
		}
		if len(act) > 8 {
			act = act[:8]
		}
		logger.ColorDebug("serviceID %d state key 0x%x is storage or unmatched lookup diff: %v", serviceID, v.Key, cmp.Diff(exp, act))
	}
	return nil
}

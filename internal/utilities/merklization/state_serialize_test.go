package merklization

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/google/go-cmp/cmp"
)

func hex2Bytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v\n", err)
	}
	return bytes
}

func parseServiceAccountInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})

	// separate by | symbol
	parts := strings.Split(info, "|")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// handle multiple key-value pairs
		subParts := strings.Split(part, " ")
		for _, subPart := range subParts {
			if subPart == "" {
				continue
			}

			// parse key-value pairs
			kv := strings.SplitN(subPart, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key := kv[0]
			value := kv[1]

			// handle different types of values
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				// handle array
				arrayStr := value[1 : len(value)-1]
				elements := strings.Split(arrayStr, ",")

				array := make([]interface{}, 0)
				for _, elem := range elements {
					elem = strings.TrimSpace(elem)
					if elem == "" {
						continue
					}

					// try to convert to number
					if num, err := strconv.Atoi(elem); err == nil {
						array = append(array, num)
					} else {
						array = append(array, elem)
					}
				}

				result[key] = array
			} else if val, err := strconv.Atoi(value); err == nil {
				// try to convert to number
				result[key] = val
			} else if strings.HasPrefix(value, "0x") {
				// handle hex value
				result[key] = value
			} else {
				// handle string
				result[key] = value
			}
		}
	}

	return result
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Test serialize alpha to xi only, delta(accounts) is not tested
func TestStateSerialize(t *testing.T) {
	directories := []string{
		"../../../pkg/test_data/jamtestnet/data/orderedaccumulation/state_snapshots",
		"../../../pkg/test_data/jamtestnet/data/assurances/state_snapshots",
		"../../../pkg/test_data/jamtestnet/data/fallback/state_snapshots",
		"../../../pkg/test_data/jamtestnet/data/safrole/state_snapshots",
	}

	for _, dir := range directories {
		jsonPattern := filepath.Join(dir, "*.json")
		binPattern := filepath.Join(dir, "*.bin")

		jsonFiles, err := filepath.Glob(jsonPattern)
		if err != nil {
			fmt.Println("Error finding JSON files:", err)
			continue
		}

		binFiles, err := filepath.Glob(binPattern)
		if err != nil {
			fmt.Println("Error finding BIN files:", err)
			continue
		}

		for _, jsonFile := range jsonFiles {
			binFile := strings.Replace(jsonFile, ".json", ".bin", 1)

			if !contains(binFiles, binFile) {
				fmt.Printf("BIN file not found for %s\n", jsonFile)
				continue
			}

			state, err := LoadStateFromFile(jsonFile)
			if err != nil {
				fmt.Printf("Error loading state from %s: %v\n", jsonFile, err)
				continue
			}

			data, err := readFile(binFile)
			if err != nil {
				fmt.Printf("Error reading BIN file %s: %v\n", binFile, err)
				continue
			}

			output := SerializeAlpha(state.Alpha)
			output = append(output, SerializeVarphi(state.Varphi)...)
			output = append(output, SerializeBeta(state.Beta)...)
			output = append(output, SerializeGamma(state.Gamma)...)
			output = append(output, SerializePsi(state.Psi)...)
			output = append(output, SerializeEta(state.Eta)...)
			output = append(output, SerializeIota(state.Iota)...)
			output = append(output, SerializeKappa(state.Kappa)...)
			output = append(output, SerializeLambda(state.Lambda)...)
			output = append(output, SerializeRho(state.Rho)...)
			output = append(output, SerializeTau(state.Tau)...)
			output = append(output, SerializeChi(state.Chi)...)
			output = append(output, SerializePi(state.Pi)...)
			output = append(output, SerializeTheta(state.Theta)...)
			output = append(output, SerializeXi(state.Xi)...)
			if !bytes.Equal(data[:len(output)], output[:]) {
				t.Error("serialize failed")
			}
			fmt.Println(jsonFile, "serialize success")
		}
	}
}

func LoadStateFromFile(filename string) (types.State, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.State{}, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return types.State{}, fmt.Errorf("failed to read file: %v", err)
	}

	var state types.State
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return types.State{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return state, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestJamTestNetStateEncodeAndSerialization(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet state test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		"fallback",
		"orderedaccumulation",
		"safrole",
	}

	for _, dirName := range dirNames {
		snapshotsDir := filepath.Join("../../../pkg/test_data/jamtestnet/data", dirName, "state_snapshots")

		files, err := utils.GetTargetExtensionFiles(snapshotsDir, utils.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			// Read the binary file
			binPath := filepath.Join(snapshotsDir, file)
			// data, err := utils.GetBytesFromFile(binPath)
			// if err != nil {
			// 	t.Errorf("failed to read file: %v", err)
			// }
			var state types.State
			err = utils.GetTestFromBin(binPath, &state)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Read the json file
			baseFileName := filepath.Base(binPath)
			baseFileName = baseFileName[:len(baseFileName)-len(utils.BIN_EXTENTION)]
			jsonFilePath := filepath.Join(snapshotsDir, baseFileName+utils.JSON_EXTENTION)
			jsonState, err := utils.GetTestFromJson(jsonFilePath, &types.State{})
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Compare the two structs -- run this to check both json and bin have same data
			// if !reflect.DeepEqual(state, jsonState) {
			// 	log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
			// 	t.Errorf("Error: %v", err)
			// } else {
			// 	log.Printf("✅ [%s] [%s] %s", types.TEST_MODE, dirName, file)
			// }
			// Compare the two serialized data
			var (
				serializedAlpha  = SerializeAlpha(jsonState.Alpha)
				serializedVarphi = SerializeVarphi(jsonState.Varphi)
				serializedBeta   = SerializeBeta(jsonState.Beta)
				serializedGamma  = SerializeGamma(jsonState.Gamma)
				serializedPsi    = SerializePsi(jsonState.Psi)
				serializedEta    = SerializeEta(jsonState.Eta)
				serializedIota   = SerializeIota(jsonState.Iota)
				serializedKappa  = SerializeKappa(jsonState.Kappa)
				serializedLambda = SerializeLambda(jsonState.Lambda)
				serializedRho    = SerializeRho(jsonState.Rho)
				serializedTau    = SerializeTau(jsonState.Tau)
				serializedChi    = SerializeChi(jsonState.Chi)
				serializedPi     = SerializePi(jsonState.Pi)
				serializedTheta  = SerializeTheta(jsonState.Theta)
				serializedXi     = SerializeXi(jsonState.Xi)
			)
			var (
				encodedAlpha  = EncodeAlpha(jsonState.Alpha)
				encodedVarphi = EncodeVarphi(jsonState.Varphi)
				encodedBeta   = EncodeBeta(jsonState.Beta)
				encodedGamma  = EncodeGamma(jsonState.Gamma)
				encodedPsi    = EncodePsi(jsonState.Psi)
				encodedEta    = EncodeEta(jsonState.Eta)
				encodedIota   = EncodeIota(jsonState.Iota)
				encodedKappa  = EncodeKappa(jsonState.Kappa)
				encodedLambda = EncodeLambda(jsonState.Lambda)
				encodedRho    = EncodeRho(jsonState.Rho)
				encodedTau    = EncodeTau(jsonState.Tau)
				encodedChi    = EncodeChi(jsonState.Chi)
				encodedPi     = EncodePi(jsonState.Pi)
				encodedTheta  = EncodeTheta(jsonState.Theta)
				encodedXi     = EncodeXi(jsonState.Xi)
			)
			for _, v := range jsonState.Delta {
				serializedDelta1 := SerializeDelta1(v)
				encodedDelta1 := EncodeDelta1(v)
				if !reflect.DeepEqual(encodedDelta1, serializedDelta1) {
					diff := cmp.Diff(encodedDelta1, serializedDelta1)
					t.Error(dirName, binPath, "serialize Delta1 failed", diff)
				}
			}
			if !reflect.DeepEqual(encodedAlpha, serializedAlpha) {
				t.Error(dirName, binPath, "serialize Alpha failed")
			} else if !reflect.DeepEqual(encodedVarphi, serializedVarphi) {
				t.Error(dirName, binPath, "serialize Varphi failed")
			} else if !reflect.DeepEqual(encodedBeta, serializedBeta) {
				t.Error(dirName, binPath, "serialize Beta failed")
			} else if !reflect.DeepEqual(encodedGamma, serializedGamma) {
				t.Error(dirName, binPath, "serialize Gamma failed")
			} else if !reflect.DeepEqual(encodedPsi, serializedPsi) {
				t.Error(dirName, binPath, "serialize Psi failed")
			} else if !reflect.DeepEqual(encodedEta, serializedEta) {
				t.Error(dirName, binPath, "serialize Eta failed")
			} else if !reflect.DeepEqual(encodedIota, serializedIota) {
				t.Error(dirName, binPath, "serialize Iota failed")
			} else if !reflect.DeepEqual(encodedKappa, serializedKappa) {
				t.Error(dirName, binPath, "serialize Kappa failed")
			} else if !reflect.DeepEqual(encodedLambda, serializedLambda) {
				t.Error(dirName, binPath, "serialize Lambda failed")
			} else if !reflect.DeepEqual(encodedRho, serializedRho) {
				diff := cmp.Diff(encodedRho, serializedRho)
				t.Error(dirName, binPath, "serialize Rho failed", diff)
			} else if !reflect.DeepEqual(encodedTau, serializedTau) {
				t.Error(dirName, binPath, "serialize Tau failed")
			} else if !reflect.DeepEqual(encodedChi, serializedChi) {
				t.Error(dirName, binPath, "serialize Chi failed")
			} else if !reflect.DeepEqual(encodedPi, serializedPi) {
				t.Log("Pi Services", jsonState.Pi.Services)
				diff := cmp.Diff(encodedPi, serializedPi)
				t.Error(dirName, binPath, "serialize Pi failed", diff)
			} else if !reflect.DeepEqual(encodedTheta, serializedTheta) {
				diff := cmp.Diff(encodedTheta, serializedTheta)
				t.Error(dirName, binPath, "serialize Theta failed", diff)
			} else if !reflect.DeepEqual(encodedXi, serializedXi) {
				diff := cmp.Diff(encodedXi, serializedXi)
				t.Error(dirName, binPath, "serialize Xi failed", diff)
			} else {
				t.Log(binPath, "serialize(before delta1) success")
			}

			// === (each state finish, test all state) ===

			encodedState, err := StateEncoder(jsonState)
			if err != nil {
				t.Error("encodedState raised error", err)
			}
			serializedState, err := StateSerialize(jsonState)
			if err != nil {
				t.Error("StateSerialize raised error", err)
			}
			if !reflect.DeepEqual(encodedState, serializedState) {
				diff := cmp.Diff(encodedState, serializedState)
				t.Error(dirName, binPath, "serialize State failed", diff)
			} else {
				t.Log(binPath, "state serialize success")
			}
		}

		// Reset the test mode
		if BACKUP_TEST_MODE == "tiny" {
			types.SetTinyMode()
		} else {
			types.SetFullMode()
		}
	}
}

func TestJamTestNetStateRoot(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet state test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		// "fallback",
		// "orderedaccumulation",
		// "safrole",
	}

	for _, dirName := range dirNames {
		snapshotsDir := filepath.Join("../../../pkg/test_data/jamtestnet/data", dirName, "state_snapshots")
		transitionsDir := filepath.Join("../../../pkg/test_data/jamtestnet/data", dirName, "state_transitions")

		snapshotsFiles, err := utils.GetTargetExtensionFiles(snapshotsDir, utils.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		transitionsFiles, err := utils.GetTargetExtensionFiles(transitionsDir, utils.JSON_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		// create a map to store the relationship between file names and paths
		fileMap := make(map[string]struct {
			snapshotPath   string
			transitionPath string
			hasSnapshot    bool
			hasTransition  bool
		})

		// process snapshots files
		for _, file := range snapshotsFiles {
			baseName := file[:len(file)-len(utils.BIN_EXTENTION)]
			entry, exists := fileMap[baseName]
			if !exists {
				entry = struct {
					snapshotPath   string
					transitionPath string
					hasSnapshot    bool
					hasTransition  bool
				}{}
			}
			entry.snapshotPath = filepath.Join(snapshotsDir, file)
			entry.hasSnapshot = true
			fileMap[baseName] = entry
		}

		// process transitions files
		for _, file := range transitionsFiles {
			if !strings.HasSuffix(file, utils.JSON_EXTENTION) {
				continue
			}
			baseName := file[:len(file)-len(utils.JSON_EXTENTION)]
			entry, exists := fileMap[baseName]
			if !exists {
				entry = struct {
					snapshotPath   string
					transitionPath string
					hasSnapshot    bool
					hasTransition  bool
				}{}
			}
			entry.transitionPath = filepath.Join(transitionsDir, file)
			entry.hasTransition = true
			fileMap[baseName] = entry
		}

		// sort the map by key
		var keys []string
		for k := range fileMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, baseName := range keys {
			entry := fileMap[baseName]
			t.Logf("Processing %s", baseName)
			if !entry.hasSnapshot || !entry.hasTransition {
				t.Errorf("Missing snapshot or transition for %s", baseName)
				continue
			}

			// Read the binary file
			var state types.State
			err = utils.GetTestFromBin(entry.snapshotPath, &state)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			var (
				encodedAlpha  = EncodeAlpha(state.Alpha)
				encodedVarphi = EncodeVarphi(state.Varphi)
				encodedBeta   = EncodeBeta(state.Beta)
				encodedGamma  = EncodeGamma(state.Gamma)
				encodedPsi    = EncodePsi(state.Psi)
				encodedEta    = EncodeEta(state.Eta)
				encodedIota   = EncodeIota(state.Iota)
				encodedKappa  = EncodeKappa(state.Kappa)
				encodedLambda = EncodeLambda(state.Lambda)
				encodedRho    = EncodeRho(state.Rho)
				encodedTau    = EncodeTau(state.Tau)
				encodedChi    = EncodeChi(state.Chi)
				encodedPi     = EncodePi(state.Pi)
				encodedTheta  = EncodeTheta(state.Theta)
				encodedXi     = EncodeXi(state.Xi)
			)

			// Read the json file
			jsonPath := entry.transitionPath

			type TransitionStateJson struct {
				PreState struct {
					StateRoot string     `json:"state_root"`
					KeyVals   [][]string `json:"keyvals"`
				} `json:"pre_state"`
				PostState struct {
					StateRoot string     `json:"state_root"`
					KeyVals   [][]string `json:"keyvals"`
				} `json:"post_state"`
				Block struct {
					Header struct {
						ParentHash string `json:"parent"`
					} `json:"header"`
				} `json:"block"`
			}

			jsonData, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Logf("failed to read file: %v", err)
				continue
			}

			var transitionState TransitionStateJson
			err = json.Unmarshal(jsonData, &transitionState)
			if err != nil {
				t.Logf("failed to parse JSON file: %v", err)
				continue
			}

			// === Compare keyVals ===
			for _, keyVal := range transitionState.PostState.KeyVals {
				if keyVal[2] == "account_lookup" {
					accountInfoMap := parseServiceAccountInfo(keyVal[3])
					// ∀(s ↦ a) ∈ δ
					if accountID, exist := accountInfoMap["s"]; exist {
						accountID := types.ServiceId(accountID.(int))
						if account, exists := state.Delta[accountID]; exists {
							// (h, l) ↦ t ∈ al
							if hashStr, exist := accountInfoMap["h"]; exist {
								if lengthValue, lengthExist := accountInfoMap["l"]; lengthExist {
									targetKey := types.LookupMetaMapkey{
										Hash:   types.OpaqueHash(hex2Bytes(hashStr.(string))),
										Length: types.U32(lengthValue.(int)),
									}

									if val, keyExists := account.LookupDict[targetKey]; keyExists {
										key16, delta4Output := EncodeDelta4KeyVal(accountID, account, targetKey, val)

										if !reflect.DeepEqual(types.OpaqueHash(hex2Bytes(keyVal[0])), key16) {
											diff := cmp.Diff(types.OpaqueHash(hex2Bytes(keyVal[0])), key16)
											t.Error(dirName, jsonPath, "key16 does not match", diff)
										}

										if !reflect.DeepEqual(types.ByteSequence(hex2Bytes(keyVal[1])), delta4Output) {
											diff := cmp.Diff(types.ByteSequence(hex2Bytes(keyVal[1])), delta4Output)
											t.Error(dirName, jsonPath, "delta4 value does not match", diff)
										}
									} else {
										t.Logf("LookupDict key with hash %s and length %d not found for account %v", targetKey.Hash, targetKey.Length, accountID)
									}
								}
							}
						}
					}
				}
				if keyVal[2] == "account_preimage" {
					accountInfoMap := parseServiceAccountInfo(keyVal[3])
					// ∀(s ↦ a) ∈ δ
					if accountID, exist := accountInfoMap["s"]; exist {
						accountID := types.ServiceId(accountID.(int))
						if account, exists := state.Delta[accountID]; exists {
							// (h ↦ p) ∈ ap
							if hashStr, exist := accountInfoMap["h"]; exist {
								targetKey := types.OpaqueHash(hex2Bytes(hashStr.(string)))
								if val, keyExists := account.PreimageLookup[targetKey]; keyExists {
									key16, delta3Output := EncodeDelta3KeyVal(accountID, account, targetKey, val)

									if !reflect.DeepEqual(types.OpaqueHash(hex2Bytes(keyVal[0])), key16) {
										diff := cmp.Diff(types.OpaqueHash(hex2Bytes(keyVal[0])), key16)
										t.Error(dirName, jsonPath, "key16 for preimage does not match", diff)
									}

									if !reflect.DeepEqual(types.ByteSequence(hex2Bytes(keyVal[1])), delta3Output) {
										diff := cmp.Diff(types.ByteSequence(hex2Bytes(keyVal[1])), delta3Output)
										t.Error(dirName, jsonPath, "delta3 value does not match", diff)
									}
								} else {
									t.Logf("PreimageLookup key with hash %s not found for account %v", targetKey, accountID)
								}
							}
						}
					}
				}
				if keyVal[2] == "account_storage" {
					accountInfoMap := parseServiceAccountInfo(keyVal[3])
					// ∀(s ↦ a) ∈ δ
					if accountID, exist := accountInfoMap["s"]; exist {
						accountID := types.ServiceId(accountID.(int))
						if account, exists := state.Delta[accountID]; exists {
							// (k ↦ v) ∈ ap
							if hashStr, exist := accountInfoMap["h"]; exist {
								targetKey := types.OpaqueHash(hex2Bytes(hashStr.(string)))

								if val, keyExists := account.StorageDict[targetKey]; keyExists {
									key16, delta2Output := EncodeDelta2KeyVal(accountID, account, targetKey, val)

									if !reflect.DeepEqual(types.OpaqueHash(hex2Bytes(keyVal[0])), key16) {
										diff := cmp.Diff(types.OpaqueHash(hex2Bytes(keyVal[0])), key16)
										t.Error(dirName, jsonPath, "key16 for storage does not match", diff)
									}
									if !reflect.DeepEqual(types.ByteSequence(hex2Bytes(keyVal[1])), delta2Output) {
										diff := cmp.Diff(types.ByteSequence(hex2Bytes(keyVal[1])), delta2Output)
										t.Error(dirName, jsonPath, "delta2 value does not match", diff)
									}
								} else {
									t.Logf("StorageDict key with hash %s not found for account %v", targetKey, accountID)
								}
							}
						}
					}
				}
				if keyVal[2] == "c1" {
					c1 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c1, EncodeAlphaKey()) {
						diff := cmp.Diff(c1, EncodeAlphaKey())
						t.Error("c1 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedAlpha) {
						diff := cmp.Diff(val, encodedAlpha)
						t.Error("c1 value does not match", diff)
					}
				}
				if keyVal[2] == "c2" {
					c2 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c2, EncodeVarphiKey()) {
						diff := cmp.Diff(c2, EncodeVarphiKey())
						t.Error("c2 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedVarphi) {
						diff := cmp.Diff(val, encodedVarphi)
						t.Error("c2 value does not match", diff)
					}
				}
				if keyVal[2] == "c3" {
					c3 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c3, EncodeBetaKey()) {
						diff := cmp.Diff(c3, EncodeBetaKey())
						t.Error("c3 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedBeta) {
						diff := cmp.Diff(val, encodedBeta)
						t.Error("c3 value does not match", diff)
					}
				}
				if keyVal[2] == "c4" {
					c4 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c4, EncodeGammaKey()) {
						diff := cmp.Diff(c4, EncodeGammaKey())
						t.Error("c4 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedGamma) {
						diff := cmp.Diff(val, encodedGamma)
						t.Error("c4 value does not match", diff)
					}
				}
				if keyVal[2] == "c5" {
					c5 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c5, EncodePsiKey()) {
						diff := cmp.Diff(c5, EncodePsiKey())
						t.Error("c5 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedPsi) {
						diff := cmp.Diff(val, encodedPsi)
						t.Error("c5 value does not match", diff)
					}
				}
				if keyVal[2] == "c6" {
					c6 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c6, EncodeEtaKey()) {
						diff := cmp.Diff(c6, EncodeEtaKey())
						t.Error("c6 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedEta) {
						diff := cmp.Diff(val, encodedEta)
						t.Error("c6 value does not match", diff)
					}
				}
				if keyVal[2] == "c7" {
					c7 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c7, EncodeIotaKey()) {
						diff := cmp.Diff(c7, EncodeIotaKey())
						t.Error("c7 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedIota) {
						diff := cmp.Diff(val, encodedIota)
						t.Error("c7 value does not match", diff)
					}
				}
				if keyVal[2] == "c8" {
					c8 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c8, EncodeKappaKey()) {
						diff := cmp.Diff(c8, EncodeKappaKey())
						t.Error("c8 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedKappa) {
						diff := cmp.Diff(val, encodedKappa)
						t.Error("c8 value does not match", diff)
					}
				}
				if keyVal[2] == "c9" {
					c9 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c9, EncodeLambdaKey()) {
						diff := cmp.Diff(c9, EncodeLambdaKey())
						t.Error("c9 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedLambda) {
						diff := cmp.Diff(val, encodedLambda)
						t.Error("c9 value does not match", diff)
					}
				}
				if keyVal[2] == "c10" {
					c10 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c10, EncodeRhoKey()) {
						diff := cmp.Diff(c10, EncodeRhoKey())
						t.Error("c10 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedRho) {
						diff := cmp.Diff(val, encodedRho)
						t.Error("c10 value does not match", diff)
					}
				}
				if keyVal[2] == "c11" {
					c11 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c11, EncodeTauKey()) {
						diff := cmp.Diff(c11, EncodeTauKey())
						t.Error("c11 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedTau) {
						diff := cmp.Diff(val, encodedTau)
						t.Error("c11 value does not match", diff)
					}
				}
				if keyVal[2] == "c12" {
					c12 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c12, EncodeChiKey()) {
						diff := cmp.Diff(c12, EncodeChiKey())
						t.Error("c12 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedChi) {
						diff := cmp.Diff(val, encodedChi)
						t.Error("c12 value does not match", diff)
					}
				}
				if keyVal[2] == "c13" {
					c13 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c13, EncodePiKey()) {
						diff := cmp.Diff(c13, EncodePiKey())
						t.Error("c13 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedPi) {
						diff := cmp.Diff(val, encodedPi)
						t.Error("c13 value does not match", diff)
					}
				}
				if keyVal[2] == "c14" {
					c14 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c14, EncodeThetaKey()) {
						diff := cmp.Diff(c14, EncodeThetaKey())
						t.Error("c14 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedTheta) {
						diff := cmp.Diff(val, encodedTheta)
						t.Error("c14 value does not match", diff)
					}
				}
				if keyVal[2] == "c15" {
					c15 := types.OpaqueHash(hex2Bytes(keyVal[0]))
					if !reflect.DeepEqual(c15, EncodeXiKey()) {
						diff := cmp.Diff(c15, EncodeXiKey())
						t.Error("c15 key does not match", diff)
					}
					val := types.ByteSequence(hex2Bytes(keyVal[1]))
					if !reflect.DeepEqual(val, encodedXi) {
						diff := cmp.Diff(val, encodedXi)
						t.Error("c15 value does not match", diff)
					}
				}
				if keyVal[2] == "service_account" {
					accountInfoMap := parseServiceAccountInfo(keyVal[3])
					// ∀(s ↦ a) ∈ δ
					if accountID, exist := accountInfoMap["s"]; exist {
						accountID := types.ServiceId(accountID.(int))
						if account, exists := state.Delta[accountID]; exists {
							key16, delta1Output := EncodeDelta1KeyVal(accountID, account)
							if !reflect.DeepEqual(key16, types.OpaqueHash(hex2Bytes(keyVal[0]))) {
								diff := cmp.Diff(key16, types.OpaqueHash(hex2Bytes(keyVal[0])))
								t.Error(dirName, jsonPath, "key16 does not match", diff)
							}
							val := types.ByteSequence(hex2Bytes(keyVal[1]))
							if !reflect.DeepEqual(val, delta1Output) {
								diff := cmp.Diff(val, delta1Output)
								t.Error(dirName, jsonPath, "delta1 value does not match", diff)
							}
						}
					}
				}
			}
			// === Compare state_root ===
			stateRoot := transitionState.PostState.StateRoot
			hexToOpaqueHash := types.OpaqueHash(hex2Bytes(stateRoot))
			ourStateRoot, err := MerklizationState(state)
			if err != nil {
				t.Error("MerklizationState raised error", err)
			}
			if !reflect.DeepEqual(ourStateRoot, hexToOpaqueHash) {
				diff := cmp.Diff(ourStateRoot, hexToOpaqueHash)
				t.Error("MerklizationState failed", diff)
			} else {
				t.Log("MerklizationState success")
			}
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

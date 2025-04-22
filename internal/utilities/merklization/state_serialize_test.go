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

func TestJamTestNetStateRoot(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet state test cases only support tiny mode")
	}

	dirNames := []string{
		// "assurances",
		"fallback",
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

		for baseName, entry := range fileMap {
			if !entry.hasSnapshot || !entry.hasTransition {
				t.Errorf("Missing snapshot or transition for %s", baseName)
				continue
			}

			// Read the binary file
			binPath := entry.snapshotPath
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
			baseFileName := filepath.Base(entry.snapshotPath)
			baseFileName = baseFileName[:len(baseFileName)-len(utils.BIN_EXTENTION)]
			jsonFilePath := filepath.Join(snapshotsDir, baseFileName+utils.JSON_EXTENTION)
			jsonState, err := utils.GetTestFromJson(jsonFilePath, &types.State{})
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Compare the two structs
			// if !reflect.DeepEqual(state, jsonState) {
			// 	log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
			// 	t.Errorf("Error: %v", err)
			// } else {
			// 	log.Printf("✅ [%s] [%s] %s", types.TEST_MODE, dirName, file)
			// }
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
					t.Error(dirName, entry.snapshotPath, "serialize Delta1 failed", diff)
				}
			}
			if !reflect.DeepEqual(encodedAlpha, serializedAlpha) {
				t.Error(dirName, entry.snapshotPath, "serialize Alpha failed")
			} else if !reflect.DeepEqual(encodedVarphi, serializedVarphi) {
				t.Error(dirName, entry.snapshotPath, "serialize Varphi failed")
			} else if !reflect.DeepEqual(encodedBeta, serializedBeta) {
				t.Error(dirName, entry.snapshotPath, "serialize Beta failed")
			} else if !reflect.DeepEqual(encodedGamma, serializedGamma) {
				t.Error(dirName, entry.snapshotPath, "serialize Gamma failed")
			} else if !reflect.DeepEqual(encodedPsi, serializedPsi) {
				t.Error(dirName, entry.snapshotPath, "serialize Psi failed")
			} else if !reflect.DeepEqual(encodedEta, serializedEta) {
				t.Error(dirName, entry.snapshotPath, "serialize Eta failed")
			} else if !reflect.DeepEqual(encodedIota, serializedIota) {
				t.Error(dirName, entry.snapshotPath, "serialize Iota failed")
			} else if !reflect.DeepEqual(encodedKappa, serializedKappa) {
				t.Error(dirName, entry.snapshotPath, "serialize Kappa failed")
			} else if !reflect.DeepEqual(encodedLambda, serializedLambda) {
				t.Error(dirName, entry.snapshotPath, "serialize Lambda failed")
			} else if !reflect.DeepEqual(encodedRho, serializedRho) {
				diff := cmp.Diff(encodedRho, serializedRho)
				t.Error(dirName, entry.snapshotPath, "serialize Rho failed", diff)
			} else if !reflect.DeepEqual(encodedTau, serializedTau) {
				t.Error(dirName, entry.snapshotPath, "serialize Tau failed")
			} else if !reflect.DeepEqual(encodedChi, serializedChi) {
				t.Error(dirName, entry.snapshotPath, "serialize Chi failed")
			} else if !reflect.DeepEqual(encodedPi, serializedPi) {
				t.Log("Pi Services", jsonState.Pi.Services)
				diff := cmp.Diff(encodedPi, serializedPi)
				t.Error(dirName, entry.snapshotPath, "serialize Pi failed", diff)
			} else if !reflect.DeepEqual(encodedTheta, serializedTheta) {
				diff := cmp.Diff(encodedTheta, serializedTheta)
				t.Error(dirName, entry.snapshotPath, "serialize Theta failed", diff)
			} else if !reflect.DeepEqual(encodedXi, serializedXi) {
				diff := cmp.Diff(encodedXi, serializedXi)
				t.Error(dirName, entry.snapshotPath, "serialize Xi failed", diff)
			} else {
				fmt.Println(entry.snapshotPath, "serialize success")
			}

			//=== (each state finish, test all state) ===

			encodedState, err := StateEncoder(jsonState)
			if err != nil {
				t.Error("encodedState raised error", err)
			}
			// fmt.Println("encodedState", encodedState)
			serializedState, err := StateSerialize(jsonState)
			if err != nil {
				t.Error("StateSerialize raised error", err)
			}
			if !reflect.DeepEqual(encodedState, serializedState) {
				diff := cmp.Diff(encodedState, serializedState)
				t.Error(dirName, entry.snapshotPath, "serialize State failed", diff)
			}
			// fmt.Println("serializedState", serializedState)

			// ========
			// Read the binary file
			jsonPath := entry.transitionPath

			// in state_transitions we only need state_root
			type SimpleStateJson struct {
				PreState struct {
					StateRoot string `json:"state_root"`
				} `json:"pre_state"`
			}

			jsonData, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Logf("failed to read file: %v", err)
				continue
			}

			var simpleState SimpleStateJson
			err = json.Unmarshal(jsonData, &simpleState)
			if err != nil {
				t.Logf("failed to parse JSON file: %v", err)
				continue
			}

			stateRoot := simpleState.PreState.StateRoot
			hexToOpaqueHash := hex2Bytes(stateRoot)
			ourStateRoot, err := MerklizationState(jsonState)
			if err != nil {
				t.Error("MerklizationState raised error", err)
			}
			if !reflect.DeepEqual(ourStateRoot, hexToOpaqueHash) {
				diff := cmp.Diff(ourStateRoot, hexToOpaqueHash)
				t.Error("MerklizationState failed", diff)
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

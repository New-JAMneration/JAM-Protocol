package merklization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

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

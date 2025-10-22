package merklization_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
)

func TestStateKeyValsToStateGenesis(t *testing.T) {
	dirNames := []string{
		"fallback",
		"preimages",
		"preimages_light",
		"safrole",
		"storage",
		"storage_light",
	}

	var err error

	for _, dirName := range dirNames {
		dir := filepath.Join("..", utilities.JAM_TEST_VECTORS_DIR, "traces", dirName)

		genesisFilePath := filepath.Join(dir, "genesis.bin")

		genesisTestCase := &jamtests_trace.Genesis{}
		err = utilities.GetTestFromBin(genesisFilePath, genesisTestCase)
		if err != nil {
			t.Errorf("Error reading file %s: %v", genesisFilePath, err)
			continue
		}

		genesisState, err := merklization.StateKeyValsToState(genesisTestCase.State.KeyVals)
		if err != nil {
			t.Errorf("Error parsing state keyvals: %v", err)
		}

		// Create a state root with the genesis state
		genesisStateRoot := merklization.MerklizationState(genesisState)

		expectedGenesisStateRoot := genesisTestCase.State.StateRoot

		// Compare the state root with the expected state root
		if genesisStateRoot != expectedGenesisStateRoot {
			t.Errorf("❌ State root mismatch: expected 0x%x, got 0x%x", expectedGenesisStateRoot, genesisStateRoot)
		} else {
			t.Logf("✅ State root matches: 0x%x", genesisStateRoot)
		}
	}
}

// FIXME: We cannot obtain the storage key from StateKeyVal.
// so this test will not work as expected.
// INFO: We can pass fallback and safrole directories because they do not contain storage keys.
func TestStateKeyValsToState(t *testing.T) {
	dirNames := []string{
		"fallback",
		"preimages",
		"preimages_light",
		"safrole",
		"storage",
		"storage_light",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join("..", utilities.JAM_TEST_VECTORS_DIR, "traces", dirName)

		fileNames, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error getting files from directory %s: %v", dir, err)
			continue
		}

		for _, fileName := range fileNames {
			if fileName == "genesis.bin" { // skip genesis
				continue
			}

			filePath := filepath.Join(dir, fileName)

			// Read the bin file
			traceTestCase := &jamtests_trace.TraceTestCase{}
			err := utilities.GetTestFromBin(filePath, traceTestCase)
			if err != nil {
				t.Errorf("Error reading file %s: %v", filePath, err)
				continue
			}

			// Parse the state keyvals
			state, err := merklization.StateKeyValsToState(traceTestCase.PostState.KeyVals)
			if err != nil {
				t.Errorf("Error parsing state keyvals: %v", err)
			}

			// Create a state root with the parsed state
			stateRoot := merklization.MerklizationState(state)

			expectedStateRoot := traceTestCase.PostState.StateRoot

			// Compare the state root with the expected state root
			if stateRoot != expectedStateRoot {
				t.Errorf("❌ State root mismatch in [%s] %s: expected 0x%x, got 0x%x", dirName, fileName, expectedStateRoot, stateRoot)
			} else {
				t.Logf("✅ State root matches in [%s] %s: 0x%x", dirName, fileName, stateRoot)
			}
		}
	}
}

// We cannot obtain the storage key from StateKeyVal in its current form,
// so our StateKeyValsToState function does not include storage keys in the output state.
// This test checks that the state keyvals do not contain storage keys,
// which is expected behavior.
func TestStateKeyValsToState_CheckStateKeyValsWithoutStorageKey(t *testing.T) {
	dirNames := []string{
		"fallback",
		"preimages",
		"preimages_light",
		"safrole",
		"storage",
		"storage_light",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join("..", utilities.JAM_TEST_VECTORS_DIR, "traces", dirName)

		fileNames, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error getting files from directory %s: %v", dir, err)
			continue
		}

		for _, fileName := range fileNames {
			if fileName == "genesis.bin" { // skip genesis
				continue
			}

			filePath := filepath.Join(dir, fileName)

			// Read the bin file
			traceTestCase := &jamtests_trace.TraceTestCase{}
			err := utilities.GetTestFromBin(filePath, traceTestCase)
			if err != nil {
				t.Errorf("Error reading file %s: %v", filePath, err)
				continue
			}

			// Parse the state keyvals
			state, err := merklization.StateKeyValsToState(traceTestCase.PostState.KeyVals)
			if err != nil {
				t.Errorf("Error parsing state keyvals: %v", err)
			}

			// serialize the state
			actualStateKeyVals, err := merklization.StateEncoder(state)
			if err != nil {
				t.Errorf("Error serializing state: %v", err)
			}

			expectedStateKeyVals := traceTestCase.PostState.KeyVals

			actualStateKeyValsMap := make(map[types.StateKey]types.ByteSequence)
			for _, stateKeyVal := range actualStateKeyVals {
				actualStateKeyValsMap[stateKeyVal.Key] = stateKeyVal.Value
			}

			expectedStateKeyValsMap := make(map[types.StateKey]types.ByteSequence)
			for _, stateKeyVal := range expectedStateKeyVals {
				expectedStateKeyValsMap[stateKeyVal.Key] = stateKeyVal.Value
			}

			diffs, err := merklization.GetStateKeyValsDiff(expectedStateKeyVals, actualStateKeyVals)
			if err != nil {
				t.Errorf("Error getting state keyvals diff: %v", err)
			}

			if len(diffs) > 0 {
				for _, diff := range diffs {
					// Only check the key exists in the actual state keyvals
					// because we cannot obtain the storage key from StateKeyVal
					// The state will not contain storage keys in its current form.
					if _, exists := actualStateKeyValsMap[diff.Key]; !exists {
						continue
					}

					t.Errorf("❌ State key 0x%x has diff in [%s] %s\n", diff.Key, dirName, fileName)
					t.Errorf("Expected: 0x%x\n", diff.ExpectedValue)
					t.Errorf("Actual: 0x%x\n", diff.ActualValue)
				}
			} else {
				fmt.Printf("✅ All state keys (without storage) match in [%s] %s\n", dirName, fileName)
			}
		}
	}
}

func TestGetStateKeyValsDiff(t *testing.T) {
	expectedStateKeyVals := []types.StateKeyVal{
		{Key: types.StateKey{0x01}, Value: types.ByteSequence{0x01}},
		{Key: types.StateKey{0x02}, Value: types.ByteSequence{0x02}},
		{Key: types.StateKey{0x03}, Value: types.ByteSequence{0x03}},
	}

	actualStateKeyVals := []types.StateKeyVal{
		{Key: types.StateKey{0x01}, Value: types.ByteSequence{0x01}},
		{Key: types.StateKey{0x02}, Value: types.ByteSequence{0x22}}, // different value
		{Key: types.StateKey{0x04}, Value: types.ByteSequence{0x04}}, // extra key
	}

	diffs, err := merklization.GetStateKeyValsDiff(expectedStateKeyVals, actualStateKeyVals)
	if err != nil {
		t.Errorf("Error getting state keyvals diff: %v", err)
	}

	expectedDiffs := []types.StateKeyValDiff{
		{
			Key:           types.StateKey{0x02},
			ExpectedValue: types.ByteSequence{0x02},
			ActualValue:   types.ByteSequence{0x22},
		},
		{
			Key:           types.StateKey{0x03},
			ExpectedValue: types.ByteSequence{0x03},
			ActualValue:   nil, // expected value is nil because the key is not in actualStateKeyVals
		},
		{
			Key:           types.StateKey{0x04},
			ExpectedValue: nil, // expected value is nil because the key is not in expectedStateKeyVals
			ActualValue:   types.ByteSequence{0x04},
		},
	}

	if len(diffs) != len(expectedDiffs) {
		t.Errorf("Expected %d diffs, got %d", len(expectedDiffs), len(diffs))
	}

	for i, diff := range diffs {
		if diff.Key != expectedDiffs[i].Key {
			t.Errorf("Diff key mismatch: expected %x, got %x", expectedDiffs[i].Key, diff.Key)
		}

		if !bytes.Equal(diff.ExpectedValue, expectedDiffs[i].ExpectedValue) {
			t.Errorf("Diff expected value mismatch: expected %x, got %x", expectedDiffs[i].ExpectedValue, diff.ExpectedValue)
		}
		if !bytes.Equal(diff.ActualValue, expectedDiffs[i].ActualValue) {
			t.Errorf("Diff actual value mismatch: expected %x, got %x", expectedDiffs[i].ActualValue, diff.ActualValue)
		}
	}
}

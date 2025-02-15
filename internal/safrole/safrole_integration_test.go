package safrole

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
)

func LoadSafroleTestCase(filename string) (jamtests.SafroleTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests.SafroleTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
	}

	return testCases, nil
}

func GetTestJsonFiles(dir string) []string {
	jsonFiles := []string{}

	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		return nil
	}

	extension := ".json"
	for _, file := range files {
		if filepath.Ext(file.Name()) == extension {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	return jsonFiles
}

func TestStatistics(t *testing.T) {
	dir := "../../pkg/test_data/jam-test-vectors/safrole/tiny/"
	jsonFiles := GetTestJsonFiles(dir)
	for _, file := range jsonFiles {
		filename := dir + file
		safroleTestCase, err := LoadSafroleTestCase(filename)
		if err != nil {
			t.Errorf("Error loading statistics test case: %v", err)
			return
		}

		// Set input to store
		s := store.GetInstance()
		s.GetProcessingBlockPointer().SetAuthorIndex(safroleTestCase.Input.AuthorIndex)
		s.GetPriorStates().SetTau(safroleTestCase.PreState.Tau)
		s.GetPosteriorStates().SetTau(safroleTestCase.Input.Slot)
		s.GetPriorStates().SetPi(safroleTestCase.PreState.Pi)
		s.GetPosteriorStates().SetKappa(safroleTestCase.PreState.Kappa)

		UpdateValidatorActivityStatistics(safroleTestCase.Input.Extrinsic)

		// Get statistics
		statistics := s.GetPosteriorStates().GetPi()

		// Expected statistics
		expectedStatistics := safroleTestCase.PostState

		// Compare statistics struct
		if !reflect.DeepEqual(statistics, expectedStatistics.Pi) {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedStatistics.Pi, statistics)
		}
	}
}

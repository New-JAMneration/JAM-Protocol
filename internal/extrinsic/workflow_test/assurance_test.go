package extrinsic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	assurance "github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
)

func TestMain(m *testing.M) {
	types.SetTestMode()
	m.Run()
}

func LoadAssuranceTestCase(filename string) (jamtests.AssuranceTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests.AssuranceTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests.AssuranceTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests.AssuranceTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests.AssuranceTestCase{}, err
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

func TestAssurances(t *testing.T) {
	dir := "../../pkg/test_data/jam-test-vectors/assurances/" + os.Getenv("TEST_MODE") + "/"
	jsonFiles := GetTestJsonFiles(dir)

	for _, file := range jsonFiles {
		fmt.Println("file : ", file)
		fileName := dir + file
		assurancesTestCase, err := LoadAssuranceTestCase(fileName)
		if err != nil {
			t.Errorf("Error loading assurance test case: %v", err)
			return
		}
		s := store.GetInstance()
		GenerateBlockForHeader(assurancesTestCase.Input)
		s.GetPosteriorStates().SetKappa(assurancesTestCase.PreState.CurrValidators)
		s.GetIntermediateStates().SetRhoDagger(assurancesTestCase.PreState.AvailAssignments)

		assurancesState := jamtests.AssuranceState{}

		err = assurance.Assurance(assurancesTestCase.Input.Assurances)
		if err != nil {
			s.GetIntermediateStates().SetRhoDoubleDagger(assurancesTestCase.PreState.AvailAssignments)
			s.GetPosteriorStates().SetKappa(assurancesTestCase.PreState.CurrValidators)
		}

		// Get assurances states
		assurancesState.CurrValidators = s.GetPosteriorStates().GetKappa()
		assurancesState.AvailAssignments = s.GetIntermediateStates().GetRhoDoubleDagger()

		// Expected assurances
		expectedAssurances := assurancesTestCase.PostState

		if !reflect.DeepEqual(assurancesState, expectedAssurances) {
			t.Errorf("Test case %v failed: ", file)
		}
	}
}

func GenerateBlockForHeader(input jamtests.AssuranceInput) {
	s := store.GetInstance()

	header := types.Header{Slot: input.Slot, Parent: input.Parent}
	block := types.Block{Header: header}
	s.AddBlock(block)
}

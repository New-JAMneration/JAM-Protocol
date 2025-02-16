package extrinsic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
)

func TestMain(m *testing.M) {
	// types.SetTestMode()
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
	dir := "../../pkg/test_data/jam-test-vectors/assurances/tiny/"
	jsonFiles := GetTestJsonFiles(dir)

	for _, file := range jsonFiles {
		fmt.Println("file : ", file)
		fileName := dir + file
		fmt.Println("fileName : ", fileName)
		assurancesTestCase, err := LoadAssuranceTestCase(fileName)
		if err != nil {
			t.Errorf("Error loading assurance test case: %v", err)
			return
		}

		Assurance(assurancesTestCase.Input.Assurances)

		// Set input to store
		s := store.GetInstance()

		// Get assurances
		assurances := s.GetIntermediateStates().GetRhoDoubleDagger()
		fmt.Println("------------------------")
		fmt.Println("RhoDoubleDagger : ", assurances)
		// Expected assurances
		expectedAssurances := assurancesTestCase.PostState
		fmt.Println("expectedAssurances : ", expectedAssurances)
		// Compare statistics struct
		if !reflect.DeepEqual(assurances, expectedAssurances) {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedAssurances, assurances)
		}
	}
}

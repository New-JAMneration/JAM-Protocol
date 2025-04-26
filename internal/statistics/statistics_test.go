package statistics

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
)

var JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"

func TestMain(m *testing.M) {
	types.SetTestMode()
	m.Run()
}

func LoadStatisticsTestCase(filename string) (jamtests.StatisticsTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests.StatisticsTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests.StatisticsTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests.StatisticsTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests.StatisticsTestCase{}, err
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
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "statistics", types.TEST_MODE)
	jsonFiles := GetTestJsonFiles(dir)
	for _, file := range jsonFiles {
		filename := filepath.Join(dir, file)
		statisticsTestCase, err := LoadStatisticsTestCase(filename)
		if err != nil {
			t.Errorf("Error loading statistics test case: %v", err)
			return
		}

		// Set input to store
		s := store.GetInstance()
		s.GetProcessingBlockPointer().SetAuthorIndex(statisticsTestCase.Input.AuthorIndex)
		s.GetPriorStates().SetTau(types.TimeSlot(statisticsTestCase.PreState.Slot))
		s.GetPosteriorStates().SetTau(types.TimeSlot(statisticsTestCase.Input.Slot))
		s.GetPriorStates().SetPi(statisticsTestCase.PreState.Statistics)
		s.GetPosteriorStates().SetKappa(statisticsTestCase.PreState.CurrValidators)

		UpdateValidatorActivityStatistics(statisticsTestCase.Input.Extrinsic)

		// Get statistics
		statistics := s.GetPosteriorStates().GetPi()

		// Expected statistics
		expectedStatistics := statisticsTestCase.PostState

		// Compare statistics struct
		if !reflect.DeepEqual(statistics, expectedStatistics.Statistics) {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedStatistics.Statistics, statistics)
		}
	}
}

package statistics

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"reflect"
// 	"testing"

// 	"github.com/New-JAMneration/JAM-Protocol/internal/store"
// 	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
// )

// func LoadStatisticsTestCase(filename string) (jamtests.StatisticsTestCase, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return jamtests.StatisticsTestCase{}, err
// 	}
// 	defer file.Close()

// 	// Read the file content
// 	byteValue, err := io.ReadAll(file)
// 	if err != nil {
// 		return jamtests.StatisticsTestCase{}, err
// 	}

// 	// Unmarshal the JSON data
// 	var testCases jamtests.StatisticsTestCase
// 	err = json.Unmarshal(byteValue, &testCases)
// 	if err != nil {
// 		return jamtests.StatisticsTestCase{}, err
// 	}

// 	return testCases, nil
// }

// func GetTestJsonFiles(dir string) []string {
// 	jsonFiles := []string{}

// 	f, err := os.Open(dir)
// 	if err != nil {
// 		return nil
// 	}
// 	defer f.Close()

// 	files, err := f.Readdir(-1)
// 	if err != nil {
// 		return nil
// 	}

// 	extension := ".json"
// 	for _, file := range files {
// 		if filepath.Ext(file.Name()) == extension {
// 			jsonFiles = append(jsonFiles, file.Name())
// 		}
// 	}

// 	return jsonFiles
// }

// func TestStatistics(t *testing.T) {
// 	dir := "../../pkg/test_data/jam-test-vectors/statistics/tiny/"
// 	jsonFiles := GetTestJsonFiles(dir)
// 	for _, file := range jsonFiles {
// 		fmt.Println(file)
// 		filename := dir + file
// 		statisticsTestCase, err := LoadStatisticsTestCase(filename)
// 		if err != nil {
// 			t.Errorf("Error loading statistics test case: %v", err)
// 			return
// 		}

// 		// Set input to store
// 		s := store.GetInstance()
// 		s.GetIntermediateHeaders().SetAuthorIndex(statisticsTestCase.Input.AuthorIndex)
// 		s.GetPosteriorStates().SetTau(statisticsTestCase.Input.Slot)
// 		s.GetPosteriorStates().SetPi(statisticsTestCase.PreState.Pi)

// 		UpdateValidatorActivityStatistics(statisticsTestCase.Input.Extrinsic)

// 		// Get statistics
// 		statistics := s.GetPosteriorStates().GetPi()

// 		// Expected statistics
// 		expectedStatistics := statisticsTestCase.PostState

// 		// Compare statistics struct
// 		if reflect.DeepEqual(statistics, expectedStatistics.Pi) {
// 			t.Errorf("Expected statistics %v, got %v", expectedStatistics.Pi, statistics)
// 		}
// 	}
// }

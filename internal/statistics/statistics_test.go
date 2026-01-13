package statistics

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests_reports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtests_statistics "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
)

var JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"

func TestMain(m *testing.M) {
	types.SetTestMode()
	m.Run()
}

func LoadReportsTestCase(filename string) (jamtests_reports.ReportsTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests_reports.ReportsTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests_reports.ReportsTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests_reports.ReportsTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests_reports.ReportsTestCase{}, err
	}

	return testCases, nil
}

func LoadStatisticsTestCase(filename string) (jamtests_statistics.StatisticsTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests_statistics.StatisticsTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests_statistics.StatisticsTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests_statistics.StatisticsTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests_statistics.StatisticsTestCase{}, err
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

// We use the report's test vectors to test the cores and services statistics.
func TestStatisticsWithReportTestVectors(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "reports", types.TEST_MODE)
	jsonFiles := GetTestJsonFiles(dir)
	for _, file := range jsonFiles {
		filename := filepath.Join(dir, file)
		reportsTestCase, err := LoadReportsTestCase(filename)
		if err != nil {
			t.Errorf("Error loading reports test case: %v", err)
			return
		}

		// In this case, we just want to test the statistics
		// Therefore, skip the test if there is an error
		output := reportsTestCase.Output
		if output.Err != nil {
			continue
		}

		// w (present work reports) from guarantee extrinsic
		reports := []types.WorkReport{}
		for _, guarantee := range reportsTestCase.Input.Guarantees {
			reports = append(reports, guarantee.Report)
		}

		// Set input to store
		s := blockchain.GetInstance()
		s.GetIntermediateStates().SetPresentWorkReports(reports)

		// input
		s.GetProcessingBlockPointer().SetSlot(reportsTestCase.Input.Slot)
		s.GetPriorStates().SetTau(reportsTestCase.Input.Slot)

		// get the guarantee extrinsic
		guaranteeExtrinsic := reportsTestCase.Input.Guarantees
		extrinsic := types.Extrinsic{
			Guarantees: guaranteeExtrinsic,
		}
		s.GetProcessingBlockPointer().SetExtrinsics(extrinsic)

		UpdateValidatorActivityStatistics()

		// Get statistics
		statistics := s.GetPosteriorStates().GetPi()

		// Expected statistics
		expectedServicesStatistics := reportsTestCase.PostState.ServicesStatistics

		// Compare services statistics struct
		if !reflect.DeepEqual(statistics.Services, expectedServicesStatistics) {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedServicesStatistics, statistics.Services)
		}

		expectedCoreStatistics := reportsTestCase.PostState.CoresStatistics

		// Compare core statistics struct
		if !reflect.DeepEqual(statistics.Cores, expectedCoreStatistics) {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedCoreStatistics, statistics.Cores)
		}
	}
}

// This function uses the statistics test vectors to test the statistics logic.
// However, the test vector does not include the core and services statistics.
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

		blockchain.ResetInstance()
		// Set input to store
		s := blockchain.GetInstance()

		// input
		s.GetProcessingBlockPointer().SetSlot(statisticsTestCase.Input.Slot)
		s.GetProcessingBlockPointer().SetAuthorIndex(statisticsTestCase.Input.AuthorIndex)
		s.GetPriorStates().SetTau(types.TimeSlot(statisticsTestCase.PreState.Slot))
		s.GetPosteriorStates().SetTau(types.TimeSlot(statisticsTestCase.Input.Slot))
		s.GetPriorStates().SetPiCurrent(statisticsTestCase.PreState.ValsCurrStats)
		s.GetPriorStates().SetPiLast(statisticsTestCase.PreState.ValsLastStats)
		s.GetPosteriorStates().SetKappa(statisticsTestCase.PreState.CurrValidators)

		// w (present work reports) from guarantee extrinsic
		reports := []types.WorkReport{}
		for _, guarantee := range statisticsTestCase.Input.Extrinsic.Guarantees {
			reports = append(reports, guarantee.Report)
		}

		// Extrinsic
		s.GetProcessingBlockPointer().SetExtrinsics(statisticsTestCase.Input.Extrinsic)

		// Set input to store
		s.GetIntermediateStates().SetPresentWorkReports(reports)

		// pre_state
		s.GetPriorStates().SetTau(statisticsTestCase.Input.Slot)
		s.GetPriorStates().SetPiCurrent(statisticsTestCase.PreState.ValsCurrStats)
		s.GetPriorStates().SetPiLast(statisticsTestCase.PreState.ValsLastStats)
		s.GetPriorStates().SetKappa(statisticsTestCase.PreState.CurrValidators)

		// post_state
		s.GetPosteriorStates().SetTau(statisticsTestCase.PostState.Slot)

		UpdateValidatorActivityStatistics()

		// Get statistics
		statistics := s.GetPosteriorStates().GetPi()

		// Expected statistics
		expectedStatistics := statisticsTestCase.PostState

		if !reflect.DeepEqual(statistics.ValsCurr, expectedStatistics.ValsCurrStats) {
			t.Errorf("ValsCurrStats mismatch in %s: got %v, want %v", file, statistics.ValsCurr, expectedStatistics.ValsCurrStats)
		}

		if !reflect.DeepEqual(statistics.ValsLast, expectedStatistics.ValsLastStats) {
			t.Errorf("ValsLastStats mismatch in %s: got %v, want %v", file, statistics.ValsLast, expectedStatistics.ValsLastStats)
		}

		// Issue: https://github.com/davxy/jam-test-vectors/issues/39
		// Temporarily commented out the services statistics comparison

		// // Compare statistics struct
		// if !reflect.DeepEqual(statistics, expectedStatistics.Statistics) {
		// 	t.Errorf("Test case %v failed: expected %v, got %v", file, expectedStatistics.Statistics, statistics)
		// }
	}
}

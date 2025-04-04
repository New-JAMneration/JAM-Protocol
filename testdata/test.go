package testdata

import (
	"fmt"
)

// TestResult represents the result of a test
type TestResult struct {
	TestFile string
	Passed   bool
	Error    error
}

// RunTests runs tests for the specified mode and size
func RunTests(mode TestMode, data interface) ([]TestResult, error) {
	// Create a test data reader
	reader := NewTestDataReader(mode, size, JSONFormat)

	// Read all test files
	testFiles, err := reader.ReadTestData()
	if err != nil {
		return nil, fmt.Errorf("failed to read test data: %v", err)
	}

	var results []TestResult
	for _, testFile := range testFiles {
		result := TestResult{
			TestFile: testFile.Name,
		}

		// Run the appropriate test based on mode
		switch mode {
		case SafroleMode:
			err = runSafroleTest(testFile.Data)
		case AssurancesMode:
			err = runAssurancesTest(testFile.Data)
		case PreimagesMode:
			err = runPreimagesTest(testFile.Data)
		case HistoryMode:
			err = runHistoryTest(testFile.Data)
		case DisputesMode:
			err = runDisputesTest(testFile.Data)
		case AuthorizationsMode:
			err = runAuthorizationsTest(testFile.Data)
		case AccumulateMode:
			err = runAccumulateTest(testFile.Data)
		default:
			err = fmt.Errorf("unsupported test mode: %s", mode)
		}

		if err != nil {
			result.Passed = false
			result.Error = err
		} else {
			result.Passed = true
		}

		results = append(results, result)
	}

	return results, nil
}

// Dummy test functions - to be implemented by others
func runSafroleTest(data []byte) error {
	// TODO: Implement using jamtests/safrole
	return nil
}

func runAssurancesTest(data []byte) error {
	// TODO: Implement using jamtests/assurances
	return nil
}

func runPreimagesTest(data []byte) error {
	// TODO: Implement using jamtests/preimages
	return nil
}

func runHistoryTest(data []byte) error {
	// TODO: Implement using jamtests/history
	return nil
}

func runDisputesTest(data []byte) error {
	// TODO: Implement using jamtests/disputes
	return nil
}

func runAuthorizationsTest(data []byte) error {
	// TODO: Implement using jamtests/authorizations
	return nil
}

func runAccumulateTest(data []byte) error {
	// TODO: Implement using jamtests/accumulate
	return nil
}

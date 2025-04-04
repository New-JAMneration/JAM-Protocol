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

// RunTests runs tests for the specified mode with the given test data
func RunTests(mode TestMode, testData interface{}) (*TestResult, error) {
	result := &TestResult{}

	// Run the appropriate test based on mode
	switch mode {
	case SafroleMode:
		// Run safrole STF
		result.Passed = false
		result.Error = fmt.Errorf("safrole STF not implemented")

	case AssurancesMode:
		// TODO: Implement assurances STF
		result.Passed = false
		result.Error = fmt.Errorf("assurances STF not implemented")

	case PreimagesMode:
		// TODO: Implement preimages STF
		result.Passed = false
		result.Error = fmt.Errorf("preimages STF not implemented")

	case HistoryMode:
		// TODO: Implement history STF
		result.Passed = false
		result.Error = fmt.Errorf("history STF not implemented")

	case DisputesMode:
		// TODO: Implement disputes STF
		result.Passed = false
		result.Error = fmt.Errorf("disputes STF not implemented")

	case AccumulateMode:
		// TODO: Implement accumulate STF
		result.Passed = false
		result.Error = fmt.Errorf("accumulate STF not implemented")

	case AuthorizationsMode:
		// TODO: Implement authorizations STF
		result.Passed = false
		result.Error = fmt.Errorf("authorizations STF not implemented")

	default:
		result.Passed = false
		result.Error = fmt.Errorf("unsupported test mode: %s", mode)
	}

	return result, nil
}

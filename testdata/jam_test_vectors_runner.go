package testdata

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
)

// TestResult represents the result of a test
type TestResult struct {
	TestFile string
	Passed   bool
	Error    error
}

type JamTestVectorsRunner struct {
	Mode TestMode
}

func NewJamTestVectorsRunner(mode TestMode) *JamTestVectorsRunner {
	return &JamTestVectorsRunner{Mode: mode}
}

func (r *JamTestVectorsRunner) Run(data interface{}) error {
	result := &TestResult{}

	// Set data into Store
	SetTestDataToDataStore(data)

	return stf.RunSTF()
}

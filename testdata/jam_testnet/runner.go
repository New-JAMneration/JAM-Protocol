package jamtestnet

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/testdata"
)

type JamTestNetRunner struct {
	Mode testdata.TestMode
}

func NewJamTestNetRunner(mode testdata.TestMode) *JamTestNetRunner {
	return &JamTestNetRunner{Mode: mode}
}

func (r *JamTestNetRunner) Run(data interface{}, runSTF bool) error {
	fmt.Printf("Running jamtestnet with mode: %s\n", r.Mode)
	// Add logic for running jamtestnet
	return nil
}

func (r *JamTestNetRunner) Verify(data testdata.Testable) error {
	fmt.Println("Verifying jamtestnet results...")
	// Add logic for verifying jamtestnet results
	return nil
}

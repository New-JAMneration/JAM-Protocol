package testdata

import "fmt"

type JamTestNetRunner struct {
	Mode TestMode
}

func NewJamTestNetRunner(mode TestMode) *JamTestNetRunner {
	return &JamTestNetRunner{Mode: mode}
}

func (r *JamTestNetRunner) Run(data interface{}) error {
	fmt.Printf("Running jamtestnet with mode: %s\n", r.Mode)
	// Add logic for running jamtestnet
	return nil
}

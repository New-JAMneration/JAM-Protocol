package testdata

import "github.com/New-JAMneration/JAM-Protocol/internal/stf"

// TimingRunner interface for runners that support detailed timing.
type TimingRunner interface {
	RunWithTiming(data interface{}) (bool, error, stf.STFTiming)
}

type TestRunner interface {
	Run(data interface{}, runSTF bool) error
	Verify(data Testable) error
}

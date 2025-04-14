package jamtestvector

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
)

type JamTestVectorsRunner struct {
	Mode testdata.TestMode
}

func NewJamTestVectorsRunner(mode testdata.TestMode) *JamTestVectorsRunner {
	return &JamTestVectorsRunner{Mode: mode}
}

func (r *JamTestVectorsRunner) Run(data interface{}) error {
	return stf.RunSTF()
}

func (r *JamTestVectorsRunner) Verify(data testdata.Testable) error {
	return data.Validate()
}

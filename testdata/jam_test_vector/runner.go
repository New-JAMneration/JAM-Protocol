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

func (r *JamTestVectorsRunner) Run(data interface{}, runSTF bool) error {
	if runSTF {
		return stf.RunSTF()
	}

	switch r.Mode {
	case testdata.SafroleMode:
		return stf.UpdateSafrole()
	case testdata.AccumulateMode:
		return stf.UpdateAccumlate()
	case testdata.PreimagesMode:
		return stf.UpdatePreimages()
	case testdata.DisputesMode:
		return stf.UpdateDisputes()
	case testdata.HistoryMode:
		return stf.UpdateHistory()
	case testdata.AuthorizationsMode:
		return stf.UpdateAuthorizations()
	case testdata.StatisticsMode:
		return stf.UpdateStatistics()
	case testdata.ReportsMode:
		return stf.UpdateReports()
	case testdata.AssurancesMode:
		return stf.UpdateAssurances()
	default:
		return nil
	}
}

func (r *JamTestVectorsRunner) Verify(data testdata.Testable) error {
	return data.Validate()
}

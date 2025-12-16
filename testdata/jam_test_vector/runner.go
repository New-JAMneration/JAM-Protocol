package jamtestvector

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
)

type JamTestVectorsRunner struct {
	Mode testdata.TestMode
}

func NewJamTestVectorsRunner(mode testdata.TestMode) *JamTestVectorsRunner {
	return &JamTestVectorsRunner{Mode: mode}
}

func (r *JamTestVectorsRunner) RunFnnc(runSTF bool) error {
	if runSTF {
		_, err := stf.RunSTF()
		return err
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

func (r *JamTestVectorsRunner) Run(data interface{}, runSTF bool) error {
	testCase := data.(testdata.Testable)

	// Execute STF
	err := r.RunFnnc(runSTF)
	if err != nil {
		return err
	}

	// Old passed logic
	expectedErr := testCase.ExpectError()
	if expectedErr != nil {
		if err == nil {
			return fmt.Errorf("expected error but got none")
		}
		logger.Debugf("Test passed (expected error: %v)", expectedErr)
	} else {
		if err != nil {
			return fmt.Errorf("unexpected error: %v", err)
		}
		err = r.Verify(testCase)
		if err != nil {
			return err
		}
		logger.Debug("Test passed")
	}

	return nil
}

func (r *JamTestVectorsRunner) Verify(data testdata.Testable) error {
	return data.Validate()
}

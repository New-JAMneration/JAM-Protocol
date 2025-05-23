package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/statistics"
)

func UpdateStatistics() error {
	statistics.UpdateValidatorActivityStatistics()
	return nil
}

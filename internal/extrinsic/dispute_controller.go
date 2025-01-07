package extrinsic

import (
	input "github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// DisputeController is a struct that contains a slice of Dispute (for controller logic)
type DisputeController struct {
	VerdictController *VerdictController
	FaultController   *FaultController
	CulpritController *CulpritController
}

// NewDisputeController returns a new DisputeController
func NewDisputeController(VerdictController *VerdictController, FaultController *FaultController, CulpritController *CulpritController) *DisputeController {
	return &DisputeController{
		VerdictController: VerdictController,
		FaultController:   FaultController,
		CulpritController: CulpritController,
	}
}

// ValidateFaults validates the faults in the verdict | Eq. 10.13
func (d *DisputeController) ValidateFaults() {
	faultMap := make(map[types.WorkReportHash]bool)
	for _, report := range d.FaultController.Faults {
		faultMap[report.Target] = true
	}

	good := PositiveJudgmentLevel(input.ValidatorsCount*2/3 + 1)
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == PositiveJudgmentLevel(good) {
			if !faultMap[types.WorkReportHash(report.ReportHash)] {
				panic("not_enough_faults")
			}
		}
	}

}

// ValidateCulprits validates the culprits in the verdict | Eq. 10.14
func (d *DisputeController) ValidateCulprits() {
	culpritMap := make(map[types.WorkReportHash]bool)

	for _, report := range d.CulpritController.Culprits {
		culpritMap[report.Target] = true
	}

	bad := PositiveJudgmentLevel(0)
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == PositiveJudgmentLevel(bad) {
			if !culpritMap[types.WorkReportHash(report.ReportHash)] {
				panic("not_enough_culprits")
			}
		}
	}
}

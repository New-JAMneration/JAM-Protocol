package extrinsic

import (
	"fmt"
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
func (d *DisputeController) ValidateFaults() []types.WorkReportHash {
	faultMap := make(map[types.WorkReportHash]bool)
	notInFault := make([]types.WorkReportHash, 0)
	for _, report := range d.FaultController.Faults {
		faultMap[report.Target] = true
	}

	good := PositiveJudgmentLevel(input.ValidatorsCount*2/3 + 1)
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == PositiveJudgmentLevel(good) {
			if !faultMap[types.WorkReportHash(report.ReportHash)] {
				notInFault = append(notInFault, types.WorkReportHash(report.ReportHash))
				fmt.Println("good report not in fault map : ", report.ReportHash)
			}
		}
	}

	return notInFault
}

// ValidateCulprits validates the culprits in the verdict | Eq. 10.14
func (d *DisputeController) ValidateCulprits() []types.WorkReportHash {
	culpritMap := make(map[types.WorkReportHash]bool)
	notInCulprit := make([]types.WorkReportHash, 0)

	for _, report := range d.CulpritController.Culprits {
		culpritMap[report.Target] = true
	}

	bad := PositiveJudgmentLevel(0)
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == PositiveJudgmentLevel(bad) {
			if !culpritMap[types.WorkReportHash(report.ReportHash)] {
				notInCulprit = append(notInCulprit, types.WorkReportHash(report.ReportHash))
				fmt.Println("good report not in fault map : ", report.ReportHash)
			}
		}
	}
	return notInCulprit
}

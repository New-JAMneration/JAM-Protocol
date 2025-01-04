package extrinsic

import (
	"fmt"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
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
func (d *DisputeController) ValidateFaults() []jamTypes.WorkReportHash {
	faultMap := make(map[jamTypes.WorkReportHash]bool)
	notInFault := make([]jamTypes.WorkReportHash, 0)

	for _, report := range d.FaultController.Faults {
		faultMap[report.Target] = true
	}

	for _, goodReport := range d.VerdictController.goodReports {
		if !faultMap[goodReport] {
			notInFault = append(notInFault, goodReport)
			fmt.Println("good report not in fault map : ", goodReport)
		}
	}
	return notInFault
}

// ValidateCulprits validates the culprits in the verdict | Eq. 10.14
func (d *DisputeController) ValidateCulprits() []jamTypes.WorkReportHash {
	culpritMap := make(map[jamTypes.WorkReportHash]bool)
	notInCulprit := make([]jamTypes.WorkReportHash, 0)

	for _, report := range d.CulpritController.Culprits {
		culpritMap[report.Target] = true
	}

	for _, badReport := range d.VerdictController.badReports {
		if !culpritMap[badReport] {
			notInCulprit = append(notInCulprit, badReport)
			fmt.Println("bad report not in culprit map : ", badReport)
		}
	}
	return notInCulprit
}

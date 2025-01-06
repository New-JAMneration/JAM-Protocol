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

	good := input.ValidatorsCount*2/3 + 1
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == good {
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

	bad := 0
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == bad {
			if !culpritMap[types.WorkReportHash(report.ReportHash)] {
				notInCulprit = append(notInCulprit, types.WorkReportHash(report.ReportHash))
				fmt.Println("good report not in fault map : ", report.ReportHash)
			}
		}
	}
	return notInCulprit
}

// UpdatePsi updates the Dispute state | Eq. 10.16~19
func (d *DisputeController) UpdatePsi(priorStatesPsi types.DisputesRecords, newVerdicts []VerdictSummary, newCulprits []types.Culprit, newFaults []types.Fault) types.DisputesRecords {
	updateVerdict := CompareVerdictsWithPsi(priorStatesPsi, newVerdicts)

	goodMap := make(map[types.WorkReportHash]bool)
	badMap := make(map[types.WorkReportHash]bool)
	wonkyMap := make(map[types.WorkReportHash]bool)
	offendersMap := make(map[types.Ed25519Public]bool)

	for _, good := range priorStatesPsi.Good {
		goodMap[good] = true
	}
	for _, bad := range priorStatesPsi.Bad {
		badMap[bad] = true
	}
	for _, wonky := range priorStatesPsi.Wonky {
		wonkyMap[wonky] = true
	}
	for _, offenders := range priorStatesPsi.Offenders {
		offendersMap[offenders] = true
	}

	updateListAndMap(&priorStatesPsi.Good, updateVerdict.Good, goodMap)
	updateListAndMap(&priorStatesPsi.Bad, updateVerdict.Bad, badMap)
	updateListAndMap(&priorStatesPsi.Wonky, updateVerdict.Wonky, wonkyMap)

	offenders := make([]types.Ed25519Public, 0)
	for _, culprit := range newCulprits {
		offenders = append(offenders, culprit.Key)
	}

	for _, fault := range newFaults {
		offenders = append(offenders, fault.Key)
	}

	for _, newCulpritAndFault := range offenders {
		if !offendersMap[newCulpritAndFault] {
			priorStatesPsi.Offenders = append(priorStatesPsi.Offenders, newCulpritAndFault)
		}
	}
	return priorStatesPsi
}

func updateListAndMap(list *[]types.WorkReportHash, newItems []types.WorkReportHash, itemMap map[types.WorkReportHash]bool) {
	for _, item := range newItems {
		if !itemMap[item] {
			*list = append(*list, item)
			itemMap[item] = true
		}
	}
}

func CompareVerdictsWithPsi(disputeState types.DisputesRecords, verdictSumSequence []VerdictSummary) types.DisputesRecords {
	var updates types.DisputesRecords
	for _, verdict := range verdictSumSequence {
		if verdict.PositiveJudgmentsSum == types.ValidatorsCount*2/3+1 {
			updates.Good = append(updates.Good, types.WorkReportHash(verdict.ReportHash))
		} else if verdict.PositiveJudgmentsSum == 0 {
			updates.Bad = append(updates.Bad, types.WorkReportHash(verdict.ReportHash))
		} else if verdict.PositiveJudgmentsSum == types.ValidatorsCount*1/3 {
			updates.Wonky = append(updates.Wonky, types.WorkReportHash(verdict.ReportHash))
		}
	}
	return updates
}

// HeaderOffenders returns the offenders markers | Eq. 10.20
func (d *DisputeController) HeaderOffenders(newCulprits []types.Culprit, newFaults []types.Fault) []types.Ed25519Public {
	offendersMarkers := make([]types.Ed25519Public, 0)
	for _, culprit := range newCulprits {
		offendersMarkers = append(offendersMarkers, culprit.Key)
	}
	for _, fault := range newFaults {
		offendersMarkers = append(offendersMarkers, fault.Key)
	}
	return offendersMarkers
}

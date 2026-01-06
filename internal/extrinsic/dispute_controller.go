package extrinsic

import (
	"bytes"
	"errors"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
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
func (d *DisputeController) ValidateFaults() error {
	faultMap := make(map[types.WorkReportHash]bool)
	for _, report := range d.FaultController.Faults {
		faultMap[report.Target] = true
	}

	good := types.ValidatorsCount*2/3 + 1
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == good {
			if !faultMap[types.WorkReportHash(report.ReportHash)] {
				return errors.New("not_enough_faults")
			}
		}
	}
	return nil
}

// ValidateCulprits validates the culprits in the verdict | Eq. 10.14
func (d *DisputeController) ValidateCulprits() error {
	culpritMap := make(map[types.WorkReportHash]int)

	for _, report := range d.CulpritController.Culprits {
		culpritMap[report.Target]++
	}

	bad := 0
	for _, report := range d.VerdictController.VerdictSumSequence {
		if report.PositiveJudgmentsSum == bad {
			if culpritMap[types.WorkReportHash(report.ReportHash)] < 2 {
				return errors.New("not_enough_culprits")
			}
		}
	}
	return nil
}

// UpdatePsiGBW updates the PsiG, PsiB, and PsiW | Eq. 10.16, 17, 18
func (d *DisputeController) UpdatePsiGBW(newVerdicts []VerdictSummary) error {
	priorPsi := blockchain.GetInstance().GetPriorStates().GetPsi()
	updateVerdicts, err := CompareVerdictsWithPsi(priorPsi, newVerdicts)
	if err != nil {
		return err
	}

	posteriorPsiG := UpdatePsiG(priorPsi, updateVerdicts)
	posteriorPsiB := UpdatePsiB(priorPsi, updateVerdicts)
	posteriorPsiW := UpdatePsiW(priorPsi, updateVerdicts)
	posteriorState := blockchain.GetInstance().GetPosteriorStates()
	posteriorState.SetPsiG(posteriorPsiG)
	posteriorState.SetPsiB(posteriorPsiB)
	posteriorState.SetPsiW(posteriorPsiW)
	return nil
}

func CompareVerdictsWithPsi(disputeState types.DisputesRecords, verdictSumSequence []VerdictSummary) (types.DisputesRecords, error) {
	var updates types.DisputesRecords
	for _, verdict := range verdictSumSequence {
		if verdict.PositiveJudgmentsSum == types.ValidatorsCount*2/3+1 {
			updates.Good = append(updates.Good, types.WorkReportHash(verdict.ReportHash))
		} else if verdict.PositiveJudgmentsSum == 0 {
			updates.Bad = append(updates.Bad, types.WorkReportHash(verdict.ReportHash))
		} else if verdict.PositiveJudgmentsSum == types.ValidatorsCount*1/3 {
			updates.Wonky = append(updates.Wonky, types.WorkReportHash(verdict.ReportHash))
		} else {
			return types.DisputesRecords{}, errors.New("bad_vote_split")
		}
	}
	return updates, nil
}

func UpdatePsiG(priorPsi, updateVerdicts types.DisputesRecords) []types.WorkReportHash {
	goodMap := make(map[types.WorkReportHash]bool)
	for _, good := range priorPsi.Good {
		goodMap[good] = true
	}
	return updateListAndMap(priorPsi.Good, updateVerdicts.Good, goodMap)
}

func UpdatePsiB(priorPsi, updateVerdicts types.DisputesRecords) []types.WorkReportHash {
	badMap := make(map[types.WorkReportHash]bool)
	for _, bad := range priorPsi.Bad {
		badMap[bad] = true
	}
	return updateListAndMap(priorPsi.Bad, updateVerdicts.Bad, badMap)
}

func UpdatePsiW(priorPsi, updateVerdicts types.DisputesRecords) []types.WorkReportHash {
	wonkyMap := make(map[types.WorkReportHash]bool)
	for _, wonky := range priorPsi.Wonky {
		wonkyMap[wonky] = true
	}
	return updateListAndMap(priorPsi.Wonky, updateVerdicts.Wonky, wonkyMap)
}

func updateListAndMap(list []types.WorkReportHash, newItems []types.WorkReportHash, itemMap map[types.WorkReportHash]bool) []types.WorkReportHash {
	for _, item := range newItems {
		if !itemMap[item] {
			list = append(list, item)
			itemMap[item] = true
		}
	}
	return list
}

// UpdatePsiO updates the PsiO | Eq. 10.19
func (d *DisputeController) UpdatePsiO(culprits []types.Culprit, faults []types.Fault) {
	s := blockchain.GetInstance()
	priorPsi := s.GetPriorStates().GetPsi()

	offenderMap := make(map[types.Ed25519Public]bool, len(priorPsi.Offenders))
	for _, k := range priorPsi.Offenders {
		offenderMap[k] = true
	}

	posteriorPsiO := make([]types.Ed25519Public, 0, len(culprits)+len(faults))
	for _, c := range culprits {
		if !offenderMap[c.Key] {
			posteriorPsiO = append(posteriorPsiO, c.Key)
			offenderMap[c.Key] = true
		}
	}
	for _, f := range faults {
		if !offenderMap[f.Key] {
			posteriorPsiO = append(posteriorPsiO, f.Key)
			offenderMap[f.Key] = true
		}
	}

	psiO := append([]types.Ed25519Public(nil), priorPsi.Offenders...)
	psiO = append(psiO, posteriorPsiO...)
	sort.Slice(psiO, func(i, j int) bool {
		return bytes.Compare(psiO[i][:], psiO[j][:]) < 0
	})

	s.GetPosteriorStates().SetPsiO(psiO)
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

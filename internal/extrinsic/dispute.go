package extrinsic

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func Disputes(disputeExtrinsic types.DisputesExtrinsic) ([]types.Ed25519Public, error) {
	// init controllers
	verdictController := NewVerdictController()
	for _, verdict := range disputeExtrinsic.Verdicts {
		verdictController.Verdicts = append(verdictController.Verdicts, VerdictWrapper{verdict})
	}
	culpritController := NewCulpritController()
	culpritController.Culprits = disputeExtrinsic.Culprits
	faultController := NewFaultController()
	faultController.Faults = disputeExtrinsic.Faults

	// verify verdicts
	for i := 0; i < len(verdictController.Verdicts); i++ {
		VerdictPtr := &verdictController.Verdicts[i]
		invalid := VerdictPtr.VerifySignature()
		if len(invalid) > 0 {
			return nil, fmt.Errorf("invalid signature in verdict")
		}
	}
	verdictController.SortUnique()
	if err := verdictController.SetDisjoint(); err != nil {
		return nil, err
	}

	verdictController.GenerateVerdictSumSequence()
	disputeController := NewDisputeController(verdictController, faultController, culpritController)

	if err := disputeController.ValidateCulprits(); err != nil {
		return nil, err
	}
	if err := disputeController.ValidateFaults(); err != nil {
		return nil, err
	}

	culpritController.SortUnique()
	faultController.SortUnique()

	// update state
	verdictController.ClearWorkReports(verdictController.VerdictSumSequence)
	disputeController.UpdatePsiGBW(verdictController.VerdictSumSequence)

	if err := culpritController.VerifyCulpritValidity(); err != nil {
		return nil, err
	}
	if err := faultController.VerifyFaultValidity(); err != nil {
		return nil, err
	}

	disputeController.UpdatePsiO(culpritController.Culprits, faultController.Faults)
	output := disputeController.HeaderOffenders(culpritController.Culprits, faultController.Faults)

	return output, nil
}

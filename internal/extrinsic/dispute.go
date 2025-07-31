package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	DisputesErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/disputes"
)

func Disputes(disputeExtrinsic types.DisputesExtrinsic) (types.OffendersMark, error) {
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
		err := VerdictPtr.VerifySignature()
		if err != nil {
			errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
			return nil, &errCode
		}
	}

	if err := verdictController.CheckSortUnique(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}
	if err := verdictController.SetDisjoint(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}

	verdictController.GenerateVerdictSumSequence()
	disputeController := NewDisputeController(verdictController, faultController, culpritController)

	if err := disputeController.ValidateCulprits(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}
	if err := disputeController.ValidateFaults(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}

	if err := culpritController.CheckSortUnique(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}
	if err := faultController.CheckSortUnique(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}

	// update state
	verdictController.ClearWorkReports(verdictController.VerdictSumSequence)
	err := disputeController.UpdatePsiGBW(verdictController.VerdictSumSequence)
	if err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}

	if err := culpritController.VerifyCulpritValidity(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}
	if err := faultController.VerifyFaultValidity(); err != nil {
		errCode := DisputesErrorCode.DisputesErrorMap[err.Error()]
		return nil, &errCode
	}

	disputeController.UpdatePsiO(culpritController.Culprits, faultController.Faults)
	output := disputeController.HeaderOffenders(culpritController.Culprits, faultController.Faults)
	offendersMark := types.OffendersMark(output)
	return offendersMark, nil
}

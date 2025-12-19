package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// Assurance is a struct that contains a slice of Assurance
func Assurance() (err *types.ErrorCode) {
	block := store.GetInstance().GetLatestBlock()
	assuranceExtrinsic := block.Extrinsic.Assurances
	assurances := AvailAssuranceController{AvailAssurances: assuranceExtrinsic}

	err = assurances.ValidateAnchor()
	if err != nil {
		logger.Errorf("ValidateAnchor failed: %v", err)
		return err
	}

	err = assurances.CheckValidatorIndex()
	if err != nil {
		logger.Errorf("CheckValidatorIndex failed: %v", err)
		return err
	}

	err = assurances.SortUnique()
	if err != nil {
		logger.Errorf("SortUnique failed: %v", err)
		return err
	}

	err = assurances.ValidateSignature()
	if err != nil {
		logger.Errorf("ValidateSignature failed: %v", err)
		return err
	}

	err = assurances.ValidateBitField()
	if err != nil {
		logger.Errorf("ValidateBitField failed: %v", err)
		return err
	}

	err = assurances.FilterAvailableReports()
	if err != nil {
		logger.Errorf("FilterAvailableReports failed: %v", err)
		return err
	}
	return nil
}

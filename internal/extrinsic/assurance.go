package extrinsic

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

// Assurance is a struct that contains a slice of Assurance
func Assurance() (err error) {
	assuranceExtrinsic := store.GetInstance().GetProcessingBlockPointer().GetAssurancesExtrinsic()
	assurances := AvailAssuranceController{AvailAssurances: assuranceExtrinsic}

	err = assurances.ValidateAnchor()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = assurances.CheckValidatorIndex()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = assurances.SortUnique()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = assurances.ValidateSignature()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = assurances.ValidateBitField()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = assurances.FilterAvailableReports()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

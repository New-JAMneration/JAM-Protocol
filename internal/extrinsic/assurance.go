package extrinsic

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Assurance is a struct that contains a slice of Assurance
func Assurance(assuranceExtrinsic types.AssurancesExtrinsic) (err error) {
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

	if err == nil || err.Error() != "bad_validator_index" {
		err = assurances.ValidateSignature()
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	err = assurances.ValidateBitField()
	if err != nil {
		fmt.Println(err)
		return err
	}

	assurances.FilterAvailableReports()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

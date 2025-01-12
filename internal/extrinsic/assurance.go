package extrinsic

import (
	"fmt"
)

func Assurance() {
	availAssureance := NewAvailAssuranceController{}

	availAssureance.ValidateAnchor()
	availAssurance.SortUnique()
	availAssurance.ValidateSignature()

	if err := availAssurance.ValidateBitField(); err != nil {
		fmt.Println(err)
	}

	availAssurance.FilterAvailableReports()
}

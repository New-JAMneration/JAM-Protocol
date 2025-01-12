package extrinsic

import (
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Assurance is a struct that contains a slice of Assurance
func Assurance(assuranceExtrinsic types.AssurancesExtrinsic) {
	assurances := NewAvailAssuranceController() // input data
	assurances.AvailAssurances = assuranceExtrinsic

	assurances.ValidateAnchor()

	assurances.SortUnique()

	assurances.ValidateSignature()

	assurances.BitfieldOctetSequenceToBinarySequence()

	if err := assurances.ValidateBitField(); err != nil {
		fmt.Println(err)
	}

	assurances.FilterAvailableReports()
}

package extrinsic

import (
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

// VerdictController is a struct that contains a slice of Verdict
type VerdictController struct {
	Verdicts []jamTypes.Verdict `json:"verdicts,omitempty"`
}

// NewVerdictController returns a new VerdictController
func NewVerdictController() *VerdictController {
	return &VerdictController{
		Verdicts: make([]jamTypes.Verdict, 0),
	}
}

// CulpritController is a struct that contains a slice of Culprit
type CulpritController struct {
	Culprits []jamTypes.Culprit `json:"culprits,omitempty"`
}

// NewCulpritController returns a new CulpritController
func NewCulpritController() *CulpritController {
	return &CulpritController{
		Culprits: make([]jamTypes.Culprit, 0),
	}
}

// FaultController is a struct that contains a slice of Fault
type FaultController struct {
	Faults []jamTypes.Fault `json:"faults,omitempty"`
}

// NewFaultController returns a new FaultController
func NewFaultController() *FaultController {
	return &FaultController{
		Faults: make([]jamTypes.Fault, 0),
	}
}

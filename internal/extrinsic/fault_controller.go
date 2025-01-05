package extrinsic

import (
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// FaultController is a struct that contains a slice of Fault
type FaultController struct {
	Faults []types.Fault `json:"faults,omitempty"`
	/*
		type Fault struct {
			Target    WorkReportHash   `json:"target,omitempty"`		r
			Vote      bool             `json:"vote,omitempty"`			v
			Key       Ed25519Public    `json:"key,omitempty"`			k
			Signature Ed25519Signature `json:"signature,omitempty"`		s
		}
	*/
}

// NewFaultController returns a new FaultController
func NewFaultController() *FaultController {
	return &FaultController{
		Faults: make([]types.Fault, 0),
	}
}

func (f *FaultController) VerifyFaultValidity() bool {
	states := store.GetInstance().GetStates()
	psiBad := states.GetState().Psi.Bad //  psi_b (bad report) will first update using verdicts in Eq. 10.17

	f.Faults = f.VerifyReportHashValidty(&psiBad)
	f.Faults = f.ExcludeOffenders()

	// the testvectors do not have the case of verifySignature at Eq. 10.6. Besides, do not how to handle the invalid signature

	return true
}

func (f *FaultController) VerifyReportHashValidty(psiBad *[]types.WorkReportHash) []types.Fault {
	checkMap := make(map[types.WorkReportHash]bool)
	for _, report := range *psiBad {
		checkMap[report] = true
	}

	var out []types.Fault
	length := len(f.Faults)
	for i := 0; i < length; i++ {
		vote := f.Faults[i].Vote

		if !vote && checkMap[f.Faults[i].Target] { // normal condition  : r have to be in psi_b (bad) and it is in psi_b (bad)
			out = append(out, f.Faults[i])
		}
	}

	return out
}

// ExcludeOffenders excludes the offenders from the validator set  Eq. 10.6  exclude psi_o will be used in verdict, fault, culprit
// Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
func (f *FaultController) ExcludeOffenders() []types.Fault {

	exclude := store.GetInstance().GetState().Psi.Offenders
	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(f.Faults)

	var out []types.Fault
	for i := 0; i < length; i++ { // culprit index

		if !excludeMap[f.Faults[i].Key] {
			out = append(out, f.Faults[i])
		}
	}
	return out
}

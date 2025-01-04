package extrinsic

import (
	"bytes"
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"sort"
)

// FaultController is a struct that contains a slice of Fault
type FaultController struct {
	Faults []jamTypes.Fault `json:"faults,omitempty"`
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
		Faults: make([]jamTypes.Fault, 0),
	}
}

func (f *FaultController) VerifyFaultValidity() bool {
	postStates := store.GetInstance().GetPosteriorStates()
	psiBad := postStates.GetState().Psi.Bad //  psi_b (bad report) will first update using verdicts in Eq. 10.17

	f.Faults = f.VerifyReportHashValidty(&psiBad)
	f.Faults = f.ExcludeOffenders()

	// the testvectors do not have the case of verifySignature at Eq. 10.6. Besides, do not how to handle the invalid signature

	return true
}

func (f *FaultController) VerifyReportHashValidty(psiBad *[]jamTypes.WorkReportHash) []jamTypes.Fault {
	checkMap := make(map[jamTypes.WorkReportHash]bool)
	for _, report := range *psiBad {
		checkMap[report] = true
	}

	var out []jamTypes.Fault
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
func (f *FaultController) ExcludeOffenders() []jamTypes.Fault {

	exclude := store.GetInstance().GetPriorState().Psi.Offenders
	excludeMap := make(map[jamTypes.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(f.Faults)

	var out []jamTypes.Fault
	for i := 0; i < length; i++ { // culprit index

		if !excludeMap[f.Faults[i].Key] {
			out = append(out, f.Faults[i])
		}
	}
	return out
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.8
func (f *FaultController) SortUnique() {
	f.Sort()
	f.Unique()
}

func (f *FaultController) Unique() {
	if len(f.Faults) == 0 {
		return
	}
}

func (f *FaultController) Sort() {
	sort.Sort(f)
}

func (f *FaultController) Less(i, j int) bool {
	return bytes.Compare(f.Faults[i].Key[:], f.Faults[j].Key[:]) < 0
}

func (f *FaultController) Swap(i, j int) {
	f.Faults[i], f.Faults[j] = f.Faults[j], f.Faults[i]
}

func (f *FaultController) Len() int {
	return len(f.Faults)
}

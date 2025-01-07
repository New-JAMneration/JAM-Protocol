package extrinsic

import (
	"bytes"
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"sort"
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

// VerifyFaultValidity verifies the validity of the faults | Eq. 10.6
func (f *FaultController) VerifyFaultValidity() bool {
	// if the faults are not valid, panic
	f.VerifyReportHashValidty()
	f.ExcludeOffenders()
	return true
}

// VerifyReportHashValidty verifies the validity of the reports
func (f *FaultController) VerifyReportHashValidty() {
	psiBad := store.GetInstance().GetPosteriorStates().GetState().Psi.Bad
	psiGood := store.GetInstance().GetPosteriorStates().GetState().Psi.Good

	badMap := make(map[types.WorkReportHash]bool)
	for _, report := range psiBad {
		badMap[report] = true
	}

	goodMap := make(map[types.WorkReportHash]bool)
	for _, report := range psiGood {
		goodMap[report] = true
	}

	length := len(f.Faults)
	for i := 0; i < length; i++ {
		vote := f.Faults[i].Vote
		// if vote : true  || the report is not in the bad list || the report is in the good list => panic
		if vote || !badMap[f.Faults[i].Target] || goodMap[f.Faults[i].Target] {
			panic("fault_verdict_wrong")
		}
	}
}

// ExcludeOffenders excludes the offenders from the validator set
func (f *FaultController) ExcludeOffenders() {

	exclude := store.GetInstance().GetPriorState().Psi.Offenders
	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(f.Faults)
	for i := 0; i < length; i++ { // culprit index
		if !excludeMap[f.Faults[i].Key] {
			panic("offenders_already_judged")
		}
	}
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

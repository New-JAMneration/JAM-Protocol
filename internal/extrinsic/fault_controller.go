package extrinsic

import (
	"bytes"
	"fmt"
	"sort"

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

// VerifyFaultValidity verifies the validity of the faults | Eq. 10.6
func (f *FaultController) VerifyFaultValidity() error {
	// if the faults are not valid, return error
	if err := f.VerifyReportHashValidty(); err != nil {
		return err
	}
	if err := f.ExcludeOffenders(); err != nil {
		return err
	}
	return nil
}

// VerifyReportHashValidty verifies the validity of the reports
func (f *FaultController) VerifyReportHashValidty() error {
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
		// if vote not contradict verdict, should not be in faults
		inGood := goodMap[f.Faults[i].Target] && !badMap[f.Faults[i].Target]
		inBad := !goodMap[f.Faults[i].Target] && badMap[f.Faults[i].Target]
		if (vote && inGood) || (!vote && inBad) {
			return fmt.Errorf("FaultController.VerifyReportHashValidty failed : fault_verdict_wrong")
		}
	}
	return nil
}

// ExcludeOffenders excludes the offenders from the validator set
func (f *FaultController) ExcludeOffenders() error {

	exclude := store.GetInstance().GetPriorState().Psi.Offenders
	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(f.Faults)
	for i := 0; i < length; i++ { // culprit index
		if excludeMap[f.Faults[i].Key] {
			return fmt.Errorf("FaultController.ExcludeOffenders failed : offenders_already_judged")
		}
	}
	return nil
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.8
func (f *FaultController) SortUnique() {
	f.Sort()
	f.Unique()
}

// Unique removes duplicates
func (f *FaultController) Unique() {
	if len(f.Faults) == 0 {
		return
	}
}

// Sort sorts the faults
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

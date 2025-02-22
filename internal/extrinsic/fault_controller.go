package extrinsic

import (
	"bytes"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
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
	posteriorStates := store.GetInstance().GetPosteriorStates()
	psiBad := posteriorStates.GetPsiB()
	psiGood := posteriorStates.GetPsiG()

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
			return fmt.Errorf("fault_verdict_wrong")
		}
	}
	return nil
}

// ExcludeOffenders excludes the offenders from the validator set
func (f *FaultController) ExcludeOffenders() error {

	exclude := store.GetInstance().GetPriorStates().GetPsiO()
	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(f.Faults)
	for i := 0; i < length; i++ { // culprit index
		if excludeMap[f.Faults[i].Key] {
			return fmt.Errorf("offender_already_reported")
		}
	}
	return nil
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.8
func (f *FaultController) CheckSortUnique() error {
	if err := f.CheckUnique(); err != nil {
		return err
	}
	if err := f.CheckSorted(); err != nil {
		return err
	}
	return nil
}

// Unique removes duplicates
func (f *FaultController) CheckUnique() error {
	if len(f.Faults) == 0 {
		return nil
	}
	uniqueMap := make(map[types.Ed25519Public]bool)
	uniqueFaults := make([]types.Fault, 0)
	for _, fault := range f.Faults {
		if uniqueMap[fault.Key] {
			return fmt.Errorf("faults_key_not_unique")
		}
		uniqueMap[fault.Key] = true
		uniqueFaults = append(uniqueFaults, fault)
	}
	f.Faults = uniqueFaults
	return nil
}

func (f *FaultController) CheckSorted() error {
	for i := 1; i < len(f.Faults); i++ {
		if bytes.Compare(f.Faults[i-1].Key[:], f.Faults[i].Key[:]) > 0 {
			return fmt.Errorf("faults_not_sorted")
		}
	}
	return nil
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

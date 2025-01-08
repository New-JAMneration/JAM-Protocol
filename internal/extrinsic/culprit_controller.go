package extrinsic

import (
	"bytes"
	"fmt"
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"sort"
)

// CulpritController is a struct that contains a slice of Culprit
type CulpritController struct {
	Culprits []types.Culprit `json:"culprits,omitempty"`
	/*
		type Culprit struct {
			Target    WorkReportHash   `json:"target,omitempty"`		r
			Key       Ed25519Public    `json:"key,omitempty"`			k
			Signature Ed25519Signature `json:"signature,omitempty"`		s
		}
	*/
}

// NewCulpritController returns a new CulpritController
func NewCulpritController() *CulpritController {

	return &CulpritController{
		Culprits: make([]types.Culprit, 0),
	}
}

// VerifyCulpritValidity verifies the validity of the culprits | Eq. 10.5
func (c *CulpritController) VerifyCulpritValidity() {
	// if the culprits are not valid, panic
	c.VerifyReportHashValidty()
	c.ExcludeOffenders()
}

// VerifyReportHashValidty verifies the validity of the reports
func (c *CulpritController) VerifyReportHashValidty() error {
	psiBad := store.GetInstance().GetPosteriorStates().GetState().Psi.Bad
	checkMap := make(map[types.WorkReportHash]bool)

	for _, report := range psiBad {
		checkMap[report] = true
	}

	for _, report := range c.Culprits {
		if !checkMap[report.Target] {
			return fmt.Errorf("culprits_not_in_bad")
		}
	}
	return nil
}

// ExcludeOffenders excludes the offenders from the validator set
// Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
func (c *CulpritController) ExcludeOffenders() error {

	exclude := store.GetInstance().GetPriorState().Psi.Offenders

	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(c.Culprits)

	for i := 0; i < length; i++ { // culprit index
		if !excludeMap[c.Culprits[i].Key] {
			return fmt.Errorf("offenders_already_judged")
		}
	}
	return nil
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.8
func (c *CulpritController) SortUnique() {
	c.Unique()
	c.Sort()
}

// Unique removes duplicates
func (c *CulpritController) Unique() {
	if len(c.Culprits) == 0 {
		return
	}

	uniqueMap := make(map[types.Ed25519Public]bool)
	result := make([]types.Culprit, 0)

	for _, culprit := range c.Culprits {
		if !uniqueMap[culprit.Key] {
			uniqueMap[culprit.Key] = true
			result = append(result, culprit)
		}
	}
	c.Culprits = result
}

// Sort sorts the slice
func (c *CulpritController) Sort() {
	sort.Sort(c)
}

func (c *CulpritController) Less(i, j int) bool {
	return bytes.Compare(c.Culprits[i].Key[:], c.Culprits[j].Key[:]) < 0
}

func (c *CulpritController) Swap(i, j int) {
	c.Culprits[i], c.Culprits[j] = c.Culprits[j], c.Culprits[i]
}

func (c *CulpritController) Len() int {
	return len(c.Culprits)
}

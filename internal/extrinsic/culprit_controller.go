package extrinsic

import (
	"bytes"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/hdevalence/ed25519consensus"
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
func (c *CulpritController) VerifyCulpritValidity() error {
	// if the culprits are not valid return error
	if err := c.VerifyReportHashValidty(); err != nil {
		return err
	}
	if err := c.VerifyCulpritSignature(); err != nil {
		return err
	}
	if err := c.ExcludeOffenders(); err != nil {
		return err
	}
	return nil
}

func (c *CulpritController) VerifyCulpritSignature() error {
	state := store.GetInstance().GetPriorStates()
	posterior := store.GetInstance().GetPosteriorStates()

	validators := append(state.GetKappa(), state.GetLambda()...)
	validKeySet := make(map[types.Ed25519Public]struct{})
	for _, v := range validators {
		validKeySet[v.Ed25519] = struct{}{}
	}

	psiO := posterior.GetPsiO()
	for _, offender := range psiO {
		delete(validKeySet, offender)
	}

	for _, culprit := range c.Culprits {
		if _, ok := validKeySet[culprit.Key]; !ok {
			return fmt.Errorf("bad_guarantor_key")
		}
		msg := []byte(types.JamGuarantee)
		msg = append(msg, culprit.Target[:]...)
		if !ed25519consensus.Verify(culprit.Key[:], msg, culprit.Signature[:]) {
			return fmt.Errorf("bad_signature")
		}
	}
	return nil
}

// VerifyReportHashValidty verifies the validity of the reports
func (c *CulpritController) VerifyReportHashValidty() error {
	psiBad := store.GetInstance().GetPosteriorStates().GetPsiB()
	checkMap := make(map[types.WorkReportHash]bool)

	for _, report := range psiBad {
		checkMap[report] = true
	}

	for _, report := range c.Culprits {
		if !checkMap[report.Target] {
			return fmt.Errorf("culprits_verdict_not_bad")
		}
	}
	return nil
}

// ExcludeOffenders excludes the offenders from the validator set
// Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
func (c *CulpritController) ExcludeOffenders() error {

	exclude := store.GetInstance().GetPriorStates().GetPsiO()

	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(c.Culprits)

	for i := 0; i < length; i++ { // culprit index
		if excludeMap[c.Culprits[i].Key] {
			return fmt.Errorf("offender_already_reported")
		}
	}
	return nil
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.8
func (c *CulpritController) CheckSortUnique() error {
	if err := c.CheckUnique(); err != nil {
		return err
	}
	if err := c.CheckSorted(); err != nil {
		return err
	}
	return nil
}

// Unique removes duplicates
func (c *CulpritController) CheckUnique() error {
	if len(c.Culprits) == 0 {
		return nil
	}

	uniqueKeyMap := make(map[types.Ed25519Public]bool)
	result := make([]types.Culprit, 0)
	for _, culprit := range c.Culprits {
		if uniqueKeyMap[culprit.Key] {
			return fmt.Errorf("culprits_not_sorted_unique")
		}
		uniqueKeyMap[culprit.Key] = true
		result = append(result, culprit)
	}
	c.Culprits = result
	return nil
}

func (v *CulpritController) CheckSorted() error {
	for i := 1; i < len(v.Culprits); i++ {
		if bytes.Compare(v.Culprits[i-1].Key[:], v.Culprits[i].Key[:]) > 0 {
			return fmt.Errorf("culprits_not_sorted_unique")
		}
	}

	return nil
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

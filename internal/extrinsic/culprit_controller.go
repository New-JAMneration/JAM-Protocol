package extrinsic

import (
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

func (c *CulpritController) VerifyCulpritValidity() bool {
	states := store.GetInstance().GetStates()
	psiBad := states.GetState().Psi.Bad //  psi_b (bad report) will first update using verdicts in Eq. 10.17

	c.Culprits = c.VerifyReportHashValidty(&psiBad)
	c.Culprits = c.ExcludeOffenders()

	// the testvectors do not have the case of verifySignature at Eq. 10.5. Besides, do not how to handle the invalid signature

	return true
}

func (c *CulpritController) VerifyReportHashValidty(psiBad *[]types.WorkReportHash) []types.Culprit {
	checkMap := make(map[types.WorkReportHash]bool)

	for _, report := range *psiBad {
		checkMap[report] = true
	}

	var out []types.Culprit
	for _, report := range c.Culprits {
		if checkMap[report.Target] {
			out = append(out, report)
		}
	}
	return out
}

// ExcludeOffenders excludes the offenders from the validator set  Eq. 10.6  exclude psi_o will be used in verdict, fault, culprit
// Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
func (c *CulpritController) ExcludeOffenders() []types.Culprit {
	exclude := store.GetInstance().GetState().Psi.Offenders
	excludeMap := make(map[types.Ed25519Public]bool)
	for _, offenderEd25519 := range exclude {
		excludeMap[offenderEd25519] = true // true : the offender is in the exclude list
	}

	length := len(c.Culprits)

	var out []types.Culprit
	for i := 0; i < length; i++ { // culprit index

		if !excludeMap[c.Culprits[i].Key] {
			out = append(out, c.Culprits[i])
		}
	}
	return out
}

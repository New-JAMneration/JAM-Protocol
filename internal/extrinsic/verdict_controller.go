package extrinsic

import (
	"crypto/ed25519"

	input "github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type VerdictWrapper struct {
	Verdict jamTypes.Verdict
}

// VerdictController is a struct that contains a slice of Verdict
type VerdictController struct {
	Verdicts []VerdictWrapper
	/*
		type Verdict struct {
			Target OpaqueHash  `json:"target,omitempty"`
			Age    U32         `json:"age,omitempty"`
			Votes  []Judgement `json:"votes,omitempty"`
		}
			type Judgement struct {
				Vote      bool             `json:"vote,omitempty"`
				Index     ValidatorIndex   `json:"index,omitempty"`
				Signature Ed25519Signature `json:"signature,omitempty"`
			}
	*/
}

// NewVerdictController returns a new VerdictController
func NewVerdictController() *VerdictController {
	return &VerdictController{
		Verdicts: make([]VerdictWrapper, 0),
	}
}

// VerifySignature verifies the signatures of the judgement in the verdict   , Eq. 10.3
// currently return []int to check the test, it might change after connect other components in Ch.10
func (v *VerdictWrapper) VerifySignature() []int {
	state := store.GetInstance().GetState()
	a := jamTypes.U32(state.Tau) / jamTypes.U32(jamTypes.EpochLength)
	var k jamTypes.ValidatorsData
	if v.Verdict.Age == a {
		k = state.Kappa
	} else {
		k = state.Lambda
	}

	// check if the judgement is valid
	VoteNum := len(v.Verdict.Votes)
	target := v.Verdict.Target[:]

	// store the index of votes with invalid signature
	invalidVotes := make([]int, 0)

	for i := 0; i < VoteNum; i++ {
		publicKey := k[v.Verdict.Votes[i].Index].Ed25519[:]
		var message []byte
		if v.Verdict.Votes[i].Vote {
			message = []byte(input.JamValid)
		} else {
			message = []byte(input.JamInvalid)
		}

		message = append(message, target...)

		if !ed25519.Verify(publicKey, message, v.Verdict.Votes[i].Signature[:]) {
			invalidVotes = append(invalidVotes, i)
		}
	}

	return invalidVotes
}

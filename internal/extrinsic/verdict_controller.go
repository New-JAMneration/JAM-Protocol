package extrinsic

import (
	"bytes"
	"crypto/ed25519"
	"sort"

	input "github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type VerdictWrapper struct {
	Verdict types.Verdict
}

// VerdictController is a struct that contains a slice of Verdict
type VerdictController struct {
	Verdicts     []VerdictWrapper
	goodReports  []types.WorkReportHash
	badReports   []types.WorkReportHash
	wonkyReports []types.WorkReportHash

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
	state := store.GetInstance().GetPriorState()
	a := types.U32(state.Tau) / types.U32(types.EpochLength)
	var k types.ValidatorsData
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

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.7, Eq. 10.10
func (v *VerdictController) SortUnique() {
	v.Unique()
	v.Sort()

}

// Unique removes duplicates
func (v *VerdictController) Unique() {
	if len(v.Verdicts) == 0 {
		return
	}
	// Eq. 10.7 unique
	uniqueMap := make(map[types.OpaqueHash]bool)
	result := make([]VerdictWrapper, 0)

	for _, v := range v.Verdicts {
		if !uniqueMap[v.Verdict.Target] {
			uniqueMap[v.Verdict.Target] = true
			result = append(result, v)
		}
	}
	(*v).Verdicts = result
	// Eq. 10.10 unique
	for _, v := range v.Verdicts {
		for i := 0; i < len(v.Verdict.Votes); i++ {
			uniqueJudgementMap := make(map[types.ValidatorIndex]bool)
			votesResult := make([]types.Judgement, 0)
			for j := 0; j < len(v.Verdict.Votes); j++ {
				if !uniqueJudgementMap[v.Verdict.Votes[j].Index] {
					uniqueJudgementMap[v.Verdict.Votes[j].Index] = true
					votesResult = append(votesResult, v.Verdict.Votes[j])
				}
				v.Verdict.Votes = votesResult
			}
		}
	}
}

// Sort sorts the verdicts
func (v *VerdictController) Sort() {
	sort.Sort(v)
	for _, v := range v.Verdicts {
		votes := VoteWrapper(v.Verdict.Votes)
		sort.Sort(&votes)
	}
}

func (v *VerdictController) Less(i, j int) bool {
	return bytes.Compare(v.Verdicts[i].Verdict.Target[:], v.Verdicts[j].Verdict.Target[:]) < 0
}

func (v *VerdictController) Len() int {
	return len(v.Verdicts)
}

func (v *VerdictController) Swap(i, j int) {
	v.Verdicts[i], v.Verdicts[j] = v.Verdicts[j], v.Verdicts[i]
}

// Sort sorts the judgements
type VoteWrapper []types.Judgement

func (v *VoteWrapper) Less(i, j int) bool {
	return (*v)[i].Index < (*v)[j].Index
}

func (v *VoteWrapper) Len() int {
	return len(*v)
}

func (v *VoteWrapper) Swap(i, j int) {
	(*v)[i], (*v)[j] = (*v)[j], (*v)[i]
}

// SetDisjoint is disjoint with psi_g, psi_b, psi_w | Eq. 10.7
func (v *VerdictController) SetDisjoint() {
	// not in psi_g, psi_b, psi_w
	// if in psi_g, psi_b, psi_w, remove it (probably duplicate submit verdict)
	psi := store.GetInstance().GetPriorState().Psi
	psiGood := psi.Good
	psiBad := psi.Bad
	psiWonky := psi.Wonky

	uniqueMap := make(map[types.OpaqueHash]bool)

	for _, v := range psiGood {
		uniqueMap[types.OpaqueHash(v)] = true
	}
	for _, v := range psiBad {
		uniqueMap[types.OpaqueHash(v)] = true
	}
	for _, v := range psiWonky {
		uniqueMap[types.OpaqueHash(v)] = true
	}

	result := make([]VerdictWrapper, 0)

	for _, v := range v.Verdicts {
		if !uniqueMap[v.Verdict.Target] {
			result = append(result, v)
		}
	}

	(*v).Verdicts = result
}

// JudgeVerdictStates judges the states of the verdicts | Eq. 10.11
func (v *VerdictController) JudgeVerdictStates() ([]types.WorkReportHash, []types.WorkReportHash, []types.WorkReportHash) {
	goodVotesNum := types.ValidatorsCount*2/3 + 1
	wonkyVotesNum := types.ValidatorsCount / 3

	goodReports := make([]types.WorkReportHash, 0)
	badReports := make([]types.WorkReportHash, 0)
	wonkyReports := make([]types.WorkReportHash, 0)

	for _, verdict := range v.Verdicts {
		positiveVotes := 0

		for _, votes := range verdict.Verdict.Votes {
			if votes.Vote {
				positiveVotes++
			}
		}

		if positiveVotes == goodVotesNum {
			goodReports = append(goodReports, types.WorkReportHash(verdict.Verdict.Target))
		} else if positiveVotes == wonkyVotesNum {
			wonkyReports = append(wonkyReports, types.WorkReportHash(verdict.Verdict.Target))
		} else if positiveVotes == 0 {
			badReports = append(badReports, types.WorkReportHash(verdict.Verdict.Target))
		}
	}
	return goodReports, badReports, wonkyReports
}

package extrinsic

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// VerdictWrapper is a struct that contains a Verdict
type VerdictWrapper struct {
	Verdict types.Verdict
}

// VerdictSummary is a struct that contains a Verdict
type VerdictSummary struct {
	ReportHash           types.OpaqueHash `json:"target,omitempty"`
	PositiveJudgmentsSum int
}

// VerdictController is a struct that contains a slice of Verdict
type VerdictController struct {
	Verdicts           []VerdictWrapper
	VerdictSumSequence []VerdictSummary

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
		Verdicts:           make([]VerdictWrapper, 0),
		VerdictSumSequence: make([]VerdictSummary, 0),
	}
}

// VerifySignature verifies the signatures of the judgement in the verdict   , Eq. 10.3
// currently return []int to check the test, it might change after connect other components in Ch.10
func (v *VerdictWrapper) VerifySignature() error {

	state := store.GetInstance().GetPriorStates()

	a := types.U32(state.GetTau()) / types.U32(types.EpochLength)
	if v.Verdict.Age != a && v.Verdict.Age != a-1 {
		return fmt.Errorf("bad_judgement_age")
	}

	var k types.ValidatorsData
	if v.Verdict.Age == a {
		k = state.GetKappa()
	} else {
		k = state.GetLambda()
	}

	// check if the judgement is valid
	VoteNum := len(v.Verdict.Votes)
	target := v.Verdict.Target[:]

	// store the index of votes with invalid signature
	invalidVotes := make([]int, 0)

	for i := 0; i < VoteNum; i++ {
		if int(v.Verdict.Votes[i].Index) >= len(k) {
			return fmt.Errorf("bad_guarantor_key")
		}
		publicKey := k[v.Verdict.Votes[i].Index].Ed25519[:]
		var message []byte
		if v.Verdict.Votes[i].Vote {
			message = []byte(types.JamValid)
		} else {
			message = []byte(types.JamInvalid)
		}

		message = append(message, target...)

		if !ed25519.Verify(publicKey, message, v.Verdict.Votes[i].Signature[:]) {
			invalidVotes = append(invalidVotes, i)
		}
	}

	if len(invalidVotes) > 0 {
		return fmt.Errorf("bad_signature")
	}
	return nil
}

// SortUnique sorts the verdicts and removes duplicates | Eq. 10.7, Eq. 10.10
func (v *VerdictController) CheckSortUnique() error {
	if err := v.CheckUnique(); err != nil {
		return err
	}
	if err := v.CheckSorted(); err != nil {
		return err
	}
	return nil
}

// Unique removes duplicates
func (v *VerdictController) CheckUnique() error {
	if len(v.Verdicts) == 0 {
		return nil
	}
	// Eq. 10.7 unique
	uniqueMap := make(map[types.OpaqueHash]bool)
	result := make([]VerdictWrapper, 0)

	for _, v := range v.Verdicts {
		if uniqueMap[v.Verdict.Target] {
			return fmt.Errorf("verdicts_not_sorted_unique")
		}
		uniqueMap[v.Verdict.Target] = true
		result = append(result, v)
	}
	(*v).Verdicts = result
	// Eq. 10.10 unique
	for _, v := range v.Verdicts {
		uniqueJudgementMap := make(map[types.ValidatorIndex]bool)
		votesResult := make([]types.Judgement, 0)
		for _, vote := range v.Verdict.Votes {
			if uniqueJudgementMap[vote.Index] {
				return fmt.Errorf("judgements_not_sorted_unique")
			}
			uniqueJudgementMap[vote.Index] = true
			votesResult = append(votesResult, vote)
		}
		v.Verdict.Votes = votesResult
	}
	return nil
}

func (v *VerdictController) CheckSorted() error {
	// Check if Verdicts are sorted by target
	for i := 1; i < len(v.Verdicts); i++ {
		if CompareOpaqueHash(v.Verdicts[i-1].Verdict.Target, v.Verdicts[i].Verdict.Target) > 0 {
			return fmt.Errorf("verdicts_not_sorted_unique")
		}
	}

	// Check if Votes are sorted by index
	for _, verdict := range v.Verdicts {
		votes := VoteWrapper(verdict.Verdict.Votes)
		if !sort.IsSorted(&votes) {
			return fmt.Errorf("judgements_not_sorted_unique")
		}
	}

	return nil
}

func CompareOpaqueHash(a, b types.OpaqueHash) int {
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
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

// SetDisjoint is disjoint with psi_g, psi_b, psi_w | Eq. 10.9
func (v *VerdictController) SetDisjoint() error {
	// not in psi_g, psi_b, psi_w
	// if in psi_g, psi_b, psi_w, remove it (probably duplicate submit verdict)
	states := store.GetInstance().GetPriorStates()
	psiGood := states.GetPsiG()
	psiBad := states.GetPsiB()
	psiWonky := states.GetPsiW()

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

	for _, v := range v.Verdicts {
		if uniqueMap[v.Verdict.Target] {
			return fmt.Errorf("already_judged")
		}
	}
	return nil
}

// GenerateVerdictSumSequence generates verdict only with report hash and votes | Eq. 10.11
func (v *VerdictController) GenerateVerdictSumSequence() {

	for _, verdict := range v.Verdicts {
		verdictSummary := VerdictSummary{}

		positiveVotes := 0
		verdictSummary.ReportHash = verdict.Verdict.Target
		for _, votes := range verdict.Verdict.Votes {
			if votes.Vote {
				positiveVotes++
			}
		}
		verdictSummary.PositiveJudgmentsSum = positiveVotes
		v.VerdictSumSequence = append(v.VerdictSumSequence, verdictSummary)
	}
}

// ClearWorkReports clear uncertain or invalid work reports from core | Eq. 10.15
func (v *VerdictController) ClearWorkReports(verdictSumSequence []VerdictSummary) {
	priorStatesRho := store.GetInstance().GetPriorStates().GetRho()
	clearReports := make(map[types.OpaqueHash]bool)
	for _, verdict := range verdictSumSequence {
		if verdict.PositiveJudgmentsSum < types.ValidatorsCount*2/3 {
			clearReports[types.OpaqueHash(verdict.ReportHash)] = true
		}
	}
	for i := range priorStatesRho {
		if priorStatesRho[i] == nil {
			continue
		}
		hashReport := hash.Blake2bHash(utilities.WorkReportSerialization(priorStatesRho[i].Report))
		if clearReports[hashReport] {
			priorStatesRho[i] = nil
		}
	}
	store.GetInstance().GetIntermediateStates().SetRhoDagger(priorStatesRho)
}

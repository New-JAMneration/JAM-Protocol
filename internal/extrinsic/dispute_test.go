package extrinsic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
)

var disputeErrorReverseMap = map[jamtests.DisputeErrorCode]string{
	jamtests.AlreadyJudged:             "already_judged",
	jamtests.BadVoteSplit:              "bad_vote_split",
	jamtests.VerdictsNotSortedUnique:   "verdicts_not_sorted_unique",
	jamtests.JudgementsNotSortedUnique: "judgements_not_sorted_unique",
	jamtests.CulpritsNotSortedUnique:   "culprits_not_sorted_unique",
	jamtests.FaultsNotSortedUnique:     "faults_not_sorted_unique",
	jamtests.NotEnoughCulprits:         "not_enough_culprits",
	jamtests.NotEnoughFaults:           "not_enough_faults",
	jamtests.CulpritsVerdictNotBad:     "culprits_verdict_not_bad",
	jamtests.FaultVerdictWrong:         "fault_verdict_wrong",
	jamtests.OffenderAlreadyReported:   "offender_already_reported",
	jamtests.BadJudgementAge:           "bad_judgement_age",
	jamtests.BadValidatorIndex:         "bad_validator_index",
	jamtests.BadSignature:              "bad_signature",
}

var JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"

func LoadDisputesTestCase(filename string) (jamtests.DisputeTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests.DisputeTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests.DisputeTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests.DisputeTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests.DisputeTestCase{}, err
	}

	return testCases, nil
}

func GetTestJsonFiles(dir string) []string {
	jsonFiles := []string{}

	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		return nil
	}

	extension := ".json"
	for _, file := range files {
		if filepath.Ext(file.Name()) == extension {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	return jsonFiles
}

func TestDisputes(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "disputes", types.TEST_MODE)
	jsonFiles := GetTestJsonFiles(dir)
	for _, file := range jsonFiles {
		filename := filepath.Join(dir, file)

		disputesTestCase, err := LoadDisputesTestCase(filename)
		if err != nil {
			t.Errorf("Error loading disputes test case: %v", err)
			return
		}
		s := store.GetInstance()
		s.GetPriorStates().SetKappa(disputesTestCase.PreState.Kappa)
		s.GetPriorStates().SetLambda(disputesTestCase.PreState.Lambda)
		s.GetPriorStates().SetRho(disputesTestCase.PreState.Rho)
		s.GetPriorStates().SetPsi(disputesTestCase.PreState.Psi)
		s.GetPriorStates().SetTau(disputesTestCase.PreState.Tau)
		s.GetPosteriorStates().SetPsi(types.DisputesRecords{})
		// disputeExtrinsic := disputesTestCase.Input.Disputes
		// output, disputeErr := Disputes(disputeExtrinsic)
		output, disputeErr := Disputes()
		if disputeErr != nil {
			copyPriorToPosterior()
		}
		psi := s.GetPosteriorStates().GetPsi()
		expectedPsi := disputesTestCase.PostState.Psi
		if err := compareDisputesRecords(psi, expectedPsi); err != nil {
			t.Errorf("Test case %v failed: expected %v, got %v", file, expectedPsi, psi)
		}
		expectedOutput := disputesTestCase.Output

		if expectedOutput.IsError() {
			expectedErrStr := disputeErrorReverseMap[jamtests.DisputeErrorCode(*expectedOutput.Err)]
			if disputeErr != nil {
				outputErrStr := disputeErr.Error()
				if outputErrStr != expectedErrStr {
					fmt.Println(filename)
					fmt.Println("Output error", outputErrStr)
					fmt.Println("Expected error", expectedErrStr)
				}
			} else {
				t.Errorf("Expected error %v, but got no error", expectedOutput.Err)
			}
		} else {
			if disputeErr != nil {
				t.Errorf("Expected no error, but got %v", disputeErr)
			} else {
				if expectedOutput.Ok.OffendersMark == nil {
					expectedOutput.Ok.OffendersMark = types.OffendersMark{}
				}

				if !reflect.DeepEqual(output, expectedOutput.Ok.OffendersMark) {
					t.Errorf("Expected ok %v, got %v", expectedOutput.Ok.OffendersMark, output)
				}
			}
		}
	}
}

func copyPriorToPosterior() {
	priorState := store.GetInstance().GetPriorStates()
	posteriorState := store.GetInstance().GetPosteriorStates()
	posteriorState.SetPsiG(priorState.GetPsiG())
	posteriorState.SetPsiB(priorState.GetPsiB())
	posteriorState.SetPsiW(priorState.GetPsiW())
	posteriorState.SetPsiO(priorState.GetPsiO())
	posteriorState.SetKappa(priorState.GetKappa())
	posteriorState.SetLambda(priorState.GetLambda())
}

func compareDisputesRecords(a, b types.DisputesRecords) error {
	if len(a.Good) != len(b.Good) {
		return fmt.Errorf("length mismatch in Good: %d != %d", len(a.Good), len(b.Good))
	}
	if len(a.Bad) != len(b.Bad) {
		return fmt.Errorf("length mismatch in Bad: %d != %d", len(a.Bad), len(b.Bad))
	}
	if len(a.Wonky) != len(b.Wonky) {
		return fmt.Errorf("length mismatch in Wonky: %d != %d", len(a.Wonky), len(b.Wonky))
	}

	for i := range a.Good {
		if a.Good[i] != b.Good[i] {
			return fmt.Errorf("mismatch in Good at index %d: %x != %x", i, a.Good[i], b.Good[i])
		}
	}

	for i := range a.Bad {
		if a.Bad[i] != b.Bad[i] {
			return fmt.Errorf("mismatch in Bad at index %d: %x != %x", i, a.Bad[i], b.Bad[i])
		}
	}

	for i := range a.Wonky {
		if a.Wonky[i] != b.Wonky[i] {
			return fmt.Errorf("mismatch in Wonky at index %d: %x != %x", i, a.Wonky[i], b.Wonky[i])
		}
	}

	if !arraysContainSameElements(a.Offenders, b.Offenders) {
		return fmt.Errorf("arrays do not contain the same elements")
	}
	return nil
}

func arraysContainSameElements(a, b []types.Ed25519Public) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[types.Ed25519Public]int)
	for _, item := range a {
		counts[item]++
	}
	for _, item := range b {
		if counts[item] == 0 {
			return false
		}
		counts[item]--
	}
	return true
}

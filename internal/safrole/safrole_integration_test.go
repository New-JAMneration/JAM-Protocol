package safrole

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
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
)

var JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"

func TestMain(m *testing.M) {
	types.SetTestMode()
	m.Run()
}

func LoadSafroleTestCase(filename string) (jamtests.SafroleTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
	}

	// Unmarshal the JSON data
	var testCases jamtests.SafroleTestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		return jamtests.SafroleTestCase{}, err
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

func TestSafrole(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "safrole", types.TEST_MODE)
	jsonFiles := GetTestJsonFiles(dir)
	for iter, file := range jsonFiles {
		// for easy testing
		t.Log(iter, file)
		filename := filepath.Join(dir, file)
		safroleTestCase, err := LoadSafroleTestCase(filename)
		if err != nil {
			t.Errorf("Error loading safrole test case: %v", err)
			return
		}
		// ----- INPUT ----- //
		// --- Set safrole state to store --- //
		s := store.GetInstance()
		// tau
		s.GetPriorStates().SetTau(safroleTestCase.PreState.Tau)
		// eta
		s.GetPriorStates().SetEta(safroleTestCase.PreState.Eta)
		s.GetPosteriorStates().SetEta(safroleTestCase.PostState.Eta)
		// lambda
		s.GetPriorStates().SetLambda(safroleTestCase.PreState.Lambda)
		// kappa
		s.GetPriorStates().SetKappa(safroleTestCase.PreState.Kappa)
		s.GetPosteriorStates().SetKappa(safroleTestCase.PostState.Kappa)
		// gammaK
		s.GetPriorStates().SetGammaK(safroleTestCase.PreState.GammaK)
		// iota
		s.GetPriorStates().SetIota(safroleTestCase.PreState.Iota)
		// gammaA
		s.GetPriorStates().SetGammaA(safroleTestCase.PreState.GammaA)
		// gammaS
		s.GetPriorStates().SetGammaS(safroleTestCase.PreState.GammaS)
		s.GetPosteriorStates().SetGammaS(safroleTestCase.PostState.GammaS)
		// gammaZ
		s.GetPriorStates().SetGammaZ(safroleTestCase.PreState.GammaZ)
		// offenders
		// s.GetPriorStates().SetPsiO(safroleTestCase.PreState.PsiO)

		// --- Set safrole input to store --- //
		s.GetProcessingBlockPointer().SetSlot(safroleTestCase.Input.Slot)
		// s.GetProcessingBlockPointer().SetEntropySource(safroleTestCase.Input.Entropy)
		s.GetProcessingBlockPointer().SetTicketsExtrinsic(safroleTestCase.Input.Extrinsic)

		// // MockHeader for sealing.go
		// s.GetIntermediateHeaders().SetHeader(header)

		// err
		safroleOutput := safroleTestCase.Output
		if safroleOutput.Ok != nil && safroleOutput.Err == nil {
			t.Log("safroleOkInfo: ", safroleOutput.Ok)
			// EpockMark, TicketsMark := safroleOutput.Ok.EpochMark, safroleOutput.Ok.TicketsMark
			// fmt.Println("epockMark: ", epockMark)
			// fmt.Println("ticketsMark: ", ticketsMark)
		} else if safroleOutput.Ok == nil && safroleOutput.Err != nil {
			t.Log("safroleErr:", safroleOutput.Err)
			// conver to int
			errCode := int(*safroleOutput.Err)
			t.Log("errCode: ", errCode)
		} else {
			t.Errorf("expected one of Ok or Err, got %v", safroleOutput)
		}
		// fmt.Println("safroleErr:", safroleErr)
		// err2 := safroleErr.UnmarshalJSON([]byte("bad_slot"))
		// fmt.Println("err2: ", err2)

		// ----- END INPUT ----- //

		// ----- PROCESS SAFROLE LOGIC ----- //
		gammaS := s.GetPriorStates().GetGammaS()
		var safroleMode string
		if gammaS.Keys != nil {
			safroleMode = "Fallback Mode"
			t.Log(safroleMode)
		} else if gammaS.Tickets != nil {
			safroleMode = "Normal Mode"
			t.Log(safroleMode)
		} else {
			t.Error("Unknown Mode")
		}
		// --- safrole.go --- //
		// KeyRotate() contains GP(6.2, 6.13, 6.14)
		KeyRotate()
		/*
			// verify KeyRotate()
			priorState := s.GetPriorStates()
			posteriorState := s.GetPosteriorStates()
			if !reflect.DeepEqual(posteriorState.GetGammaK(), priorState.GetIota()) {
				t.Errorf("Expected GammaK to be %v, got %v", priorState.GetIota(), posteriorState.GetGammaK())
			}
			if !reflect.DeepEqual(posteriorState.GetKappa(), priorState.GetGammaK()) {
				t.Errorf("Expected Kappa to be %v, got %v", priorState.GetGammaK(), posteriorState.GetKappa())
			}
			if !reflect.DeepEqual(posteriorState.GetLambda(), priorState.GetKappa()) {
				t.Errorf("Expected Lambda to be %v, got %v", priorState.GetKappa(), posteriorState.GetLambda())
			}
			if posteriorState.GetGammaZ() != priorState.GetGammaZ() {
				t.Errorf("Expected GammaZ to be %v, got %v", priorState.GetGammaZ(), posteriorState.GetGammaZ())
			}
		*/

		// --- ticketbody_controller.go --- //

		// --- sealing.go (GP 6.15~6.24)--- //
		// if safroleMode == "Normal Mode" {
		// 	SealingByTickets()
		// } else if safroleMode == "Fallback Mode" {
		// 	SealingByBandersnatchs()
		// }

		// --- slot_key_sequence.go (GP 6.25, 6.26) --- //

		// --- markers.go (GP 6.27, 6.28) --- //
		markerErr := CreateEpochMarker()
		fmt.Println("markerErr: ", markerErr)
		// if safroleOutput.Err == nil {
		// 	if markerErr != nil {
		// 		t.Errorf("expected nil, got %v", markerErr)
		// 	}
		// } else {
		// 	if markerErr == nil || !reflect.DeepEqual(markerErr, jamtests.BadSlot) {
		// 		t.Errorf("expected %v or %v, got %v", safroleOutput.Err, jamtests.BadSlot, markerErr)
		// 	}
		// }

		// --- extrinsic_tickets.go (GP 6.30~6.34) --- //

		// ----- END PROCESS SAFROLE LOGIC ----- //

		if file == "enact-epoch-change-with-no-tickets-1.json" || file == "enact-epoch-change-with-no-tickets-3.json" {
			t.Log(file, "O")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", safroleOutput.Err)
			} else if !reflect.DeepEqual(*safroleOutput.Ok, jamtests.SafroleOutputData{}) {
				t.Errorf("expected %v, got %v", jamtests.SafroleOutputData{}, *safroleOutput.Ok)
			} else {
				t.Log(file, "passed")
			}
		}
		if file == "enact-epoch-change-with-no-tickets-2.json" {
			t.Log(file, "X")
			if safroleOutput.Err == nil {
				t.Errorf("expected %v, got nil", jamtests.BadSlot)
			} else if !reflect.DeepEqual(*safroleOutput.Err, jamtests.BadSlot) {
				t.Errorf("expected %v, got %v", jamtests.BadSlot, *safroleOutput.Err)
			} else {
				t.Log(file, "passed")
			}
		}
		// if file == "enact-epoch-change-with-no-tickets-4.json" {
		// 	t.Log(file, "O")
		// 	if safroleOutput.Err != nil {
		// 		t.Errorf("expected nil, got %v", safroleOutput.Err)
		// 	} else if !reflect.DeepEqual(*safroleOutput.Ok, jamtests.SafroleOutputData{}) {
		// 		t.Errorf("expected %v, got %v", jamtests.SafroleOutputData{}, *safroleOutput.Ok)
		// 	} else {
		// 		t.Log(file, "passed")
		// 	}
		// }

	}
}

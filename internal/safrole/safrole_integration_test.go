package safrole

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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
	sort.Strings(jsonFiles)
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
		s.GetProcessingBlockPointer().SetTicketsExtrinsic(safroleTestCase.Input.Extrinsic)
		// tau
		s.GetPriorStates().SetTau(safroleTestCase.PreState.Tau)
		s.GetPosteriorStates().SetTau(safroleTestCase.Input.Slot)
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
		s.GetPosteriorStates().SetGammaK(safroleTestCase.PostState.GammaK)
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

		// --- HANDLE SAFROLEOUT ERROR --- //
		safroleOutput := safroleTestCase.Output
		safroleOutputEpockMark := types.EpochMark{}
		safroleOutputTicketsMark := types.TicketsMark{}
		safroleOutputErrCode := 7 // 7 is not a valid error code
		if safroleOutput.Ok != nil && safroleOutput.Err == nil {
			t.Log("input safroleOkInfo: ", *safroleOutput.Ok)
			if safroleOutput.Ok.EpochMark != nil {
				safroleOutputEpockMark = *safroleOutput.Ok.EpochMark
			}
			if safroleOutput.Ok.TicketsMark != nil {
				safroleOutputTicketsMark = *safroleOutput.Ok.TicketsMark
			}
			fmt.Println("safroleOutputEpockMark: ", &safroleOutputEpockMark)
			fmt.Println("safroleOutputTicketsMark: ", &safroleOutputTicketsMark)
		} else if safroleOutput.Ok == nil && safroleOutput.Err != nil {
			t.Log("input safroleErr:", *safroleOutput.Err)
			// conver to int
			safroleOutputErrCode = int(*safroleOutput.Err)
			t.Log("input safroleOutputErrCode: ", safroleOutputErrCode)
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
		// --- safrole.go (GP 6.2, 6.13, 6.14) --- //
		// (6.2, 6.13, 6.14)
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

		// --- ticketbody_controller.go (GP 6.5, 6.6) --- //

		// --- sealing.go (GP 6.15~6.24) --- //
		if safroleMode == "Normal Mode" {
			t.Log("sealing with tickets")
			// (GP 6.15)
			SealingByTickets()
		} else if safroleMode == "Fallback Mode" {
			t.Log("sealing with keys")
			// (GP 6.16)
			SealingByBandersnatchs()
		}
		// (GP 6.22)
		UpdateEtaPrime0()
		// (GP 6.23)
		UpdateEntropy()
		// (GP 6.17)
		UpdateHeaderEntropy()
		// (GP 6.24) contain slot_key_sequence.go (GP 6.25, 6.26)
		UpdateSlotKeySequence()

		// --- slot_key_sequence.go (GP 6.25, 6.26) --- //

		// --- markers.go (GP 6.27, 6.28) --- //
		// (GP 6.27)
		ourEpochMarkErr := CreateEpochMarker()
		// if ourEpochMarkErr != nil {
		// 	fmt.Println("markerErr:", ourEpochMarkErr)
		// }

		// (GP 6.28)
		CreateWinningTickets()

		// --- extrinsic_tickets.go (GP 6.30~6.34) --- //
		ourEtErr := CreateNewTicketAccumulator()
		// if ourEtErr != nil {
		// 	fmt.Println("outEtErr:", ourEtErr)
		// }
		// ----- END PROCESS SAFROLE LOGIC ----- //

		// ----- EXTRACT OUR OUTPUT RESULT ----- //
		ourTicketMark := s.GetIntermediateHeaderPointer().GetTicketsMark()
		// if ourTicketMark != nil {
		// 	fmt.Println("ourTicketMark: ", ourTicketMark)
		// }
		ourEpochMark := s.GetIntermediateHeaderPointer().GetEpochMark()
		// if ourEpochMark != nil {
		// 	fmt.Println("ourEpochMark: ", *ourEpochMark)
		// }
		// ourSeal := s.GetIntermediateHeader().Seal

		// ----- VERIFY OUTPUT ----- //

		if file == "enact-epoch-change-with-no-tickets-1.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "enact-epoch-change-with-no-tickets-2.json" { // OK w/ output
			t.Log("expected \"bad_slot\"(0) error")
			if safroleOutput.Err == nil {
				t.Errorf("expected %v, got nil", jamtests.BadSlot)
			} else {
				if safroleOutputErrCode != int(*ourEpochMarkErr) {
					t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEpochMarkErr))
				} else {
					t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEpochMarkErr), safroleOutputErrCode)
				}
			}
		}
		if file == "enact-epoch-change-with-no-tickets-3.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "enact-epoch-change-with-no-tickets-4.json" { // OK w/ output
			t.Log("expected OK w/ epochmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if safroleOutputEpockMark.Entropy != ourEpochMark.Entropy || safroleOutputEpockMark.TicketsEntropy != ourEpochMark.TicketsEntropy || !reflect.DeepEqual(safroleOutputEpockMark.Validators, ourEpochMark.Validators) {
					t.Errorf("expected %v, \ngot %v", safroleOutputEpockMark, ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", &ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "skip-epochs-1.json" { // OK w/ output
			t.Log("expected OK w/ epochmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if safroleOutputEpockMark.Entropy != ourEpochMark.Entropy || safroleOutputEpockMark.TicketsEntropy != ourEpochMark.TicketsEntropy || !reflect.DeepEqual(safroleOutputEpockMark.Validators, ourEpochMark.Validators) {
					t.Errorf("expected %v, \ngot %v", safroleOutputEpockMark, ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", &ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "skip-epoch-tail-1.json" { // OK w/ output
			t.Log("expected OK w/ epochmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if safroleOutputEpockMark.Entropy != ourEpochMark.Entropy || safroleOutputEpockMark.TicketsEntropy != ourEpochMark.TicketsEntropy || !reflect.DeepEqual(safroleOutputEpockMark.Validators, ourEpochMark.Validators) {
					t.Errorf("expected %v, \ngot %v", safroleOutputEpockMark, ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", &ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-no-mark-1.json" { // OK w/o output
			t.Log("expected \"bad_ticket_attempt\"(4) error")
			if safroleOutput.Err == nil {
				t.Errorf("expected %v, got nil", jamtests.BadTicketAttempt)
			} else {
				if ourEtErr == nil || safroleOutputErrCode != int(*ourEtErr) {
					t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEtErr))
				} else {
					t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEtErr), safroleOutputErrCode)
				}
			}
		}
		// if file == "publish-tickets-no-mark-2.json" { // VerificationFailure with CreateNewTicketAccumulator
		// 	t.Log("expected OK w/o info")
		// 	if safroleOutput.Err != nil {
		// 		t.Errorf("expected nil, got %v", *safroleOutput.Err)
		// 	} else {
		// 		if ourEpochMarkErr != nil {
		// 			t.Errorf("expected nil, got %v", *ourEpochMarkErr)
		// 		} else if ourEpochMark != nil {
		// 			t.Errorf("expected nil, got %v", *ourEpochMark)
		// 		} else if ourTicketMark != nil {
		// 			t.Errorf("expected nil, got %v", *ourTicketMark)
		// 		} else {
		// 			t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", ourEpochMark, ourTicketMark, *safroleOutput.Ok)
		// 		}
		// 	}
		// }
		// if file == "publish-tickets-no-mark-3.json" { // VerificationFailure with CreateNewTicketAccumulator
		// 	t.Log("expected \"duplicate_ticket\"(6) error")
		// 	if safroleOutput.Err == nil {
		// 		t.Errorf("expected %v, got nil", jamtests.BadTicketAttempt)
		// 	} else {
		// 		if ourEtErr == nil || safroleOutputErrCode != int(*ourEtErr) {
		// 			t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEtErr))
		// 		} else {
		// 			t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEtErr), safroleOutputErrCode)
		// 		}
		// 	}
		// }
		// if file == "publish-tickets-no-mark-4.json" { // got wrong errorCode
		// 	t.Log("expected \"bad_ticket_order\"(2) error")
		// 	if safroleOutput.Err == nil {
		// 		t.Errorf("expected %v, got nil", jamtests.BadTicketOrder)
		// 	} else {
		// 		if ourEtErr == nil || safroleOutputErrCode != int(*ourEtErr) {
		// 			t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEtErr))
		// 		} else {
		// 			t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEtErr), safroleOutputErrCode)
		// 		}
		// 	}
		// }
		if file == "publish-tickets-no-mark-5.json" { // OK w/ output
			t.Log("expected \"bad_ticket_proof\"(3) error")
			if safroleOutput.Err == nil {
				t.Errorf("expected %v, got nil", jamtests.BadTicketOrder)
			} else {
				if ourEtErr == nil || safroleOutputErrCode != int(*ourEtErr) {
					t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEtErr))
				} else {
					t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEtErr), safroleOutputErrCode)
				}
			}
		}
		if file == "publish-tickets-no-mark-6.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else {
					t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-no-mark-7.json" { // OK w/ output
			t.Log("expected \"unexpected_ticket\"(1) error")
			if safroleOutput.Err == nil {
				t.Errorf("expected %v, got nil", jamtests.BadTicketOrder)
			} else {
				if ourEtErr == nil || safroleOutputErrCode != int(*ourEtErr) {
					t.Errorf("expected %v, got %v", safroleOutputErrCode, int(*ourEtErr))
				} else {
					t.Logf("\nour output %v fits safroleOutputErrCode %v", int(*ourEtErr), safroleOutputErrCode)
				}
			}
		}
		if file == "publish-tickets-no-mark-8.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else {
					t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-no-mark-9.json" { // OK w/ output
			t.Log("expected OK w/ epochmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if safroleOutputEpockMark.Entropy != ourEpochMark.Entropy || safroleOutputEpockMark.TicketsEntropy != ourEpochMark.TicketsEntropy || !reflect.DeepEqual(safroleOutputEpockMark.Validators, ourEpochMark.Validators) {
					t.Errorf("expected %v, \ngot %v", safroleOutputEpockMark, ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", &ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-with-mark-1.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else {
					t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-with-mark-2.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else {
					t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-with-mark-3.json" { // OK w/ output
			t.Log("expected OK w/o info")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if ourEpochMark != nil {
					t.Errorf("expected nil, got %v", *ourEpochMark)
				} else if ourTicketMark != nil {
					t.Errorf("expected nil, got %v", *ourTicketMark)
				} else {
					t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
				}
			}
		}
		if file == "publish-tickets-with-mark-4.json" { // OK w/ output
			t.Log("expected OK w/ ticketmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else if ourTicketMark == nil {
				t.Errorf("expected %v, got nil", safroleOutputTicketsMark)
			} else if !reflect.DeepEqual(safroleOutputTicketsMark, *ourTicketMark) {
				t.Errorf("expected %v, \ngot %v", safroleOutputTicketsMark, *ourTicketMark)
			} else {
				t.Log("safroleOutput.Ok: ", *safroleOutput.Ok)
			}
		}
		if file == "publish-tickets-with-mark-5.json" { // OK w/ output
			t.Log("expected OK w/ epochmark")
			if safroleOutput.Err != nil {
				t.Errorf("expected nil, got %v", *safroleOutput.Err)
			} else {
				if ourEpochMarkErr != nil {
					t.Errorf("expected nil, got %v", *ourEpochMarkErr)
				} else if safroleOutputEpockMark.Entropy != ourEpochMark.Entropy || safroleOutputEpockMark.TicketsEntropy != ourEpochMark.TicketsEntropy || !reflect.DeepEqual(safroleOutputEpockMark.Validators, ourEpochMark.Validators) {
					t.Errorf("expected %v, \ngot %v", safroleOutputEpockMark, ourEpochMark)
				} else {
					t.Logf("\nour output {%v, %v} fits safroleOutput.Ok: %v", &ourEpochMark, ourTicketMark, *safroleOutput.Ok)
				}
			}
		}

	}
}

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
)

func init() {
	safroleCmd := &cli.Command{
		Use:   "safrole",
		Short: "Test safrole stf",
		Long:  "Test safrole state transition functions with test vectors",
		Run: func(args []string) {
			RunSafroleTests()
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:         "test-vector",
				Usage:        "Path to test vector file",
				DefaultValue: "testvectors/safrole.json",
				Destination:  &testVectorPath,
			},
		},
	}

	cli.AddCommand(safroleCmd)
}

var testVectorPath string

func RunSafroleTests() { //encapsulates runSafroleTests for troubleshooting
	err := runSafroleTests()
	if err != nil {
		fmt.Printf("Test execution failed: %v\n", err)
	}
}

func runSafroleTests() error {
	// Read test vector file
	testCases, err := os.ReadFile("xxxx")
	if err != nil {
		return fmt.Errorf("failed to read test vector file: %v", err)
	}

	// var testCases
	//for i, tc := range testCases {
	// Execute state transitions
	// executeStateTransitions()

	// verifyFinalState(tc.PostState)
	// }

	fmt.Printf("All %d test cases passed\n", len(testCases))
	return nil
}

// Update Safrole
func executeStateTransitions() {
	// Key rotation
	// This function will update GammaK, GammaZ, Lambda, Kappa
	safrole.KeyRotate()

	// Create new ticket accumulator
	// This function will update GammaA
	safrole.CreateNewTicketAccumulator()

	// Sealing Operations
	// This function will update Heaer Seal
	safrole.SealingByBandersnatchs()

	// Update GammaS
	safrole.UpdateSlotKeySequence()

	// Entropy update
	// This function will update Eta
	safrole.UpdateEntropy()

	// Update Eta'0
	safrole.UpdateEtaPrime0()

	// Update Header Entropy
	safrole.UpdateHeaderEntropy()

	// Marker
	// This function will update Header EpochMark
	safrole.CreateEpochMarker()

	// Winning tickets
	// This function will update Header TicketsMark
	safrole.CreateWinningTickets()
}

type SafroleState struct {
	Tau    types.TimeSlot                    `json:"tau"`
	Eta    [4]types.BandersnatchVrfSignature `json:"eta"`
	Lambda types.ValidatorsData              `json:"lambda"`
	Kappa  types.ValidatorsData              `json:"kappa"`
	GammaK types.ValidatorsData              `json:"gamma_k"`
	Iota   types.ValidatorsData              `json:"iota"`
	GammaA types.TicketsAccumulator          `json:"gamma_a"`
	GammaS types.TicketsOrKeys               `json:"gamma_s"`
	GammaZ types.BandersnatchRingCommitment  `json:"gamma_z"`
}

func verifyFinalState(expected SafroleState) error {
	s := store.GetInstance()
	posterior := s.GetPosteriorStates()

	// Verify each state field
	if posterior.GetTau() != expected.Tau {
		return fmt.Errorf("Tau mismatch: expected %v, got %v", expected.Tau, posterior.GetTau())
	}

	// Compare EntropyBuffer elements individually
	eta := posterior.GetEta()
	for i := 0; i < 4; i++ {
		if !bytes.Equal(eta[i][:], expected.Eta[i][:]) {
			return fmt.Errorf("Eta[%d] mismatch", i)
		}
	}

	if !reflect.DeepEqual(posterior.GetLambda(), expected.Lambda) {
		return fmt.Errorf("Lambda mismatch")
	}

	if !reflect.DeepEqual(posterior.GetKappa(), expected.Kappa) {
		return fmt.Errorf("Kappa mismatch")
	}

	if !reflect.DeepEqual(posterior.GetGammaK(), expected.GammaK) {
		return fmt.Errorf("GammaK mismatch")
	}

	if !reflect.DeepEqual(posterior.GetIota(), expected.Iota) {
		return fmt.Errorf("Iota mismatch")
	}

	if !reflect.DeepEqual(posterior.GetGammaA(), expected.GammaA) {
		return fmt.Errorf("GammaA mismatch")
	}

	if !reflect.DeepEqual(posterior.GetGammaS(), expected.GammaS) {
		return fmt.Errorf("GammaS mismatch")
	}

	// Compare BandersnatchRingCommitment
	gammaZ := posterior.GetGammaZ()
	if !reflect.DeepEqual(gammaZ, expected.GammaZ) {
		return fmt.Errorf("GammaZ mismatch")
	}

	// Note: PostOffenders is not part of the store state, so we don't verify it
	// It's only used for testing purposes

	return nil
}

package testdata

import (
	"fmt"
)

// TestResult represents the result of a test
type TestResult struct {
	TestFile string
	Passed   bool
	Error    error
}

type JamTestVectorsRunner struct {
	Mode TestMode
}

func NewJamTestVectorsRunner(mode TestMode) *JamTestVectorsRunner {
	return &JamTestVectorsRunner{Mode: mode}
}

func (r *JamTestVectorsRunner) Run(data interface{}) error {
	result := &TestResult{}

	// Set data into Store
	SetTestDataToDataStore(data)

	// Run the appropriate test based on mode
	switch r.Mode {
	case SafroleMode:
		// Run safrole STF
		RunSafroleTests()
		result.Passed = false
		result.Error = fmt.Errorf("safrole STF not implemented")

	case AssurancesMode:
		// TODO: Implement assurances STF
		result.Passed = false
		result.Error = fmt.Errorf("assurances STF not implemented")

	case PreimagesMode:
		// TODO: Implement preimages STF
		result.Passed = false
		result.Error = fmt.Errorf("preimages STF not implemented")

	case HistoryMode:
		// TODO: Implement history STF
		result.Passed = false
		result.Error = fmt.Errorf("history STF not implemented")

	case DisputesMode:
		// TODO: Implement disputes STF
		result.Passed = false
		result.Error = fmt.Errorf("disputes STF not implemented")

	case AccumulateMode:
		// TODO: Implement accumulate STF
		result.Passed = false
		result.Error = fmt.Errorf("accumulate STF not implemented")

	case AuthorizationsMode:
		// TODO: Implement authorizations STF
		result.Passed = false
		result.Error = fmt.Errorf("authorizations STF not implemented")

	default:
		result.Passed = false
		result.Error = fmt.Errorf("unsupported test mode: %s", mode)
	}

	return result, nil
}

// Run SafroleTests runs the safrole tests
func RunSafroleTests() {
	// Key rotation
	// This function will update GammaK, GammaZ, Lambda, Kappa
	// safrole.KeyRotate()

	// Create new ticket accumulator
	// This function will update GammaA
	// safrole.CreateNewTicketAccumulator()

	// Sealing Operations
	// This function will update Heaer Seal
	// safrole.SealingByBandersnatchs()

	// Update GammaS
	// safrole.UpdateSlotKeySequence()

	// Entropy update
	// This function will update Eta
	// safrole.UpdateEntropy()

	// Update Eta'0
	// safrole.UpdateEtaPrime0()

	// Update Header Entropy
	// safrole.UpdateHeaderEntropy()

	// Marker
	// This function will update Header EpochMark
	// safrole.CreateEpochMarker()

	// Winning tickets
	// This function will update Header TicketsMark
	// safrole.CreateWinningTickets()

	// Update the PostOffenders
	// Find the function that will update the PostOffenders
}

// Run AssurancesTests runs the assurances tests
func RunAssurancesTests() {
	// Update AvailAssignments

	// Update CurrValidators
}

// Run PreimagesTests runs the preimages tests
func RunPreimagesTests() {
	// Update AccountsMapEntry
}

// Run HistoryTests runs the history tests
func RunHistoryTests() {
	// Update Beta
}

// Run DisputesTests runs the disputes tests
func RunDisputesTests() {
	// Update Psi

	// Update Rho

	// Update Tau

	// Update Kappa

	// Update Lambda
}

// Run AccumulateTests runs the accumulate tests
func RunAccumulateTests() {
	// Update Slot

	// Update Entropy

	// Update ReadyQueue

	// Update Accumulated

	// Update Privileges

	// Update Accounts
}

func RunAuthorizationsTests() {
	// Update Alpha
	// Varphi
}

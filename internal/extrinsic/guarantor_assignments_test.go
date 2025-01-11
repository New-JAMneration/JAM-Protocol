package extrinsic

import (
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestRotateCores verifies that the rotateCores function properly rotates
// the slice of cores by a given offset n.
func TestRotateCores(t *testing.T) {
	// arrange
	in := []types.U32{0, 1, 2, 3, 4, 5, 6}
	var n types.U32 = 2

	// act
	out := rotateCores(in, n)

	// assert
	expected := []types.U32{0, 1, 0, 1, 0, 1, 0}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("rotateCores failed.\nExpected: %v\nGot:      %v", expected, out)
	}
}

// TestPermute verifies that permute returns a properly permuted list of cores
// based on some dummy entropy and current slot.
func TestPermute(t *testing.T) {
	// arrange
	// Use dummy data for entropy (32 bytes)
	dummyEntropy := types.Entropy([32]byte{0xAA, 0xBB, 0xCC, 0xDD})
	var dummySlot types.TimeSlot = 100

	// act
	out := permute(dummyEntropy, dummySlot) // same note about unexported

	if len(out) != int(V) {
		t.Errorf("permute output length mismatch.\nExpected: %d\nGot:      %d", V, len(out))
	}

	expected := []types.CoreIndex{0, 0, 1, 0, 1, 1}

	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("permute failed.\nExpected: %v\nGot:      %v", expected, out)
	}
}

// TestNewGuranatorAssignments checks that GuranatorAssignments is created properly.
func TestNewGuranatorAssignments(t *testing.T) {
	// arrange
	dummyEntropy := types.Entropy([32]byte{0x01, 0x02, 0x03})
	var dummySlot types.TimeSlot = 50

	// Construct dummy validators data.
	// Suppose your types.ValidatorsData is something like:
	//   type ValidatorsData []Validator
	//   type Validator struct {
	//       Ed25519  Ed25519Public
	//       SomeFlag bool
	//   }
	// Provide a small sample of them:
	dummyValidators := types.ValidatorsData{
		{Ed25519: [32]byte{0x11, 0x22, 0x33}},
		{Ed25519: [32]byte{0x44, 0x55, 0x66}},
	}

	// act
	gAssignments := NewGuranatorAssignments(dummyEntropy, dummySlot, dummyValidators)

	// assert
	// Check the length of assignments
	if len(gAssignments.CoreAssignments) != int(V) {
		t.Errorf("expected %d core assignments, got %d", V, len(gAssignments.CoreAssignments))
	}

	// Check the length of public keys
	if len(gAssignments.PublicKeys) != len(dummyValidators) {
		t.Errorf("expected %d public keys, got %d", len(dummyValidators), len(gAssignments.PublicKeys))
	}

	// Optionally check if the public keys match the dummyValidators
	for i, pubKey := range gAssignments.PublicKeys {
		if pubKey != dummyValidators[i].Ed25519 {
			t.Errorf("mismatch in public key at index %d.\nExpected: %v\nGot:      %v",
				i, dummyValidators[i].Ed25519, pubKey)
		}
	}
}

// TestGStar checks GStar logic by verifying that the correct epoch entropy and
// validator set (Kappa vs Lambda) are used, as well as the correct offset on Tau.
func TestGStarLambda(t *testing.T) {
	// arrange
	// Suppose state.Tau = 120, E=20, R=10. Then subEpoch boundary checks can be tested.
	// We'll create a dummy State with known eta and kappa/lambda.
	dummyEta := [4]types.Entropy{
		[32]byte{0x01}, // eta_prime[0]
		[32]byte{0x02}, // eta_prime[1]
		[32]byte{0x03}, // eta_prime[2]
		[32]byte{0x04}, // eta_prime[3]
	}
	// Also define dummyKappa and dummyLambda
	dummyKappa := types.ValidatorsData{
		{Ed25519: [32]byte{0xAA}},
		{Ed25519: [32]byte{0xBB}},
	}
	dummyLambda := types.ValidatorsData{
		{Ed25519: [32]byte{0xCC}},
		{Ed25519: [32]byte{0xDD}},
	}

	state := types.State{
		Tau:    120,
		Eta:    dummyEta,
		Kappa:  dummyKappa,
		Lambda: dummyLambda,
	}

	// act
	gStarVal := GStar(state)

	// assert
	// Because Tau=120, E=12, and R=10, we check which epoch segment it falls into:
	//  floor((120-10)/12) = 9  != floor(120/12) = 10
	//  it should go to else branch => (η′3, λ′).
	// So expected entropy is dummyEta[3] = [32]byte{0x04}
	// and expected validators set is dummyLambda.

	if gStarVal.PublicKeys[0] != dummyLambda[0].Ed25519 {
		t.Errorf("expected G* to use lambda's public key[0], got something else")
	}
	if gStarVal.PublicKeys[1] != dummyLambda[1].Ed25519 {
		t.Errorf("expected G* to use lambda's public key[1], got something else")
	}

	expected := []types.CoreIndex{1, 0, 1, 0, 0, 1}

	if !reflect.DeepEqual(gStarVal.CoreAssignments, expected) {
		t.Fatalf("GStar failed.\nExpected: %v\nGot:      %v", expected, gStarVal.CoreAssignments)
	}

}

// TestGStar checks GStar logic by verifying that the correct epoch entropy and
// validator set (Kappa vs Lambda) are used, as well as the correct offset on Tau.
func TestGStarKappa(t *testing.T) {
	dummyEta := [4]types.Entropy{
		[32]byte{0x01}, // eta_prime[0]
		[32]byte{0x02}, // eta_prime[1]
		[32]byte{0x03}, // eta_prime[2]
		[32]byte{0x04}, // eta_prime[3]
	}
	// Also define dummyKappa and dummyLambda
	dummyKappa := types.ValidatorsData{
		{Ed25519: [32]byte{0xAA}},
		{Ed25519: [32]byte{0xBB}},
	}
	dummyLambda := types.ValidatorsData{
		{Ed25519: [32]byte{0xCC}},
		{Ed25519: [32]byte{0xDD}},
	}

	state := types.State{
		Tau:    130,
		Eta:    dummyEta,
		Kappa:  dummyKappa,
		Lambda: dummyLambda,
	}

	// act
	gStarVal := GStar(state)

	// assert
	// Because Tau=130, E=12, and R=10, we check which epoch segment it falls into:
	//  floor((130-10)/12) = 10  == floor(130/12) = 10
	//  it should go to else branch => (η′2, κ′)
	// So expected entropy is dummyEta[2] = [32]byte{0x03}
	// and expected validators set is dummyLambda.

	if gStarVal.PublicKeys[0] != dummyKappa[0].Ed25519 {
		t.Errorf("expected G* to use lambda's public key[0], got something else")
	}
	if gStarVal.PublicKeys[1] != dummyKappa[1].Ed25519 {
		t.Errorf("expected G* to use lambda's public key[1], got something else")
	}

	expected := []types.CoreIndex{0, 0, 0, 1, 1, 1}

	if !reflect.DeepEqual(gStarVal.CoreAssignments, expected) {
		t.Fatalf("GStar failed.\nExpected: %v\nGot:      %v", expected, gStarVal.CoreAssignments)
	}
}

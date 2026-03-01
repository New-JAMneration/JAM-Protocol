package extrinsic

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_assurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func hexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

func TestAvailAssuranceController(t *testing.T) {
	AvailAssurances := NewAvailAssuranceController()

	if len(AvailAssurances.AvailAssurances) != 0 {
		t.Errorf("Expected %d assurances, got %d", 0, len(AvailAssurances.AvailAssurances))
	}
}

type InputWrapper[T any] struct {
	Input T
}

func ParseData[t any](fileName string) (InputWrapper[t], error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fileName, err)
		return InputWrapper[t]{}, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %s: %v\n", fileName, err)
		return InputWrapper[t]{}, err
	}
	var wrapper InputWrapper[t]
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON:: %v\n", err)
	}
	return wrapper, nil
}

func DecodeJSONByte(input []byte) []byte {
	toJSON, _ := json.Marshal(input)
	out := string(toJSON)[1 : len(string(toJSON))-1]
	return hexToBytes(out)
}

// need to package structure in JSON file
type Input struct {
	AssurancesExtrinsic types.AssurancesExtrinsic `json:"assurances,omitempty"`
	Slot                types.TimeSlot            `json:"slot,omitempty"`
	Parent              types.HeaderHash          `json:"parent,omitempty"`
	PreStateTest        PreStateTest              `json:"pre_state,omitempty"`
	PostStateTest       PreStateTest              `json:"post_state,omitempty"`
}

type PreStateTest struct {
	Rho   types.AvailabilityAssignments `json:"avail_assignments"`
	Kappa types.ValidatorsData          `json:"curr_validators"`
}

type PostStateTest struct {
	Rho   types.AvailabilityAssignments `json:"avail_assignments"`
	Kappa types.ValidatorsData          `json:"curr_validators"`
}

// 11.11
func TestValidateAnchor(t *testing.T) {
	wrapper, err := ParseData[Input]("assurance_data/assurance_with_bad_attestation_parent-1.json")
	if err != nil {
		t.Error(err)
		return
	}

	input := wrapper.Input

	assuranceExtrinsic := NewAvailAssuranceController()
	for _, availAssurance := range input.AssurancesExtrinsic {
		assuranceExtrinsic.AvailAssurances = append(assuranceExtrinsic.AvailAssurances,
			types.AvailAssurance{
				ValidatorIndex: availAssurance.ValidatorIndex,
				Anchor:         availAssurance.Anchor,
				Bitfield:       availAssurance.Bitfield,
				Signature:      availAssurance.Signature,
			})
	}

	blockchain.GetInstance().AddBlock(types.Block{
		Header: types.Header{
			Parent: input.Parent,
		},
	})

	err = assuranceExtrinsic.ValidateAnchor()
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}
}

// 11.13
func TestValidateSignature(t *testing.T) {
	wrapper, err := ParseData[Input]("assurance_data/assurances_with_bad_signature-1.json")
	if err != nil {
		t.Error(err)
		return
	}
	input := wrapper.Input
	assuranceExtrinsic := NewAvailAssuranceController()

	for _, availAssurance := range input.AssurancesExtrinsic {
		assuranceExtrinsic.AvailAssurances = append(assuranceExtrinsic.AvailAssurances,
			types.AvailAssurance{
				ValidatorIndex: availAssurance.ValidatorIndex,
				Anchor:         availAssurance.Anchor,
				Bitfield:       availAssurance.Bitfield,
				Signature:      availAssurance.Signature,
			})
	}

	kappa := types.ValidatorsData{
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		types.Validator{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
			Bls:          types.BlsPublic(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	blockchain.GetInstance().GetPriorStates().SetKappa(kappa)

	err = assuranceExtrinsic.ValidateSignature()
	if err == nil {
		t.Errorf("Expected failed, but validate the signature")
	}
}

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestAssuranceTestVectors(t *testing.T) {
	dir := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "stf", "assurances", types.TEST_MODE)

	// Read binary files
	binFiles, err := utils.GetTargetExtensionFiles(dir, utils.BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		if binFile == "no_assurances_with_stale_report-1.bin" {
			continue
		}
		// Read the binary file
		binPath := filepath.Join(dir, binFile)

		// Load preimages test case
		a := &jamtests_assurances.AssuranceTestCase{}

		err := utils.GetTestFromBin(binPath, a)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Get blockchain instance and required states
		blockchain.ResetInstance()
		s := blockchain.GetInstance()

		// Add block
		s.GetPriorStates().SetKappa(a.PreState.CurrValidators)
		s.GetPriorStates().SetRho(a.PreState.AvailAssignments)
		s.GetIntermediateStates().SetRhoDagger(a.PreState.AvailAssignments)
		block := types.Block{
			Header: types.Header{
				Slot:   a.Input.Slot,
				Parent: a.Input.Parent,
			},
			Extrinsic: types.Extrinsic{
				Assurances: a.Input.Assurances,
			},
		}
		s.AddBlock(block)

		assuranceErr := Assurance()
		t.Logf("assuranceErr: %v", assuranceErr)

		// Get output state
		rhoDoubleDagger := s.GetIntermediateStates().GetRhoDoubleDagger()

		// Filter reports: compare rhoDagger with rhoDoubleDagger
		ourOutput := make([]types.WorkReport, 0)

		// rhoDagger >= rhoDoubleDagger
		if !reflect.DeepEqual(a.PreState.AvailAssignments, rhoDoubleDagger) {
			for i := 0; i < types.CoresCount; i++ {
				if !reflect.DeepEqual(a.PreState.AvailAssignments[i], rhoDoubleDagger[i]) {
					// For debugging
					// t.Logf("rhoDagger[%d]: %v", i, a.PreState.AvailAssignments[i])
					// t.Logf("rhoDoubleDagger[%d]: %v", i, rhoDoubleDagger[i])
					ourOutput = append(ourOutput, a.PreState.AvailAssignments[i].Report)
				}
			}
		}

		// For debugging
		// t.Logf("ourOutput: %v", ourOutput)

		// Validate output state
		if a.Output.Err != nil {
			if assuranceErr == nil || assuranceErr.Error() != a.Output.Err.Error() {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Should raise Error %v but got %v", a.Output.Err, assuranceErr)
			} else {
				t.Logf("ErrorCode matched: expected %v, got %v", a.Output.Err, assuranceErr)
				t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
			}
		} else {
			if assuranceErr != nil {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("No Error expected but got %v", assuranceErr)
			} else if !cmp.Equal(ourOutput, a.Output.Ok.Reported, cmpopts.EquateEmpty(), cmpopts.IgnoreUnexported()) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				diff := cmp.Diff(ourOutput, a.Output.Ok.Reported, cmpopts.EquateEmpty(), cmpopts.IgnoreUnexported())
				t.Fatalf("Result Outputs are not equal: %v", diff)
			} else if !reflect.DeepEqual(rhoDoubleDagger, a.PostState.AvailAssignments) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				diff := cmp.Diff(a.PostState.AvailAssignments, rhoDoubleDagger)
				t.Fatalf("Result States are not equal: %v", diff)
			} else {
				t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
			}
		}
	}
}

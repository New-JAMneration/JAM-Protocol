package extrinsic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"io/ioutil"
	"os"
	"testing"
)

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

	bytes, err := ioutil.ReadAll(file)
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
	PostStateTest       PreStateTest              `json:"post_state,omitemtpy"`
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
				Bitfield:       DecodeJSONByte(availAssurance.Bitfield),
				Signature:      availAssurance.Signature,
			})
	}

	store.GetInstance().AddBlock(types.Block{
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

// 11.12
func TestAssuranceSortUnique(t *testing.T) {
	wrapper, err := ParseData[Input]("assurance_data/assurers_not_sorted_or_unique-1.json")
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
				Bitfield:       DecodeJSONByte(availAssurance.Bitfield),
				Signature:      availAssurance.Signature,
			})
	}

	assuranceExtrinsic.SortUnique()

	expected := NewAvailAssuranceController()

	expected.AvailAssurances = append(expected.AvailAssurances,
		types.AvailAssurance{
			ValidatorIndex: 0,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0xeab18c2630b3debf04f7141686097131e4983a818e4d8c281c269132e4fc15c0d7bbe963adc39ce2251bf9080c06f731c9e16eda9eff0dd6919a8e12f5cd9b00")),
		},
		types.AvailAssurance{
			ValidatorIndex: 1,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0x4a8fa6859dab41d76fa801ad076e2eda3f1e06a26e13df362b5da57f53c5ae4eef6454eb2d9327c4e94b3827f03ad3e8f796007571ddff898f394cac75bbdb0d")),
		},
		types.AvailAssurance{
			ValidatorIndex: 2,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0xdbd50734b049bcc9e25f5c4d2d2b635e22ec1d4eefcc324863de9e1673bacb4b7ac4424a946abae83755908a3f77470776c160e7d5b42991c1b8914bfc16b700")),
		},
		types.AvailAssurance{
			ValidatorIndex: 3,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0x2e1c0fe5ada7046355c7a8b23320dea86edf0df6410d13126f738755dec8f45652fd8c7ac2c84e682d745d2273977d03916865236fa93c9484bc41ed4318d30a")),
		},
		types.AvailAssurance{
			ValidatorIndex: 4,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0xa3afee85825aefb49cfe10000b72d22321f6d562f89f57f56da813f62761130774e2540b2c0ce33da3c28fcbffe52ea0d1eccfbd859be46835128c4cc87fb50c")),
		},
		types.AvailAssurance{
			ValidatorIndex: 5,
			Anchor:         types.OpaqueHash(hexToBytes("0xd61a38a0f73beda90e8c1dfba731f65003742539f4260694f44e22cabef24a8e")),
			Bitfield:       hexToBytes("0x03"),
			Signature:      types.Ed25519Signature(hexToBytes("0xed0d7e4258c6feeecac6ef70db5c866b7fd21af3e409315c79a83a040c50f5f4404b2ab59a1752101ce1af03b2ebd41bfdb1595a2df83f88d937a974b81f2709")),
		},
	)

	for i := 0; i < len(expected.AvailAssurances); i++ {
		if expected.AvailAssurances[i].ValidatorIndex != assuranceExtrinsic.AvailAssurances[i].ValidatorIndex {
			t.Errorf("AvailAssuranceController.SortUnique failed : expected.AvailAssurances[%d].ValidatorIndex != assuranceExtrinsic.AvailAssurances[%d].ValidatorIndex", i, i)
		}

		for j := 0; j < len(expected.AvailAssurances[i].Bitfield); j++ {
			if expected.AvailAssurances[i].Bitfield[j] != assuranceExtrinsic.AvailAssurances[i].Bitfield[j] {
				t.Errorf("AvailAssuranceController.SortUnique failed : expected.AvailAssurances[%d].Bitfield[%d] != assuranceExtrinsic.AvailAssurances[%d].Bitfield[%d]", i, j, i, j)
			}
		}

		for j := 0; j < len(expected.AvailAssurances[i].Signature); j++ {
			if expected.AvailAssurances[i].Signature[j] != assuranceExtrinsic.AvailAssurances[i].Signature[j] {
				t.Errorf("AvailAssuranceController.SortUnique failed : expected.AvailAssurances[%d].Signature[%d] != assuranceExtrinsic.AvailAssurances[%d].Signature[%d]", i, j, i, j)
			}
		}

		for j := 0; j < len(expected.AvailAssurances[i].Anchor); j++ {
			if expected.AvailAssurances[i].Anchor[j] != assuranceExtrinsic.AvailAssurances[i].Anchor[j] {
				t.Errorf("AvailAssuranceController.SortUnique failed : expected.AvailAssurances[%d].Anchor[%d] != assuranceExtrinsic.AvailAssurances[%d].Anchor[%d]", i, j, i, j)
			}
		}
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
				Bitfield:       DecodeJSONByte(availAssurance.Bitfield),
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

	store.GetInstance().GetPosteriorStates().SetKappa(kappa)

	err = assuranceExtrinsic.ValidateSignature()
	if err == nil {
		t.Errorf("Expected failed, but validate the signature")
	}
}

func TestValidateBitField(t *testing.T) {
	wrapper, err := ParseData[Input]("assurance_data/assurance_for_not_engaged_core-1.json")
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
				Bitfield:       DecodeJSONByte(availAssurance.Bitfield),
				Signature:      availAssurance.Signature,
			})
	}

	rhoDagger := make(types.AvailabilityAssignments, 2)
	rhoDagger[0] = &types.AvailabilityAssignment{
		Report: types.WorkReport{
			PackageSpec: types.WorkPackageSpec{
				Hash:         types.WorkPackageHash(hexToBytes("0x63c03371b9dad9f1c60473ec0326c970984e9c90c0b5ed90eba6ada471ba4d86")),
				Length:       types.U32(12345),
				ErasureRoot:  types.ErasureRoot(hexToBytes("0x58e5c51934af8039cde6c9683669a9802021c0e9fc3bda4e9ecc986def429389")),
				ExportsRoot:  types.ExportsRoot(hexToBytes("0xc74f0ee9bf7e8531eae672a7995b9a209153d1891610d032572ecea56cc11d9b")),
				ExportsCount: types.U16(3),
			},
			Context:           types.RefineContext{},
			CoreIndex:         types.CoreIndex(0),
			AuthorizerHash:    types.OpaqueHash(hexToBytes("0x022e5e165cc8bd586404257f5cd6f5a31177b5c951eb076c7c10174f90006eef")),
			AuthOutput:        types.ByteSequence(hexToBytes("0x")),
			SegmentRootLookup: types.SegmentRootLookup{},
			Results: []types.WorkResult{
				{
					ServiceId:     types.ServiceId(129),
					CodeHash:      types.OpaqueHash(hexToBytes("0x8178abf4f459e8ed591be1f7f629168213a5ac2a487c28c0ef1a806198096c7a")),
					PayloadHash:   types.OpaqueHash(hexToBytes("0xfa99b97e72fcfaef616108de981a59dc3310e2a9f5e73cd44d702ecaaccd8696")),
					AccumulateGas: types.Gas(120),
					Result: types.WorkExecResult{
						"ok": hexToBytes("0x64756d6d792d726573756c74"),
					},
				},
			},
		},
		Timeout: types.TimeSlot(11),
	}
	assuranceExtrinsic.BitfieldOctetSequenceToBinarySequence()
	store.GetInstance().GetIntermediateStates().SetRhoDagger(rhoDagger)
	err = assuranceExtrinsic.ValidateBitField()
	if err == nil {
		t.Errorf("Expected error, but validate")
	}
}

func TestFilterAvailableReports(t *testing.T) {
	wrapper, err := ParseData[Input]("assurance_data/some_assurances-1.json")
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
				Bitfield:       DecodeJSONByte(availAssurance.Bitfield),
				Signature:      availAssurance.Signature,
			})
	}

	rhoDagger := make(types.AvailabilityAssignments, 2)
	rhoDagger[0] = &types.AvailabilityAssignment{
		Report: types.WorkReport{
			PackageSpec: types.WorkPackageSpec{
				Hash:         types.WorkPackageHash(hexToBytes("0x63c03371b9dad9f1c60473ec0326c970984e9c90c0b5ed90eba6ada471ba4d86")),
				Length:       types.U32(12345),
				ErasureRoot:  types.ErasureRoot(hexToBytes("0x58e5c51934af8039cde6c9683669a9802021c0e9fc3bda4e9ecc986def429389")),
				ExportsRoot:  types.ExportsRoot(hexToBytes("0xc74f0ee9bf7e8531eae672a7995b9a209153d1891610d032572ecea56cc11d9b")),
				ExportsCount: types.U16(3),
			},
			Context: types.RefineContext{
				Anchor:           types.HeaderHash(hexToBytes("0xc0564c5e0de0942589df4343ad1956da66797240e2a2f2d6f8116b5047768986")),
				StateRoot:        types.StateRoot(hexToBytes("0xf6967658df626fa39cbfb6014b50196d23bc2cfbfa71a7591ca7715472dd2b48")),
				BeefyRoot:        types.BeefyRoot(hexToBytes("0x9329de635d4bbb8c47cdccbbc1285e48bf9dbad365af44b205343e99dea298f3")),
				LookupAnchor:     types.HeaderHash(hexToBytes("0x168490e085497fcb6cbe3b220e2fa32456f30c1570412edd76ccb93be9254fef")),
				LookupAnchorSlot: types.TimeSlot(4),
				Prerequisites:    []types.OpaqueHash{},
			},
			CoreIndex:         types.CoreIndex(0),
			AuthorizerHash:    types.OpaqueHash(hexToBytes("0x022e5e165cc8bd586404257f5cd6f5a31177b5c951eb076c7c10174f90006eef")),
			AuthOutput:        types.ByteSequence(hexToBytes("0x")),
			SegmentRootLookup: types.SegmentRootLookup{},
			Results: []types.WorkResult{
				{
					ServiceId:     types.ServiceId(129),
					CodeHash:      types.OpaqueHash(hexToBytes("0x8178abf4f459e8ed591be1f7f629168213a5ac2a487c28c0ef1a806198096c7a")),
					PayloadHash:   types.OpaqueHash(hexToBytes("0xfa99b97e72fcfaef616108de981a59dc3310e2a9f5e73cd44d702ecaaccd8696")),
					AccumulateGas: types.Gas(120),
					Result: types.WorkExecResult{
						"ok": hexToBytes("0x64756d6d792d726573756c74"),
					},
				},
			},
		},
		Timeout: types.TimeSlot(11),
	}
	rhoDagger[1] = &types.AvailabilityAssignment{
		Report: types.WorkReport{
			PackageSpec: types.WorkPackageSpec{
				Hash:         types.WorkPackageHash(hexToBytes("0xc7e675b7e3450cf70d436172644faef291c2aa905b1fe81c068785cd6e1e44e5")),
				Length:       types.U32(12345),
				ErasureRoot:  types.ErasureRoot(hexToBytes("0x093a290c71d39876d2c5bf3cb8815defc25444166d989459cddf0ac3f90f37df")),
				ExportsRoot:  types.ExportsRoot(hexToBytes("0x2b9b9bc0fceabf4a91373a8a333d7a6d59e2151a9ccc020e418ce226f4977fb3")),
				ExportsCount: types.U16(3),
			},
			Context: types.RefineContext{
				Anchor:           types.HeaderHash(hexToBytes("0xc0564c5e0de0942589df4343ad1956da66797240e2a2f2d6f8116b5047768986")),
				StateRoot:        types.StateRoot(hexToBytes("0xf6967658df626fa39cbfb6014b50196d23bc2cfbfa71a7591ca7715472dd2b48")),
				BeefyRoot:        types.BeefyRoot(hexToBytes("0x9329de635d4bbb8c47cdccbbc1285e48bf9dbad365af44b205343e99dea298f3")),
				LookupAnchor:     types.HeaderHash(hexToBytes("0x168490e085497fcb6cbe3b220e2fa32456f30c1570412edd76ccb93be9254fef")),
				LookupAnchorSlot: types.TimeSlot(4),
				Prerequisites:    []types.OpaqueHash{},
			},
			CoreIndex:         types.CoreIndex(1),
			AuthorizerHash:    types.OpaqueHash(hexToBytes("0x022e5e165cc8bd586404257f5cd6f5a31177b5c951eb076c7c10174f90006eef")),
			AuthOutput:        types.ByteSequence(hexToBytes("0x")),
			SegmentRootLookup: types.SegmentRootLookup{},
			Results: []types.WorkResult{
				{
					ServiceId:     types.ServiceId(198),
					CodeHash:      types.OpaqueHash(hexToBytes("0xc69d09eaac7e1af761c2448987e69929cf86960d091556dab5caf7ded9b3e766")),
					PayloadHash:   types.OpaqueHash(hexToBytes("0xd55e07438aeeeb0d6509ab28af8a758d1fb70424db6b27c7e1ef6473e721c328")),
					AccumulateGas: types.Gas(157),
					Result: types.WorkExecResult{
						"ok": hexToBytes("0x64756d6d792d726573756c74"),
					},
				},
			},
		},
		Timeout: types.TimeSlot(11),
	}
	assuranceExtrinsic.BitfieldOctetSequenceToBinarySequence()
	store.GetInstance().GetIntermediateStates().SetRhoDagger(rhoDagger)
	store.GetInstance().AddBlock(types.Block{
		Header: types.Header{
			Slot: input.Slot,
		},
	})
	assuranceExtrinsic.FilterAvailableReports()

	rhoDoubleDagger := store.GetInstance().GetIntermediateStates().GetRhoDoubleDagger()

	expected := rhoDagger
	expected[0] = nil

	if len(expected) != len(rhoDoubleDagger) {
		t.Errorf("FilterAvailableReports failed : length not fit")
	} else {
		for i := 0; i < len(rhoDagger); i++ {
			if (rhoDoubleDagger[i] == nil && expected[i] != nil) || (rhoDoubleDagger[i] != nil && expected[i] == nil) {
				t.Errorf("FilterAvailableReports failed ")
			}
		}
	}
}

func TestBitfieldOctetSequenceToBinarySequence(t *testing.T) {
	assuranceExtrinsic := NewAvailAssuranceController()

	assuranceExtrinsic.AvailAssurances = append(assuranceExtrinsic.AvailAssurances,
		types.AvailAssurance{
			Bitfield: []byte{0x03}})

	assuranceExtrinsic.BitfieldOctetSequenceToBinarySequence()
	expected := []byte{1, 1, 0, 0, 0, 0, 0, 0}

	if bytes.Compare(assuranceExtrinsic.AvailAssurances[0].Bitfield, expected) != 0 {
		t.Errorf("BitfieldOctetSequenceToBinarySequence failed")
	}

}

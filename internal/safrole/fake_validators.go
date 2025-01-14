package safrole

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// [JAM Testnet Key generation](https://github.com/jam-duna/jamtestnet/tree/main/key#jam-testnet-key-generation)
// We will create 6 validators for testing purposes.

type (
	BandersnatchPrivate [32]byte
	Ed25519Private      [32]byte
	BlsPrivate          [32]byte
)

type FakeValidatorDTO struct {
	Bandersnatch        string `json:"bandersnatch"`
	Ed25519             string `json:"ed25519"`
	BLS                 string `json:"bls"`
	BandersnatchPrivate string `json:"bandersnatch_priv"`
	Ed25519Private      string `json:"ed25519_priv"`
	BlsPrivate          string `json:"bls_priv"`
}

type FakeValidatorDTOs []FakeValidatorDTO

type FakeValidator struct {
	Bandersnatch        types.BandersnatchPublic
	Ed25519             types.Ed25519Public
	BLS                 types.BlsPublic
	BandersnatchPrivate BandersnatchPrivate
	Ed25519Private      Ed25519Private
	BlsPrivate          BlsPrivate
}

type FakeValidators []FakeValidator

func hex2Bytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v\n", err)
	}
	return bytes
}

func LoadRawFakeValidators() FakeValidatorDTOs {
	// Open the JSON file
	file, err := os.Open("../input/validator/fake_validators.json")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Unmarshal the JSON data
	var fakeValidatorDTOs FakeValidatorDTOs
	err = json.Unmarshal(byteValue, &fakeValidatorDTOs)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
	}

	return fakeValidatorDTOs
}

func MapFakeValidators(fakeValidatorDTOs FakeValidatorDTOs) FakeValidators {
	var fakeValidators FakeValidators

	for _, fakeValidatorDTO := range fakeValidatorDTOs {
		fakeValidators = append(fakeValidators, FakeValidator{
			Bandersnatch:        types.BandersnatchPublic(hex2Bytes(fakeValidatorDTO.Bandersnatch)),
			Ed25519:             types.Ed25519Public(hex2Bytes(fakeValidatorDTO.Ed25519)),
			BLS:                 types.BlsPublic(hex2Bytes(fakeValidatorDTO.BLS)),
			BandersnatchPrivate: BandersnatchPrivate(hex2Bytes(fakeValidatorDTO.BandersnatchPrivate)),
			Ed25519Private:      Ed25519Private(hex2Bytes(fakeValidatorDTO.Ed25519Private)),
			BlsPrivate:          BlsPrivate(hex2Bytes(fakeValidatorDTO.BlsPrivate)),
		})
	}

	return fakeValidators
}

func LoadFakeValidators() FakeValidators {
	fakeValidatorDTOs := LoadRawFakeValidators()
	return MapFakeValidators(fakeValidatorDTOs)
}

// INFO: This function only for GetBandersnatchRingRootCommmitment
func CreateRingVRFHandler(bandersnatchKeys []types.BandersnatchPublic, proverIdx uint) (*vrf.Handler, error) {
	if proverIdx >= uint(len(bandersnatchKeys)) {
		return nil, fmt.Errorf("proverIdx is out of range: %v", proverIdx)
	}

	// Only the secret key of the prover is needed
	fakeValidators := LoadFakeValidators()
	skBytes := fakeValidators[proverIdx].BandersnatchPrivate[:]

	// Use input bandersnatch keys to create the ring
	ringBytes := []byte{}
	ringSize := uint(len(bandersnatchKeys))

	for _, bandersnatch := range bandersnatchKeys {
		ringBytes = append(ringBytes, bandersnatch[:]...)
	}

	return vrf.NewHandler(ringBytes, skBytes, ringSize, proverIdx)
}

func CreateVRFHandler(bandersnatchKey types.BandersnatchPublic) (*vrf.Handler, error) {
	// Only the secret key of the prover is needed
	fakeValidators := LoadFakeValidators()
	var private BandersnatchPrivate
	for _, bandersnatch := range fakeValidators {
		//fmt.Println(bandersnatchKey)
		//fmt.Println(bandersnatch.Bandersnatch)
		if bandersnatch.Bandersnatch == bandersnatchKey {
			private = bandersnatch.BandersnatchPrivate
			// proverIdx = uint(idx)
			// fmt.Println("SUCCESS")
		}
	}
	skBytes := private[:]
	// fmt.Println(skBytes)
	// Use input bandersnatch keys to create the ring
	ringBytes := []byte{}
	ringSize := uint(1)
	ringBytes = append(ringBytes, bandersnatchKey[:]...)
	return vrf.NewHandler(ringBytes, skBytes, ringSize, 0)
}

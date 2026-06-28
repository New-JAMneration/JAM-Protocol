package safrole

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// TinyValidatorCount is the number of validators in tiny test mode (V=6).
const TinyValidatorCount = 6

// DeriveTinyValidator builds validator public data and the JIP-5 bandersnatch
// secret seed for trivial_seed(index).
func DeriveTinyValidator(index uint32) (types.Validator, []byte, error) {
	seed := keystore.TrivialSeed(index)
	_, bandersnatchSecret, ed25519Public, bandersnatchPublic, err := keystore.DeriveValidatorKeys(seed[:])
	if err != nil {
		return types.Validator{}, nil, err
	}

	return types.Validator{
		Bandersnatch: bandersnatchPublic,
		Ed25519:      ed25519Public,
		// BLS derivation is not part of JIP-5 in this codebase yet.
		Bls:      types.BlsPublic{},
		Metadata: types.ValidatorMetadata{},
	}, append([]byte(nil), bandersnatchSecret...), nil
}

// LoadTinyValidatorsData returns the tiny-mode validator set using JIP-5
// derivation (trivial_seed(0)..trivial_seed(5)).
func LoadTinyValidatorsData() (types.ValidatorsData, error) {
	data := make(types.ValidatorsData, 0, TinyValidatorCount)
	for i := uint32(0); i < TinyValidatorCount; i++ {
		validator, _, err := DeriveTinyValidator(i)
		if err != nil {
			return nil, err
		}
		data = append(data, validator)
	}
	return data, nil
}

// LookupBandersnatchSecretSeed returns the JIP-5 bandersnatch secret seed for a
// known tiny-mode validator public key. Production nodes should resolve keys
// from keystore instead.
func LookupBandersnatchSecretSeed(bandersnatchKey types.BandersnatchPublic) ([]byte, error) {
	for i := uint32(0); i < TinyValidatorCount; i++ {
		validator, secretSeed, err := DeriveTinyValidator(i)
		if err != nil {
			return nil, err
		}
		if validator.Bandersnatch == bandersnatchKey {
			return secretSeed, nil
		}
	}
	return nil, fmt.Errorf("bandersnatch secret seed not found for public key")
}

// CreateVRFHandler builds a ring VRF handler for ring-sign paths (e.g. auditing).
// IETF-only paths (seal, entropy) use vrf.IETFSign via ietfSignForBandersnatchKey or SignHeaderEntropy.
func CreateVRFHandler(bandersnatchKey types.BandersnatchPublic) (*vrf.Handler, error) {
	skBytes, err := LookupBandersnatchSecretSeed(bandersnatchKey)
	if err != nil {
		return nil, err
	}
	ringBytes := make([]byte, 0, 32)
	ringSize := uint(1)
	ringBytes = append(ringBytes, bandersnatchKey[:]...)
	return vrf.NewHandler(ringBytes, skBytes, ringSize, 0)
}

// Hex2Bytes decodes a 0x-prefixed hex string into bytes.
func Hex2Bytes(hexString string) []byte {
	if len(hexString) >= 2 && hexString[0:2] == "0x" {
		hexString = hexString[2:]
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		logger.Errorf("failed to decode hex string: %v", err)
	}
	return bytes
}

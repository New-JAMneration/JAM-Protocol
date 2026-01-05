package keystore

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ImportValidatorKeysFromSeed imports Ed25519 and Bandersnatch key pairs from a 32-byte seed
// using JIP-5 derivation method. This is a convenience function that uses DeriveValidatorKeys
// and creates key pair objects suitable for use in the keystore.
func ImportValidatorKeysFromSeed(seed []byte) (
	ed25519KeyPair *Ed25519KeyPair,
	bandersnatchKeyPair *BandersnatchKeyPair,
	err error,
) {
	ed25519SecretSeed, bandersnatchSecret, ed25519Public, bandersnatchPublic, err := DeriveValidatorKeys(seed)
	if err != nil {
		return nil, nil, err
	}

	// Create Ed25519 key pair from derived seed
	ed25519KeyPair, err = ImportEd25519KeyPair(ed25519SecretSeed)
	if err != nil {
		return nil, nil, err
	}

	// Verify the public key matches
	if len(ed25519KeyPair.Public) != 32 {
		return nil, nil, errors.New("invalid Ed25519 public key size")
	}

	// Verify the public key matches the derived one
	var derivedPub types.Ed25519Public
	copy(derivedPub[:], ed25519KeyPair.Public)
	if derivedPub != ed25519Public {
		return nil, nil, errors.New("Ed25519 public key mismatch")
	}

	bandersnatchKeyPair, err = ImportBandersnatchKeyPair(bandersnatchSecret)
	if err != nil {
		return nil, nil, err
	}

	// Verify the public key matches the derived one
	if bandersnatchKeyPair.Public != bandersnatchPublic {
		return nil, nil, errors.New("Bandersnatch public key mismatch")
	}

	return ed25519KeyPair, bandersnatchKeyPair, nil
}

package keystore

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

var cryptoRand = rand.Reader

// BandersnatchKeyPair represents a Bandersnatch key pair for VRF operations
type BandersnatchKeyPair struct {
	private []byte
	Public  types.BandersnatchPublic
}

// Type returns the key type
func (k *BandersnatchKeyPair) Type() KeyType {
	return KeyTypeBandersnatch
}

// PublicKey returns the public key as bytes
func (k *BandersnatchKeyPair) PublicKey() []byte {
	return k.Public[:]
}

// PrivateKey returns the secret seed as bytes
func (k *BandersnatchKeyPair) PrivateKey() []byte {
	return k.private
}

// Sign is not implemented for Bandersnatch VRF.
// VRF operations require context and should use vrf package instead.
func (k *BandersnatchKeyPair) Sign(message []byte) ([]byte, error) {
	return nil, fmt.Errorf("Bandersnatch VRF requires context parameter. Use CreateHandler() and Handler.IETFSign(context, message) instead")
}

// Verify is not implemented for Bandersnatch VRF.
// VRF operations require context and should use vrf package instead.
func (k *BandersnatchKeyPair) Verify(message []byte, signature []byte) bool {
	return false
}

// CreateHandler creates a vrf.Handler for VRF operations with a custom ring.
// This is the recommended way to perform VRF operations with Bandersnatch.
func (k *BandersnatchKeyPair) CreateHandler(ring []byte, ringSize, proverIdx uint) (*vrf.Handler, error) {
	return vrf.NewHandler(ring, k.private, ringSize, proverIdx)
}

// NewBandersnatchKeyPair generates a new Bandersnatch key pair
// This generates a random secret seed and derives the public key
func NewBandersnatchKeyPair() (*BandersnatchKeyPair, error) {
	// Generate a random 32-byte seed
	private := make([]byte, 32)
	if _, err := cryptoRand.Read(private); err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %w", err)
	}

	// Derive public key from secret seed
	publicKeyBytes, err := vrf.GetPublicKeyFromSecret(private)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	var publicKey types.BandersnatchPublic
	copy(publicKey[:], publicKeyBytes)

	return &BandersnatchKeyPair{
		private: private,
		Public:  publicKey,
	}, nil
}

// ImportBandersnatchKeyPair imports a Bandersnatch key pair from a 32-byte secret seed
func ImportBandersnatchKeyPair(private []byte) (*BandersnatchKeyPair, error) {
	if len(private) != 32 {
		return nil, errors.New("secret seed must be exactly 32 bytes")
	}

	// Derive public key from secret seed
	publicKeyBytes, err := vrf.GetPublicKeyFromSecret(private)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	var publicKey types.BandersnatchPublic
	copy(publicKey[:], publicKeyBytes)

	return &BandersnatchKeyPair{
		private: private,
		Public:  publicKey,
	}, nil
}

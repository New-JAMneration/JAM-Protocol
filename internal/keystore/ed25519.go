package keystore

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"

	"github.com/hdevalence/ed25519consensus"
)

type Ed25519KeyPair struct {
	private ed25519.PrivateKey
	Public  ed25519.PublicKey
}

func (k *Ed25519KeyPair) Type() KeyType {
	return KeyTypeEd25519
}

func (k *Ed25519KeyPair) PublicKey() []byte {
	return k.Public
}

func (k *Ed25519KeyPair) PrivateKey() []byte {
	return k.private
}

func (k *Ed25519KeyPair) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(k.private, message), nil
}

func (k *Ed25519KeyPair) Verify(message []byte, signature []byte) bool {
	// We use ed25519consensus to verify the signature
	// which is compliant with ZIP 215.
	return ed25519consensus.Verify(k.Public, message, signature)
	// return ed25519.Verify(k.Public, message, signature)
}

func NewEd25519KeyPair() (*Ed25519KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &Ed25519KeyPair{private: priv, Public: pub}, nil
}

// imports from a 32-byte seed
func ImportEd25519KeyPair(seed []byte) (*Ed25519KeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, errors.New("invalid ed25519 seed size")
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	return &Ed25519KeyPair{private: priv, Public: pub}, nil
}

// imports from a 32-byte private key
func FromEd25519PrivateKey(priv ed25519.PrivateKey) (*Ed25519KeyPair, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid ed25519 private key size")
	}
	pub := priv.Public().(ed25519.PublicKey)
	return &Ed25519KeyPair{private: priv, Public: pub}, nil
}

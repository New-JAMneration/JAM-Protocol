package keystore

import (
	"crypto/ed25519"
	"errors"
)

// KeyPair is a minimal signing interface used by CE handlers.
//
// This package intentionally avoids depending on `internal/keystore` because that
// package currently pulls in Bandersnatch/VRF FFI symbols which may not be present
// in all build environments, while CE handlers only require Ed25519 signing.
type KeyPair interface {
	Sign(message []byte) ([]byte, error)
}

type Ed25519KeyPair struct {
	private ed25519.PrivateKey
	Public  ed25519.PublicKey
}

func (k *Ed25519KeyPair) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(k.private, message), nil
}

func FromEd25519PrivateKey(priv ed25519.PrivateKey) (KeyPair, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid ed25519 private key size")
	}
	pub := priv.Public().(ed25519.PublicKey)
	return &Ed25519KeyPair{private: priv, Public: pub}, nil
}

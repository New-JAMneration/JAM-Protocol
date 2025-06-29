package quic

import (
	"crypto/tls"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/cert"
	"github.com/quic-go/quic-go"
)

// NewTLSConfig creates a TLS configuration based on the isServer parameter.
// In server mode, it generates a self-signed certificate. In client mode, it
// skips certificate verification (for demo purposes only).
//
// ALPN (Application Layer Protocol Negotiation) is used to identify the protocol,
// version, and chain. The protocol identifier should follow the format:
// "jamnp-s/V/H" or "jamnp-s/V/H/builder", where:
//   - V is the protocol version (currently 0),
//   - H is the first 8 nibbles of the genesis header hash of the chain (in lower-case hexadecimal),
//   - The "/builder" suffix is used by the initiator when connecting as a work-package builder.
//
// TODO: Update the ALPN logic to derive the correct protocol identifier based on chain data.
func NewTLSConfig(isServer, isBuilder bool) (*tls.Config, error) {
	// randomly generated seed for Ed25519 key generation
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}

	return cert.TLSConfigGen(seed, isServer, isBuilder)
}

// generateSelfSignedCerert
func GenerateSelfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 年有效
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return tls.X509KeyPair(certPEM, keyPEM)
}

// NewQuicConfig
func NewQuicConfig() *quic.Config {
	return &quic.Config{
		MaxIncomingStreams: 100,
	}
}

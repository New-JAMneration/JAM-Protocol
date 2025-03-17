package cert

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base32"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"

	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// Ed25519KeyGen generates a new Ed25519 key pair from 32byte seed
// Notice ed25519.PrivateKey is key pair of (sk, pk)
func Ed25519KeyGen(seed []byte) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, nil, errors.New("seed must be 32 bytes long")
	}

	// Generate the private key from the seed
	sk := ed25519.NewKeyFromSeed(seed)
	pk := sk.Public().(ed25519.PublicKey)

	return sk, pk, nil
}

// EncodeBase32 encodes data using base32 with a custom alphabet
func EncodeBase32(data []byte) string {
	// Using the specified alphabet: "abcdefghijklmnopqrstuvwxyz234567"
	encoder := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)
	return strings.ToLower(encoder.EncodeToString(data))
}

// SelfSignedCertGen generates a self-signed X.509 certificate using Ed25519
func SelfSignedCertGen(sk ed25519.PrivateKey, pk ed25519.PublicKey) (tls.Certificate, error) {
	// Create a unique serial number for the certificate
	// Use 128-bit random number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %v", err)
	}

	// Encode the public key in base32
	encodedPubKey := EncodeBase32(pk)

	// Create the DNS name: "e" followed by the encoded public key
	dnsName := "e" + encodedPubKey

	// Ensure the DNS name is exactly 53 characters
	if len(dnsName) != 53 {
		return tls.Certificate{}, errors.New("DNS name must be 53 characters long")
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{dnsName},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, // localhost IPv4 and IPv6
	}

	// Create the certificate
	selfSignedCert, err := x509.CreateCertificate(rand.Reader, &template, &template, pk, sk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Encode the certificate in PEM format
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: selfSignedCert,
	})
	if pemCert == nil {
		return tls.Certificate{}, errors.New("failed to encode certificate to PEM")
	}

	// Encode private key to PKCS8 and then to PEM
	skByte, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to marshal private key: %v", err)
	}

	pemSk := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: skByte,
	})
	if pemSk == nil {
		return tls.Certificate{}, errors.New("failed to encode private key to PEM")
	}
	tlsCert, err := tls.X509KeyPair(pemCert, pemSk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create X509 key pair: %v", err)
	}

	// validate TLS cert before generation
	err = ValidateTlsCertificate(tlsCert)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tlsCert, nil
}

// ALPNGen generates an ALPN string based on a genesis header and builder flag
// Example outputs: "jamnp-s/0/H" or "jamnp-s/0/H/builder"
func ALPNGen(isBuilder bool) ([]string, error) {
	genesisBlock, err := store.GetGenesisBlockFromBin()
	if err != nil {
		return nil, err
	}
	genesisBlockHeaderHash := hash.Blake2bHashPartial(utils.HeaderSerialization(genesisBlock.Header), 8)
	baseALPN := "jamnp-s/0/" + string(genesisBlockHeaderHash)
	builderALPN := baseALPN + "/builder"

	nextProtos := []string{baseALPN}
	if isBuilder {
		nextProtos = []string{builderALPN, baseALPN}
	}
	// tlsConfig := &tls.Config{
	// 	Certificates: []tls.Certificate{cert},
	// 	NextProtos:   nextProtos,
	// }

	return nextProtos, nil
}

// NewTLSConfig creates a new TLS configuration with a self-signed certificate.
// isBuilder make sure have a slot reserved for work package builder
func NewTLSConfig(seed []byte, isServer bool, isBuilder bool) (*tls.Config, error) {
	// default ALPN
	alpn, err := ALPNGen(false)
	if err != nil {
		return nil, err
	}

	if isServer {
		sk, pk, err := Ed25519KeyGen(seed)
		if err != nil {
			return nil, err
		}
		cert, err := SelfSignedCertGen(sk, pk)
		if err != nil {
			return nil, err
		}
		alpn, err := ALPNGen(isBuilder)
		if err != nil {
			return nil, err
		}
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.VerifyClientCertIfGiven,

			// Only TLS1.3 allowed
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
			// explicitly set curve to ED25519 since we're not relying on default curve selection
			CurvePreferences: []tls.CurveID{tls.CurveID(tls.Ed25519)},
			NextProtos:       alpn,
		}, nil
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         alpn,
	}, nil
}

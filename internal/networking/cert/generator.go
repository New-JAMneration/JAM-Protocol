package network

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

// GenerateEd25519Key generates a new Ed25519 key pair from 32byte seed
// Notice ed25519.PrivateKey is key pair of (sk, pk)
func GenerateEd25519Key(seed []byte) (ed25519.PrivateKey, ed25519.PublicKey, error) {
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

// GenerateSelfSignedCertificate generates a self-signed X.509 certificate using Ed25519
func GenerateSelfSignedCertificate(sk ed25519.PrivateKey, pk ed25519.PublicKey) (tls.Certificate, error) {
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
	return tlsCert, nil
}

// GenerateALPN generates an ALPN string based on a genesis header and builder flag
// Example outputs: "jamnp-s/0/H" or "jamnp-s/0/H/builder"
func GenerateALPN(cert tls.Certificate, isBuilder bool) (*tls.Config, error) {
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
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   nextProtos,
	}

	return tlsConfig, nil
}

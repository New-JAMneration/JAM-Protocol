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
	if sk == nil {
		return tls.Certificate{}, errors.New("sk is nil")
	} else if pk == nil {
		pk = sk.Public().(ed25519.PublicKey)
	}

	// Create a unique serial number for the certificate
	// Use 128-bit random number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %v", err)
	}

	// Encode the public key in base32
	encodedPk := EncodeBase32(pk)

	// Create the DNS name: "e" followed by the encoded public key
	dnsName := "e" + encodedPk

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

	// Validate TLS cert before generation
	err = ValidateTlsCertificate(tlsCert)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}

// ALPNGen generates an ALPN string based on a genesis header and builder flag
// Example outputs: "jamnp-s/0/H" or "jamnp-s/0/H/builder"
func ALPNGen(isBuilder bool) ([]string, error) {
	// TODO:This is a temporary fix, should be changed to use genesis block from redis
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("error getting redis backend: %v", err)
	}
	genesisBlock, err := redisBackend.GetGenesisBlock(nil)
	if err != nil {
		return nil, err
	}

	genesisBlockHeaderHash := hash.Blake2bHashPartial(utils.HeaderSerialization(genesisBlock.Header), 8)
	// Currently Version is set to 0
	baseALPN := "jamnp-s/0/" + string(genesisBlockHeaderHash)

	nextProtos := []string{baseALPN}
	if isBuilder {
		builderALPN := baseALPN + "/builder"
		nextProtos = []string{builderALPN, baseALPN}
	}

	return nextProtos, nil
}

// TLSConfigGen creates a new TLS configuration with a self-signed certificate.
// isBuilder make sure have a slot reserved for work package builder
func TLSConfigGen(seed []byte, isServer bool, isBuilder bool) (*tls.Config, error) {
	sk, pk, err := Ed25519KeyGen(seed)
	if err != nil {
		return nil, fmt.Errorf("error generating Ed25519 key pair: %v", err)
	}

	selfSignedCert, err := SelfSignedCertGen(sk, pk)
	if err != nil {
		return nil, fmt.Errorf("error generating self-signed certificate: %v", err)
	}

	clientALPN, err := ALPNGen(false)
	if err != nil {
		return nil, fmt.Errorf("error generating ALPN: %v", err)
	}

	// Default TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{selfSignedCert},
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("no certificate provided")
			}
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return err
			}

			return ValidateX509Certificate(cert)
		},
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
		// TODO: ask if specific curve is needed in DH key exchange
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		NextProtos:       clientALPN,
	}

	if isServer {
		// Server ALPN
		serverALPN, err := ALPNGen(isBuilder)
		if err != nil {
			return nil, fmt.Errorf("error generating ALPN: %v", err)
		}

		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return errors.New("peer did not present any certificates")
			}

			if len(state.NegotiatedProtocol) == 0 {
				return errors.New("failed to negotiate protocol (ALPN)")
			}

			if state.Version != tls.VersionTLS13 {
				return errors.New("TLS version is not 1.3")
			}

			return nil
		}
		tlsConfig.NextProtos = serverALPN

		return tlsConfig, nil
	}

	return tlsConfig, nil
}

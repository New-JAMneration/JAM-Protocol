package cert

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
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

// AlternativeName returns a string with prefix "e" and followed by
// the result of import Ed25519 public key base function
func AlternativeName(pk ed25519.PublicKey) string {
	// E^{-1}_{32} deserialize function for 256-bit unsigned integers (serialization codec appendix)
	deserialize := func(pk ed25519.PublicKey) *big.Int {
		x := big.NewInt(0)
		for i, byte := range pk {
			tmp := big.NewInt(int64(byte))
			tmp.Lsh(tmp, uint(8*i))
			x.Add(x, tmp)
		}
		return x
	}

	n := deserialize(pk)

	// Using the specified alphabet: "abcdefghijklmnopqrstuvwxyz234567"
	alphabet := "abcdefghijklmnopqrstuvwxyz234567"

	// B(n, l) encodes the deserialized integer, where n is the integer
	// to base32 and l is the length of the output
	var encode func(n *big.Int, l int) string
	encode = func(n *big.Int, l int) string {
		if l == 0 {
			return ""
		}

		// n mod 32
		mod := new(big.Int).Mod(n, big.NewInt(32))
		// n / 32
		div := new(big.Int).Div(n, big.NewInt(32))

		// Get the character at position (n mod 32)
		char := string(alphabet[mod.Int64()])

		// Recursively encode the remaining part
		return char + encode(div, l-1)
	}

	return "e" + encode(n, 52)
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

	// Create the DNS name: "e" followed by the encoded public key
	dnsName := AlternativeName(pk)

	// Ensure the DNS name is exactly 53 characters
	if len(dnsName) != 53 {
		return tls.Certificate{}, errors.New("DNS name must be 53 characters long")
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{dnsName},
		SignatureAlgorithm:    x509.PureEd25519,                                       // Use Ed25519 signature algorithm
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, // localhost IPv4 and IPv6
		BasicConstraintsValid: true,
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
	store := blockchain.GetInstance()
	genesisBlock := store.GetGenesisBlock()

	// Get first 4 bytes (8 nibbles) of the genesis header hash
	genesisBlockHeaderHash, err := utils.HeaderSerialization(genesisBlock.Header)
	if err != nil {
		return nil, fmt.Errorf("error serializing genesis block header: %v", err)
	}
	genesisBlockHeaderHash = hash.Blake2bHashPartial(genesisBlockHeaderHash, 4)

	// Convert to lowercase hexadecimal string
	hashHex := hex.EncodeToString(genesisBlockHeaderHash)

	// Currently Version is set to 0
	baseALPN := "jamnp-s/0/" + hashHex

	nextProtos := []string{baseALPN}
	if isBuilder {
		builderALPN := baseALPN + "/builder"
		nextProtos = []string{builderALPN}
	}

	return nextProtos, nil
}

// generateTLSCertificate is a helper function that generates Ed25519 key pair and self-signed certificate
func generateTLSCertificate(seed []byte) (tls.Certificate, error) {
	sk, pk, err := Ed25519KeyGen(seed)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("error generating Ed25519 key pair: %v", err)
	}

	selfSignedCert, err := SelfSignedCertGen(sk, pk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("error generating self-signed certificate: %v", err)
	}

	return selfSignedCert, nil
}

// ServerTLSConfigGen creates a TLS configuration for servers.
// Servers always accept both base and builder ALPN protocols.
func ServerTLSConfigGen(seed []byte) (*tls.Config, error) {
	selfSignedCert, err := generateTLSCertificate(seed)
	if err != nil {
		return nil, err
	}

	// The /builder suffix should always be permitted by the side accepting the connection (server)
	// https://jam-docs.onrender.com/knowledge/advanced/simple-networking/spec#alpn
	baseALPN, err := ALPNGen(false)
	if err != nil {
		return nil, fmt.Errorf("error generating base ALPN: %v", err)
	}
	builderALPN, err := ALPNGen(true)
	if err != nil {
		return nil, fmt.Errorf("error generating builder ALPN: %v", err)
	}

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
			log.Printf("peer certificate subject!")
			return ValidateX509Certificate(cert)
		},
		MinVersion:       tls.VersionTLS13,
		MaxVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		NextProtos:       append(builderALPN, baseALPN...),
		ClientAuth:       tls.RequireAnyClientCert,
		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				log.Printf("peer did not present any certificates")
				return errors.New("peer did not present any certificates")
			}

			if len(state.NegotiatedProtocol) == 0 {
				log.Printf("ALPN negotiation failed, no protocol negotiated")
				return errors.New("failed to negotiate protocol (ALPN)")
			}

			if state.Version != tls.VersionTLS13 {
				log.Printf("TLS version is not 1.3, got: %s", tls.VersionName(state.Version))
				return errors.New("TLS version is not 1.3")
			}

			log.Printf("TLS connection established with protocol: %s", state.NegotiatedProtocol)
			return nil
		},
	}

	return tlsConfig, nil
}

// ClientTLSConfigGen creates a TLS configuration for clients.
// isBuilder determines whether to use builder ALPN or base ALPN.
func ClientTLSConfigGen(seed []byte, isBuilder bool) (*tls.Config, error) {
	selfSignedCert, err := generateTLSCertificate(seed)
	if err != nil {
		return nil, err
	}

	clientALPN, err := ALPNGen(isBuilder) // Use builder ALPN if isBuilder is true
	if err != nil {
		return nil, fmt.Errorf("error generating ALPN: %v", err)
	}

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
			log.Printf("peer certificate subject!")
			return ValidateX509Certificate(cert)
		},
		MinVersion:       tls.VersionTLS13,
		MaxVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		NextProtos:       clientALPN,
	}

	return tlsConfig, nil
}

// TLSConfigGen creates a new TLS configuration with a self-signed certificate.
// isBuilder make sure have a slot reserved for work package builder
func TLSConfigGen(seed []byte, isServer bool, isBuilder bool) (*tls.Config, error) {
	if isServer {
		return ServerTLSConfigGen(seed)
	}
	return ClientTLSConfigGen(seed, isBuilder)
}

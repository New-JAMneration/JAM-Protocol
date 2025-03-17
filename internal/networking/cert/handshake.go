package cert

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"time"
)

// TLSHandshakeConfig contains configuration for TLS handshake
type TLSHandshakeConfig struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Address    string
	Timeout    time.Duration
	IsServer   bool
}

// DefaultTLSHandshakeConfig returns default configuration for TLS handshake
func DefaultTLSHandshakeConfig(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, address string) TLSHandshakeConfig {
	return TLSHandshakeConfig{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		Timeout:    30 * time.Second,
		IsServer:   false,
	}
}

// TLSHandshake performs a TLS handshake with the given configuration
// It integrates certificate generation, parsing, and validation
func TLSHandshake(config TLSHandshakeConfig) (interface{}, error) {
	// Step 1: Generate self-signed certificate
	certificate, err := GenSelfSignedCert(config.PrivateKey, config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %v", err)
	}

	// Step 2: Create TLS configuration with certificate validation
	tlsConfig := createTLSConfig(certificate)

	// Step 3: Perform handshake (either as server or client)
	if config.IsServer {
		return performServerHandshake(config.Address, tlsConfig)
	} else {
		return performClientHandshake(config.Address, tlsConfig, config.Timeout)
	}
}

// createTLSConfig creates a TLS configuration with certificate validation
func createTLSConfig(certificate tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAnyClientCert,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("no certificate provided")
			}

			// Parse the certificate
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %v", err)
			}

			// Validate signature algorithm
			if err := ValidateX509SignatureAlgorithm(*cert); err != nil {
				return fmt.Errorf("invalid signature algorithm: %v", err)
			}

			// Validate DNS names
			if err := ValidateX509DNSNames(*cert); err != nil {
				return fmt.Errorf("invalid DNS names: %v", err)
			}

			// Validate public key matches SAN
			if err := ValidateX509PubKeyMatchesSAN(*cert); err != nil {
				return fmt.Errorf("public key does not match SAN: %v", err)
			}

			return nil
		},
	}
}

// performServerHandshake starts a TLS server and waits for connections
func performServerHandshake(address string, tlsConfig *tls.Config) (net.Listener, error) {
	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start TLS server: %v", err)
	}

	log.Printf("TLS server started on %s", address)
	return listener, nil
}

// performClientHandshake connects to a TLS server
func performClientHandshake(address string, tlsConfig *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	dialer := &tls.Dialer{
		Config: tlsConfig,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	log.Printf("Connected to TLS server at %s", address)
	return conn.(*tls.Conn), nil
}

// HandleTLSConnection processes a TLS connection after successful handshake
func HandleTLSConnection(conn *tls.Conn) {
	// Get connection state
	state := conn.ConnectionState()

	// Log certificate information
	for i, cert := range state.PeerCertificates {
		log.Printf("Certificate #%d: Subject=%v, Issuer=%v", i, cert.Subject, cert.Issuer)

		// Extract public key
		pubKey, ok := cert.PublicKey.(ed25519.PublicKey)
		if ok {
			log.Printf("Public Key: %x", pubKey)
		}
	}

	// Additional connection handling logic can be added here
}

// StartTLSServer starts a TLS server with the given configuration
func StartTLSServer(config TLSHandshakeConfig) error {
	// Ensure config is set to server mode
	config.IsServer = true

	// Perform handshake to get listener
	listener, err := TLSHandshake(config)
	if err != nil {
		return err
	}

	// Type assertion
	tlsListener, ok := listener.(net.Listener)
	if !ok {
		return fmt.Errorf("invalid listener type")
	}
	defer tlsListener.Close()

	// Accept connections
	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle connection in a goroutine
		go func(c net.Conn) {
			defer c.Close()
			HandleTLSConnection(c.(*tls.Conn))
		}(conn)
	}
}

// ConnectToTLSServer connects to a TLS server with the given configuration
func ConnectToTLSServer(config TLSHandshakeConfig) (*tls.Conn, error) {
	// Ensure config is set to client mode
	config.IsServer = false

	// Perform handshake to get connection
	conn, err := TLSHandshake(config)
	if err != nil {
		return nil, err
	}

	// Type assertion
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("invalid connection type")
	}

	return tlsConn, nil
}

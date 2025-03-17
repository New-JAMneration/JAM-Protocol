package cert

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

// The certificate uses Ed25519 as the signature algorithm.
func ValidateX509SignatureAlgorithm(cert x509.Certificate) error {
	if cert.SignatureAlgorithm != x509.PureEd25519 {
		return errors.New("the signature algorithm is not Ed25519")
	}

	return nil
}

// The SAN is exactly 53 characters, starting with "e" followed by a base32 encoded string (using the specified alphabet).
func ValidateX509DNSNames(cert x509.Certificate) error {
	dnsNames := cert.DNSNames

	for _, dnsName := range dnsNames {
		if len(dnsName) != 53 || dnsName[0] != 'e' {
			return fmt.Errorf("invalid DNS name: %v", dnsName)
		}
	}

	return nil
}

// The certificate’s public key matches the information encoded in the SAN.
func ValidateX509PubKeyMatchesSAN(cert x509.Certificate) error {
	pk := cert.PublicKey
	pkEd25519, ok := pk.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("invalid public key type: %T", pk)
	}

	expectedEncodedPubKey := EncodeBase32(pkEd25519)

	dnsNames := cert.DNSNames

	for _, dnsName := range dnsNames {
		if dnsName != "e"+expectedEncodedPubKey {
			return fmt.Errorf("invalid DNS name: %v", dnsName)
		}
	}

	return nil
}

func ParseCertificateFromPem(pemCerts []byte) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate(pemCerts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

// Validate the TLS certificate.
func ValidateTlsCertificate(cert tls.Certificate) error {
	x509Cert, err := ParseCertificateFromPem(cert.Certificate[0])
	if err != nil {
		return err
	}

	if x509Cert == nil {
		return fmt.Errorf("the x509 certificate is nil")
	}

	if err := ValidateX509SignatureAlgorithm(*x509Cert); err != nil {
		return err
	}

	if err := ValidateX509DNSNames(*x509Cert); err != nil {
		return err
	}

	if err := ValidateX509PubKeyMatchesSAN(*x509Cert); err != nil {
		return err
	}

	return nil
}

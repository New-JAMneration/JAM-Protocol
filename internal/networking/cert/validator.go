package cert

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

const AlternativeNameLength = 53

// The certificate uses Ed25519 as the signature algorithm.
func ValidateX509SignatureAlgorithm(cert x509.Certificate) error {
	if cert.SignatureAlgorithm != x509.PureEd25519 {
		return errors.New("the signature algorithm is not Ed25519")
	}

	return nil
}

// The SAN must contain exactly one Ed25519 alternative name and no IP SANs.
func ValidateX509DNSNames(cert x509.Certificate) error {
	dnsNames := cert.DNSNames
	if len(dnsNames) != 1 {
		return fmt.Errorf("expected exactly one DNS alternative name, got %d: %v", len(dnsNames), dnsNames)
	}
	if len(cert.IPAddresses) != 0 {
		return fmt.Errorf("unexpected IP alternative names: %v", cert.IPAddresses)
	}

	dnsName := dnsNames[0]
	if len(dnsName) != AlternativeNameLength || dnsName[0] != 'e' {
		return fmt.Errorf("invalid DNS name format: got %q (length=%d), expected 53 chars starting with 'e'", dnsName, len(dnsName))
	}

	return nil
}

// The certificate's public key matches the information encoded in the SAN.
func ValidateX509PubKeyMatchesSAN(cert x509.Certificate) error {
	pk := cert.PublicKey
	pkEd25519, ok := pk.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("invalid public key type: %T", pk)
	}

	expectedEncodedPubKey := AlternativeName(pkEd25519)

	dnsNames := cert.DNSNames
	if len(dnsNames) != 1 {
		return fmt.Errorf("expected exactly one DNS alternative name, got %d: %v", len(dnsNames), dnsNames)
	}

	if dnsNames[0] != expectedEncodedPubKey {
		return fmt.Errorf("DNS SAN does not match public key: got %q, expected %q", dnsNames[0], expectedEncodedPubKey)
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

// Validate the X509 certificate.
func ValidateX509Certificate(x509Cert *x509.Certificate) error {
	if x509Cert == nil {
		return errors.New("the x509 certificate is nil")
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

// Validate the TLS certificate.
func ValidateTlsCertificate(cert tls.Certificate) error {
	x509Cert, err := ParseCertificateFromPem(cert.Certificate[0])
	if err != nil {
		return err
	}

	if err := ValidateX509Certificate(x509Cert); err != nil {
		return err
	}

	return nil
}

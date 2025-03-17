package cert

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

func ValidateTlsCertificateIsNotEmpty(cert tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return errors.New("the certificate is empty")
	}

	return nil
}

func ValidateTlsSignatureAlgorithm(cert tls.Certificate) error {
	x509Cert := cert.Leaf

	if x509Cert == nil {
		return errors.New("the x509 certificate is nil")
	}

	return ValidateX509SignatureAlgorithm(*x509Cert)
}

// The certificate uses Ed25519 as the signature algorithm.
func ValidateX509SignatureAlgorithm(cert x509.Certificate) error {
	if cert.SignatureAlgorithm != x509.PureEd25519 {
		return errors.New("the signature algorithm is not Ed25519")
	}

	return nil
}

func ValidateTlsDNSNames(cert tls.Certificate) error {
	x509Cert := cert.Leaf

	if x509Cert == nil {
		return errors.New("the x509 certificate is nil")
	}

	return ValidateX509DNSNames(*x509Cert)
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

func ValidateTlsPubKeyMatchesSAN(cert tls.Certificate) error {
	x509Cert := cert.Leaf

	if x509Cert == nil {
		return fmt.Errorf("the x509 certificate is nil")
	}

	return ValidateX509PubKeyMatchesSAN(*x509Cert)
}

// Validate the TLS certificate.
func ValidateTlsCertificate(cert tls.Certificate) error {
	if err := ValidateTlsCertificateIsNotEmpty(cert); err != nil {
		return err
	}

	if err := ValidateTlsSignatureAlgorithm(cert); err != nil {
		return err
	}

	if err := ValidateTlsDNSNames(cert); err != nil {
		return err
	}

	if err := ValidateTlsPubKeyMatchesSAN(cert); err != nil {
		return err
	}

	return nil
}

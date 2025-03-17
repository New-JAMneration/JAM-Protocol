package network

import (
	"crypto/ed25519"
	"testing"
)

func TestValidateX509SignatureAlgorithm(t *testing.T) {
	type PkSk struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}

	pksk := PkSk{
		pk: strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		sk: strToHex("0x00000000000000000000000000000000000000000000000000000000000000003b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
	}

	tlsCert, err := GenerateSelfSignedCertificate(pksk.sk, pksk.pk)
	if err != nil {
		t.Errorf("GenerateSelfSignedCertificate() error = %v", err)
		return
	}

	err = ValidateX509SignatureAlgorithm(*tlsCert.Leaf)
	if err != nil {
		t.Errorf("ValidateX509SignatureAlgorithm() error = %v", err)
		return
	}
}

func TestValidateX509DNSNames(t *testing.T) {
	type PkSk struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}

	pksk := PkSk{
		pk: strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		sk: strToHex("0x00000000000000000000000000000000000000000000000000000000000000003b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
	}

	tlsCert, err := GenerateSelfSignedCertificate(pksk.sk, pksk.pk)
	if err != nil {
		t.Errorf("GenerateSelfSignedCertificate() error = %v", err)
		return
	}

	err = ValidateX509DNSNames(*tlsCert.Leaf)
	if err != nil {
		t.Errorf("ValidateX509DNSNames() error = %v", err)
		return
	}
}

func TestValidateX509PubKeyMatchesSAN(t *testing.T) {
	type PkSk struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}

	pksk := PkSk{
		pk: strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		sk: strToHex("0x00000000000000000000000000000000000000000000000000000000000000003b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
	}

	tlsCert, err := GenerateSelfSignedCertificate(pksk.sk, pksk.pk)
	if err != nil {
		t.Errorf("GenerateSelfSignedCertificate() error = %v", err)
		return
	}

	err = ValidateX509PubKeyMatchesSAN(*tlsCert.Leaf)
	if err != nil {
		t.Errorf("ValidateX509PubKeyMatchesSAN() error = %v", err)
		return
	}
}

// TODO: Test something that fails

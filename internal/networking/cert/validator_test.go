package cert

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func testTLSCert(t *testing.T) *x509.Certificate {
	t.Helper()
	pksk := struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}{
		pk: strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		sk: strToHex("0x00000000000000000000000000000000000000000000000000000000000000003b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
	}
	tlsCert, err := SelfSignedCertGen(pksk.sk, pksk.pk)
	require.NoError(t, err)
	require.NotNil(t, tlsCert.Leaf)
	return tlsCert.Leaf
}

func TestValidateX509SignatureAlgorithm(t *testing.T) {
	require.NoError(t, ValidateX509SignatureAlgorithm(*testTLSCert(t)))
}

func TestValidateX509DNSNames(t *testing.T) {
	require.NoError(t, ValidateX509DNSNames(*testTLSCert(t)))
}

func TestValidateX509PubKeyMatchesSAN(t *testing.T) {
	require.NoError(t, ValidateX509PubKeyMatchesSAN(*testTLSCert(t)))
}

func TestValidateX509CertificateRejectsEmptySAN(t *testing.T) {
	cert := *testTLSCert(t)
	cert.DNSNames = nil
	require.Error(t, ValidateX509Certificate(&cert))
}

func TestValidateX509CertificateRejectsMismatchedSAN(t *testing.T) {
	cert := *testTLSCert(t)
	cert.DNSNames = []string{"eaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	require.Error(t, ValidateX509Certificate(&cert))
}

func TestValidateX509CertificateRejectsExtraDNSNames(t *testing.T) {
	cert := *testTLSCert(t)
	cert.DNSNames = append(cert.DNSNames, cert.DNSNames[0])
	require.Error(t, ValidateX509Certificate(&cert))
}

func TestValidateX509CertificateRejectsIPSANs(t *testing.T) {
	cert := *testTLSCert(t)
	cert.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	require.Error(t, ValidateX509Certificate(&cert))
}

func TestValidateX509CertificateRejectsNonEd25519PublicKey(t *testing.T) {
	cert := *testTLSCert(t)
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	cert.PublicKey = &rsaKey.PublicKey
	require.Error(t, ValidateX509Certificate(&cert))
}

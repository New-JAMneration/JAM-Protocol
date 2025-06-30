package cert

import (
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func strToHex(str string) []byte {
	hexStr, err := hex.DecodeString(str[2:]) // 0x prefix
	if err != nil {
		panic(err)
	}
	return hexStr
}

func TestGenerateEd25519PrivateKey(t *testing.T) {
	tests := []struct {
		name     string
		seed     []byte
		wantPriv ed25519.PrivateKey
		wantPub  ed25519.PublicKey
		wantErr  bool
	}{
		{
			name:     "Alice",
			seed:     strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
			wantPriv: strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
			wantPub:  strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		}, {
			name:     "Bob",
			seed:     strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"),
			wantPriv: strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"),
			wantPub:  strToHex("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862"),
		}, {
			name:     "Carol",
			seed:     strToHex("0x0200000002000000020000000200000002000000020000000200000002000000"),
			wantPriv: strToHex("0x0200000002000000020000000200000002000000020000000200000002000000"),
			wantPub:  strToHex("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00"),
		}, {
			name:     "David",
			seed:     strToHex("0x0300000003000000030000000300000003000000030000000300000003000000"),
			wantPriv: strToHex("0x0300000003000000030000000300000003000000030000000300000003000000"),
			wantPub:  strToHex("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9"),
		}, {
			name:     "Eve",
			seed:     strToHex("0x0400000004000000040000000400000004000000040000000400000004000000"),
			wantPriv: strToHex("0x0400000004000000040000000400000004000000040000000400000004000000"),
			wantPub:  strToHex("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff"),
		}, {
			name:     "Fergie",
			seed:     strToHex("0x0500000005000000050000000500000005000000050000000500000005000000"),
			wantPriv: strToHex("0x0500000005000000050000000500000005000000050000000500000005000000"),
			wantPub:  strToHex("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPriv, gotPub, err := Ed25519KeyGen(tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEd25519Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPriv[:32], tt.wantPriv) {
				t.Errorf("gotPriv = %v, \nwant %v", gotPriv[:32], tt.wantPriv)
			}
			if !reflect.DeepEqual(gotPub, tt.wantPub) {
				t.Errorf("gotPub = %v, \nwant %v", gotPub, tt.wantPub)
			}
		})
	}
}

func TestAlternativeName(t *testing.T) {
	type args struct {
		data []byte
	}
	// test cases are given in https://docs.jamcha.in/basics/dev-accounts
	// ed25519_public, dns_alt_name
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "alice",
			args: args{
				data: strToHex("0x4418fb8c85bb3985394a8c2756d3643457ce614546202a2f50b093d762499ace"),
			},
			want: "eecgwpgwq3noky4ijm4jmvjtmuzv44qvigciusxakq5epnrfj2utb",
		},
		{
			name: "bob",
			args: args{
				data: strToHex("0xad93247bd01307550ec7acd757ce6fb805fcf73db364063265b30a949e90d933"),
			},
			want: "en5ejs5b2tybkfh4ym5vpfh7nynby73xhtfzmazumtvcijpcsz6ma",
		},
		{
			name: "carol",
			args: args{
				data: strToHex("0xcab2b9ff25c2410fbe9b8a717abb298c716a03983c98ceb4def2087500b8e341"),
			},
			want: "ekwmt37xecoq6a7otkm4ux5gfmm4uwbat4bg5m223shckhaaxdpqa",
		},
		{
			name: "david",
			args: args{
				data: strToHex("0xf30aa5444688b3cab47697b37d5cac5707bb3289e986b19b17db437206931a8d"),
			},
			want: "etxckkczii4mvm22ox4m3horvx2bwlzerjxbd3n6c36qehdms2idb",
		},
		{
			name: "eve",
			args: args{
				data: strToHex("0x8b8c5d436f92ecf605421e873a99ec528761eb52a88a2f9a057b3b3003e6f32a"),
			},
			want: "eled3vb5nse3n7cii6ybvtms5s2bdwvlkivc7cnwa33oatby4txka",
		},
		{
			name: "fergie",
			args: args{
				data: strToHex("0xab0084d01534b31c1dd87c81645fd762482a90027754041ca1b56133d0466c06"),
			},
			want: "elfaiiixcuzmzroa34lajwp52cdsucikaxdviaoeuvnygdi3imtba",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlternativeName(tt.args.data); got != tt.want {
				t.Errorf("EncodeBase32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelfSignedCertGen(t *testing.T) {
	// Generate keys for testing
	privKey1, pubKey1, err := Ed25519KeyGen(strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"))
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	privKey2, _, err := Ed25519KeyGen(strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"))
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	type args struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		check   func(t *testing.T, cert tls.Certificate)
	}{
		{
			name: "Valid case - with both private and public keys",
			args: args{
				sk: privKey1,
				pk: pubKey1,
			},
			wantErr: false,
			check: func(t *testing.T, cert tls.Certificate) {
				if len(cert.Certificate) == 0 {
					t.Error("Expected certificate chain to have at least one certificate")
				}
				if cert.PrivateKey == nil {
					t.Error("Expected private key to not be nil")
				}
			},
		},
		{
			name: "Valid case - with only private key",
			args: args{
				sk: privKey2,
				pk: nil, // Public key will be derived from private key
			},
			wantErr: false,
			check: func(t *testing.T, cert tls.Certificate) {
				if len(cert.Certificate) == 0 {
					t.Error("Expected certificate chain to have at least one certificate")
				}
				if cert.PrivateKey == nil {
					t.Error("Expected private key to not be nil")
				}
			},
		},
		{
			name: "Error case - nil private key",
			args: args{
				sk: nil,
				pk: pubKey1,
			},
			wantErr: true,
			check: func(t *testing.T, cert tls.Certificate) {
				// No need to check certificate properties for error case
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelfSignedCertGen(tt.args.sk, tt.args.pk)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelfSignedCertGen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.check(t, got)

				// Validate the generated certificate
				err = ValidateTlsCertificate(got)
				if err != nil {
					t.Errorf("Generated certificate failed validation: %v", err)
				}
			}
		})
	}
}

func TestTLSConfigGen(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer os.Unsetenv("USE_MINI_REDIS") // Cleanup after test
	defer store.CloseMiniRedis()

	tests := []struct {
		name      string
		seed      []byte
		isServer  bool
		isBuilder bool
		wantErr   bool
		checkFunc func(t *testing.T, config *tls.Config)
	}{
		{
			name:      "Client configuration",
			seed:      strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
			isServer:  false,
			isBuilder: false,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				if config == nil {
					t.Error("Expected non-nil TLS config")
					return
				}

				// Client config checks
				if len(config.Certificates) != 1 {
					t.Error("Expected exactly one certificate")
				}

				if config.ClientAuth != 0 {
					t.Error("Expected no client auth for client config")
				}

				if config.VerifyConnection != nil {
					t.Error("VerifyConnection should be nil for client config")
				}

				if config.MinVersion != tls.VersionTLS13 || config.MaxVersion != tls.VersionTLS13 {
					t.Error("Expected TLS 1.3 version")
				}

				if len(config.CurvePreferences) != 2 {
					t.Error("Expected 2 curve preferences")
				}

				if len(config.NextProtos) == 0 {
					t.Error("Expected at least one ALPN protocol")
				}

				if config.VerifyPeerCertificate == nil {
					t.Error("Expected VerifyPeerCertificate to be set")
				}
			},
		},
		{
			name:      "Server configuration",
			seed:      strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"),
			isServer:  true,
			isBuilder: false,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				if config == nil {
					t.Error("Expected non-nil TLS config")
					return
				}

				// Server config checks
				if len(config.Certificates) != 1 {
					t.Error("Expected exactly one certificate")
				}

				if config.ClientAuth != tls.RequireAnyClientCert {
					t.Error("Expected RequireAnyClientCert for server config")
				}

				if config.VerifyConnection == nil {
					t.Error("VerifyConnection should not be nil for server config")
				}

				if config.MinVersion != tls.VersionTLS13 || config.MaxVersion != tls.VersionTLS13 {
					t.Error("Expected TLS 1.3 version")
				}

				if len(config.CurvePreferences) != 2 {
					t.Error("Expected 2 curve preferences")
				}

				if len(config.NextProtos) == 0 {
					t.Error("Expected at least one ALPN protocol")
				}

				// Check no builder-specific protocol in NextProtos
				if strings.HasSuffix(config.NextProtos[0], "/builder") {
					t.Error("Found builder-specific protocol when isBuilder is false")
				}
			},
		},
		{
			name:      "Builder server configuration",
			seed:      strToHex("0x0200000002000000020000000200000002000000020000000200000002000000"),
			isServer:  true,
			isBuilder: true,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				if config == nil {
					t.Error("Expected non-nil TLS config")
					return
				}

				// Builder server config checks
				if len(config.Certificates) != 1 {
					t.Error("Expected exactly one certificate")
				}

				if config.ClientAuth != tls.RequireAnyClientCert {
					t.Error("Expected RequireAnyClientCert for server config")
				}

				if config.VerifyConnection == nil {
					t.Error("VerifyConnection should not be nil for server config")
				}

				// Check for builder-specific protocol in NextProtos
				builderProtoFound := false
				for _, proto := range config.NextProtos {
					if strings.HasSuffix(proto, "/builder") {
						builderProtoFound = true
						break
					}
				}
				if !builderProtoFound {
					t.Error("Expected to find builder-specific protocol")
				}
			},
		},
		{
			name:      "Invalid seed",
			seed:      nil,
			isServer:  false,
			isBuilder: false,
			wantErr:   true,
			checkFunc: nil, // No need to check config for error case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TLSConfigGen(tt.seed, tt.isServer, tt.isBuilder)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("TLSConfigGen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Skip additional checks if we expected an error
			if tt.wantErr {
				return
			}

			// Run specific checks for this test case
			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

package cert

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/require"
)

func strToHex(str string) []byte {
	hexStr, err := hex.DecodeString(str[2:]) // 0x prefix
	if err != nil {
		panic(err)
	}
	return hexStr
}

// setupTestGenesis initializes a genesis block in the blockchain instance for tests
// that require ALPN generation. Returns cleanup function and computed genesis hash prefix.
func setupTestGenesis(t *testing.T) (cleanup func(), hashPrefix string) {
	t.Helper()
	blockchain.ResetInstance()

	cs := blockchain.GetInstance()
	genesis := types.Block{
		Header: types.Header{
			Slot: 0,
		},
		Extrinsic: types.Extrinsic{},
	}
	require.NoError(t, cs.GenerateGenesisBlock(genesis))

	genesisHash, err := hash.ComputeBlockHeaderHash(genesis.Header)
	require.NoError(t, err)
	prefix := hex.EncodeToString(genesisHash[:4])

	return func() { blockchain.ResetInstance() }, prefix
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantPriv, gotPriv[:32])
			require.Equal(t, tt.wantPub, gotPub)
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
			require.Equal(t, tt.want, AlternativeName(tt.args.data))
		})
	}
}

func TestSelfSignedCertGen(t *testing.T) {
	// Generate keys for testing
	privKey1, pubKey1, err := Ed25519KeyGen(strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"))
	require.NoError(t, err)

	privKey2, _, err := Ed25519KeyGen(strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"))
	require.NoError(t, err)

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
				require.NotEmpty(t, cert.Certificate)
				require.NotNil(t, cert.PrivateKey)
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
				require.NotEmpty(t, cert.Certificate)
				require.NotNil(t, cert.PrivateKey)
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			tt.check(t, got)
			require.NoError(t, ValidateTlsCertificate(got))
		})
	}
}

func TestTLSConfigGen(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer os.Unsetenv("USE_MINI_REDIS") // Cleanup after test
	defer blockchain.CloseMiniRedis()

	cleanup, _ := setupTestGenesis(t)
	defer cleanup()

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
				require.NotNil(t, config)

				// Client config checks
				require.Len(t, config.Certificates, 1)
				require.Equal(t, tls.ClientAuthType(0), config.ClientAuth)
				require.Nil(t, config.VerifyConnection)
				require.Equal(t, uint16(tls.VersionTLS13), config.MinVersion)
				require.Equal(t, uint16(tls.VersionTLS13), config.MaxVersion)
				require.Len(t, config.CurvePreferences, 2)
				require.NotEmpty(t, config.NextProtos)
				require.NotNil(t, config.VerifyPeerCertificate)
			},
		},
		{
			name:      "Server configuration",
			seed:      strToHex("0x0100000001000000010000000100000001000000010000000100000001000000"),
			isServer:  true,
			isBuilder: false,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				require.NotNil(t, config)

				// Server config checks
				require.Len(t, config.Certificates, 1)
				require.Equal(t, tls.RequireAnyClientCert, config.ClientAuth)
				require.NotNil(t, config.VerifyConnection)
				require.Equal(t, uint16(tls.VersionTLS13), config.MinVersion)
				require.Equal(t, uint16(tls.VersionTLS13), config.MaxVersion)
				require.Len(t, config.CurvePreferences, 2)
				require.NotEmpty(t, config.NextProtos)
			},
		},
		{
			name:      "Client is Builder",
			seed:      strToHex("0x0200000002000000020000000200000002000000020000000200000002000000"),
			isServer:  false,
			isBuilder: true,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				require.NotNil(t, config)
				require.Len(t, config.Certificates, 1)
				require.Equal(t, tls.ClientAuthType(0), config.ClientAuth)
				require.Nil(t, config.VerifyConnection)

				builderProtoFound := false
				for _, proto := range config.NextProtos {
					if strings.HasSuffix(proto, "/builder") {
						builderProtoFound = true
						break
					}
				}
				require.True(t, builderProtoFound)
			},
		},
		{
			name:      "Builder server configuration",
			seed:      strToHex("0x0200000002000000020000000200000002000000020000000200000002000000"),
			isServer:  true,
			isBuilder: true,
			wantErr:   false,
			checkFunc: func(t *testing.T, config *tls.Config) {
				require.NotNil(t, config)

				// Builder server config checks
				require.Len(t, config.Certificates, 1)
				require.Equal(t, tls.RequireAnyClientCert, config.ClientAuth)
				require.NotNil(t, config.VerifyConnection)

				// Check for builder-specific protocol in NextProtos
				builderProtoFound := false
				for _, proto := range config.NextProtos {
					if strings.HasSuffix(proto, "/builder") {
						builderProtoFound = true
						break
					}
				}
				require.True(t, builderProtoFound)
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

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestTLSConfigFromPrivateKey_leafPublicKeyMatches(t *testing.T) {
	cleanup, _ := setupTestGenesis(t)
	defer cleanup()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wantPub := priv.Public().(ed25519.PublicKey)

	for _, isServer := range []bool{true, false} {
		for _, isBuilder := range []bool{false, true} {
			cfg, err := TLSConfigFromPrivateKey(priv, isServer, isBuilder)
			require.NoError(t, err, "TLSConfigFromPrivateKey(server=%v builder=%v)", isServer, isBuilder)
			require.Len(t, cfg.Certificates, 1)

			leaf, err := x509.ParseCertificate(cfg.Certificates[0].Certificate[0])
			require.NoError(t, err)

			gotPub, ok := leaf.PublicKey.(ed25519.PublicKey)
			require.True(t, ok)
			require.True(t, wantPub.Equal(gotPub), "leaf public key mismatch (server=%v builder=%v)", isServer, isBuilder)
		}
	}
}

func TestALPNFromGenesisHashPrefix(t *testing.T) {
	base := alpnFromGenesisHashPrefix("11223344", false)
	require.Equal(t, []string{"jamnp-s/0/11223344"}, base)

	builder := alpnFromGenesisHashPrefix("11223344", true)
	require.Equal(t, []string{"jamnp-s/0/11223344/builder"}, builder)
}

func TestALPNGenUsesGenesisHeaderHashPrefix(t *testing.T) {
	cleanup, prefix := setupTestGenesis(t)
	defer cleanup()

	base, err := ALPNGen(false)
	require.NoError(t, err)
	expectedBase := fmt.Sprintf("jamnp-s/0/%s", prefix)
	require.Equal(t, []string{expectedBase}, base)
	require.True(t, strings.HasPrefix(base[0], "jamnp-s/0/"), "ALPN should start with jamnp-s/0/")
	require.Len(t, prefix, 8, "genesis hash prefix should be 8 hex chars (4 bytes)")

	builder, err := ALPNGen(true)
	require.NoError(t, err)
	expectedBuilder := fmt.Sprintf("jamnp-s/0/%s/builder", prefix)
	require.Equal(t, []string{expectedBuilder}, builder)
	require.True(t, strings.HasSuffix(builder[0], "/builder"), "builder ALPN should end with /builder")
}

func TestMutualAuthenticatedQUICWithFixedEd25519Keys(t *testing.T) {
	cleanup, hashPrefix := setupTestGenesis(t)
	defer cleanup()

	_, serverPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, clientPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	serverTLS, err := TLSConfigFromPrivateKey(serverPriv, true, false)
	require.NoError(t, err)

	clientTLS, err := TLSConfigFromPrivateKey(clientPriv, false, false)
	require.NoError(t, err)

	quicCfg := &quic.Config{
		HandshakeIdleTimeout: 5 * time.Second,
		MaxIdleTimeout:       5 * time.Second,
	}
	ln, err := quic.ListenAddr("127.0.0.1:0", serverTLS, quicCfg)
	require.NoError(t, err)
	defer ln.Close()

	expectedALPN := fmt.Sprintf("jamnp-s/0/%s", hashPrefix)

	serverDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err := ln.Accept(ctx)
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.CloseWithError(0, "done")

		state := conn.ConnectionState().TLS
		if len(state.PeerCertificates) == 0 {
			serverDone <- fmt.Errorf("server saw no peer certificates")
			return
		}
		if err := ValidateX509Certificate(state.PeerCertificates[0]); err != nil {
			serverDone <- err
			return
		}
		if state.NegotiatedProtocol == "" {
			serverDone <- fmt.Errorf("server negotiated protocol is empty")
			return
		}
		if state.NegotiatedProtocol != expectedALPN {
			serverDone <- fmt.Errorf("server ALPN mismatch: got %q, expected %q", state.NegotiatedProtocol, expectedALPN)
			return
		}
		serverDone <- nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := quic.DialAddr(ctx, ln.Addr().String(), clientTLS, quicCfg)
	require.NoError(t, err)
	defer conn.CloseWithError(0, "done")

	cstate := conn.ConnectionState().TLS
	require.NotEmpty(t, cstate.PeerCertificates)
	require.NoError(t, ValidateX509Certificate(cstate.PeerCertificates[0]))
	require.Equal(t, expectedALPN, cstate.NegotiatedProtocol, "client ALPN should match expected format jamnp-s/0/HASH_PREFIX")
	require.True(t, strings.HasPrefix(cstate.NegotiatedProtocol, "jamnp-s/0/"), "ALPN should start with jamnp-s/0/")

	require.NoError(t, <-serverDone)
}

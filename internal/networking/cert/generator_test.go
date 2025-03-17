package cert

import (
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"reflect"
	"testing"
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

func TestEncodeBase32(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				data: []byte("test"),
			},
			want: "orsxg5a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBase32(tt.args.data); got != tt.want {
				t.Errorf("EncodeBase32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSelfSignedCertificate(t *testing.T) {
	type args struct {
		sk ed25519.PrivateKey
		pk ed25519.PublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    tls.Certificate
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelfSignedCertGen(tt.args.sk, tt.args.pk)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelfSignedCertGen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelfSignedCertGen() = %v, want %v", got, tt.want)
			}
		})
	}
}

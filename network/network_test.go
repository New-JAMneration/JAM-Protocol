package network

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
			name:     "alice",
			seed:     strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
			wantPriv: strToHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
			wantPub:  strToHex("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPriv, gotPub, err := GenerateEd25519Key(tt.seed)
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
		// TODO: Add test cases.
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
			got, err := GenerateSelfSignedCertificate(tt.args.sk, tt.args.pk)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSelfSignedCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSelfSignedCertificate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateALPN(t *testing.T) {
	type args struct {
		cert      tls.Certificate
		isBuilder bool
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateALPN(tt.args.cert, tt.args.isBuilder)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateALPN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateALPN() = %v, want %v", got, tt.want)
			}
		})
	}
}

package keystore

import (
	"encoding/hex"
	"testing"
)

func TestTrivialSeed(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{
			name:     "trivial_seed(0)",
			input:    0,
			expected: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "trivial_seed(1)",
			input:    1,
			expected: "0100000001000000010000000100000001000000010000000100000001000000",
		},
		{
			name:     "trivial_seed(2)",
			input:    2,
			expected: "0200000002000000020000000200000002000000020000000200000002000000",
		},
		{
			name:     "trivial_seed(3)",
			input:    3,
			expected: "0300000003000000030000000300000003000000030000000300000003000000",
		},
		{
			name:     "trivial_seed(4)",
			input:    4,
			expected: "0400000004000000040000000400000004000000040000000400000004000000",
		},
		{
			name:     "trivial_seed(5)",
			input:    5,
			expected: "0500000005000000050000000500000005000000050000000500000005000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := TrivialSeed(tt.input)
			got := hex.EncodeToString(seed[:])
			if got != tt.expected {
				t.Errorf("TrivialSeed(%d) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDeriveValidatorKeys(t *testing.T) {
	tests := []struct {
		name                           string
		seed                           string
		expectedEd25519SecretSeed      string
		expectedEd25519Public          string
		expectedBandersnatchSecretSeed string
		expectedBandersnatchPublic     string
	}{
		{
			name:                           "seed = trivial_seed(0)",
			seed:                           "0000000000000000000000000000000000000000000000000000000000000000",
			expectedEd25519SecretSeed:      "996542becdf1e78278dc795679c825faca2e9ed2bf101bf3c4a236d3ed79cf59",
			expectedEd25519Public:          "4418fb8c85bb3985394a8c2756d3643457ce614546202a2f50b093d762499ace",
			expectedBandersnatchSecretSeed: "007596986419e027e65499cc87027a236bf4a78b5e8bd7f675759d73e7a9c799",
			expectedBandersnatchPublic:     "ff71c6c03ff88adb5ed52c9681de1629a54e702fc14729f6b50d2f0a76f185b3",
		},
		{
			name:                           "seed = trivial_seed(1)",
			seed:                           "0100000001000000010000000100000001000000010000000100000001000000",
			expectedEd25519SecretSeed:      "b81e308145d97464d2bc92d35d227a9e62241a16451af6da5053e309be4f91d7",
			expectedEd25519Public:          "ad93247bd01307550ec7acd757ce6fb805fcf73db364063265b30a949e90d933",
			expectedBandersnatchSecretSeed: "12ca375c9242101c99ad5fafe8997411f112ae10e0e5b7c4589e107c433700ac",
			expectedBandersnatchPublic:     "dee6d555b82024f1ccf8a1e37e60fa60fd40b1958c4bb3006af78647950e1b91",
		},
		{
			name:                           "seed = trivial_seed(2)",
			seed:                           "0200000002000000020000000200000002000000020000000200000002000000",
			expectedEd25519SecretSeed:      "0093c8c10a88ebbc99b35b72897a26d259313ee9bad97436a437d2e43aaafa0f",
			expectedEd25519Public:          "cab2b9ff25c2410fbe9b8a717abb298c716a03983c98ceb4def2087500b8e341",
			expectedBandersnatchSecretSeed: "3d71dc0ffd02d90524fda3e4a220e7ec514a258c59457d3077ce4d4f003fd98a",
			expectedBandersnatchPublic:     "9326edb21e5541717fde24ec085000b28709847b8aab1ac51f84e94b37ca1b66",
		},
		{
			name:                           "seed = trivial_seed(3)",
			seed:                           "0300000003000000030000000300000003000000030000000300000003000000",
			expectedEd25519SecretSeed:      "69b3a7031787e12bfbdcac1b7a737b3e5a9f9450c37e215f6d3b57730e21001a",
			expectedEd25519Public:          "f30aa5444688b3cab47697b37d5cac5707bb3289e986b19b17db437206931a8d",
			expectedBandersnatchSecretSeed: "107a9148b39a1099eeaee13ac0e3c6b9c256258b51c967747af0f8749398a276",
			expectedBandersnatchPublic:     "0746846d17469fb2f95ef365efcab9f4e22fa1feb53111c995376be8019981cc",
		},
		{
			name:                           "seed = trivial_seed(4)",
			seed:                           "0400000004000000040000000400000004000000040000000400000004000000",
			expectedEd25519SecretSeed:      "b4de9ebf8db5428930baa5a98d26679ab2a03eae7c791d582e6b75b7f018d0d4",
			expectedEd25519Public:          "8b8c5d436f92ecf605421e873a99ec528761eb52a88a2f9a057b3b3003e6f32a",
			expectedBandersnatchSecretSeed: "0bb36f5ba8e3ba602781bb714e67182410440ce18aa800c4cb4dd22525b70409",
			expectedBandersnatchPublic:     "151e5c8fe2b9d8a606966a79edd2f9e5db47e83947ce368ccba53bf6ba20a40b",
		},
		{
			name:                           "seed = trivial_seed(5)",
			seed:                           "0500000005000000050000000500000005000000050000000500000005000000",
			expectedEd25519SecretSeed:      "4a6482f8f479e3ba2b845f8cef284f4b3208ba3241ed82caa1b5ce9fc6281730",
			expectedEd25519Public:          "ab0084d01534b31c1dd87c81645fd762482a90027754041ca1b56133d0466c06",
			expectedBandersnatchSecretSeed: "75e73b8364bf4753c5802021c6aa6548cddb63fe668e3cacf7b48cdb6824bb09",
			expectedBandersnatchPublic:     "2105650944fcd101621fd5bb3124c9fd191d114b7ad936c1d79d734f9f21392e",
		},
		{
			name:                           "seed = f92d680ea3f0ac06307795490d8a03c5c0d4572b5e0a8cffec87e1294855d9d1",
			seed:                           "f92d680ea3f0ac06307795490d8a03c5c0d4572b5e0a8cffec87e1294855d9d1",
			expectedEd25519SecretSeed:      "f21e2d96a51387f9a7e5b90203654913dde7fa1044e3eba5631ed19f327d6126",
			expectedEd25519Public:          "11a695f674de95ff3daaff9a5b88c18448b10156bf88bc04200e48d5155c7243",
			expectedBandersnatchSecretSeed: "06154d857537a9b622a9a94b1aeee7d588db912bfc914a8a9707148bfba3b9d1",
			expectedBandersnatchPublic:     "299bdfd8d615aadd9e6c58718f9893a5144d60e897bc9da1f3d73c935715c650",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedBytes, err := hex.DecodeString(tt.seed)
			if err != nil {
				t.Fatalf("Failed to decode seed: %v", err)
			}

			ed25519SecretSeed, bandersnatchSecretSeed, ed25519Public, bandersnatchPublic, err := DeriveValidatorKeys(seedBytes)
			if err != nil {
				t.Fatalf("DeriveValidatorKeys failed: %v", err)
			}

			// Check Ed25519 secret seed
			gotEd25519SecretSeed := hex.EncodeToString(ed25519SecretSeed)
			if gotEd25519SecretSeed != tt.expectedEd25519SecretSeed {
				t.Errorf("Ed25519 secret seed = %s, want %s", gotEd25519SecretSeed, tt.expectedEd25519SecretSeed)
			}

			// Check Ed25519 public key
			gotEd25519Public := hex.EncodeToString(ed25519Public[:])
			if gotEd25519Public != tt.expectedEd25519Public {
				t.Errorf("Ed25519 public key = %s, want %s", gotEd25519Public, tt.expectedEd25519Public)
			}

			// Check Bandersnatch secret seed
			gotBandersnatchSecretSeed := hex.EncodeToString(bandersnatchSecretSeed)
			if gotBandersnatchSecretSeed != tt.expectedBandersnatchSecretSeed {
				t.Errorf("Bandersnatch secret seed = %s, want %s", gotBandersnatchSecretSeed, tt.expectedBandersnatchSecretSeed)
			}

			// Check Bandersnatch public key
			gotBandersnatchPublic := hex.EncodeToString(bandersnatchPublic[:])
			if gotBandersnatchPublic != tt.expectedBandersnatchPublic {
				t.Errorf("Bandersnatch public key = %s, want %s", gotBandersnatchPublic, tt.expectedBandersnatchPublic)
			}
		})
	}
}

package safrole

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

/*
	func CalculateNewEntropy(eta types.EntropyBuffer, public_key types.BandersnatchPublic, entropy_source types.BandersnatchVrfSignature) types.Entropy {
		handler, _ := CreateVRFHandler(public_key, 0)
		vrfOutput, _ := handler.VRFOutput(entropy_source[:])
		hash_input := append(eta[0][:], vrfOutput...)
		return types.Entropy(hash.Blake2bHash(hash_input))
	}
*/

func hexToByteArray32(hexString string) types.ByteArray32 {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		return types.ByteArray32{}
	}

	// Check if the decoded byte slice has the correct length
	if len(bytes) != 32 {
		return types.ByteArray32{}
	}

	var result types.ByteArray32
	copy(result[:], bytes[:])

	return result
}

func TestCalculateNewEntropy(t *testing.T) {
	eta := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa")),
		types.Entropy(hexToByteArray32("0x491db955568b306018eee09f2966f873bb665a20b703584ad868454b81b17e76")),
		types.Entropy(hexToByteArray32("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
	}

	publicKey := types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc"))
	entropySource := types.BandersnatchVrfSignature(hex2Bytes("0x4cc4f71f2a4e3404632ba128c1de4a56e961d50d408d3bba17e31f99fb082a8a998e6e03d797bed2fd1e0c50f777c0df5d6c9e19720ea85d7ad106bbb9a7141ad530fa180d36b0e6fc4fdae845113377e12a0117943c91ac4751eeb4eb66e416")) // next epoch first block header

	expectedEntropy := types.Entropy(hexToByteArray32("0x8c89c318db72cd8b6edf1cd42e2e5501afc76d7fc4d84d7af95d84f64a96daf0"))
	// fmt.Println((len(expectedEntropy)))
	actualEntropy := CalculateNewEntropy(publicKey, entropySource, eta)
	if !bytes.Equal(actualEntropy[:], expectedEntropy[:]) {
		t.Errorf("CalculateHeaderEntropy() = %v, want %v", actualEntropy, expectedEntropy)
	}
}

func TestCalculateHeaderEntropy(t *testing.T) {
	publicKey := types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc"))
	seal := types.BandersnatchVrfSignature(hex2Bytes("0xa70cd5b81c52bb33540d75a2c68ba482f54758f347d1ef97b6f4f9708199f51c451c97b60fc5a218249ed7c0b2bfb589ac6dbc5776bd10e20bb067257daf410c2ab890dec90b4e0dc96d99058721e6fb0a1543d83a8d3d56d1125d1162fd9208"))

	expectedHeaderEntropy := types.BandersnatchVrfSignature(hex2Bytes("0xa81d92d83542ddfe0f34a81ce306ad3032740a75c14c6e99890690caa647986bf0b815a3ca5543ad24fe67d723a9180266ff65dcc3fc4824f9d688ab1e528301c76c5290d655cc1d1ad079ce696b5a9368241e2f0d16bc5cd413830f5c4c100a"))

	actualHeaderEntropy := CalculateHeaderEntropy(publicKey, seal)

	if !bytes.Equal(actualHeaderEntropy[:], expectedHeaderEntropy[:]) {
		t.Errorf("CalculateHeaderEntrop() = %v, want %v", actualHeaderEntropy, expectedHeaderEntropy)
	}
}

func TestHeaderU(t *testing.T) {
}

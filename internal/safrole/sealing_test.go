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

/*
func TestSeal(t *testing.T) {
	"header": {
        "parent": "0xa6abed6e5112f9a912aa2eae7230038e7df0f2f3f88e9a1558b84f743246920a",
        "parent_state_root": "0x25f8086b89b911077d7eb0c62d04984e65ee01dbddd9d0ebef69779f2aa9338a",
        "extrinsic_hash": "0xe06438f6d29edace8e2f0594aef34ab0da579122a54f104983dd97bc039b5e8e",
        "slot": 5106408,
        "epoch_mark": {
            "entropy": "0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7",
            "tickets_entropy": "0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa",
            "validators": [
                "0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d",
                "0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0",
                "0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc",
                "0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33",
                "0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3",
                "0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d"
            ]
        },
        "tickets_mark": null,
        "offenders_mark": [],
        "author_index": 4,
        "entropy_source": "0x4cc4f71f2a4e3404632ba128c1de4a56e961d50d408d3bba17e31f99fb082a8a998e6e03d797bed2fd1e0c50f777c0df5d6c9e19720ea85d7ad106bbb9a7141ad530fa180d36b0e6fc4fdae845113377e12a0117943c91ac4751eeb4eb66e416",
        "seal": "0x31ccb8f7f7d2822a67398b4f512a4fb56a0d0df51be7b664a28824da34f0aa0d879f9e2daca096b9cb323afd4841944072dbb5f1f1257ca2bbaa8a3ecb42070ac29779c9d9dfe351954d3a46e7644f70ab664e56d7d577216c2fa3497067971c"
    }
}*/
/*
[
  {
    "bandersnatch": "0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d",
    "ed25519": "0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29",
    "bls": "0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e",
    "bandersnatch_priv": "0x51c1537c18eea5c5969cb2ae45c1224cc245de5c5b8e6e25f48fb99f2786ee05",
    "ed25519_priv": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "bls_priv": "0x25f137a62c84a5adc12c8159d678a80b51f81bbe85d41e144c7a4e1edbdc5f44"
  },
  {
    "bandersnatch": "0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0",
    "ed25519": "0x4cb5abf6ad79fbf5abbccafcc269d85cd2651ed4b885b5869f241aedf0a5ba29",
    "bls": "0x8b8a096ada14a51df7e2067007bf6c24d7568d88bf89816c1287ba2784b4188c3536d70b1a1cbc8ab438056e457e2aa0ab48d30d6279373652d19269f7260624d0965c3dc00ed944d1b6ff6db06bb73dc1314164e9fed6020108487897ac3a9814eca841aedc47f504a848513166ffe39f89c9f3e7c6729dc99207f863a10bda142d5a24ba42b90d99d6d6df3fa6d780",
    "bandersnatch_priv": "0x1bf0237a2bdc1e47ffb24c5d08cc60944be0072c2a851118bf1943afd1a5cc05",
    "ed25519_priv": "0x0000000000000000000000000000000000000000000000000000000000000001",
    "bls_priv": "0x53568c52695a29fafb59ae33d2cb35b71ccd64f0dad5d887d53df1efb021306a"
  },
  {
    "bandersnatch": "0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc",
    "ed25519": "0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00",
    "bls": "0x8faee314528448651e50bea6d2e7e5d3176698fea0b932405a4ec0a19775e72325e44a6d28f99fba887e04eb818f13d1b73f75f0161644283df61e7fbaad7713fae0ef79fe92499202834c97f512d744515a57971badf2df62e23697e9fe347f168fed0adb9ace131f49bbd500a324e2469569423f37c5d3b430990204ae17b383fcd582cb864168c8b46be8d779e7ca",
    "bandersnatch_priv": "0x91ebd09c591e41858a7a2a45c671642708f546c163b76eef0991b755017e7412",
    "ed25519_priv": "0x0200000002000000020000000200000002000000020000000200000002000000",
    "bls_priv": "0x78f22d094f700292fa218f05ef42f4bc209c665b77aa2be3e80952139eb1cd54"
  },
  {
    "bandersnatch": "0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33",
    "ed25519": "0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9",
    "bls": "0x8dfdac3e2f604ecda637e4969a139ceb70c534bd5edc4210eb5ac71178c1d62f0c977197a2c6a9e8ada6a14395bc9aa3a384d35f40c3493e20cb7efaa799f66d1cedd5b2928f8e34438b07072bbae404d7dfcee3f457f9103173805ee163ba550854e4660ccec49e25fafdb00e6adfbc8e875de1a9541e1721e956b972ef2b135cc7f71682615e12bb7d6acd353d7681",
    "bandersnatch_priv": "0x40ad858dd0abe3016f7834831c93ae02764e0bb99ee204ffc6777b01c946ac0c",
    "ed25519_priv": "0x0300000003000000030000000300000003000000030000000300000003000000",
    "bls_priv": "0x99b1c8ea82f8d8f8019b5e7c368abce2aa578af6f722050947336c29034a4660"
  },
  {
    "bandersnatch": "0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3",
    "ed25519": "0xfd50b8e3b144ea244fbf7737f550bc8dd0c2650bbc1aada833ca17ff8dbf329b",
    "bls": "0xadd763d2421b4e78c2f48f60427f0c876b1e3d82b2c7e0afb970f7ab67e4f1d501afb777ed572b2372bbc17fea7fea71ae14d35dfed7ab67740de26f7283a4f0c7b29d801647e0222fb6d0803b10c54c6d1de4df460beb8ca6bb88e5210cc1ca0a83b9eab9dcbe83fdf2addc2f85c1af1a07c3d082f99216f4cd56fb5218b26e4dff130a24306ab523eff5d98279e549",
    "bandersnatch_priv": "0x33cc9b6271c4fc5293aaaafaa25d807b05c6f4465d0697fd99bd2d3602856006",
    "ed25519_priv": "0x0000000000000000000000000000000000000000000000000000000000000004",
    "bls_priv": "0x63ee607bff36cf4b72f25253a1c25a940fa8b30ea9a10f947e062527934af01d"
  },
  {
    "bandersnatch": "0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d",
    "ed25519": "0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02",
    "bls": "0xb0b9121622bf8a9a9e811ee926740a876dd0d9036f2f3060ebfab0c7c489a338a7728ee2da4a265696edcc389fe02b2caf20b5b83aeb64aaf4184bedf127f4eea1d737875854411d58ca4a2b69b066b0a0c09d2a0b7121ade517687c51954df913fe930c227723dd8f58aa2415946044dc3fb15c367a2185d0fc1f7d2bb102ff14a230d5f81cfc8ad445e51efddbf426",
    "bandersnatch_priv": "0x68f5494ec1c3d3cd8ff2a3cb285abf0e826b2c762d95fc2e953eaef666315403",
    "ed25519_priv": "0x0500000005000000050000000500000005000000050000000500000005000000",
    "bls_priv": "0x870ac1a7736dc9d533f42711fed07f22ac317052d390bce13e02a7cc0038c126"
  }
]
*/

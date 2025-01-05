package extrinsic

import (
	"encoding/hex"
	"fmt"
	"testing"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func hexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

func TestNewVerdictController(t *testing.T) {
	verdictController := NewVerdictController()
	verdictController.Verdicts = append(verdictController.Verdicts, VerdictWrapper{
		types.Verdict{
			Target: types.OpaqueHash(hexToBytes("0x3c6680931983df80bbd1cb77a0db0303e61550ee0a868d8edd69302d8f45c29f")),
			Age:    0,
			Votes: []types.Judgement{
				{
					Vote:      true,
					Index:     0,
					Signature: types.Ed25519Signature(hexToBytes("0x4d3e576eaa0d449465115bedc8640c452813bd74385003e212b0ff679b9fbf4951ff1a624fcdf4539926cc00f719078d5828f7aecee8918a62c720ead07c240f")),
				},
				{
					Vote:      true,
					Index:     1,
					Signature: types.Ed25519Signature(hexToBytes("0xc316345dbf3e4de359e4ddbb94ae8460230e9df90be6f57d92bc177f74da74337fb5461d51389ade2a3cfb55d8980b5e7ffb3063f5d6997b4b115dc9a369f301")),
				},
				{
					Vote:      true,
					Index:     2,
					Signature: types.Ed25519Signature(hexToBytes("0x8ef8805d2d8ad355c1805c6cbc378bfd549ed2e84ebcda971cc783a967f76a247d6ca435c1188c704f6045bdd080885a9611c3fe812a910c61a4f6016134d508")),
				},
				{
					Vote:      true,
					Index:     3,
					Signature: types.Ed25519Signature(hexToBytes("0xc354b7fca4817783c68c2ca47c45028a1fac29cf6ed7b6df53951bf29b85871d3d1ab00163f085198efbbeb27328ab1af2c1b50ebddfc39475e6a0678e093a0d")),
				},
				{
					Vote:      true,
					Index:     4,
					Signature: types.Ed25519Signature(hexToBytes("0xf36e93889f6df8103404f019c0b408a5079c85cb06a680c4fd70acc087cae0d563a5d4e5cdd7ff2fb1260b885b23ee1fbfc58c80be0301a95f9f271964eeba02")),
				},
			},
		},
	})
	verdictController.Verdicts = append(verdictController.Verdicts, VerdictWrapper{
		types.Verdict{
			Target: types.OpaqueHash(hexToBytes("0x94a9424dec0e513afb0a9187c0456b7760f02ee5969130fde29c8683d62f74fb")),
			Age:    0,
			Votes: []types.Judgement{
				{
					Vote:      false,
					Index:     0,
					Signature: types.Ed25519Signature(hexToBytes("0xccb6a65353f79cc9d65ba9e8b14cd19f24104cd5194c9d2cc81798d982dd6f50777e368364a9df062bb70c0bc9f66ce1391e2bb98cd76f891167cb2f90be770e")),
				},
				{
					Vote:      false,
					Index:     1,
					Signature: types.Ed25519Signature(hexToBytes("0x3b487b6e435b45f11d01fa9bd2d3a34863b2ec4d2c2b63f99cebd6262270e642cce938a7f52e6b5eabd3eb80ca4f4c4b88dc6f147de42e7f7abf8bd0ce0ed00f")),
				},
				{
					Vote:      false,
					Index:     2,
					Signature: types.Ed25519Signature(hexToBytes("0x3070c1929f539dd8251542cc4b9ef79511684614d520410ca080c75be23dbd3d9f8308b68e78879354de377e66c413dcd170d742ed234a794478d6c50dc84801")),
				},
				{
					Vote:      false,
					Index:     3,
					Signature: types.Ed25519Signature(hexToBytes("0x9d8e8608c4a581ae559f2ed2265347183aac86888988a4f233e78b753b50181e153fc718354c0d964a3b0399886c18d0343ed7cfbda722aa51b3c18e3d1b4a09")),
				},
				{
					Vote:      false,
					Index:     4,
					Signature: types.Ed25519Signature(hexToBytes("0x73c4c5f659227fd0b63e0929c764490726e1346bb6540c6cf9ead7bc74473c41829f92b43f78e0516603787ed80766e98e4daeb51187723b2f3a40f2c06b3209")),
				},
			},
		},
	})
}

func TestVerifySignature(t *testing.T) {
	// test data 1 : https://github.com/davxy/jam-test-vectors/blob/polkajam-vectors/disputes/tiny/progress_with_verdicts-1.json
	verdictController := NewVerdictController()
	verdictController.Verdicts = append(verdictController.Verdicts, VerdictWrapper{
		types.Verdict{
			Target: types.OpaqueHash(hexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
			Age:    0,
			Votes: []types.Judgement{
				{
					Vote:      true,
					Index:     0,
					Signature: types.Ed25519Signature(hexToBytes("0x0b1e29dbda5e3bba5dde21c81a8178b115ebf0cf5920fe1a38e897ecadd91718e34bf01c9fc7fdd0df31d83020231b6e8338c8dc204b618cbde16a03cb269d05")),
				},
				{
					Vote:      true,
					Index:     1,
					Signature: types.Ed25519Signature(hexToBytes("0x0d44746706e09ff6b6f2929e736c2f868a4d17939af6d37ca7d3c7f6d4914bd095a6fd4ff48c320b673e2de92bfdb5ed9f5c0c40749816ab4171a2272386fc05")),
				},
				{
					Vote:      true,
					Index:     2,
					Signature: types.Ed25519Signature(hexToBytes("0x0d5d39f2239b775b22aff53b74a0d708a9b9363ed5017170f0abebc8ffd97fc1cc3cf597c578b555ad5abab26e09ecda727c2909feae99587c6354b86e4cc50c")),
				},
				{
					Vote:      true,
					Index:     3,
					Signature: types.Ed25519Signature(hexToBytes("0x701d277fa78993b343a5d4367f1c2a2fb7ddb77f0246bf9028196feccbb7c0f2bd994966b3e9b1e51ff5dd63d8aa5e2331432b9cca4a125552c4700d51814a04")),
				},
				{
					Vote:      true,
					Index:     4,
					Signature: types.Ed25519Signature(hexToBytes("0x08d96d2e49546931dc3de989a69aa0ae3547d67a038bdaa84f7e549da8318d48aab72b4b30ecc0c588696305fce3e2c4657f409463f6a05c52bf641f2684460f")),
				},
			},
		},
	})

	kappa := []types.Validator{
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			Ed25519: types.Ed25519Public(hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"))},
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			Ed25519: types.Ed25519Public(hexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862"))},
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			Ed25519: types.Ed25519Public(hexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00"))},
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			Ed25519: types.Ed25519Public(hexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9"))},
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			Ed25519: types.Ed25519Public(hexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff"))},
		{Bandersnatch: types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			Ed25519: types.Ed25519Public(hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02"))},
	}

	states := store.GetInstance().GetStates()
	states.SetKappa(kappa)
	// test data 1 : https://github.com/davxy/jam-test-vectors/blob/polkajam-vectors/disputes/tiny/progress_with_verdicts-1.json
	fmt.Println("Data 1 : valid signatures in verdicts")
	for i := 0; i < len(verdictController.Verdicts); i++ {
		//_ = VerifyJudgementSignature(&verdictController.Verdicts[i])
		VerdictPtr := &verdictController.Verdicts[i]
		invalid := VerdictPtr.VerifySignature()
		if len(invalid) > 0 {
			t.Errorf("invalid signature in verdict %d", i)
		} else {
			fmt.Println("All signatures are valid")
		}
	}
	fmt.Println("-----------------------------")
	// test data 2 : https://github.com/davxy/jam-test-vectors/blob/polkajam-vectors/disputes/tiny/progress_with_bad_signatures-1.json
	verdictController = NewVerdictController()
	verdictController.Verdicts = append(verdictController.Verdicts, VerdictWrapper{
		types.Verdict{
			Target: types.OpaqueHash(hexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
			Age:    0,
			Votes: []types.Judgement{
				{
					Vote:      true,
					Index:     0,
					Signature: types.Ed25519Signature(hexToBytes("0x647c04630e911a432f99e6c1108bcf4c06496754033b77c5eb8271a5d06a85e1884db0fb977e232e416643ccfaf4f334e99f3b8d9cdfc65a8e4ecbc9db284005")),
				},
				{
					Vote:      true,
					Index:     1,
					Signature: types.Ed25519Signature(hexToBytes("0xb95288deca20fcb649a3515ece2f1d147f8c0f3acef20967cd05a7b770c96e2e0928a056af1aa233c45b0a154e31dae842a50ff48f249cc364af20f282db950a")),
				},
				{
					Vote:      true,
					Index:     2,
					Signature: types.Ed25519Signature(hexToBytes("0x2f8476e2c06dec1fd24130363f922f5419d91cd250d0d9448e8db471e08377a8ee927c8ca9c45ff47796cf0dfb35e4aff4fee96cae8dbe40fbd48d1cd59c7f0c")),
				},
				{
					Vote:      true,
					Index:     3,
					Signature: types.Ed25519Signature(hexToBytes("0xdf20eab8438b43d774ac84d4225607cdd4159ee495991c89c9dc302e7d826b53e23b1f266a2dcf915dee277a0bfa0b93c957504503213c3f57a4c08b5ce07f0b")),
				},
				{
					Vote:      true,
					Index:     4,
					Signature: types.Ed25519Signature(hexToBytes("0xdf20eab8438b43d774ac84d4225607cdd4159ee495991c89c9dc302e7d826b53e23b1f266a2dcf915dee277a0bfa0b93c957504503213c3f57a4c08b5ce07f0b")),
				},
			},
		},
	})
	// kappa is the same

	// test data 2 : https://github.com/davxy/jam-test-vectors/blob/polkajam-vectors/disputes/tiny/progress_with_bad_signatures-1.json
	fmt.Println("Data 2 : invalid signatures in verdict")
	for i := 0; i < len(verdictController.Verdicts); i++ {
		//_ = VerifyJudgementSignature(&verdictController.Verdicts[i])
		VerdictPtr := &verdictController.Verdicts[i]
		invalid := VerdictPtr.VerifySignature()
		if len(invalid) > 0 {
			fmt.Println("Invalid signature at index", invalid)
		}
	}
}

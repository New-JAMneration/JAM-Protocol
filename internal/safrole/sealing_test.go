package safrole

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
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

func TestUpdateEtaPrime0(t *testing.T) {
	// input from safrole/state_snapshots/425530_011.json
	eta := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa")),
		types.Entropy(hexToByteArray32("0x491db955568b306018eee09f2966f873bb665a20b703584ad868454b81b17e76")),
		types.Entropy(hexToByteArray32("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
	}

	kappa := types.ValidatorsData{
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			Bls:          types.BlsPublic(hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			Bls:          types.BlsPublic(hexToBytes("0xa2534be5b2f761dc898160a9b4762eb46bd171222f6cdf87f5127a9e8970a54c44fe7b2e12dda098854a9aaab03c3a47953085668673a84b0cedb4b0391ed6ae2deb1c3e04f0bc618a2bc1287d8599e8a1c47ff715cd4cbd3fe80e2607744d4514b491ed2ef76ae114ecb1af99ba6af32189bf0471c06aa3e6acdaf82e7a959cb24a5c1444cac3a6678f5182459fd8ce")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
			Bls:          types.BlsPublic(hexToBytes("0x8faee314528448651e50bea6d2e7e5d3176698fea0b932405a4ec0a19775e72325e44a6d28f99fba887e04eb818f13d1b73f75f0161644283df61e7fbaad7713fae0ef79fe92499202834c97f512d744515a57971badf2df62e23697e9fe347f168fed0adb9ace131f49bbd500a324e2469569423f37c5d3b430990204ae17b383fcd582cb864168c8b46be8d779e7ca")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
			Bls:          types.BlsPublic(hexToBytes("0x8dfdac3e2f604ecda637e4969a139ceb70c534bd5edc4210eb5ac71178c1d62f0c977197a2c6a9e8ada6a14395bc9aa3a384d35f40c3493e20cb7efaa799f66d1cedd5b2928f8e34438b07072bbae404d7dfcee3f457f9103173805ee163ba550854e4660ccec49e25fafdb00e6adfbc8e875de1a9541e1721e956b972ef2b135cc7f71682615e12bb7d6acd353d7681")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
			Bls:          types.BlsPublic(hexToBytes("0xb78a95d81f6c7cdc517a36d81191b6f7718dcf44e76c0ce9fb724d3aea39fdb3c5f4ee31eb1f45e55b783b687b1e9087b092a18341c7cda102b4100685b0a014d55f1ccdb7600ec0db14bb90f7fc3126dc2625945bb44f302fc80df0c225546c06fa1952ef05bdc83ceb7a23373de0637cd9914272e3e3d1a455db6c48cc6b2b2c17e1dcf7cd1586a235821308aee001")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
			Bls:          types.BlsPublic(hexToBytes("0xb0b9121622bf8a9a9e811ee926740a876dd0d9036f2f3060ebfab0c7c489a338a7728ee2da4a265696edcc389fe02b2caf20b5b83aeb64aaf4184bedf127f4eea1d737875854411d58ca4a2b69b066b0a0c09d2a0b7121ade517687c51954df913fe930c227723dd8f58aa2415946044dc3fb15c367a2185d0fc1f7d2bb102ff14a230d5f81cfc8ad445e51efddbf426")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	s := store.GetInstance()

	s.GetPriorStates().SetEta(eta)
	s.GetPosteriorStates().SetKappa(kappa)
	s.GetIntermediateHeaders().SetAuthorIndex(2)
	s.GetIntermediateHeaders().SetEntropySource(types.BandersnatchVrfSignature(hex2Bytes("0x4cc4f71f2a4e3404632ba128c1de4a56e961d50d408d3bba17e31f99fb082a8a998e6e03d797bed2fd1e0c50f777c0df5d6c9e19720ea85d7ad106bbb9a7141ad530fa180d36b0e6fc4fdae845113377e12a0117943c91ac4751eeb4eb66e416"))) // next epoch first block header

	UpdateEtaPrime0()

	expectedEntropy := types.Entropy(hexToByteArray32("0x8c89c318db72cd8b6edf1cd42e2e5501afc76d7fc4d84d7af95d84f64a96daf0"))
	actualEntropy := s.GetPosteriorStates().GetState().Eta[0]

	if !bytes.Equal(actualEntropy[:], expectedEntropy[:]) {
		t.Errorf("CalculateHeaderEntropy() = %v, want %v", actualEntropy, expectedEntropy)
	}
}

func TestUpdateHeaderEntropy(t *testing.T) {
	kappa := types.ValidatorsData{
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			Bls:          types.BlsPublic(hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			Bls:          types.BlsPublic(hexToBytes("0xa2534be5b2f761dc898160a9b4762eb46bd171222f6cdf87f5127a9e8970a54c44fe7b2e12dda098854a9aaab03c3a47953085668673a84b0cedb4b0391ed6ae2deb1c3e04f0bc618a2bc1287d8599e8a1c47ff715cd4cbd3fe80e2607744d4514b491ed2ef76ae114ecb1af99ba6af32189bf0471c06aa3e6acdaf82e7a959cb24a5c1444cac3a6678f5182459fd8ce")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
			Bls:          types.BlsPublic(hexToBytes("0x8faee314528448651e50bea6d2e7e5d3176698fea0b932405a4ec0a19775e72325e44a6d28f99fba887e04eb818f13d1b73f75f0161644283df61e7fbaad7713fae0ef79fe92499202834c97f512d744515a57971badf2df62e23697e9fe347f168fed0adb9ace131f49bbd500a324e2469569423f37c5d3b430990204ae17b383fcd582cb864168c8b46be8d779e7ca")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
			Bls:          types.BlsPublic(hexToBytes("0x8dfdac3e2f604ecda637e4969a139ceb70c534bd5edc4210eb5ac71178c1d62f0c977197a2c6a9e8ada6a14395bc9aa3a384d35f40c3493e20cb7efaa799f66d1cedd5b2928f8e34438b07072bbae404d7dfcee3f457f9103173805ee163ba550854e4660ccec49e25fafdb00e6adfbc8e875de1a9541e1721e956b972ef2b135cc7f71682615e12bb7d6acd353d7681")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
			Bls:          types.BlsPublic(hexToBytes("0xb78a95d81f6c7cdc517a36d81191b6f7718dcf44e76c0ce9fb724d3aea39fdb3c5f4ee31eb1f45e55b783b687b1e9087b092a18341c7cda102b4100685b0a014d55f1ccdb7600ec0db14bb90f7fc3126dc2625945bb44f302fc80df0c225546c06fa1952ef05bdc83ceb7a23373de0637cd9914272e3e3d1a455db6c48cc6b2b2c17e1dcf7cd1586a235821308aee001")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
			Bls:          types.BlsPublic(hexToBytes("0xb0b9121622bf8a9a9e811ee926740a876dd0d9036f2f3060ebfab0c7c489a338a7728ee2da4a265696edcc389fe02b2caf20b5b83aeb64aaf4184bedf127f4eea1d737875854411d58ca4a2b69b066b0a0c09d2a0b7121ade517687c51954df913fe930c227723dd8f58aa2415946044dc3fb15c367a2185d0fc1f7d2bb102ff14a230d5f81cfc8ad445e51efddbf426")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	s := store.GetInstance()

	s.GetPosteriorStates().SetKappa(kappa)
	s.GetIntermediateHeaders().SetAuthorIndex(2)
	s.GetIntermediateHeaders().SetSeal(types.BandersnatchVrfSignature(hex2Bytes("0xa70cd5b81c52bb33540d75a2c68ba482f54758f347d1ef97b6f4f9708199f51c451c97b60fc5a218249ed7c0b2bfb589ac6dbc5776bd10e20bb067257daf410c2ab890dec90b4e0dc96d99058721e6fb0a1543d83a8d3d56d1125d1162fd9208")))
	UpdateHeaderEntropy()

	expectedHeaderEntropy := types.BandersnatchVrfSignature(hex2Bytes("0xa81d92d83542ddfe0f34a81ce306ad3032740a75c14c6e99890690caa647986bf0b815a3ca5543ad24fe67d723a9180266ff65dcc3fc4824f9d688ab1e528301c76c5290d655cc1d1ad079ce696b5a9368241e2f0d16bc5cd413830f5c4c100a"))
	actualHeaderEntropy := s.GetIntermediateHeader().EntropySource

	if !bytes.Equal(actualHeaderEntropy[:], expectedHeaderEntropy[:]) {
		t.Errorf("CalculateHeaderEntropy() = %v, want %v", actualHeaderEntropy, expectedHeaderEntropy)
	}
}

func TestUpdateEntropy(t *testing.T) {
	s := store.GetInstance()
	eta := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa")),
		types.Entropy(hexToByteArray32("0x491db955568b306018eee09f2966f873bb665a20b703584ad868454b81b17e76")),
		types.Entropy(hexToByteArray32("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
	}
	s.GetPriorStates().SetTau(types.TimeSlot(types.EpochLength - 1))
	s.GetPriorStates().SetEta(eta)
	s.GetPosteriorStates().SetTau(types.TimeSlot(types.EpochLength))
	UpdateEntropy()
	etaPrime := s.GetPosteriorStates().GetEta()
	expect_etaPrime := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa")),
		types.Entropy(hexToByteArray32("0x491db955568b306018eee09f2966f873bb665a20b703584ad868454b81b17e76")),
	}
	for i := 0; i < 4; i++ {
		if !bytes.Equal(etaPrime[i][:], expect_etaPrime[i][:]) {
			t.Errorf("UpdateEntropy() = %v, want %v", etaPrime[i], expect_etaPrime[i])
		}
	}
}

func createTicketBody(id string, attempt uint8) types.TicketBody {
	idBytes := hexToByteArray32(id)
	return types.TicketBody{Id: types.TicketId(idBytes), Attempt: types.TicketAttempt(attempt)}
}

func TestUpdateSlotKeySequence(t *testing.T) {
	s := store.GetInstance()
	eta := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x64e9065b8ed901f4fe6b04ce75c5e4f116de1b632090027b39bea2bfdf5453d7")),
		types.Entropy(hexToByteArray32("0x4346d1d2300d8a705e8d0165384f2b778e114a7498fbf881343b2f59b4450efa")),
		types.Entropy(hexToByteArray32("0x491db955568b306018eee09f2966f873bb665a20b703584ad868454b81b17e76")),
		types.Entropy(hexToByteArray32("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
	}
	s.GetPriorStates().SetTau(types.TimeSlot(types.EpochLength - 1))
	s.GetPriorStates().SetEta(eta)
	s.GetPosteriorStates().SetTau(types.TimeSlot(types.EpochLength))

	old_gamma_a := []types.TicketBody{
		createTicketBody("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d", 2),
		createTicketBody("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331", 1),
		createTicketBody("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef", 0),
		createTicketBody("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8", 2),
		createTicketBody("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2", 1),
		createTicketBody("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4", 1),
		createTicketBody("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c", 0),
		createTicketBody("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a", 0),
		createTicketBody("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074", 1),
		createTicketBody("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec", 2),
		createTicketBody("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7", 1),
		createTicketBody("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c", 2),
	}

	expected_new_gamma_s := []types.TicketBody{
		createTicketBody("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d", 2),
		createTicketBody("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c", 2),
		createTicketBody("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331", 1),
		createTicketBody("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7", 1),
		createTicketBody("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef", 0),
		createTicketBody("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec", 2),
		createTicketBody("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8", 2),
		createTicketBody("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074", 1),
		createTicketBody("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2", 1),
		createTicketBody("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a", 0),
		createTicketBody("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4", 1),
		createTicketBody("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c", 0),
	}

	s.GetPriorStates().SetGammaA(old_gamma_a)

	UpdateSlotKeySequence()

	gamma_s := s.GetPosteriorStates().GetGammaS().Tickets

	for i := 0; i < len(expected_new_gamma_s); i++ {
		if !bytes.Equal(gamma_s[i].Id[:], expected_new_gamma_s[i].Id[:]) {
			t.Errorf("UpdateEntropy() = %v, want %v", gamma_s[i].Id[i], expected_new_gamma_s[i].Id[i])
		}
	}
}

func TestSealingByBender(t *testing.T) {
	s := store.GetInstance()
	s.GetIntermediateHeaders().SetParent(types.HeaderHash(hex2Bytes("0x0000000000000000000000000000000000000000000000000000000000000000")))
	s.GetIntermediateHeaders().SetParentStateRoot(types.StateRoot(hex2Bytes("0x57b3782db21156854f214dc7cc88da9bdad250e90813437a1279a9c4af9b5003")))
	s.GetIntermediateHeaders().SetExtrinsicHash(types.OpaqueHash(hex2Bytes("0xdc080ad182cb9ff052a1ca8ecbc51164264efc7dd6debaaa45764950f843acb8")))
	s.GetIntermediateHeaders().SetSlot(types.TimeSlot(4888824))

	validators := []types.BandersnatchPublic{
		types.BandersnatchPublic(hexToByteArray32("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		types.BandersnatchPublic(hexToByteArray32("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		types.BandersnatchPublic(hexToByteArray32("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
		types.BandersnatchPublic(hexToByteArray32("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
		types.BandersnatchPublic(hexToByteArray32("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
	}

	epochMark := types.EpochMark{
		Validators:     validators,
		Entropy:        types.Entropy(hex2Bytes("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")),
		TicketsEntropy: types.Entropy(hex2Bytes("0x835ac82bfa2ce8390bb50680d4b7a73dfa2a4cff6d8c30694b24a605f9574eaf")),
	}
	s.GetIntermediateHeaders().SetEpochMark(epochMark)
	s.GetIntermediateHeaders().SetAuthorIndex(2)
	s.GetIntermediateHeaders().SetEntropySource(types.BandersnatchVrfSignature(hex2Bytes("0x8fc34b4c24f74f16ef7b13fff47ab7b84e6ce6ccae23870ee63d0b9c39261eb84dfed75e12c38d7599bf39f443f5fa5ce583a10b96d1184b450752a93c8c6a1566a41c9aa8ac0645aef85b6f6c0c6c00bc7421cd472cac7011abc190116e8b0d")))
	var gamma_s types.TicketsOrKeys
	gamma_s.Keys = []types.BandersnatchPublic{
		types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		types.BandersnatchPublic(hexToByteArray32("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		types.BandersnatchPublic(hexToByteArray32("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
		types.BandersnatchPublic(hexToByteArray32("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
		types.BandersnatchPublic(hexToByteArray32("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		types.BandersnatchPublic(hexToByteArray32("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		types.BandersnatchPublic(hexToByteArray32("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
		types.BandersnatchPublic(hexToByteArray32("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		types.BandersnatchPublic(hexToByteArray32("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		types.BandersnatchPublic(hexToByteArray32("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
	}
	s.GetPosteriorStates().SetGammaS(gamma_s)
	/*
			"slot": 4888824,
		        "epoch_mark": {
		            "entropy": "0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a",
		            "tickets_entropy": "0x835ac82bfa2ce8390bb50680d4b7a73dfa2a4cff6d8c30694b24a605f9574eaf",
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
		        "author_index": 2,
		        "entropy_source": "0x8fc34b4c24f74f16ef7b13fff47ab7b84e6ce6ccae23870ee63d0b9c39261eb84dfed75e12c38d7599bf39f443f5fa5ce583a10b96d1184b450752a93c8c6a1566a41c9aa8ac0645aef85b6f6c0c6c00bc7421cd472cac7011abc190116e8b0d",
		        "seal": "0xc0909e69767bdce7476a497eabc11c3ab0892703755c7ed244c8144424fa3971e79470cfa9b2f4c8af8cb90a7a1bb6e24b00decb827fca424a1abf325a0838166c38544ea446844ec3d7f2893630597cf825848b9c5009cd494e231eb5571210"

	*/

	kappa := types.ValidatorsData{
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			Bls:          types.BlsPublic(hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			Bls:          types.BlsPublic(hexToBytes("0xa2534be5b2f761dc898160a9b4762eb46bd171222f6cdf87f5127a9e8970a54c44fe7b2e12dda098854a9aaab03c3a47953085668673a84b0cedb4b0391ed6ae2deb1c3e04f0bc618a2bc1287d8599e8a1c47ff715cd4cbd3fe80e2607744d4514b491ed2ef76ae114ecb1af99ba6af32189bf0471c06aa3e6acdaf82e7a959cb24a5c1444cac3a6678f5182459fd8ce")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
			Bls:          types.BlsPublic(hexToBytes("0x8faee314528448651e50bea6d2e7e5d3176698fea0b932405a4ec0a19775e72325e44a6d28f99fba887e04eb818f13d1b73f75f0161644283df61e7fbaad7713fae0ef79fe92499202834c97f512d744515a57971badf2df62e23697e9fe347f168fed0adb9ace131f49bbd500a324e2469569423f37c5d3b430990204ae17b383fcd582cb864168c8b46be8d779e7ca")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			Ed25519:      types.Ed25519Public(hexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
			Bls:          types.BlsPublic(hexToBytes("0x8dfdac3e2f604ecda637e4969a139ceb70c534bd5edc4210eb5ac71178c1d62f0c977197a2c6a9e8ada6a14395bc9aa3a384d35f40c3493e20cb7efaa799f66d1cedd5b2928f8e34438b07072bbae404d7dfcee3f457f9103173805ee163ba550854e4660ccec49e25fafdb00e6adfbc8e875de1a9541e1721e956b972ef2b135cc7f71682615e12bb7d6acd353d7681")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
			Bls:          types.BlsPublic(hexToBytes("0xb78a95d81f6c7cdc517a36d81191b6f7718dcf44e76c0ce9fb724d3aea39fdb3c5f4ee31eb1f45e55b783b687b1e9087b092a18341c7cda102b4100685b0a014d55f1ccdb7600ec0db14bb90f7fc3126dc2625945bb44f302fc80df0c225546c06fa1952ef05bdc83ceb7a23373de0637cd9914272e3e3d1a455db6c48cc6b2b2c17e1dcf7cd1586a235821308aee001")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		{
			Bandersnatch: types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			Ed25519:      types.Ed25519Public(hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
			Bls:          types.BlsPublic(hexToBytes("0xb0b9121622bf8a9a9e811ee926740a876dd0d9036f2f3060ebfab0c7c489a338a7728ee2da4a265696edcc389fe02b2caf20b5b83aeb64aaf4184bedf127f4eea1d737875854411d58ca4a2b69b066b0a0c09d2a0b7121ade517687c51954df913fe930c227723dd8f58aa2415946044dc3fb15c367a2185d0fc1f7d2bb102ff14a230d5f81cfc8ad445e51efddbf426")),
			Metadata:     types.ValidatorMetadata(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	eta := types.EntropyBuffer{
		types.Entropy(hexToByteArray32("0x02ab839dfe0af0aed416957b9065515734c4af180b82b7b49d6b9bb02b38fd7d")),
		types.Entropy(hexToByteArray32("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")),
		types.Entropy(hexToByteArray32("0x835ac82bfa2ce8390bb50680d4b7a73dfa2a4cff6d8c30694b24a605f9574eaf")),
		types.Entropy(hexToByteArray32("0xd2d34655ebcad804c56d2fd5f932c575b6a5dbb3f5652c5202bcc75ab9c2cc95")),
	}

	s.GetPosteriorStates().SetEta(eta)
	s.GetPosteriorStates().SetKappa(kappa)

	SealingByBandersnatchs()

	expectedSeal := types.BandersnatchVrfSignature(hex2Bytes("0xc0909e69767bdce7476a497eabc11c3ab0892703755c7ed244c8144424fa3971e79470cfa9b2f4c8af8cb90a7a1bb6e24b00decb827fca424a1abf325a0838166c38544ea446844ec3d7f2893630597cf825848b9c5009cd494e231eb5571210"))
	fmt.Println(expectedSeal)
	fmt.Println(s.GetIntermediateHeader().Seal)
}

package safrole

import (
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestCreateEpochMarkerNewEpoch(t *testing.T) {
	s := store.GetInstance()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Simulate previous time slot to trigger create epoch marker
	priorTau := tauPrime - types.TimeSlot(types.EpochLength)

	s.GetPriorStates().SetTau(priorTau)

	// Set gamma_k into posterior state
	// Load fake validators
	fakeValidators := LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	s.GetPosteriorStates().SetGammaK(validatorsData)

	// Prepare eta_0, eta_1
	eta0 := types.Entropy(hash.Blake2bHash([]byte("eta0")))
	eta1 := types.Entropy(hash.Blake2bHash([]byte("eta1")))

	// Get Eta from prior state
	priorEta := s.GetPriorStates().GetEta()

	// Update Eta
	priorEta[0] = eta0
	priorEta[1] = eta1

	// Set Eta
	s.GetPriorStates().SetEta(priorEta)

	CreateEpochMarker()

	// Check if epoch marker is created
	epochMarker := s.GetIntermediateHeaderPointer().GetEpochMark()

	if epochMarker == nil {
		t.Errorf("Epoch marker should not be nil")
		return
	}

	// Check if epoch marker is correct
	if epochMarker.Entropy != eta0 {
		t.Errorf("Epoch marker entropy is incorrect")
	}

	if epochMarker.TicketsEntropy != eta1 {
		t.Errorf("Epoch marker tickets entropy is incorrect")
	}

	if len(epochMarker.Validators) != len(fakeValidators) {
		t.Errorf("Epoch marker validators length is incorrect")
	}

	// Compare each validator
	for i, validator := range epochMarker.Validators {
		if validator != fakeValidators[i].Bandersnatch {
			t.Errorf("Epoch marker validator is incorrect")
		}
	}
}

func TestCreateEpochMarkerSameEpoch(t *testing.T) {
	s := store.GetInstance()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Set the prior tau to the current tau, so that the epoch index is the same
	priorTau := tauPrime

	s.GetPriorStates().SetTau(priorTau)

	// Set gamma_k into posterior state
	// Load fake validators
	fakeValidators := LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	s.GetPosteriorStates().SetGammaK(validatorsData)

	CreateEpochMarker()

	// Check if epoch marker is nil
	if s.GetIntermediateHeaderPointer().GetEpochMark() != nil {
		t.Errorf("Epoch marker should be nil")
	}
}

func TestCreateEpochMarkerNewEpochWithJamTestNetData(t *testing.T) {
	// input from safrole/state_snapshots/425530_011.json
	eta := types.EntropyBuffer{
		types.Entropy(hexToBytes("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
		types.Entropy(hexToBytes("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")),
		types.Entropy(hexToBytes("0x835ac82bfa2ce8390bb50680d4b7a73dfa2a4cff6d8c30694b24a605f9574eaf")),
		types.Entropy(hexToBytes("0xd2d34655ebcad804c56d2fd5f932c575b6a5dbb3f5652c5202bcc75ab9c2cc95")),
	}

	gamma_k := types.ValidatorsData{
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

	// expected epoch marker from blocks/425531_000.json
	expectedEpochMarker := types.EpochMark{
		Entropy:        types.Entropy(hexToBytes("0x0de5a58e78d62b28af21e63fc7f901e1435165a1e8324fa3b20c18afd901c29b")),
		TicketsEntropy: types.Entropy(hexToBytes("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")),
		Validators: []types.BandersnatchPublic{
			types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
			types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
			types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
			types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
			types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
			types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
		},
	}

	s := store.GetInstance()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Simulate previous time slot to trigger create epoch marker
	priorTau := tauPrime - types.TimeSlot(types.EpochLength)

	s.GetPriorStates().SetTau(priorTau)

	s.GetPosteriorStates().SetGammaK(gamma_k)
	s.GetPriorStates().SetEta(eta)

	CreateEpochMarker()

	// Check if epoch marker is created
	epochMarker := s.GetIntermediateHeaderPointer().GetEpochMark()

	if epochMarker == nil {
		t.Errorf("Epoch marker should not be nil")
		return
	}

	// Check if epoch marker is correct
	if epochMarker.Entropy != expectedEpochMarker.Entropy {
		t.Errorf("Epoch marker entropy is incorrect")
	}

	if epochMarker.TicketsEntropy != expectedEpochMarker.TicketsEntropy {
		t.Errorf("Epoch marker tickets entropy is incorrect")
	}

	if len(epochMarker.Validators) != len(expectedEpochMarker.Validators) {
		t.Errorf("Epoch marker validators length is incorrect")
	}

	// Compare each validator
	for i, validator := range epochMarker.Validators {
		if validator != expectedEpochMarker.Validators[i] {
			t.Errorf("Epoch marker validator is incorrect")
		}
	}
}

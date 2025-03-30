package safrole

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SaforleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestCreateEpochMarkerNewEpoch(t *testing.T) {
	s := store.GetInstance()

	priorTau := types.TimeSlot(types.EpochLength - 1)
	posteriorTau := types.TimeSlot(types.EpochLength)

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posteriorTau)

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
		if validator.Bandersnatch != fakeValidators[i].Bandersnatch || validator.Ed25519 != fakeValidators[i].Ed25519 {
			t.Errorf("Epoch marker validator is incorrect")
		}
	}
}

func TestCreateEpochMarkerSameEpoch(t *testing.T) {
	s := store.GetInstance()

	priorTau := types.TimeSlot(types.EpochLength - 2)
	posteriorTau := types.TimeSlot(types.EpochLength - 1)

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posteriorTau)

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

/*
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

	// Simulate previous time slot to trigger create epoch marker
	priorTau := types.TimeSlot(types.EpochLength - 1)
	posteriorTau := types.TimeSlot(types.EpochLength)

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posteriorTau)

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
*/

func TestCreateWinningTicketsPassFullConditions(t *testing.T) {
	// Prepare test data
	s := store.GetInstance()

	// Simulate different epoch and pass slot index condition
	priorTau := types.TimeSlot(types.SlotSubmissionEnd - 1)
	posterTau := types.TimeSlot(types.SlotSubmissionEnd)
	gamma_a := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c")),
			Attempt: 2,
		},
	}

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posterTau)
	s.GetPriorStates().SetGammaA(gamma_a)

	CreateWinningTickets()

	if s.GetIntermediateHeaderPointer().GetTicketsMark() == nil {
		t.Errorf("Tickets mark should not be nil")
	}
}

// Different epoch, no epoch marker should be created
func TestCreateWinningTicketsDifferentEpoch(t *testing.T) {
	// Prepare test data
	s := store.GetInstance()

	// Simulate different epoch and pass slot index condition
	priorTau := types.TimeSlot(types.SlotSubmissionEnd - 1)
	posterTau := types.TimeSlot(types.SlotSubmissionEnd + types.EpochLength)
	gamma_a := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c")),
			Attempt: 2,
		},
	}

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posterTau)
	s.GetPriorStates().SetGammaA(gamma_a)

	CreateWinningTickets()

	// Check if epoch marker is nil
	if s.GetIntermediateHeaderPointer().GetTicketsMark() != nil {
		t.Errorf("Tickets mark should be nil")
	}
}

// The slot index isn't the end of the submission (m < Y <= m')
// No winning tickets should be created
func TestCreateWinningTicketsSlotIndexNotEndOfSubmission(t *testing.T) {
	// Prepare test data
	s := store.GetInstance()

	priorTau := types.TimeSlot(0)
	posterTau := types.TimeSlot(1)
	gamma_a := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c")),
			Attempt: 2,
		},
	}

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posterTau)
	s.GetPriorStates().SetGammaA(gamma_a)

	CreateWinningTickets()

	// Check if epoch marker is nil
	if s.GetIntermediateHeaderPointer().GetTicketsMark() != nil {
		t.Errorf("Tickets mark should be nil")
	}
}

// |gamma_a| != EpochLength
// No winning tickets should be created
func TestCreateWinningTicketsGammaALengthNotEqualEpochLength(t *testing.T) {
	// Prepare test data
	s := store.GetInstance()

	priorTau := types.TimeSlot(types.SlotSubmissionEnd - 1)
	posterTau := types.TimeSlot(types.SlotSubmissionEnd)
	gamma_a := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
	}

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posterTau)
	s.GetPriorStates().SetGammaA(gamma_a)

	CreateWinningTickets()

	// Check if epoch marker is nil
	if s.GetIntermediateHeaderPointer().GetTicketsMark() != nil {
		t.Errorf("Tickets mark should be nil")
	}
}

// 425530_009.json -> 425530_010.json
func TestCreateWinningTicketsWithJamTestNet(t *testing.T) {
	testPriorEpochIndex := 425530
	testPosteriorEpochIndex := 425530
	testPriorSlotIndex := 9
	testPosteriorSlotIndex := 10

	// Prepare test data
	s := store.GetInstance()

	// Simulate different epoch and pass slot index condition
	priorTau := types.TimeSlot(testPriorEpochIndex*types.EpochLength + testPriorSlotIndex)
	posterTau := types.TimeSlot(testPosteriorEpochIndex*types.EpochLength + testPosteriorSlotIndex)
	gamma_a := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c")),
			Attempt: 2,
		},
	}

	s.GetPriorStates().SetTau(priorTau)
	s.GetPosteriorStates().SetTau(posterTau)
	s.GetPriorStates().SetGammaA(gamma_a)

	// Expected tickets mark
	expectedTicketsMark := types.TicketsAccumulator{
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x0b7537993b0a700def26bb16e99ed0bfb530f616e4c13cf63ecb60bcbe83387d")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x71dd32fb8a1580b4aa3213c3616d8fbbcb9edc00467c4e4548ff8a1fd815811c")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x1912baa74049a4cad89dc3f0646144459b691b926cf8b9c1c4a5bbfa1ee0c331")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x6b5cc620ed50042cd517ec8267706c82482f07ebcb3c65bfb6288ef5984141a7")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x22fdcfa858e5195e222174597d7d33bd66d97748c413b876f7a132134ce9baef")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x665df13fd353ffe92e9bd68ae952f4511681f04bd2ffb9a6da1b1f5f706c53ec")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x23bd628fd365a0f3ecd10db746dd04ec5efe61f96da19ae070c44b97d3c9a7b8")),
			Attempt: 2,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39f7d99b86f90cada4aa3b08adfe310024813fca0bdcdff944873a2cc2e47074")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31d6a25525ff4bd6e47e611646d7b5835b94b5c0a69c225371b2b762c93095a2")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39e2d23807ff3788156eac40cc0a622a9fd23e9468bf962aebe48079c0fd2f1a")),
			Attempt: 0,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x31e9b8070f42d7c9083eca5879e5528191259a395761b8fcc068dcdd36b06be4")),
			Attempt: 1,
		},
		types.TicketBody{
			Id:      types.TicketId(hexToBytes("0x39120d5b82981c7f5aba8247925f358afb9539839b61602a0726f51efb35ef4c")),
			Attempt: 0,
		},
	}

	CreateWinningTickets()

	if s.GetIntermediateHeaderPointer().GetTicketsMark() == nil {
		t.Errorf("Tickets mark should not be nil")
	}

	if len(*s.GetIntermediateHeaderPointer().GetTicketsMark()) != len(expectedTicketsMark) {
		t.Errorf("Tickets mark length is incorrect")
	}

	// Check if tickets mark is correct
	ticketsMark := s.GetIntermediateHeaderPointer().GetTicketsMark()
	for i, ticket := range *ticketsMark {
		if ticket.Id != expectedTicketsMark[i].Id {
			t.Errorf("Tickets mark id is incorrect")
		}

		if ticket.Attempt != expectedTicketsMark[i].Attempt {
			t.Errorf("Tickets mark attempt is incorrect")
		}
	}
}

// jam-test-vectors: enact-epoch-change-with-no-tickets-2.json
// Progress from slot X to slot X.
// Timeslot must be strictly monotonic.
func TestCreateEpochMarkerGetBadSlotError(t *testing.T) {
	badSlotErr := SaforleErrorCode.BadSlot
	testCases := []struct {
		slot        types.TimeSlot
		entropy     types.Entropy
		extrinsic   []types.TicketsExtrinsic
		preTau      types.TimeSlot
		preEta      types.EntropyBuffer
		pks         []types.BandersnatchPublic // from gamma_k
		expectedErr *types.ErrorCode
	}{
		{
			slot:      1,
			entropy:   types.Entropy(hexToBytes("0xe4b188579aa828f694f769a31a965a11f2017288fbfdfa8734aadc80685ffff7")),
			extrinsic: []types.TicketsExtrinsic{},
			preTau:    1,
			preEta: types.EntropyBuffer{
				types.Entropy(hexToBytes("0xa0243a82952899598fcbc74aff0df58a71059a9882d4416919055c5d64bf2a45")),
				types.Entropy(hexToBytes("0xee155ace9c40292074cb6aff8c9ccdd273c81648ff1149ef36bcea6ebb8a3e25")),
				types.Entropy(hexToBytes("0xbb30a42c1e62f0afda5f0a4e8a562f7a13a24cea00ee81917b86b89e801314aa")),
				types.Entropy(hexToBytes("0xe88bd757ad5b9bedf372d8d3f0cf6c962a469db61a265f6418e1ffed86da29ec")),
			},
			pks: []types.BandersnatchPublic{
				types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
				types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
				types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
				types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
				types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
				types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			},
			expectedErr: &badSlotErr,
		},
		{
			slot:      1,
			entropy:   types.Entropy(hexToBytes("0x8c2e6d327dfaa6ff8195513810496949210ad20a96e2b0672a3e1b9335080801")),
			extrinsic: []types.TicketsExtrinsic{},
			preTau:    0,
			preEta: types.EntropyBuffer{
				types.Entropy(hexToBytes("0x03170a2e7597b7b7e3d84c05391d139a62b157e78786d8c082f29dcf4c111314")),
				types.Entropy(hexToBytes("0xee155ace9c40292074cb6aff8c9ccdd273c81648ff1149ef36bcea6ebb8a3e25")),
				types.Entropy(hexToBytes("0xbb30a42c1e62f0afda5f0a4e8a562f7a13a24cea00ee81917b86b89e801314aa")),
				types.Entropy(hexToBytes("0xe88bd757ad5b9bedf372d8d3f0cf6c962a469db61a265f6418e1ffed86da29ec")),
			},
			pks: []types.BandersnatchPublic{
				types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
				types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
				types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
				types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
				types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
				types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		s := store.GetInstance()

		// Set the time slot
		s.GetPriorStates().SetTau(tc.preTau)
		s.GetPosteriorStates().SetTau(tc.slot)

		gammaK := types.ValidatorsData{}
		for _, pk := range tc.pks {
			gammaK = append(gammaK, types.Validator{
				Bandersnatch: pk,
				Ed25519:      types.Ed25519Public{},
				Bls:          types.BlsPublic{},
				Metadata:     types.ValidatorMetadata{},
			})
		}

		s.GetPosteriorStates().SetGammaK(gammaK)

		// Set Eta
		s.GetPriorStates().SetEta(tc.preEta)

		err := CreateEpochMarker()

		if tc.expectedErr == nil {
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		} else {
			if err == nil || *err != *tc.expectedErr {
				t.Errorf("expected %v, got %v", *tc.expectedErr, err)
			}
		}
	}
}

package store

import (
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestGetInitGenesisState(t *testing.T) {
	testFile := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.json"
	state, err := GetInitGenesisState(testFile)

	if err != nil {
		t.Fatalf("Error loading genesis state: %v", err)
	}

	// 宣告並初始化一個 AuthPool 變數
	var authPool types.AuthPool = []types.AuthorizerHash{
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
	}

	// 宣告並初始化一個 AuthPools 變數
	var expectedAlpha types.AuthPools = []types.AuthPool{
		authPool,
		{
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	// Expected Varphi data
	var authQueue types.AuthQueue = []types.AuthorizerHash{
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
	}

	var expectedVarphi types.AuthQueues = []types.AuthQueue{
		authQueue,
		{
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			types.AuthorizerHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
		},
	}

	// Expected Beta data
	expectedBeta := types.BlocksHistory{}

	// Expected Gamma signature data
	expectedGammaKBandersnatch := hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")
	expectedGammaKedEd25519 := hexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")
	expectedGammaKBls := hexToBytes("0xb0b9121622bf8a9a9e811ee926740a876dd0d9036f2f3060ebfab0c7c489a338a7728ee2da4a265696edcc389fe02b2caf20b5b83aeb64aaf4184bedf127f4eea1d737875854411d58ca4a2b69b066b0a0c09d2a0b7121ade517687c51954df913fe930c227723dd8f58aa2415946044dc3fb15c367a2185d0fc1f7d2bb102ff14a230d5f81cfc8ad445e51efddbf426")
	expectedGammaKMetadata := hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	// Expected GammaZ signature data
	expectedGammaZ := hexToBytes("0xa949a60ad754d683d398a0fb674a9bbe525ca26b0b0b9c8d79f210291b40d286d9886a9747a4587d497f2700baee229ca72c54ad652e03e74f35f075d0189a40d41e5ee65703beb5d7ae8394da07aecf9056b98c61156714fd1d9982367bee2992e630ae2b14e758ab0960e372172203f4c9a41777dadd529971d7ab9d23ab29fe0e9c85ec450505dde7f5ac038274cf")

	// Expected GammaS keys data
	expectedGammaSKey := hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")

	// Expected GammaA data (assuming it's an empty array)
	expectedGammaA := types.TicketsAccumulator{}

	// Expected Eta data
	expectedEta := hexToBytes("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")

	// Expected Iota data
	expectedIotaBandersnatch := hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")
	expectedIotaEd25519 := hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")
	expectedIotaBls := hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")
	expectedIotaMetadata := hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	// Expected Kappa data
	expectedKappaBandersnatch := hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")
	expectedKappaEd25519 := hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")
	expectedKappaBls := hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")
	expectedKappaMetadata := hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	// Expected Lambda data
	expectedLambdaBandersnatch := hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")
	expectedLambdaEd25519 := hexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")
	expectedLambdaBls := hexToBytes("0xb27150a1f1cd24bccc792ba7ba4220a1e8c36636e35a969d1d14b4c89bce7d1d463474fb186114a89dd70e88506fefc9830756c27a7845bec1cb6ee31e07211afd0dde34f0dc5d89231993cd323973faa23d84d521fd574e840b8617c75d1a1d0102aa3c71999137001a77464ced6bb2885c460be760c709009e26395716a52c8c52e6e23906a455b4264e7d0c75466e")
	expectedLambdaMetadata := hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	// Expected Rho data
	var expectedRho types.AvailabilityAssignments = []types.AvailabilityAssignmentsItem{
		nil, nil,
	}
	// Expected Tau data
	expectedTau := 0

	// Expected Chi data
	expectedChi := types.PrivilegedServices{
		ManagerServiceIndex:     0,
		AlterPhiServiceIndex:    0,
		AlterIotaServiceIndex:   0,
		AutoAccumulateGasLimits: nil,
	}
	// Expected Pi data
	expectedPi := types.Statistics{
		Current: types.ActivityRecords{
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
		},
		Last: types.ActivityRecords{
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
			{Blocks: 0, Tickets: 0, PreImages: 0, PreImagesSize: 0, Guarantees: 0, Assurances: 0},
		},
	}

	keys := []struct {
		name     string
		actual   []byte
		expected []byte
	}{
		{"Bandersnatch", state.Gamma.GammaK[5].Bandersnatch[:], expectedGammaKBandersnatch},
		{"Ed25519", state.Gamma.GammaK[5].Ed25519[:], expectedGammaKedEd25519},
		{"Bls", state.Gamma.GammaK[5].Bls[:], expectedGammaKBls},
		{"Metadata", state.Gamma.GammaK[5].Metadata[:], expectedGammaKMetadata},
		{"GammaZ", state.Gamma.GammaZ[:], expectedGammaZ},
		{"GammaSKey", state.Gamma.GammaS.Keys[0][:], expectedGammaSKey},
		{"Eta", state.Eta[0][:], expectedEta},
		{"IotaBandersnatch", state.Iota[0].Bandersnatch[:], expectedIotaBandersnatch},
		{"IotaEd25519", state.Iota[0].Ed25519[:], expectedIotaEd25519},
		{"IotaBls", state.Iota[0].Bls[:], expectedIotaBls},
		{"IotaMetadata", state.Iota[0].Metadata[:], expectedIotaMetadata},
		{"KappaBandersnatch", state.Kappa[0].Bandersnatch[:], expectedKappaBandersnatch},
		{"KappaEd25519", state.Kappa[0].Ed25519[:], expectedKappaEd25519},
		{"KappaBls", state.Kappa[0].Bls[:], expectedKappaBls},
		{"KappaMetadata", state.Kappa[0].Metadata[:], expectedKappaMetadata},
		{"LambdaBandersnatch", state.Lambda[0].Bandersnatch[:], expectedLambdaBandersnatch},
		{"LambdaEd25519", state.Lambda[0].Ed25519[:], expectedLambdaEd25519},
		{"LambdaBls", state.Lambda[0].Bls[:], expectedLambdaBls},
		{"LambdaMetadata", state.Lambda[0].Metadata[:], expectedLambdaMetadata},
	}

	for _, key := range keys {
		if !reflect.DeepEqual(key.actual, key.expected) {
			t.Errorf("%s does not match expected value: got %v, expected %v", key.name, key.actual, key.expected)
		}
	}

	if len(state.Gamma.GammaA) != len(expectedGammaA) {
		t.Errorf("GammaA does not match expected value: got %v, expected %v", state.Gamma.GammaA, expectedGammaA)
	}

	if !reflect.DeepEqual(state.Alpha, expectedAlpha) {
		t.Errorf("Alpha does not match expected value: got %v, expected %v", state.Alpha, expectedAlpha)
	}

	if !reflect.DeepEqual(state.Varphi, expectedVarphi) {
		t.Errorf("Varphi does not match expected value: got %v, expected %v", state.Varphi, expectedVarphi)
	}

	if len(state.Beta) != len(expectedBeta) {
		t.Errorf("Beta does not match expected value: got %v, expected %v", state.Beta, expectedBeta)
	}

	if !reflect.DeepEqual(state.Rho, expectedRho) {
		t.Errorf("Rho does not match expected value: got %v, expected %v", state.Rho, expectedRho)
	}

	if state.Tau != types.TimeSlot(expectedTau) {
		t.Errorf("Tau does not match expected value: got %v, expected %v", state.Tau, expectedTau)
	}

	if !reflect.DeepEqual(state.Chi, expectedChi) {
		t.Errorf("Chi does not match expected value: got %v, expected %v", state.Chi, expectedChi)
	}

	if !reflect.DeepEqual(state.Pi, expectedPi) {
		t.Errorf("Pi does not match expected value: got %v, expected %v", state.Pi, expectedPi)
	}

}

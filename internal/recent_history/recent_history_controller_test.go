package recent_history

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_history "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	"github.com/google/go-cmp/cmp"
)

// Custom input struct for json^^

type mmr_json struct {
	Peak []string `json:"peak,omitempty"`
}

type workPackage_json struct {
	Hash        string `json:"hash,omitempty"`
	ExportsRoot string `json:"exports_root,omitempty"`
}

type blockInfo_json struct {
	HeaderHash string             `json:"header_hash,omitempty"`
	Mmr        mmr_json           `json:"mmr"`
	StateRoot  string             `json:"state_root,omitempty"`
	Reported   []workPackage_json `json:"reported,omitempty"`
}

type state_json struct {
	Beta []blockInfo_json
}
type history_json struct {
	HeaderHash      string             `json:"header_hash,omitempty"`
	ParentStateRoot string             `json:"parent_state_root,omitempty"`
	AccumulateRoot  string             `json:"accumulate_root,omitempty"`
	WorkPackages    []workPackage_json `json:"work_packages,omitempty"`
}

type vector_json struct {
	Input     history_json `json:"input,omitempty"`
	PreState  state_json   `json:"pre_state,omitempty"`
	Output    interface{}  `json:"output,omitempty"`
	PostState state_json   `json:"post_state,omitempty"`
}

// mytypes
type myHistory struct {
	HeaderHash      types.HeaderHash
	ParentStateRoot types.StateRoot
	AccumulateRoot  types.OpaqueHash
	WorkPackages    []types.ReportedWorkPackage
}

type myVector struct {
	Input     myHistory
	PreState  types.State
	Output    interface{}
	PostState types.State
}

func hexToOpaqueHash(hexStr string) ([32]byte, error) {
	// remove "0x" prefix
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// decode hex string
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to decode hex: %v", err)
	}

	if len(bytes) != 32 {
		return [32]byte{}, fmt.Errorf("invalid length: expected 32 bytes, got %d", len(bytes))
	}

	var result [32]byte
	copy(result[:], bytes)
	return result, nil
}

/*
func loadInputFromJSON(filePath string) (my myVector, err error) {
	var vector_json vector_json
	data, err := os.ReadFile(filePath)
	if err != nil {
		return my, err
	}
	err = json.Unmarshal(data, &vector_json)

	// convert json to mytypes
	// MyVector.Input: json to myHistory
	var (
		// string to HeaderHash
		inputHeaderHash, _    = hexToOpaqueHash(vector_json.Input.HeaderHash)
		inputHeaderHash_types = types.HeaderHash(inputHeaderHash)

		// string to StateRoot
		inputParentStateRoot, _    = hexToOpaqueHash(vector_json.Input.ParentStateRoot)
		inputParentStateRoot_types = types.StateRoot(inputParentStateRoot)

		// string to OpaqueHash
		inputAccumulateRoot, _    = hexToOpaqueHash(vector_json.Input.AccumulateRoot)
		inputAccumulateRoot_types = types.OpaqueHash(inputAccumulateRoot)

		// string to ReportedWorkPackage
		inputWorkPackages       = vector_json.Input.WorkPackages
		inputWorkPackages_types []types.ReportedWorkPackage
	)
	for _, inputWorkPackage := range inputWorkPackages {
		// string to WorkReportHash and ExportsRoot
		workPackageHash, _ := hexToOpaqueHash(inputWorkPackage.Hash)
		workPackageExportsRoot, _ := hexToOpaqueHash(inputWorkPackage.ExportsRoot)
		// append to inputWorkPackages_types
		inputWorkPackages_types = append(inputWorkPackages_types, types.ReportedWorkPackage{
			Hash:        types.WorkReportHash(workPackageHash),
			ExportsRoot: types.ExportsRoot(workPackageExportsRoot),
		})
	}

	// MyVector.PreState: json to []types.BlockInfo
	var (
		preStateBetas       = vector_json.PreState.Beta
		preStateBetas_types []types.BlockInfo // allow empty slice
	)
	for _, preStateBeta := range preStateBetas {
		var (
			// string to HeaderHash and StateRoot
			headerHash, _ = hexToOpaqueHash(preStateBeta.HeaderHash)
			stateRoot, _  = hexToOpaqueHash(preStateBeta.StateRoot)
		)

		blockInfo := types.BlockInfo{
			HeaderHash: types.HeaderHash(headerHash),
			// Mmr:        mmr, (Write later)
			StateRoot: types.StateRoot(stateRoot),
			Reported:  []types.ReportedWorkPackage{},
		}

		for _, mmrPeak := range preStateBeta.Mmr.Peak {
			var (
				// string to MmrPeak
				peak, _  = hexToOpaqueHash(mmrPeak)
				peakHash = types.OpaqueHash(peak)
			)
			// append to MmrPeak
			blockInfo.Mmr.Peaks = append(blockInfo.Mmr.Peaks, types.MmrPeak(&peakHash))
		}

		var (
			reportedWorkPackages = preStateBeta.Reported
		)
		for _, reportedWorkPackage := range reportedWorkPackages {
			var (
				// string to WorkReportHash and ExportsRoot
				reportedWorkPackageHash, _        = hexToOpaqueHash(reportedWorkPackage.Hash)
				reportedWorkPackageExportsRoot, _ = hexToOpaqueHash(reportedWorkPackage.ExportsRoot)
			)
			// append to ReportedWorkPackage
			blockInfo.Reported = append(blockInfo.Reported, types.ReportedWorkPackage{
				Hash:        types.WorkReportHash(reportedWorkPackageHash),
				ExportsRoot: types.ExportsRoot(reportedWorkPackageExportsRoot),
			})
		}
		preStateBetas_types = append(preStateBetas_types, blockInfo)
	}

	// MyVector.PostState: json to []types.BlockInfo
	var (
		postStateBetas       = vector_json.PostState.Beta
		postStateBetas_types []types.BlockInfo // allow empty slice
	)
	for _, postStateBeta := range postStateBetas {
		var (
			// string to HeaderHash and StateRoot
			headerHash, _ = hexToOpaqueHash(postStateBeta.HeaderHash)
			stateRoot, _  = hexToOpaqueHash(postStateBeta.StateRoot)
		)

		blockInfo := types.BlockInfo{
			HeaderHash: types.HeaderHash(headerHash),
			//Mmr:        mmr,
			StateRoot: types.StateRoot(stateRoot),
			Reported:  []types.ReportedWorkPackage{},
		}

		for _, mmrPeak := range postStateBeta.Mmr.Peak {
			var (
				peak, _  = hexToOpaqueHash(mmrPeak)
				peakHash = types.OpaqueHash(peak)
			)
			// append to MmrPeak
			blockInfo.Mmr.Peaks = append(blockInfo.Mmr.Peaks, types.MmrPeak(&peakHash))
		}

		var (
			reportedWorkPackages = postStateBeta.Reported
		)
		for _, reportedWorkPackage := range reportedWorkPackages {
			var (
				// string to WorkReportHash and ExportsRoot
				reportedWorkPackageHash, _        = hexToOpaqueHash(reportedWorkPackage.Hash)
				reportedWorkPackageExportsRoot, _ = hexToOpaqueHash(reportedWorkPackage.ExportsRoot)
			)
			// append to ReportedWorkPackage
			blockInfo.Reported = append(blockInfo.Reported, types.ReportedWorkPackage{
				Hash:        types.WorkReportHash(reportedWorkPackageHash),
				ExportsRoot: types.ExportsRoot(reportedWorkPackageExportsRoot),
			})
		}
		postStateBetas_types = append(postStateBetas_types, blockInfo)
	}

	// Finally, we construct our myVector with types
	my = myVector{
		Input: myHistory{
			HeaderHash:      inputHeaderHash_types,
			ParentStateRoot: inputParentStateRoot_types,
			AccumulateRoot:  inputAccumulateRoot_types,
			WorkPackages:    inputWorkPackages_types,
		},
		PreState: types.State{
			Beta: preStateBetas_types,
		},
		Output: vector_json.Output,
		PostState: types.State{
			Beta: postStateBetas_types,
		},
	}
	return my, err
}

func TestCheckDuplicate(t *testing.T) {
	my, err := loadInputFromJSON("./data/progress_blocks_history-2.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}
	// Create a new RecentHistoryController
	rhc := NewRecentHistoryController()
	if rhc == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
	}

	// Store initialization
	store.NewPriorStates()
	store.NewIntermediateStates()
	store.GetInstance().GetPriorStates().SetBeta(my.PreState.Beta)
	priorState := store.GetInstance().GetPriorStates()

	// Put input beta to rhc.Betas as priorState.Beta
	rhc.Betas = priorState.GetBeta()

	// Start test CheckDuplicate
	if rhc.Betas != nil { // Beta is empty -> not need to remove duplicate
		// Test existing header hashes
		for _, beta := range rhc.Betas {
			if rhc.CheckDuplicate(beta.HeaderHash) != true {
				t.Error("Expected true for existing header hash:", beta.HeaderHash)
			}
		}
		// Test non-existing header hash
		if rhc.CheckDuplicate(my.Input.HeaderHash) != false {
			t.Error("Expected false for non-existing header hash", my.Input.HeaderHash)
		}
	} else if rhc.Betas == nil {
		fmt.Print("PreState Beta is empty\n")
	}
}

func TestAddToBetaDagger(t *testing.T) {
	my, err := loadInputFromJSON("./data/progress_blocks_history-2.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	// Create a new RecentHistoryController
	rhc := NewRecentHistoryController()
	if rhc == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
	}

	// Store initialization
	s := store.GetInstance()
	s.GetPriorStates().SetBeta(my.PreState.Beta)
	priorState := s.GetPriorStates()

	// Put input beta to rhc.Betas as priorState.Beta
	rhc.Betas = priorState.GetBeta()

	// Start test AddToBetaDagger
	rhc.AddToBetaDagger(my.Input.ParentStateRoot)

	// Get result of BetaDagger from store
	betadagger := s.GetIntermediateStates().GetBetaDagger()

	// length of BetaDagger should not exceed maxBlocksHistory
	if len(betadagger) > types.MaxBlocksHistory {
		t.Errorf("Expected BetaDagger length not to greater than %d, got %d", types.MaxBlocksHistory, len(betadagger))
	}
}

func TestAddToBetaPrime(t *testing.T) {
	my, err := loadInputFromJSON("./data/progress_blocks_history-2.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	// Create a new RecentHistoryController
	rhc := NewRecentHistoryController()
	if rhc == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
	}

	// Store initialization
	s := store.GetInstance()
	s.GetPriorStates().SetBeta(my.PreState.Beta)
	priorState := s.GetPriorStates()

	// Put input beta to rhc.Betas as priorState.Beta
	rhc.Betas = priorState.GetBeta()

	// Start test AddToBetaPrime

	// r function
	// handmade AccumulatedServiceOutput
	mockC := make(types.AccumulatedServiceOutput)
	mockC[types.AccumulatedServiceHash{ServiceId: 1, Hash: my.Input.AccumulateRoot}] = true
	s.GetIntermediateStates().SetBeefyCommitmentOutput(mockC)

	accumulationResultTreeRoot := r(mockC)

	// b function
	NewMmr := rhc.b(accumulationResultTreeRoot)

	if len(NewMmr) == 0 {
		t.Error("Expected non-empty NewMmr")
	}
	if reflect.DeepEqual(NewMmr, my.PostState.Beta[len(my.PostState.Beta)-1].Mmr.Peaks) {
		t.Error("NewMmr result is not equal to BetaPrime")
	}

	// p function
	// GuaranteesExtrinsic from vector_json
	mockEg := types.GuaranteesExtrinsic{}
	for _, workPackage := range my.Input.WorkPackages {
		mockEg = append(mockEg, types.ReportGuarantee{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(workPackage.Hash),
					ExportsRoot: workPackage.ExportsRoot,
				},
			},
		})
	}

	reports := p(mockEg)

	if len(reports) == 0 {
		t.Error("Expected non-empty reports")
	}
	for i, report := range reports {
		if report.Hash != my.PostState.Beta[len(my.PostState.Beta)-1].Reported[i].Hash {
			t.Errorf("report[%d].Hash is not equal to testvector's BetaPrime", i)
		}
		if report.ExportsRoot != my.PostState.Beta[len(my.PostState.Beta)-1].Reported[i].ExportsRoot {
			t.Errorf("report[%d].ExportsRoot is not equal to testvector's BetaPrime", i)
		}
	}

	// n function
	items := rhc.N(my.Input.HeaderHash, mockEg, my.Input.AccumulateRoot)

	if items.HeaderHash != my.PostState.Beta[len(my.PostState.Beta)-1].HeaderHash {
		t.Errorf("items.HeaderHash %x is not equal to BetaPrime's %x", items.HeaderHash, my.PostState.Beta[len(my.PostState.Beta)-1].HeaderHash)
	}

	// (7.4)
	rhc.AddToBetaPrime(items)

	// Get result of (7.4), beta^prime, from store
	betaPrime := s.GetPosteriorStates().GetBeta()

	if len(betaPrime) < 1 {
		t.Errorf("Expected BetaPrime length to be 1, got %d", len(betaPrime))
	}
	if reflect.DeepEqual(betaPrime, my.PostState.Beta) {
		t.Error("BetaPrime should equal to PostState.Beta")
	}
}

func TestSTFBeta2BetaDagger(t *testing.T) {
	my, err := loadInputFromJSON("./data/progress_blocks_history-2.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	// Create a new RecentHistoryController
	rhc := NewRecentHistoryController()
	if rhc == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
	}

	// Store initialization
	s := store.GetInstance()
	s.GetPriorStates().SetBeta(my.PreState.Beta)
	priorState := s.GetPriorStates().GetState()

	// Put input beta to rhc.Betas as priorState.Beta
	rhc.Betas = priorState.Beta

	var mockBlock = types.Block{
		Header: types.Header{
			Parent:          my.Input.HeaderHash,
			ParentStateRoot: my.Input.ParentStateRoot,
		},
	}
	s.AddBlock(mockBlock)

	// Start test STFBetaDagger2BetaPrime

	STFBetaDagger2BetaPrime()

	// Get result of BetaDagger from store
	betadagger := s.GetIntermediateStates().GetBetaDagger()

	// length of BetaDagger should not exceed maxBlocksHistory
	if len(betadagger) > types.MaxBlocksHistory {
		t.Errorf("Expected BetaDagger length not to greater than %d, got %d", types.MaxBlocksHistory, len(betadagger))
	}
}

func TestSTFBetaDagger2BetaPrime(t *testing.T) {
	my, err := loadInputFromJSON("./data/progress_blocks_history-2.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	// Create a new RecentHistoryController
	rhc := NewRecentHistoryController()
	if rhc == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(rhc.Betas) != 0 {
		t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
	}

	// Store initialization
	s := store.GetInstance()
	s.GetPriorStates().SetBeta(my.PreState.Beta)
	priorState := s.GetPriorStates()

	// Put input beta to rhc.Betas as priorState.Beta
	rhc.Betas = priorState.GetBeta()

	var mockBlock = types.Block{
		Header: types.Header{
			Parent:          my.Input.HeaderHash,
			ParentStateRoot: my.Input.ParentStateRoot,
		},
	}
	s.AddBlock(mockBlock)

	// Start test STFBetaDagger2BetaPrime

	STFBetaDagger2BetaPrime()

	// Get result of (7.4), beta^prime, from store
	betaPrime := s.GetPosteriorStates().GetBeta()

	if len(betaPrime) < 1 {
		t.Errorf("Expected BetaPrime length to be 1, got %d", len(betaPrime))
	}
}

// Test recent_history.go
func TestRecentHistory(t *testing.T) {
	// Load test vectors from JSON
	vectors := []string{
		"./data/progress_blocks_history-1.json", // Empty history queue
		"./data/progress_blocks_history-2.json", // Not empty nor full history queue
		"./data/progress_blocks_history-3.json", // Fill the history queue
		"./data/progress_blocks_history-4.json", // Shift the history queue
	}

	for i, vector := range vectors {
		my, err := loadInputFromJSON(vector)
		if err != nil {
			t.Fatalf("Failed to load input from JSON[%d]: %v", i, err)
		}

		// Create a new RecentHistoryController
		rhc := NewRecentHistoryController()
		if rhc == nil {
			t.Fatal("Controller should be initialized successfully")
		}
		if len(rhc.Betas) != 0 {
			t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
		}

		// Store initialization
		s := store.GetInstance()
		s.GetPriorStates().SetBeta(my.PreState.Beta)
		priorState := s.GetPriorStates().GetState()

		// Put preState beta to rhc.Betas as priorState.Beta
		rhc.Betas = priorState.Beta

		// Test STF functions
		// β† ≺ (H, β) (4.6)
		// β′ ≺ (H, EG, β†, C) (4.7)
		var (
			// Header from vector_json
			// mockHeader = types.Header{
			// 	Parent:          my.Input.HeaderHash,
			// 	ParentStateRoot: my.Input.ParentStateRoot,
			// }

			// Handmade AccumulatedServiceOutput from vector_json
			mockC = make(types.AccumulatedServiceOutput)

			// GuaranteesExtrinsic from vector_json
			mockEg = types.GuaranteesExtrinsic{}
		)
		mockC[types.AccumulatedServiceHash{
			ServiceId: 1,
			Hash:      my.Input.AccumulateRoot,
		}] = true
		for _, workPackage := range my.Input.WorkPackages {
			mockEg = append(mockEg, types.ReportGuarantee{
				Report: types.WorkReport{
					PackageSpec: types.WorkPackageSpec{
						Hash:        types.WorkPackageHash(workPackage.Hash),
						ExportsRoot: workPackage.ExportsRoot,
					},
				},
			})
		}
		s.GetIntermediateStates().SetBeefyCommitmentOutput(mockC)

		// Test CheckDuplicate
		// Start test CheckDuplicate
		if rhc.Betas != nil { // Beta is empty -> not need to remove duplicate
			// Test existing header hashes
			for _, beta := range rhc.Betas {
				if rhc.CheckDuplicate(beta.HeaderHash) != true {
					t.Errorf("[%d] Expected true for existing header hash %x :", i, beta.HeaderHash)
				}
			}
			// Test non-existing header hash
			if rhc.CheckDuplicate(my.Input.HeaderHash) != false {
				t.Errorf("[%d] Expected false for non-existing header hash %x", i, my.Input.HeaderHash)
			}
		} else if rhc.Betas == nil {
			fmt.Printf("[%d]PreState Beta is empty\n", i)
		}

		// Test AddToBetaDagger
		// Start test AddToBetaDagger
		rhc.AddToBetaDagger(my.Input.ParentStateRoot)

		// Get result of BetaDagger from store
		betadagger := s.GetIntermediateStates().GetBetaDagger()

		// length of BetaDagger should not exceed maxBlocksHistory
		if len(betadagger) > types.MaxBlocksHistory {
			t.Errorf("[%d]Expected BetaDagger length not to greater than %d, got %d", i, types.MaxBlocksHistory, len(betadagger))
		}

		// Test AddToBetaPrime
		// Start test AddToBetaPrime

		// r function
		accumulationResultTreeRoot := r(mockC)

		// b function
		NewMmr := rhc.b(accumulationResultTreeRoot)

		if len(NewMmr) == 0 {
			t.Errorf("[%d]Expected non-empty NewMmr", i)
		}
		if reflect.DeepEqual(NewMmr, my.PostState.Beta[len(my.PostState.Beta)-1].Mmr.Peaks) {
			t.Errorf("[%d]NewMmr result is not equal to BetaPrime", i)
		}

		// p function
		reports := p(mockEg)

		if len(reports) == 0 {
			t.Errorf("[%d]Expected non-empty reports", i)
		}
		for j, report := range reports {
			if report.Hash != my.PostState.Beta[len(my.PostState.Beta)-1].Reported[j].Hash {
				t.Errorf("[%d]report[%d].Hash is not equal to testvector's BetaPrime", i, j)
			}
			if report.ExportsRoot != my.PostState.Beta[len(my.PostState.Beta)-1].Reported[j].ExportsRoot {
				t.Errorf("[%d]report[%d].ExportsRoot is not equal to testvector's BetaPrime", i, j)
			}
		}

		// n function
		items := rhc.N(my.Input.HeaderHash, mockEg, my.Input.AccumulateRoot)

		if items.HeaderHash != my.PostState.Beta[len(my.PostState.Beta)-1].HeaderHash {
			t.Errorf("[%d]items.HeaderHash %x is not equal to BetaPrime's %x", i, items.HeaderHash, my.PostState.Beta[len(my.PostState.Beta)-1].HeaderHash)
		}

		// (7.4)
		rhc.AddToBetaPrime(items)

		// Get result of (7.4), beta^prime, from store
		betaPrime := s.GetPosteriorStates().GetBeta()

		if len(betaPrime) < 1 {
			t.Errorf("[%d]Expected BetaPrime length to be 1, got %d", i, len(betaPrime))
		}
		if reflect.DeepEqual(betaPrime, my.PostState.Beta) {
			t.Errorf("[%d]BetaPrime should equal to PostState.Beta", i)
		}

	}
}

func TestOuterUsedRecentHistory(t *testing.T) {
	// Load test vectors from JSON
	vectors := []string{
		"./data/progress_blocks_history-1.json", // Empty history queue
		"./data/progress_blocks_history-2.json", // Not empty nor full history queue
		"./data/progress_blocks_history-3.json", // Fill the history queue
		"./data/progress_blocks_history-4.json", // Shift the history queue
	}

	for i, vector := range vectors {
		my, err := loadInputFromJSON(vector)
		if err != nil {
			t.Fatalf("Failed to load input from JSON[%d]: %v", i, err)
		}

		// Create a new RecentHistoryController
		rhc := NewRecentHistoryController()
		if rhc == nil {
			t.Fatal("Controller should be initialized successfully")
		}
		if len(rhc.Betas) != 0 {
			t.Fatalf("Expected controller to have no states initially, got %d", len(rhc.Betas))
		}

		// Store initialization
		s := store.GetInstance()
		s.GetPriorStates().SetBeta(my.PreState.Beta)
		priorState := s.GetPriorStates()

		// Put preState beta to rhc.Betas as priorState.Beta
		rhc.Betas = priorState.GetBeta()

		// Test STF functions
		// β† ≺ (H, β) (4.6)
		// β′ ≺ (H, EG, β†, C) (4.7)
		var (
			// Handmade AccumulatedServiceOutput from vector_json
			mockC types.AccumulatedServiceOutput

			// GuaranteesExtrinsic from vector_json
			mockEg = types.GuaranteesExtrinsic{}
		)
		mockC = make(types.AccumulatedServiceOutput)
		mockC[types.AccumulatedServiceHash{
			ServiceId: 1,
			Hash:      my.Input.AccumulateRoot,
		}] = true
		for _, workPackage := range my.Input.WorkPackages {
			mockEg = append(mockEg, types.ReportGuarantee{
				Report: types.WorkReport{
					PackageSpec: types.WorkPackageSpec{
						Hash:        types.WorkPackageHash(workPackage.Hash),
						ExportsRoot: workPackage.ExportsRoot,
					},
				},
			})
		}

		var mockBlock = types.Block{
			// Header from vector_json
			Header: types.Header{
				Parent:          my.Input.HeaderHash,
				ParentStateRoot: my.Input.ParentStateRoot,
			},
			Extrinsic: types.Extrinsic{
				Guarantees: mockEg,
			},
		}
		s.AddBlock(mockBlock)
		s.GetIntermediateStates().SetBeefyCommitmentOutput(mockC)

		// Test AddToBetaDagger
		// Start test STFBeta2BetaDagger (4.6)
		STFBeta2BetaDagger()

		// Get result of BetaDagger from store
		betadagger := s.GetIntermediateStates().GetBetaDagger()

		// length of BetaDagger should not exceed maxBlocksHistory
		if len(betadagger) > types.MaxBlocksHistory {
			t.Errorf("[%d]Expected BetaDagger length not to greater than %d, got %d", i, types.MaxBlocksHistory, len(betadagger))
		}

		// Start test STFBetaDagger2BetaPrime (4.7)
		STFBetaDagger2BetaPrime()

		// Get result of (7.4), beta^prime, from store
		betaPrime := s.GetPosteriorStates().GetBeta()

		if len(betaPrime) < 1 {
			t.Errorf("[%d]Expected BetaPrime length to be 1, got %d", i, len(betaPrime))
		}
		if reflect.DeepEqual(betaPrime, my.PostState.Beta) {
			t.Errorf("[%d]BetaPrime should equal to PostState.Beta", i)
		}

	}
}
*/

// ===== NEW TEST FOR TESTVECTOR =====

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestRecentHistoryTestVectors(t *testing.T) {

	dir := filepath.Join(utilities.JAM_TEST_VECTORS_DIR, "stf", "history", types.TEST_MODE)

	// Read binary files
	binFiles, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		t.Logf("🚀 Processing file: %s", binFile)
		// if binFile != "progress_blocks_history-4.bin" {
		// 	continue
		// }
		// Read the binary file
		binPath := filepath.Join(dir, binFile)

		// Load recent history test case
		history := &jamtests_history.HistoryTestCase{}

		err := utilities.GetTestFromBin(binPath, history)
		if err != nil {
			t.Errorf("Can't decode from bin: %v", err)
		}

		/*
			STORE
		*/
		store.ResetInstance()
		storeInstance := store.GetInstance()
		// Set prior state recent history ( beta_H )
		storeInstance.GetPriorStates().SetBetaH(history.PreState.Beta.History)
		// Set extrinsic
		mockGuarantessExtrinsic := types.GuaranteesExtrinsic{}
		for _, workPackage := range history.Input.WorkPackages {
			mockGuarantessExtrinsic = append(mockGuarantessExtrinsic, types.ReportGuarantee{
				Report: types.WorkReport{
					PackageSpec: types.WorkPackageSpec{
						Hash:        types.WorkPackageHash(workPackage.Hash),
						ExportsRoot: workPackage.ExportsRoot,
					},
				},
			})
		}
		block := types.Block{
			Header: types.Header{
				Parent:          history.Input.HeaderHash,
				ParentStateRoot: history.Input.ParentStateRoot,
			},
			Extrinsic: types.Extrinsic{
				Guarantees: mockGuarantessExtrinsic,
			},
		}
		storeInstance.AddBlock(block)

		/*
			STF
		*/
		// Start test STFBeta2BetaDagger (4.6)
		STFBetaH2BetaHDagger()

		// Validate intermediate state betaHDagger
		HistoryDagger := storeInstance.GetIntermediateStates().GetBetaHDagger()
		if HistoryDagger.Validate() != nil {
			t.Logf("❌ [data] %s", binFile)
			t.Errorf("betaHDagger validation failed: %v", HistoryDagger.Validate())
		}

		// Start test STFBetaDagger2BetaPrime (4.7)
		// For test-vector, we cannot call STFBetaHDagger2BetaHPrime(),
		// set intermediate value accumulationRoot manually
		t.Logf("mmr peaks before append: %+v", history.PreState.Beta.Mmr.Peaks)
		beefyBeltPrime, commitment := appendAndCommitMmr(history.PreState.Beta.Mmr, history.Input.AccumulateRoot)
		t.Logf("mmr peaks after append: %+v", beefyBeltPrime.Peaks)
		workReportHash := mapWorkReportFromEg(block.Extrinsic.Guarantees)
		item := newItem(history.Input.HeaderHash, workReportHash, commitment)
		historyPrime := AddItem2BetaHPrime(HistoryDagger, item)

		// Set beta_B^prime and beta_H^prime to store
		storeInstance.GetPosteriorStates().SetBetaB(beefyBeltPrime)
		storeInstance.GetPosteriorStates().SetBetaH(historyPrime)

		// Validate posterior state betaPrime
		betaPrime := storeInstance.GetPosteriorStates().GetBeta()
		if betaPrime.History.Validate() != nil {
			t.Logf("❌ [data] %s", binFile)
			t.Errorf("betaPrime validation failed: %v", betaPrime.History.Validate())
		} else if len(betaPrime.History) < 1 {
			t.Logf("❌ [data] %s", binFile)
			t.Errorf("BetaPrime.History should not be nil, got %d", len(betaPrime.History))
		}

		/*
			Validate
		*/
		if !reflect.DeepEqual(betaPrime.History, history.PostState.Beta.History) {
			t.Logf("❌ [data] %s", binFile)
			t.Logf("BetaPrime: %+#v", betaPrime)
			t.Logf("BetaPrime BeefyRoot: %+#v", betaPrime.History[len(betaPrime.History)-1].BeefyRoot)
			t.Logf("PostState.Beta BeefyRoot: %+#v", history.PostState.Beta.History[len(history.PostState.Beta.History)-1].BeefyRoot)
			diff := cmp.Diff(history.PostState.Beta.History, betaPrime.History)
			t.Errorf("BetaPrime.History should equal to PostState.Beta.History\n%s", diff)
		} else if !reflect.DeepEqual(betaPrime.Mmr.Peaks, history.PostState.Beta.Mmr.Peaks) {
			t.Logf("❌ [data] %s", binFile)
			diff := cmp.Diff(history.PostState.Beta.Mmr.Peaks, betaPrime.Mmr.Peaks)
			t.Errorf("BetaPrime.Mmr.Peaks should equal to PostState.Beta.Mmr.Peaks\n%s", diff)
		} else {
			t.Logf("🟢 [data] %s", binFile)
		}
	}

}

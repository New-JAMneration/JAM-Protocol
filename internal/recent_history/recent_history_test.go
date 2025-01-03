package recent_history

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// custom input struct for json^^
//type mmrpeak_json string

type mmr_json struct {
	Peak []string `json:"peak,omitempty"`
}

type workPackage_json struct {
	Hash        string `json:"hash,omitempty"`
	ExportsRoot string `json:"exports_root,omitempty"`
}

type blockInfo_json struct {
	HeaderHash string             `json:"header_hash,omitempty"`
	Mmr        mmr_json           `json:"mmr,omitempty"`
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
	PreState  State
	Output    interface{}
	PostState State
}

func hexToOpaqueHash(hexStr string) ([32]byte, error) {
	// 移除可能的 "0x" 前綴
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// 解碼 hex 字串
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

func loadInputFromJSON(filePath string) (m myVector, err error) {
	var vector_json vector_json
	data, err := os.ReadFile(filePath)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &vector_json)

	// convert to mytype
	// MyVector.Input: json to myHistory
	var (
		// string to HeaderHash
		//inputHeaderHash, _    = hex.DecodeString(vector_json.Input.HeaderHash)
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
		workPackageHash, _ := hexToOpaqueHash(inputWorkPackage.Hash)
		workPackageExportsRoot, _ := hexToOpaqueHash(inputWorkPackage.ExportsRoot)
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
			//Mmr:        mmr,
			StateRoot: types.StateRoot(stateRoot),
			Reported:  []types.ReportedWorkPackage{},
		}

		for _, mmrPeak := range preStateBeta.Mmr.Peak {
			var (
				peak, _  = hexToOpaqueHash(mmrPeak)
				peakHash = types.OpaqueHash(peak)
				//mmr     = types.Mmr{Peaks: append(preStateBetas_types[i].Mmr.Peaks, types.MmrPeak(peak))}
			)
			// string to MmrPeak
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
				//mmr     = types.Mmr{Peaks: append(postStateBetas_types[i].Mmr.Peaks, types.MmrPeak(peak))}
			)
			// string to MmrPeak
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
			blockInfo.Reported = append(blockInfo.Reported, types.ReportedWorkPackage{
				Hash:        types.WorkReportHash(reportedWorkPackageHash),
				ExportsRoot: types.ExportsRoot(reportedWorkPackageExportsRoot),
			})
		}
		postStateBetas_types = append(postStateBetas_types, blockInfo)
	}

	// Finally, we construct our myVector with types
	m = myVector{
		Input: myHistory{
			HeaderHash:      inputHeaderHash_types,
			ParentStateRoot: inputParentStateRoot_types,
			AccumulateRoot:  inputAccumulateRoot_types,
			WorkPackages:    inputWorkPackages_types,
		},
		PreState: State{
			Beta: preStateBetas_types,
		},
		Output: vector_json.Output,
		PostState: State{
			BetaPrime: postStateBetas_types,
		},
	}
	return m, err
}

func TestRemoveDuplicate(t *testing.T) {
	m, err := loadInputFromJSON("./data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta: m.PreState.Beta,
	}
	if state.Beta != nil { // Beta is empty -> not need to remove duplicate
		// test existing header hash
		if state.RemoveDuplicate(m.PreState.Beta[0].HeaderHash) != true {
			t.Error("Expected true for existing header hash")
		}
		// test non-existing header hash
		if state.RemoveDuplicate(m.Input.HeaderHash) != false {
			t.Error("Expected false for non-existing header hash")
		}
	}

}

func TestAddToBetaDagger(t *testing.T) {
	m, err := loadInputFromJSON("./data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta:       m.PreState.Beta,
		BetaDagger: []types.BlockInfo{},
	}

	var monkHeader = types.Header{
		ParentStateRoot: m.Input.ParentStateRoot,
	}

	state.AddToBetaDagger(monkHeader)

	// length of BetaDagger should not exceed maxBlocksHistory
	if len(state.BetaDagger) > types.MaxBlocksHistory {
		t.Errorf("Expected BetaDagger length not to greater than %d, got %d", types.MaxBlocksHistory, len(state.BetaDagger))
	}
}

func TestAddToBetaPrime(t *testing.T) {
	m, err := loadInputFromJSON("./data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta:       m.PreState.Beta,
		BetaDagger: []types.BlockInfo{},
		BetaPrime:  []types.BlockInfo{},
	}

	var monkHeader = types.Header{
		Parent:          m.Input.HeaderHash,
		ParentStateRoot: m.Input.ParentStateRoot,
	}
	// ------------init----------------
	// r function
	// handmade BeefyCommitmentOutput
	monkC := BeefyCommitmentOutput{
		{commitment: m.Input.AccumulateRoot},
	}

	accumulationResultTreeRoot := r(monkC)

	// b function
	NewMmr := state.b(accumulationResultTreeRoot)

	if len(NewMmr) == 0 {
		t.Error("Expected non-empty NewMmr")
	}
	if reflect.DeepEqual(NewMmr, m.PostState.BetaPrime[0].Mmr.Peaks) {
		t.Error("NewMmr result is not equal to BetaPrime")
	}

	// p function
	// GuaranteesExtrinsic from vector_json
	eg := types.GuaranteesExtrinsic{
		{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(m.Input.WorkPackages[0].Hash),
					ExportsRoot: m.Input.WorkPackages[0].ExportsRoot,
				},
			},
		},
		{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(m.Input.WorkPackages[1].Hash),
					ExportsRoot: m.Input.WorkPackages[1].ExportsRoot,
				},
			},
		},
	}

	reports := p(eg)
	// fmt.Print(m.PostState.BetaPrime[0].Reported)
	// fmt.Print("\n")
	// fmt.Print(reports)

	if len(reports) == 0 {
		t.Error("Expected non-empty reports")
	}
	for i, report := range reports {
		if report.Hash != m.PostState.BetaPrime[0].Reported[i].Hash {
			t.Errorf("report[%d].Hash is not equal to testvector's BetaPrime", i)
		}
		if report.ExportsRoot != m.PostState.BetaPrime[0].Reported[i].ExportsRoot {
			t.Errorf("report[%d].ExportsRoot is not equal to testvector's BetaPrime", i)
		}
	}

	// n func
	items := state.n(monkHeader, eg, monkC)

	if items.HeaderHash != m.PostState.BetaPrime[0].HeaderHash {
		t.Errorf("items.HeaderHash %x is not equal to BetaPrime's %x", items.HeaderHash, m.PostState.BetaPrime[0].HeaderHash)
	}

	// (7.4)
	state.AddToBetaPrime(items)

	if len(state.BetaPrime) < 1 {
		t.Errorf("Expected BetaPrime length to be 1, got %d", len(state.BetaPrime))
	}
	if reflect.DeepEqual(state.BetaPrime, m.PostState.BetaPrime) {
		t.Error("BetaPrime should equal to PostState.BetaPrime")
	}
}

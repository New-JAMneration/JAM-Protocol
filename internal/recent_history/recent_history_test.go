package recent_history

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type historyInput struct {
	HeaderHash      types.HeaderHash            `json:"header_hash"`
	ParentStateRoot types.StateRoot             `json:"parent_state_root"`
	AccumulateRoot  types.OpaqueHash            `json:"accumulate_root"`
	WorkPackages    []types.ReportedWorkPackage `json:"work_packages"`
}

type testVector struct {
	Input     historyInput      `json:"input"`
	PreState  []types.BlockInfo `json:"pre_state"`
	Output    interface{}       `json:"output"`
	PostState []types.BlockInfo `json:"post_state"`
}

func loadInputFromJSON(filePath string) (testVector, error) {
	var testVector testVector
	data, err := os.ReadFile("/home/yu2c/gitclone/JAM-Protocol/internal/recent_history/data/progress_blocks_history-1.json")
	if err != nil {
		return testVector, err
	}
	err = json.Unmarshal(data, testVector)
	return testVector, err

}

func TestRemoveDuplicate(t *testing.T) {
	testVector, err := loadInputFromJSON("internal/recent_history/data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta: testVector.PreState,
	}
	// test existing header hash
	if state.RemoveDuplicate(testVector.PreState[0].HeaderHash) != true {
		t.Error("Expected true for existing header hash")
	}
	// test non-existing header hash
	if state.RemoveDuplicate(testVector.Input.HeaderHash) != false {
		t.Error("Expected false for non-existing header hash")
	}
}

func TestAddToBetaDagger(t *testing.T) {
	testVector, err := loadInputFromJSON("internal/recent_history/data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta:       testVector.PreState,
		BetaDagger: []types.BlockInfo{},
	}

	var monkHeader = types.Header{
		ParentStateRoot: testVector.Input.ParentStateRoot,
	}

	state.AddToBetaDagger(monkHeader)

	// length of BetaDagger should be equal to maxBlocksHistory
	if len(state.BetaDagger) != types.MaxBlocksHistory {
		t.Errorf("Expected BetaDagger length to be %d, got %d", types.MaxBlocksHistory, len(state.BetaDagger))
	}
}

func TestAddToBetaPrime(t *testing.T) {
	testVector, err := loadInputFromJSON("internal/recent_history/data/progress_blocks_history-1.json")
	if err != nil {
		t.Fatalf("Failed to load input from JSON: %v", err)
	}

	state := &State{
		Beta:       testVector.PreState,
		BetaDagger: []types.BlockInfo{},
		BetaPrime:  testVector.PostState,
	}

	var monkHeader = types.Header{
		ParentStateRoot: testVector.Input.ParentStateRoot,
	}
	// ------------init----------------
	// r function
	// handmade BeefyCommitmentOutput
	monkC := BeefyCommitmentOutput{
		{commitment: testVector.Input.AccumulateRoot},
	}

	accumulationResultTreeRoot := r(monkC)

	// b function
	NewMmr := state.b(accumulationResultTreeRoot)

	if len(NewMmr) == 0 {
		t.Error("Expected non-empty NewMmr")
	}
	if reflect.DeepEqual(NewMmr, state.BetaPrime[0].Mmr.Peaks) {
		t.Error("NewMmr result is not equal to BetaPrime")
	}

	// p function
	// GuaranteesExtrinsic from testVector
	eg := types.GuaranteesExtrinsic{
		{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(testVector.Input.WorkPackages[0].Hash),
					ExportsRoot: testVector.Input.WorkPackages[0].ExportsRoot,
				},
			},
		},
		{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(testVector.Input.WorkPackages[1].Hash),
					ExportsRoot: testVector.Input.WorkPackages[1].ExportsRoot,
				},
			},
		},
	}

	reports := p(eg)

	if len(reports) == 0 {
		t.Error("Expected non-empty reports")
	}
	if reflect.DeepEqual(reports[0], state.BetaPrime[0].Reported) {
		t.Error("report[0] is not equal to BetaPrime")
	}
	if reflect.DeepEqual(reports[1], state.BetaPrime[1].Reported) {
		t.Error("report[1] is not equal to BetaPrime")
	}
	if reflect.DeepEqual(reports[0], state.BetaPrime[1].Reported) {
		t.Error("report[0] should not equal to BetaPrime")
	}

	// n func

	items := state.n(monkHeader, eg, monkC)

	if items.HeaderHash != state.BetaPrime[0].HeaderHash {
		t.Error("items.HeaderHash is not equal to BetaPrime")
	}

	state.AddToBetaPrime(items)

	if len(state.BetaPrime) != 1 {
		t.Errorf("Expected BetaPrime length to be 1, got %d", len(state.BetaPrime))
	}
	if reflect.DeepEqual(state.BetaPrime, testVector.PostState) {
		t.Error("BetaPrime should equal to PostState")
	}
}

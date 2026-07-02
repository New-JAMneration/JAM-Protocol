package recent_history_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_history "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestRecentHistoryTestVectors(t *testing.T) {
	// GP v0.8.0 eq:recenthistoryspec adds a timeslot to each recent-history
	// entry (#1014), so the v0.7.x history vectors' beta no longer decodes and
	// the post-state comparison no longer holds. Re-enable on official v0.8.0
	// vectors (#1012). The wire format is covered by
	// TestEncodeBlockInfo_V080Timeslot in internal/types.
	t.Skip("v0.7.x history vectors predate the GP v0.8.0 BlockInfo timeslot (#1014); re-enable on official v0.8.0 vectors")

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
		blockchain.ResetInstance()
		storeInstance := blockchain.GetInstance()
		// Set prior state recent history ( beta_H )
		storeInstance.GetPriorStates().SetBeta(history.PreState.Beta)
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
		recent_history.STFBetaH2BetaHDagger()

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
		beefyBeltPrime, commitment := recent_history.AppendAndCommitMmr(history.PreState.Beta.Mmr, history.Input.AccumulateRoot)
		t.Logf("mmr peaks after append: %+v", beefyBeltPrime.Peaks)
		workReportHash := recent_history.MapWorkReportFromEg(block.Extrinsic.Guarantees)
		// Slot 0: the v0.7.x HistoryInput has no slot; the real value comes with
		// the official v0.8.0 vector schema (#1012).
		item := recent_history.NewItem(history.Input.HeaderHash, workReportHash, commitment, 0)
		historyPrime := recent_history.AddItem2BetaHPrime(HistoryDagger, item)
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

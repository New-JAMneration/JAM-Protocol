package authorization

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_authorization "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	"github.com/google/go-cmp/cmp"
)

func TestAuthorizationTestVectors(t *testing.T) {
	if types.TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else if types.TEST_MODE == "full" {
		types.SetFullMode()
	}

	dir := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "authorizations", types.TEST_MODE)

	// Read binary files
	binFiles, err := utils.GetTargetExtensionFiles(dir, utils.BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)

		// Load authorization test case
		authorization := &jamtests_authorization.AuthorizationTestCase{}

		err := utils.GetTestFromBin(binPath, authorization)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Set up test input state
		var (
			inputSlot      = authorization.Input.Slot
			inputEg        = authorization.Input.Auths
			priorState     = authorization.PreState
			posteriorState = authorization.PostState
		)
		mockEgs := make(types.GuaranteesExtrinsic, 0, len(inputEg))
		for _, eg := range inputEg {
			mockEgs = append(mockEgs, types.ReportGuarantee{
				Report: types.WorkReport{
					CoreIndex:      eg.CoreIndex,
					AuthorizerHash: eg.AuthorizerHash,
				},
			})
		}
		// Get store instance and required states
		s := store.GetInstance()
		s.GetProcessingBlockPointer().SetSlot(inputSlot)
		s.GetProcessingBlockPointer().SetGuaranteesExtrinsic(mockEgs)
		s.GetPosteriorStates().SetVarphi(posteriorState.Varphi)
		s.GetPriorStates().SetAlpha(priorState.Alpha)

		// === Run Authorization ===
		err = Authorization()
		if err != nil {
			t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Fatalf("Error: %v", err)
		}

		// Get output state
		outputAlpha := s.GetPosteriorStates().GetAlpha()
		// Validate output state
		if !reflect.DeepEqual(posteriorState.Varphi, priorState.Varphi) {
			t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
			diff := cmp.Diff(priorState.Varphi, posteriorState.Varphi)
			t.Fatalf("Varphi State are not equal: %v", diff)
		} else if !reflect.DeepEqual(outputAlpha, posteriorState.Alpha) {
			t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
			// diff := cmp.Diff(posteriorState.Alpha, outputAlpha)
			for i := range outputAlpha {
				if !reflect.DeepEqual(outputAlpha[i], posteriorState.Alpha[i]) {
					t.Logf("len(outputAlpha[%d]): %d, len(posteriorState.Alpha[%d]): %d", i, len(outputAlpha[i]), i, len(posteriorState.Alpha[i]))
					needDiff := cmp.Diff(priorState.Alpha[i], posteriorState.Alpha[i])
					t.Logf("expected diff: %v", needDiff)
					deepDiff := cmp.Diff(posteriorState.Alpha[i], outputAlpha[i])
					t.Fatalf("Alpha State[%d] are not equal: %v", i, deepDiff)
				}
			}
		} else {
			t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
		}

	}
}

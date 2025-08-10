package merklization

import (
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
)

func TestStateKeyValsToState(t *testing.T) {
	dirNames := []string{
		// "fallback",
		"preimages",
		// "preimages_light",
		// "safrole",
		// "storage",
		// "storage_light",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join("..", utilities.JAM_TEST_VECTORS_DIR, "traces", dirName)

		fileNames, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error getting files from directory %s: %v", dir, err)
			continue
		}

		for _, fileName := range fileNames {
			if fileName != "00000050.bin" {
				continue
			}
			filePath := filepath.Join(dir, fileName)

			// Read the bin file
			traceTestCase := &jamtests_trace.TraceTestCase{}
			err := utilities.GetTestFromBin(filePath, traceTestCase)
			if err != nil {
				t.Errorf("Error reading file %s: %v", filePath, err)
				continue
			}

			// Parse the state keyvals
			// TODO: prestate or poststate
			_, err = StateKeyValsToState(traceTestCase.PreState.KeyVals)

			// TODO: 正確答案從哪裡拿?
		}
	}
}

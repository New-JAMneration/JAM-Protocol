package traces

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamteststrace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
)

var genesisStateRoot = types.StateRoot(types.OpaqueHash(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")))

type TraceRunner struct {
	// Store is the global protocol store. The runner mutates it between traces.
	Store *store.Store
}

// NewTraceRunner constructs a TraceRunner with sane defaults.
func NewTraceRunner() *TraceRunner {
	return &TraceRunner{
		Store: store.GetInstance(),
	}
}

func (tr *TraceRunner) Run(data interface{}, _ bool) error {
	testCase := data.(*jamteststrace.TraceTestCase)

	if len(tr.Store.GetBlocks()) == 0 {
		return fmt.Errorf("no blocks in the store")
	}

	// Verify the header
	if tr.Store.GetLatestBlock().Header.Parent != testCase.Block.Header.Parent {
		return fmt.Errorf("parent mismatch: got %x, want %x", tr.Store.GetLatestBlock().Header.Parent, testCase.PreState.StateRoot)
	}

	if tr.Store.GetLatestBlock().Header.ParentStateRoot != testCase.Block.Header.ParentStateRoot {
		return fmt.Errorf("state_root mismatch: got %x, want %x", tr.Store.GetLatestBlock().Header.ParentStateRoot, testCase.PreState.StateRoot)
	}

	_, err := stf.RunSTF()
	if err != nil {
		return err
	}

	return nil
}

// Verify checks the block header's parent_hash and state_root against the
// expected values in the Testable instance.
func (tr *TraceRunner) Verify(data testdata.Testable) error {
	return data.Validate()
}

func hexToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

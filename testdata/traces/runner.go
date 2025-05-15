package traces

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamteststraces "github.com/New-JAMneration/JAM-Protocol/jamtests/traces"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
)

var (
	genesisStateRoot = types.StateRoot(types.OpaqueHash(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")))
)

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
	testCase := data.(*jamteststraces.TraceTestCase)

	// Initialize the genesis if the state root is 0
	if testCase.PreState.StateRoot == genesisStateRoot {
		if err := tr.InitializeGenesis(testCase); err != nil {
			return err
		}
	}

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

	err := stf.RunSTF()
	if err != nil {
		return err
	}

	return tr.Verify(testCase)
}

// Verify checks the block header's parent_hash and state_root against the
// expected values in the Testable instance.
func (tr *TraceRunner) Verify(data testdata.Testable) error {
	return data.Validate()
}

func (tr *TraceRunner) InitializeGenesis(data *jamteststraces.TraceTestCase) error {
	log.Printf("Initializing genesis block....")
	// TODO: Initialize the genesis
	return nil
}

func hexToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

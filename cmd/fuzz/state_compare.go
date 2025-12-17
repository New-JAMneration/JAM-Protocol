package main

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type importBlockCompareInput struct {
	Client            *fuzz.FuzzClient
	Block             types.Block
	BlockHeaderHash   types.HeaderHash
	ExpectedStateRoot types.StateRoot
	ActualStateRoot   types.StateRoot
	ExpectedPostState types.StateKeyVals
}

func compareImportBlockState(input importBlockCompareInput) (bool, error) {
	if input.ActualStateRoot == input.ExpectedStateRoot {
		return false, nil
	}

	headerHash, err := input.ensureBlockHeaderHash()
	if err != nil {
		return false, fmt.Errorf("error preparing block header hash: %w", err)
	}

	logger.ColorBlue("[ImportBlock][Check] state_root mismatch: expected 0x%x, got 0x%x",
		input.ExpectedStateRoot, input.ActualStateRoot)

	if input.Client == nil || len(input.ExpectedPostState) == 0 {
		return true, nil
	}

	// Build context with block info for Diff
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(input.Block.Header.Slot) % uint32(types.EpochLength)
	epoch := slot / uint32(types.EpochLength)
	diffCtx := logger.FormatContext(hashStr, slot, epoch, "Diff")

	if err := logImportBlockDiffWithContext(input.Client, headerHash, input.ExpectedStateRoot, input.ActualStateRoot, input.ExpectedPostState, diffCtx); err != nil {
		return true, err
	}

	return true, nil
}

func (input importBlockCompareInput) ensureBlockHeaderHash() (types.HeaderHash, error) {
	if input.BlockHeaderHash != (types.HeaderHash{}) {
		return input.BlockHeaderHash, nil
	}

	serializedHeader, err := utilities.HeaderSerialization(input.Block.Header)
	if err != nil {
		return types.HeaderHash{}, err
	}

	return types.HeaderHash(hash.Blake2bHash(serializedHeader)), nil
}

func logImportBlockDiffWithContext(client *fuzz.FuzzClient, headerHash types.HeaderHash, expectedRoot, actualRoot types.StateRoot, expectedKeyVals types.StateKeyVals, ctx string) error {

	logger.ColorGreen("[GetState][Request] header_hash= 0x%v", hex.EncodeToString(headerHash[:]))
	actualStateKeyVal, err := client.GetState(headerHash)
	if err != nil {
		logger.Errorf("%s error sending get_state request: %v", ctx, err)
		return fmt.Errorf("error sending get_state request: %w", err)
	}

	diffs, err := merklization.GetStateKeyValsDiff(expectedKeyVals, actualStateKeyVal)
	if err != nil {
		logger.Errorf("%s GetStateKeyValsDiff error: %v", ctx, err)
		return fmt.Errorf("fuzzer GetStateKeyValsDiff error: %w", err)
	}

	logger.ColorYellow("[GetState][Response] %d different key-val: ", len(diffs))
	nullStateRoot := types.StateRoot{}
	if actualRoot == nullStateRoot || expectedRoot == nullStateRoot || len(actualStateKeyVal) == 0 {
		logger.ColorYellow("[GetState][Response] nil state root, imply protocol error")
		return nil // no diffs
	}
	err = merklization.DebugStateKeyValsDiff(diffs)
	if err != nil {
		logger.Errorf("%s DebugStateKeyValsDiff error: %v", ctx, err)
		return fmt.Errorf("fuzzer DebugStateKeyValsDiff error: %w", err)
	}

	return nil
}
